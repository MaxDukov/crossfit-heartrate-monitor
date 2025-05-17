#!/usr/bin/env python3
import sys
import time
import json
import threading
import logging
import signal
from flask import Flask, render_template
from flask_socketio import SocketIO
from openant.easy.node import Node
from openant.devices import ANTPLUS_NETWORK_KEY
from openant.devices.heart_rate import HeartRate
from collections import defaultdict

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
LOGGER = logging.getLogger(__name__)

APP = Flask(__name__)
SOCKETIO = SocketIO(APP)

# Глобальные переменные
SENSORS_DATA = defaultdict(lambda: {'times': [], 'heart_rates': []})
MAX_POINTS = 100
ACTIVE_SENSORS = set()
ANT_NODE = None
DEVICES = {}
INIT_TIMEOUT = 30  # Таймаут инициализации в секундах

class AntInitTimeout(Exception):
    """Исключение для таймаута инициализации ANT+"""
    pass

def timeout_handler(signum, frame):
    """Обработчик сигнала таймаута"""
    raise AntInitTimeout("Превышено время инициализации ANT+ устройства")

def setup_ant_device():
    """Инициализация ANT+ устройства"""
    global ANT_NODE
    try:
        LOGGER.info("Начало инициализации ANT+ устройства...")        
        # Устанавливаем обработчик таймаута
        signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(INIT_TIMEOUT)
        try:
            ANT_NODE = Node()
            ANT_NODE.start()
            LOGGER.info("ANT+ устройство успешно инициализировано")
        finally:
            # Отключаем таймер
            signal.alarm(0)
    except AntInitTimeout as e:
        LOGGER.error(f"Ошибка: {str(e)}")
        if ANT_NODE:
            try:
                ANT_NODE.stop()
            except:
                pass
        raise
    except Exception as e:
        LOGGER.error(f"Ошибка при инициализации ANT+ устройства: {e}")
        if ANT_NODE:
            try:
                ANT_NODE.stop()
            except:
                pass
        raise

def on_heart_rate_data(data):
    """Callback для обработки данных с датчика сердечного ритма"""
    SENSOR_ID = data.device_number
    HEART_RATE = data.heart_rate
    
    LOGGER.debug(f"Получены данные с датчика {SENSOR_ID}: пульс = {HEART_RATE} уд/мин")
  
    # Добавляем данные в соответствующий датчик
    SENSORS_DATA[SENSOR_ID]['heart_rates'].append(HEART_RATE)
    SENSORS_DATA[SENSOR_ID]['times'].append(time.time())
    
    # Ограничиваем количество точек на графике
    if len(SENSORS_DATA[SENSOR_ID]['heart_rates']) > MAX_POINTS:
        SENSORS_DATA[SENSOR_ID]['heart_rates'].pop(0)
        SENSORS_DATA[SENSOR_ID]['times'].pop(0)
    
    # Отправляем данные через WebSocket
    SOCKETIO.emit('heart_rate_data', {
        'sensor_id': SENSOR_ID,
        'heart_rate': HEART_RATE,
        'time': time.time()
    })
    
    # Добавляем новый датчик в список, если его там еще нет
    if SENSOR_ID not in ACTIVE_SENSORS:
        LOGGER.info(f"Обнаружен новый датчик: {SENSOR_ID}")
        ACTIVE_SENSORS.add(SENSOR_ID)
        SOCKETIO.emit('new_sensor', {'sensor_id': SENSOR_ID})

def continuous_scan():
    """Непрерывный поиск датчиков"""
    global DEVICES
    LOGGER.info("Запуск процесса сканирования датчиков...")
    
    while True:
        try:
            # Очищаем предыдущие устройства
            for DEVICE in DEVICES.values():
                DEVICE.close()
            DEVICES.clear()
            
            LOGGER.info("Поиск новых датчиков сердечного ритма...")
            
            # Создаем новый датчик сердечного ритма
            DEVICE = HeartRate(ANT_NODE)
            DEVICE.on_heart_rate_data = on_heart_rate_data
            DEVICE.open()
            
            DEVICES[0] = DEVICE
            SOCKETIO.emit('sensor_status', {'status': 'scanning'})
            LOGGER.info("Сканирование запущено")
            
        except Exception as e:
            ERROR_MSG = f"Ошибка при сканировании: {e}"
            LOGGER.error(ERROR_MSG)
            SOCKETIO.emit('sensor_status', {'status': 'error', 'message': str(e)})
        
        time.sleep(5)  # Пауза между сканированиями

def get_sensor_data(SENSOR_ID):
    """Получение данных конкретного датчика"""
    return SENSORS_DATA[SENSOR_ID]

def get_active_sensors():
    """Получение списка активных датчиков"""
    return list(ACTIVE_SENSORS)

def cleanup():
    """Очистка ресурсов"""
    LOGGER.info("Начало очистки ресурсов...")
    try:
        for DEVICE in DEVICES.values():
            DEVICE.close()
        if ANT_NODE:
            ANT_NODE.stop()
        LOGGER.info("Ресурсы успешно очищены")
    except Exception as e:
        LOGGER.error(f"Ошибка при очистке ресурсов: {e}")

def main():
    try:
        LOGGER.info("Запуск приложения мониторинга сердечного ритма...")
        setup_ant_device()
        SCAN_THREAD = threading.Thread(target=continuous_scan, daemon=True)
        SCAN_THREAD.start()
        LOGGER.info("Поток сканирования запущен")
        
        LOGGER.info("Запуск веб-сервера на порту 5000...")
        SOCKETIO.run(APP, host='0.0.0.0', port=5000, debug=True)
    except AntInitTimeout as e:
        LOGGER.error(f"Критическая ошибка: {str(e)}")
        sys.exit(1)
    except Exception as e:
        LOGGER.error(f"Ошибка при запуске веб-сервера: {e}")
    finally:
        cleanup()

if __name__ == '__main__':
    main()
 