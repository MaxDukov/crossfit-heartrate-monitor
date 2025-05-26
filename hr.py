#!/usr/bin/env python3
import sys
import time
import threading
import logging
import signal
from openant.easy.node import Node
from openant.devices import ANTPLUS_NETWORK_KEY
from openant.devices.heart_rate import HeartRate, HeartRateData
from web_app import update_sensor_data, emit_sensor_status, emit_new_sensor, emit_sensor_athlete

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

# Global variables
node = None
device = None
running = True
HEARTBEAT_INTERVAL = 3  # seconds

def on_found():
    """Callback when sensor is found"""
    global device
    if device:
        logger.info(f"Sensor {device} found and receiving data")
        emit_sensor_status('scanning')

def check_sensor_timeouts():
    """Check for sensor timeouts and send zero readings if needed"""
    current_time = time.time()
    # TODO: Implement timeout checking logic

def on_device_data(page: int, page_name: str, data):
    """Handle incoming data from heart rate sensor"""
    if isinstance(data, HeartRateData):
        # Получаем ID датчика из имени устройства
        sensor_id = int(str(device).split('_')[-1])  # Извлекаем ID из имени устройства (например, heart_rate_52881 -> 52881)
        heart_rate = data.heart_rate
        current_time = time.time()

        logger.debug(f"Received data from sensor {sensor_id}: heart rate = {heart_rate} bpm")

        # Update sensor data through web interface
        update_sensor_data(sensor_id, heart_rate, current_time)

        # Add new sensor to list if not already present
        emit_new_sensor(sensor_id)
        emit_sensor_status('connected', sensor_id)
            
        # Check if sensor is bound to athlete
        import sqlite3
        conn = sqlite3.connect('athletes.db')
        c = conn.cursor()
        c.execute('SELECT first_name, last_name FROM athletes WHERE sensor_id = ?', (sensor_id,))
        athlete = c.fetchone()
        conn.close()
        
        if athlete:
            emit_sensor_athlete(sensor_id, {
                'first_name': athlete[0],
                'last_name': athlete[1]
            })

def ant_worker():
    """Function for ANT+ work in separate thread"""
    global node, device, running
    
    try:
        logger.info("Initializing ANT+ device...")
        node = Node()
        node.set_network_key(0x00, ANTPLUS_NETWORK_KEY)
        
        device = HeartRate(node)
        device.on_found = on_found
        device.on_device_data = on_device_data
        
        logger.info("Starting ANT+ device...")
        node.start()
        
        # Keep thread active and check for timeouts
        while running:
            check_sensor_timeouts()
            time.sleep(0.1)
            
    except Exception as e:
        logger.error(f"Error in ANT+ device operation: {e}")
        emit_sensor_status('error', str(e))
    finally:
        cleanup()

def cleanup():
    """Clean up resources on exit"""
    global running
    running = False
    
    logger.info("Cleaning up resources...")
    if device:
        device.close_channel()
    if node:
        node.stop()

def signal_handler(signum, frame):
    """Signal handler for graceful shutdown"""
    logger.info("Received shutdown signal...")
    cleanup()
    sys.exit(0)

def main():
    # Set up signal handlers
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        # Start ANT+ in separate thread
        ant_thread = threading.Thread(target=ant_worker, daemon=True)
        ant_thread.start()
        logger.info("ANT+ thread started")

        # Start web server
        from web_app import run_web_server
        logger.info("Starting web server on port 5000...")
        run_web_server()
        
    except Exception as e:
        logger.error(f"Error starting application: {e}")
    finally:
        cleanup()

if __name__ == "__main__":
    main()