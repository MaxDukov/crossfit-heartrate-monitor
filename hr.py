#!/usr/bin/env python3
import sys
import time
import threading
import logging
import signal
import sqlite3
from flask import Flask, render_template, jsonify, request
from flask_socketio import SocketIO
from collections import defaultdict
from openant.easy.node import Node
from openant.devices import ANTPLUS_NETWORK_KEY
from openant.devices.heart_rate import HeartRate, HeartRateData

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
socketio = SocketIO(app, async_mode='threading')

# Глобальные переменные
sensors_data = defaultdict(lambda: {'times': [], 'heart_rates': []})
MAX_POINTS = 100
active_sensors = set()
node = None
device = None
running = True

# Инициализация базы данных
def init_db():
    """Initialize the SQLite database with required tables"""
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    c.execute('''
        CREATE TABLE IF NOT EXISTS athletes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            first_name TEXT NOT NULL,
            last_name TEXT NOT NULL,
            max_hr INTEGER,
            sensor_id INTEGER UNIQUE,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    conn.commit()
    conn.close()

# Маршруты Flask
@app.route('/')
def index():
    """Render main monitoring page"""
    return render_template('index.html')

@app.route('/athletes')
def athletes():
    """Render athlete management page"""
    return render_template('athletes.html')

@app.route('/api/sensors')
def get_sensors():
    """API endpoint для получения списка активных датчиков"""
    return jsonify(list(active_sensors))

@app.route('/api/sensor/<int:sensor_id>')
def get_sensor(sensor_id):
    """API endpoint для получения данных конкретного датчика"""
    if sensor_id not in sensors_data:
        return jsonify({'error': 'Sensor not found'}), 404
    return jsonify(sensors_data[sensor_id])

@app.route('/api/athletes', methods=['GET'])
def get_athletes():
    """Get list of all athletes"""
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    c.execute('SELECT * FROM athletes ORDER BY last_name, first_name')
    athletes = [{'id': row[0], 'first_name': row[1], 'last_name': row[2], 'max_hr': row[3]} 
               for row in c.fetchall()]
    conn.close()
    return jsonify(athletes)

@app.route('/api/athletes', methods=['POST'])
def add_athlete():
    """Add new athlete"""
    data = request.json
    if not all(k in data for k in ('first_name', 'last_name', 'max_hr')):
        return jsonify({'error': 'Missing required fields'}), 400
    
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    try:
        c.execute('''
            INSERT INTO athletes (first_name, last_name, max_hr)
            VALUES (?, ?, ?)
        ''', (data['first_name'], data['last_name'], data['max_hr']))
        conn.commit()
        athlete_id = c.lastrowid
        conn.close()
        return jsonify({'id': athlete_id, 'message': 'Athlete added successfully'})
    except Exception as e:
        conn.close()
        return jsonify({'error': str(e)}), 500

@app.route('/api/athletes/<int:athlete_id>', methods=['PUT'])
def update_athlete(athlete_id):
    """Update athlete information"""
    data = request.json
    if not all(k in data for k in ('first_name', 'last_name', 'max_hr')):
        return jsonify({'error': 'Missing required fields'}), 400
    
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    try:
        c.execute('''
            UPDATE athletes 
            SET first_name = ?, last_name = ?, max_hr = ?
            WHERE id = ?
        ''', (data['first_name'], data['last_name'], data['max_hr'], athlete_id))
        conn.commit()
        conn.close()
        return jsonify({'message': 'Athlete updated successfully'})
    except Exception as e:
        conn.close()
        return jsonify({'error': str(e)}), 500

@app.route('/api/athletes/<int:athlete_id>', methods=['DELETE'])
def delete_athlete(athlete_id):
    """Delete athlete"""
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    try:
        c.execute('DELETE FROM athletes WHERE id = ?', (athlete_id,))
        conn.commit()
        conn.close()
        return jsonify({'message': 'Athlete deleted successfully'})
    except Exception as e:
        conn.close()
        return jsonify({'error': str(e)}), 500

@app.route('/api/athletes/<int:athlete_id>/sensor', methods=['POST'])
def bind_sensor(athlete_id):
    """Bind sensor to athlete"""
    data = request.json
    if 'sensor_id' not in data:
        return jsonify({'error': 'Missing sensor_id'}), 400
    
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    try:
        # First, unbind the sensor from any athlete
        c.execute('UPDATE athletes SET sensor_id = NULL WHERE sensor_id = ?', (data['sensor_id'],))
        # Then bind it to the new athlete
        c.execute('UPDATE athletes SET sensor_id = ? WHERE id = ?', (data['sensor_id'], athlete_id))
        conn.commit()
        conn.close()
        return jsonify({'message': 'Sensor bound successfully'})
    except Exception as e:
        conn.close()
        return jsonify({'error': str(e)}), 500

@app.route('/api/athletes/<int:athlete_id>/sensor', methods=['DELETE'])
def unbind_sensor(athlete_id):
    """Unbind sensor from athlete"""
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    try:
        c.execute('UPDATE athletes SET sensor_id = NULL WHERE id = ?', (athlete_id,))
        conn.commit()
        conn.close()
        return jsonify({'message': 'Sensor unbound successfully'})
    except Exception as e:
        conn.close()
        return jsonify({'error': str(e)}), 500

@app.route('/api/sensor/<int:sensor_id>/athlete')
def get_sensor_athlete(sensor_id):
    """Get athlete bound to sensor"""
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    c.execute('SELECT id, first_name, last_name, max_hr FROM athletes WHERE sensor_id = ?', (sensor_id,))
    athlete = c.fetchone()
    conn.close()
    
    if athlete:
        return jsonify({
            'id': athlete[0],
            'first_name': athlete[1],
            'last_name': athlete[2],
            'max_hr': athlete[3]
        })
    return jsonify(None)

def on_found():
    """Callback при обнаружении датчика"""
    global device
    if device:
        logger.info(f"Датчик {device} обнаружен и принимает данные")
        socketio.emit('sensor_status', {'status': 'scanning'})

def on_device_data(page: int, page_name: str, data):
    """Handle incoming data from heart rate sensor"""
    if isinstance(data, HeartRateData):
        sensor_id = 1  # data.device_number
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
            # Отправляем событие только один раз при первом обнаружении
            socketio.emit('new_sensor', {'sensor_id': sensor_id})
            # Отправляем начальные данные датчика
            socketio.emit('sensor_status', {'status': 'connected', 'sensor_id': sensor_id})
            
            # Проверяем, привязан ли датчик к спортсмену
            conn = sqlite3.connect('athletes.db')
            c = conn.cursor()
            c.execute('SELECT first_name, last_name FROM athletes WHERE sensor_id = ?', (sensor_id,))
            athlete = c.fetchone()
            conn.close()
            
            if athlete:
                socketio.emit('sensor_athlete', {
                    'sensor_id': sensor_id,
                    'athlete': {
                        'first_name': athlete[0],
                        'last_name': athlete[1]
                    }
                })

def ant_worker():
    """Функция для работы с ANT+ в отдельном потоке"""
    global node, device, running
    
    try:
        logger.info("Инициализация ANT+ устройства...")
        node = Node()
        node.set_network_key(0x00, ANTPLUS_NETWORK_KEY)
        
        device = HeartRate(node)
        device.on_found = on_found
        device.on_device_data = on_device_data
        
        logger.info("Запуск ANT+ устройства...")
        node.start()
        
        # Держим поток активным
        while running:
            time.sleep(0.1)
            
    except Exception as e:
        logger.error(f"Ошибка в работе ANT+ устройства: {e}")
        socketio.emit('sensor_status', {'status': 'error', 'message': str(e)})
    finally:
        cleanup()

def cleanup():
    """Очистка ресурсов при завершении"""
    global running
    running = False
    
    logger.info("Очистка ресурсов...")
    if device:
        device.close_channel()
    if node:
        node.stop()

def signal_handler(signum, frame):
    """Обработчик сигналов завершения"""
    logger.info("Получен сигнал завершения работы...")
    cleanup()
    sys.exit(0)

def main():
    # Инициализация базы данных
    init_db()
    
    # Устанавливаем обработчики сигналов
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        # Запускаем ANT+ в отдельном потоке
        ant_thread = threading.Thread(target=ant_worker, daemon=True)
        ant_thread.start()
        logger.info("ANT+ поток запущен")

        # Запускаем веб-сервер
        logger.info("Запуск веб-сервера на порту 5000...")
        socketio.run(app, host='0.0.0.0', port=5000, debug=False)
        
    except Exception as e:
        logger.error(f"Ошибка при запуске приложения: {e}")
    finally:
        cleanup()

if __name__ == "__main__":
    main()