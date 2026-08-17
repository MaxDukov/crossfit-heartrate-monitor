"""ANT+ коллектор — обнаружение датчиков и сбор ЧСС.

Переиспользует подход из ant_hr_monitor.py, но вместо print()
отправляет данные через callback для интеграции с FastAPI.
"""

import logging
import threading
import time
from typing import Callable, Optional

from openant.devices import ANTPLUS_NETWORK_KEY
from openant.devices.heart_rate import HeartRate, HeartRateData
from openant.easy.node import Node

from . import sensor_db

_logger = logging.getLogger(__name__)


def _get_device_id(device: HeartRate) -> str:
    """Извлекает реальный ANT+ Device ID из строкового представления."""
    return str(device).rsplit("_", 1)[-1]


class AntCollector:
    """Фоновый ANT+ коллектор, работающий в отдельном потоке.

    Обнаруживает датчики ЧСС и вызывает callback при получении данных.
    Управляет жизненным циклом Node и каналов.
    """

    def __init__(
        self,
        max_sensors: int = 8,
        on_hr_data: Optional[Callable] = None,
        on_new_sensor: Optional[Callable] = None,
    ):
        self._max_sensors = max_sensors
        self._on_hr_data = on_hr_data
        self._on_new_sensor = on_new_sensor
        self._node: Optional[Node] = None
        self._devices: list[HeartRate] = []
        self._known_ids: set[int] = set()
        self._thread: Optional[threading.Thread] = None
        self._running = False
        self._lock = threading.Lock()

    def start(self):
        """Запускает ANT+ collector в фоновом потоке."""
        if self._running:
            return
        self._running = True
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()
        _logger.info("ANT+ collector started")

    def stop(self):
        """Останавливает collector и закрывает все каналы."""
        self._running = False
        for d in self._devices:
            try:
                d.close_channel()
            except Exception:
                pass
        if self._node:
            try:
                self._node.stop()
            except Exception:
                pass
        _logger.info("ANT+ collector stopped")

    def _run(self):
        """Основной цикл collector'а (работает в отдельном потоке)."""
        try:
            self._node = Node()
            self._node.set_network_key(0x00, ANTPLUS_NETWORK_KEY)

            for i in range(self._max_sensors):
                hr = HeartRate(self._node, device_id=0)
                hr.on_found = self._make_on_found(i, hr)
                hr.on_device_data = self._make_on_data(i, hr)
                self._devices.append(hr)

            self._node.start()
        except Exception as e:
            _logger.error("ANT+ collector error: %s", e)
            if self._running:
                _logger.info("Restarting ANT+ collector in 5s...")
                time.sleep(5)
                if self._running:
                    self._run()

    def _make_on_found(self, index: int, device: HeartRate):
        """Создаёт callback обнаружения датчика."""
        def on_found():
            dev_id = _get_device_id(device)
            numeric_id = int(dev_id)
            with self._lock:
                is_new = numeric_id not in self._known_ids
                self._known_ids.add(numeric_id)

            _logger.info("Sensor #%d found (ID: %s)", index + 1, dev_id)

            sensor_db.upsert_sensor(numeric_id)

            if (
                is_new
                and self._on_new_sensor
                and not sensor_db.is_sensor_assigned(numeric_id)
                and not sensor_db.is_sensor_ignored(numeric_id)
            ):
                try:
                    self._on_new_sensor(numeric_id)
                except Exception as e:
                    _logger.error("on_new_sensor callback error: %s", e)

        return on_found

    def _make_on_data(self, index: int, device: HeartRate):
        """Создаёт callback обработки ЧСС данных."""
        last_hr = [None]
        last_time = [0.0]

        def on_data(page, page_name, data):
            if not isinstance(data, HeartRateData):
                return

            hr = data.heart_rate
            if hr == 0:
                return

            now = time.time()
            if hr == last_hr[0] and (now - last_time[0]) < 2.0:
                return
            last_hr[0] = hr
            last_time[0] = now

            dev_id = _get_device_id(device)
            numeric_id = int(dev_id)

            sensor_db.update_sensor_hr(numeric_id, hr, data.battery_percentage)

            if self._on_hr_data:
                try:
                    self._on_hr_data(numeric_id, hr, data.battery_percentage)
                except Exception as e:
                    _logger.error("on_hr_data callback error: %s", e)

        return on_data
