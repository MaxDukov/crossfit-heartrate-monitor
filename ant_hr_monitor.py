#!/usr/bin/env python3
"""
ANT+ монитор сердечного ритма — поддержка до 8 датчиков одновременно.

Использует библиотеку openant. Установка зависимостей:
    pip install openant
    brew install libusb  # macOS

Использование:
    python ant_hr_monitor.py              # авто-подключение 2 датчиков
    python ant_hr_monitor.py -n 4         # авто-подключение 4 датчиков
    python ant_hr_monitor.py --ids 45231 78912  # конкретные Device ID
    python ant_hr_monitor.py --log        # с отладочным логированием
"""

import argparse
import logging
import sys
from datetime import datetime

from openant.devices import ANTPLUS_NETWORK_KEY
from openant.devices.heart_rate import HeartRate, HeartRateData
from openant.easy.node import Node


def create_node():
    """Создаёт и настраивает ANT+ Node с сетевым ключом ANT+.

    Node — основной объект для взаимодействия с USB ANT-стиком.
    При создании автоматически запускается фоновый поток для обмена
    данными с USB-устройством.

    Returns:
        Node: настроенный ANT+ узел, готовый к созданию каналов.
    """
    node = Node()
    node.set_network_key(0x00, ANTPLUS_NETWORK_KEY)
    return node


def _get_device_id(device):
    """Извлекает реальный Device ID из объекта HeartRate.

    При wildcard-подключении (device_id=0) реальный ID назначается
    после обнаружения датчика. Извлекается из строкового представления
    устройства в формате 'heart_rate_XXXXX'.

    Args:
        device: объект HeartRate или AntPlusDevice.

    Returns:
        str: реальный Device ID датчика (например, '01531').
    """
    return str(device).rsplit("_", 1)[-1]


def make_on_sensor_found(sensor_index, device):
    """Создаёт колбэк, вызываемый при обнаружении датчика ANT+.

    Генерирует функцию-замыкание, которая привязана к конкретному
    индексу датчика и выводит информацию о найденном устройстве
    с его реальным Device ID.

    Args:
        sensor_index: порядковый номер датчика (0-based), используется
                      для идентификации в консольном выводе.
        device: объект HeartRate для извлечения Device ID.

    Returns:
        callable: функция-колбэк без аргументов для привязки к device.on_found.
    """
    def on_sensor_found():
        dev_id = _get_device_id(device)
        print(
            f"[{datetime.now().strftime('%H:%M:%S')}] "
            f"Датчик #{sensor_index + 1} (ID: {dev_id}) найден"
        )

    return on_sensor_found


def make_on_device_data(sensor_index, device):
    """Создаёт колбэк для обработки данных от датчика сердечного ритма.

    Генерирует функцию-замыкание, которая вызывается при получении
    новых данных от ANT+ датчика. Парсит HeartRateData и выводит
    текущий пульс и (при наличии) уровень заряда батареи.

    Args:
        sensor_index: порядковый номер датчика (0-based) для вывода.
        device: объект HeartRate, из которого читается device_id.

    Returns:
        callable: функция-колбэк с сигнатурой (page, page_name, data),
                  совместимая с device.on_device_data.
    """
    last_hr = [None]

    def on_device_data(page, page_name, data):
        if not isinstance(data, HeartRateData):
            return

        timestamp = datetime.now().strftime("%H:%M:%S")
        dev_id = _get_device_id(device)
        hr = data.heart_rate

        if hr != last_hr[0]:
            last_hr[0] = hr
            print(
                f"[{timestamp}] Датчик #{sensor_index + 1} (ID: {dev_id}): "
                f"{hr} bpm"
            )

        if data.battery_percentage != 0xFF:
            print(
                f"[{timestamp}] Датчик #{sensor_index + 1} (ID: {dev_id}): "
                f"батарея {data.battery_percentage}%"
            )

    return on_device_data


def setup_sensor(node, device_id=0):
    """Создаёт и настраивает экземпляр HeartRate на отдельном канале ANT+.

    Каждый вызов занимает один канал на ANT-стике (максимум 8 каналов).
    При device_id=0 используется wildcard-режим — подключение к первому
    найденному HR-датчику.

    Args:
        node: настроенный ANT+ Node от create_node().
        device_id: Device ID датчика (0 = авто-подключение к любому).

    Returns:
        HeartRate: объект датчика с выделенным каналом.
    """
    return HeartRate(node, device_id=device_id)


def main(max_sensors=2, device_ids=None, enable_log=False):
    """Основная функция — создаёт датчики и запускает event loop.

    Последовательно создаёт N HeartRate-устройств на одном ANT+ Node,
    привязывает к каждому колбэки для вывода данных и запускает
    блокирующий event loop. Корректно завершает работу по Ctrl-C.

    Args:
        max_sensors: количество датчиков при auto-detect (default: 2).
        device_ids: список конкретных Device ID (None = auto-detect).
        enable_log: включить DEBUG-логирование openant (default: False).
    """
    if enable_log:
        logging.basicConfig(level=logging.DEBUG)

    sensors = []

    try:
        node = create_node()

        if device_ids:
            for i, did in enumerate(device_ids):
                hr_device = setup_sensor(node, device_id=did)
                hr_device.on_found = make_on_sensor_found(i, hr_device)
                hr_device.on_device_data = make_on_device_data(i, hr_device)
                sensors.append(hr_device)
        else:
            for i in range(max_sensors):
                hr_device = setup_sensor(node, device_id=0)
                hr_device.on_found = make_on_sensor_found(i, hr_device)
                hr_device.on_device_data = make_on_device_data(i, hr_device)
                sensors.append(hr_device)

        count = len(sensors)
        ids_info = ", ".join(str(d) for d in device_ids) if device_ids else "auto"
        print(f"Мониторинг {count} датчиков (ID: {ids_info}). Нажмите Ctrl-C для выхода.")

        node.start()
    except KeyboardInterrupt:
        print("\nОстановка мониторинга...")
    finally:
        for s in sensors:
            try:
                s.close_channel()
            except Exception:
                pass
        try:
            node.stop()
        except Exception:
            pass

    print("Завершено.")


def parse_args():
    """Разбирает аргументы командной строки.

    Supports:
        -n, --max-sensors  количество датчиков при auto-detect (default: 2)
        --ids              конкретные Device ID датчиков
        --log              включить отладочное логирование

    Returns:
        argparse.Namespace: распознанные аргументы.
    """
    parser = argparse.ArgumentParser(
        description="ANT+ монитор сердечного ритма (до 8 датчиков)"
    )
    parser.add_argument(
        "-n",
        "--max-sensors",
        type=int,
        default=2,
        help="Количество датчиков при авто-подключении (default: 2)",
    )
    parser.add_argument(
        "--ids",
        type=int,
        nargs="+",
        help="Device ID датчиков (если известны)",
    )
    parser.add_argument(
        "--log",
        action="store_true",
        help="Включить отладочное логирование openant",
    )
    return parser.parse_args()


if __name__ == "__main__":
    args = parse_args()
    main(
        max_sensors=args.max_sensors,
        device_ids=args.ids,
        enable_log=args.log,
    )
