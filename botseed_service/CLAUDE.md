# Botseed

Companion at **botseed.net** — harvests entropy from ACPWB live traffic → random integers via WebSocket + HTTP API.

- `botseed_processor` subscribes to `request_stream`, SHA-256 mixes with `secrets.token_bytes(32)`, publishes to `botseed_stream` at ≤20 events/sec
- `botseed_service/ws_server.py` — asyncio WS (port 8766) + HTTP API (port 8767)
- Static files in `botseed/` — served by acpwb_edge (botseed.net vhost in `acpwb_edge/router`)
