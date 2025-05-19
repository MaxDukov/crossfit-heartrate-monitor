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
logger = logging.getLogger(__name__)

app = Flask(__name__)
socketio = SocketIO(app)

# Глобальные переменные
sensors_data = defaultdict(lambda: {'times': [], 'heart_rates': []})
MAX_POINTS = 100
active_sensors = set()
ant_node = None
devices = {}
INIT_TIMEOUT = 30  # Таймаут инициализации в секундах

class AntInitTimeout(Exception):
    """Исключение для таймаута инициализации ANT+"""
    pass

def timeout_handler(signum, frame):
    """Обработчик сигнала таймаута"""
    raise AntInitTimeout("Превышено время инициализации ANT+ устройства")

def setup_ant_device():
    """Инициализация ANT+ устройства"""
    global ant_node
    try:
        logger.info("Начало инициализации ANT+ устройства...")        
        # Устанавливаем обработчик таймаута
        signal.signal(signal.SIGALRM, timeout_handler)
        signal.alarm(INIT_TIMEOUT)
        try:
            ant_node = Node()
            ant_node.set_network_key(0x00, ANTPLUS_NETWORK_KEY)
            ant_node.start()
            logger.info("ANT+ устройство успешно инициализировано")
        finally:
            # Отключаем таймер
            signal.alarm(0)
    except AntInitTimeout as e:
        logger.error(f"Ошибка: {str(e)}")
        if ant_node:
            try:
                ant_node.stop()
            except:
                pass
        raise
    except Exception as e:
        logger.error(f"Ошибка при инициализации ANT+ устройства: {e}")
        if ant_node:
            try:
                ant_node.stop()
            except:
                pass
        raise

def on_heart_rate_data(data):
    """Callback для обработки данных с датчика сердечного ритма"""
    sensor_id = data.device_number
    heart_rate = data.heart_rate
    
    logger.debug(f"Получены данные с датчика {sensor_id}: пульс = {heart_rate} уд/мин")
  
    # Добавляем данные в соответствующий датчик
    sensors_data[sensor_id]['heart_rates'].append(heart_rate)
    sensors_data[sensor_id]['times'].append(time.time())
    
    # Ограничиваем количество точек на графике
    if len(sensors_data[sensor_id]['heart_rates']) > MAX_POINTS:
        sensors_data[sensor_id]['heart_rates'].pop(0)
        sensors_data[sensor_id]['times'].pop(0)
    
    # Отправляем данные через WebSocket
    socketio.emit('heart_rate_data', {
        'sensor_id': sensor_id,
        'heart_rate': heart_rate,
        'time': time.time()
    })
    
    # Добавляем новый датчик в список, если его там еще нет
    if sensor_id not in active_sensors:
        logger.info(f"Обнаружен новый датчик: {sensor_id}")
        active_sensors.add(sensor_id)
        socketio.emit('new_sensor', {'sensor_id': sensor_id})

def continuous_scan():
    """Непрерывный поиск датчиков"""
    global devices
    logger.info("Запуск процесса сканирования датчиков...")
    
    while True:
        try:
            # Очищаем предыдущие устройства
            for device in devices.values():
                device.close()
            devices.clear()
            
            logger.info("Поиск новых датчиков сердечного ритма...")
            
            # Создаем новый датчик сердечного ритма
            device = HeartRate(ant_node)
            device.on_heart_rate_data = on_heart_rate_data
            device.open()
            
            devices[0] = device
            socketio.emit('sensor_status', {'status': 'scanning'})
            logger.info("Сканирование запущено")
            
            # Ждем данные от датчика
            while True:
                time.sleep(1)
                
        except Exception as e:
            error_msg = f"Ошибка при сканировании: {e}"
            logger.error(error_msg)
            socketio.emit('sensor_status', {'status': 'error', 'message': str(e)})
            time.sleep(5)  # Пауза перед повторной попыткой

def get_sensor_data(sensor_id):
    """Получение данных конкретного датчика"""
    return sensors_data[sensor_id]

def get_active_sensors():
    """Получение списка активных датчиков"""
    return list(active_sensors)

def cleanup():
    """Очистка ресурсов"""
    logger.info("Начало очистки ресурсов...")
    try:
        for device in devices.values():
            device.close()
        if ant_node:
            ant_node.stop()
        logger.info("Ресурсы успешно очищены")
    except Exception as e:
        logger.error(f"Ошибка при очистке ресурсов: {e}")

def main():
    try:
        logger.info("Запуск приложения мониторинга сердечного ритма...")
        setup_ant_device()
        scan_thread = threading.Thread(target=continuous_scan, daemon=True)
        scan_thread.start()
        logger.info("Поток сканирования запущен")
        
        logger.info("Запуск веб-сервера на порту 5000...")
        socketio.run(app, host='0.0.0.0', port=5000, debug=True)
    except AntInitTimeout as e:
        logger.error(f"Критическая ошибка: {str(e)}")
        sys.exit(1)
    except Exception as e:
        logger.error(f"Ошибка при запуске веб-сервера: {e}")
    finally:
        cleanup()

if __name__ == '__main__':
    main()
 