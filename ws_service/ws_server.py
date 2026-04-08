"""
Standalone asyncio WebSocket server for real-time request streaming.

Subscribes to Redis pub/sub channel "request_stream" and broadcasts
every JSON message to all connected WebSocket clients.

Environment variables:
  REDIS_URL         Redis URL (default: redis://redis:6379/0)
  WS_HOST           Bind host (default: 0.0.0.0)
  WS_PORT           Bind port (default: 8765)
  STREAM_WS_TOKEN   Required token in ?token= query param (optional; if unset, open access)
"""

import asyncio
import logging
import os
import urllib.parse

import redis.asyncio as aioredis
import websockets
from websockets.asyncio.server import serve

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger(__name__)

REDIS_URL = os.environ.get("REDIS_URL", "redis://redis:6379/0")
WS_HOST = os.environ.get("WS_HOST", "0.0.0.0")
WS_PORT = int(os.environ.get("WS_PORT", "8765"))
STREAM_WS_TOKEN = os.environ.get("STREAM_WS_TOKEN", "")

connected_clients: set = set()


def _parse_token(path: str) -> str:
    """Extract ?token= from the WebSocket request path."""
    qs = urllib.parse.urlparse(path).query
    return urllib.parse.parse_qs(qs).get("token", [""])[0]


async def handle_client(websocket):
    """Accept a WebSocket connection, auth-check it, then hold it open."""
    if STREAM_WS_TOKEN:
        token = _parse_token(websocket.request.path)
        if token != STREAM_WS_TOKEN:
            await websocket.close(4401, "Unauthorized")
            return

    connected_clients.add(websocket)
    logger.info("Client connected: %s (total: %d)", websocket.remote_address, len(connected_clients))
    try:
        await websocket.wait_closed()
    finally:
        connected_clients.discard(websocket)
        logger.info("Client disconnected (total: %d)", len(connected_clients))


async def redis_listener():
    """Subscribe to Redis and broadcast messages to all connected clients."""
    delay = 5
    while True:
        try:
            client = aioredis.from_url(REDIS_URL, decode_responses=True)
            pubsub = client.pubsub()
            await pubsub.subscribe("request_stream")
            logger.info("Connected to Redis, listening on 'request_stream'")
            delay = 5  # reset backoff on successful connect

            async for message in pubsub.listen():
                if message["type"] != "message":
                    continue
                data = message["data"]
                if not connected_clients:
                    continue
                # Snapshot the set before iterating — clients may disconnect mid-broadcast
                snapshot = set(connected_clients)
                results = await asyncio.gather(
                    *[ws.send(data) for ws in snapshot],
                    return_exceptions=True,
                )
                for ws, result in zip(snapshot, results):
                    if isinstance(result, Exception):
                        connected_clients.discard(ws)

        except Exception as exc:
            logger.warning("Redis connection failed (%s), retrying in %ds", exc, delay)
            await asyncio.sleep(delay)
            delay = min(delay * 2, 60)


async def main():
    listener_task = asyncio.create_task(redis_listener())
    async with serve(handle_client, WS_HOST, WS_PORT) as server:
        logger.info("WebSocket server listening on %s:%d", WS_HOST, WS_PORT)
        await server.serve_forever()
    listener_task.cancel()


if __name__ == "__main__":
    asyncio.run(main())
