"""
Botseed WebSocket + HTTP API server.

WebSocket  (port 8766): subscribes to Redis 'botseed_stream', broadcasts JSON
                        to all connected browser clients.
HTTP API   (port 8767): GET /api/v1/current → returns Redis 'botseed:latest'
                        as JSON; 503 if not yet populated.

Environment variables:
  REDIS_URL           Redis URL (default: redis://redis:6379/0)
  WS_HOST             Bind host (default: 0.0.0.0)
  WS_PORT             WebSocket port (default: 8766)
  API_PORT            HTTP API port (default: 8767)
  BOTSEED_WS_TOKEN    Required ?token= value for WebSocket auth (optional)
"""

import asyncio
import json
import logging
import os
import threading
import urllib.parse
from http.server import BaseHTTPRequestHandler, HTTPServer

import redis
import redis.asyncio as aioredis
from websockets.asyncio.server import serve

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

REDIS_URL = os.environ.get("REDIS_URL", "redis://redis:6379/0")
WS_HOST = os.environ.get("WS_HOST", "0.0.0.0")
WS_PORT = int(os.environ.get("WS_PORT", "8766"))
API_PORT = int(os.environ.get("API_PORT", "8767"))
BOTSEED_WS_TOKEN = os.environ.get("BOTSEED_WS_TOKEN", "")

connected_clients: set = set()


def _parse_token(path: str) -> str:
    qs = urllib.parse.urlparse(path).query
    return urllib.parse.parse_qs(qs).get("token", [""])[0]


async def handle_client(websocket):
    if BOTSEED_WS_TOKEN:
        token = _parse_token(websocket.request.path)
        if token != BOTSEED_WS_TOKEN:
            await websocket.close(4401, "Unauthorized")
            return

    connected_clients.add(websocket)
    logger.info("WS client connected: %s (total: %d)", websocket.remote_address, len(connected_clients))
    try:
        await websocket.wait_closed()
    finally:
        connected_clients.discard(websocket)
        logger.info("WS client disconnected (total: %d)", len(connected_clients))


async def redis_listener():
    """Subscribe to botseed_stream and fan out to all WebSocket clients."""
    delay = 5
    while True:
        try:
            client = aioredis.from_url(REDIS_URL, decode_responses=True)
            pubsub = client.pubsub()
            await pubsub.subscribe("botseed_stream")
            logger.info("Connected to Redis, listening on 'botseed_stream'")
            delay = 5

            async for message in pubsub.listen():
                if message["type"] != "message":
                    continue
                data = message["data"]
                if not connected_clients:
                    continue
                snapshot = set(connected_clients)
                results = await asyncio.gather(
                    *[ws.send(data) for ws in snapshot],
                    return_exceptions=True,
                )
                for ws, result in zip(snapshot, results):
                    if isinstance(result, Exception):
                        connected_clients.discard(ws)

        except Exception as exc:
            logger.warning("Redis WS listener error (%s), retrying in %ds", exc, delay)
            await asyncio.sleep(delay)
            delay = min(delay * 2, 60)


# ── HTTP API server ─────────────────────────────────────────────────────────────

def _make_api_handler():
    sync_redis = redis.from_url(
        REDIS_URL,
        decode_responses=True,
        socket_connect_timeout=2,
        socket_timeout=1,
    )

    class APIHandler(BaseHTTPRequestHandler):
        def log_message(self, fmt, *args):
            logger.debug("API %s - %s", self.address_string(), fmt % args)

        def do_GET(self):
            if self.path != "/api/v1/current":
                self._respond(404, {"error": "not found — the only endpoint is /api/v1/current"})
                return
            try:
                data = sync_redis.get("botseed:latest")
            except Exception as exc:
                logger.warning("API Redis error: %s", exc)
                self._respond(503, {"error": "redis unavailable"})
                return

            if data is None:
                self._respond(503, {"error": "no data yet — botseed_processor may still be starting"})
                return

            # data is already a JSON string — write it directly
            body = data.encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.send_header("Access-Control-Allow-Origin", "*")
            self.end_headers()
            self.wfile.write(body)

        def _respond(self, status, obj):
            body = json.dumps(obj).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.send_header("Access-Control-Allow-Origin", "*")
            self.end_headers()
            self.wfile.write(body)

    return APIHandler


def _start_api_server():
    handler = _make_api_handler()
    server = HTTPServer(("0.0.0.0", API_PORT), handler)
    logger.info("HTTP API listening on 0.0.0.0:%d  (GET /api/v1/current)", API_PORT)
    server.serve_forever()


async def main():
    api_thread = threading.Thread(target=_start_api_server, daemon=True)
    api_thread.start()

    listener_task = asyncio.create_task(redis_listener())
    async with serve(handle_client, WS_HOST, WS_PORT) as server:
        logger.info("WebSocket server listening on %s:%d", WS_HOST, WS_PORT)
        await server.serve_forever()
    listener_task.cancel()


if __name__ == "__main__":
    asyncio.run(main())
