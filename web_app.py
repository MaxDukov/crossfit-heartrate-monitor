#!/usr/bin/env python3
import sqlite3
import logging
from flask import Flask, render_template, jsonify, request
from flask_socketio import SocketIO

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

app = Flask(__name__)
socketio = SocketIO(app, async_mode='threading')

# Глобальная переменная для хранения данных от датчиков
sensors_data = {}

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

# Flask routes
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
    """API endpoint to get list of active sensors"""
    return jsonify(list(sensors_data.keys()))

@app.route('/api/sensor/<int:sensor_id>')
def get_sensor(sensor_id):
    """API endpoint to get data for specific sensor"""
    if sensor_id not in sensors_data:
        return jsonify({'error': 'Sensor not found'}), 404
    return jsonify(sensors_data[sensor_id])

@app.route('/api/athletes', methods=['GET'])
def get_athletes():
    """Get list of all athletes"""
    conn = sqlite3.connect('athletes.db')
    c = conn.cursor()
    c.execute('SELECT id, first_name, last_name, max_hr, sensor_id FROM athletes ORDER BY last_name, first_name')
    athletes = [{'id': row[0], 'first_name': row[1], 'last_name': row[2], 'max_hr': row[3], 'sensor_id': row[4]} 
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
        # Получаем текущий sensor_id перед обновлением
        c.execute('SELECT sensor_id FROM athletes WHERE id = ?', (athlete_id,))
        current_sensor = c.fetchone()
        
        # Обновляем данные спортсмена
        c.execute('''
            UPDATE athletes 
            SET first_name = ?, last_name = ?, max_hr = ?
            WHERE id = ?
        ''', (data['first_name'], data['last_name'], data['max_hr'], athlete_id))
        
        # Если у спортсмена был привязан датчик, сохраняем эту привязку
        if current_sensor and current_sensor[0]:
            c.execute('UPDATE athletes SET sensor_id = ? WHERE id = ?', (current_sensor[0], athlete_id))
            
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

def update_sensor_data(sensor_id, heart_rate, timestamp):
    """Update sensor data and emit through WebSocket"""
    if sensor_id not in sensors_data:
        sensors_data[sensor_id] = {'times': [], 'heart_rates': []}
    
    sensors_data[sensor_id]['heart_rates'].append(heart_rate)
    sensors_data[sensor_id]['times'].append(timestamp)
    
    # Limit number of points
    if len(sensors_data[sensor_id]['heart_rates']) > 100:
        sensors_data[sensor_id]['heart_rates'].pop(0)
        sensors_data[sensor_id]['times'].pop(0)
    
    socketio.emit('heart_rate_data', {
        'sensor_id': sensor_id,
        'heart_rate': heart_rate,
        'time': timestamp
    })

def emit_sensor_status(status, sensor_id=None):
    """Emit sensor status through WebSocket"""
    data = {'status': status}
    if sensor_id is not None:
        data['sensor_id'] = sensor_id
    socketio.emit('sensor_status', data)

def emit_new_sensor(sensor_id):
    """Emit new sensor event through WebSocket"""
    socketio.emit('new_sensor', {'sensor_id': sensor_id})

def emit_sensor_athlete(sensor_id, athlete):
    """Emit sensor-athlete binding through WebSocket"""
    socketio.emit('sensor_athlete', {
        'sensor_id': sensor_id,
        'athlete': athlete
    })

def run_web_server():
    """Run the web server"""
    init_db()
    socketio.run(app, host='0.0.0.0', port=5000, debug=False)

if __name__ == "__main__":
    run_web_server() 