"""Менеджер WebSocket-соединений — broadcast HR данных всем клиентам."""

import asyncio
import json
import logging
from typing import Set

from fastapi import WebSocket

_logger = logging.getLogger(__name__)


class ConnectionManager:
    """Хранит активные WS-подключения и рассылает HR-обновления."""

    def __init__(self):
        self._connections: Set[WebSocket] = set()

    async def connect(self, ws: WebSocket):
        """Принимает новое WebSocket-подключение."""
        await ws.accept()
        self._connections.add(ws)
        _logger.info(f"WS client connected, total: {len(self._connections)}")

    def disconnect(self, ws: WebSocket):
        """Удаляет отключённого клиента."""
        self._connections.discard(ws)
        _logger.info(f"WS client disconnected, total: {len(self._connections)}")

    async def broadcast(self, data: dict):
        """Рассылает dict как JSON всем подключённым клиентам."""
        if not self._connections:
            return
        payload = json.dumps(data, default=str)
        dead = set()
        for ws in self._connections:
            try:
                await ws.send_text(payload)
            except Exception:
                dead.add(ws)
        self._connections -= dead


manager = ConnectionManager()
