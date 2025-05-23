# CrossFit Heart Rate Monitor

A real-time heart rate monitoring system for CrossFit training sessions. The application supports multiple ANT+ heart rate sensors and provides real-time visualization of heart rate data for multiple athletes.

## Features

- Real-time heart rate monitoring
- Support for multiple ANT+ sensors
- Athlete management with individual maximum heart rate thresholds
- Sensor binding to athletes
- Real-time data visualization with color-coded heart rate zones
- Data storage in SQLite database
- Dark/Light theme support
- Lightweight and resource-efficient design
- Cross-platform compatibility
- Raspberry Pi ready

## Key Advantages

- **Lightweight**: The application is designed to run efficiently on low-power devices like Raspberry Pi, making it perfect for small gyms and training facilities
- **Cross-Platform**: Works seamlessly on Linux, Windows, and macOS
- **Resource Efficient**: Minimal system requirements and optimized performance
- **Easy Deployment**: Simple setup process with no complex dependencies
- **Scalable**: Can handle multiple sensors simultaneously without performance degradation
- **Web-Based Interface**: Accessible from any device with a web browser
- **Offline Capable**: Works without internet connection once deployed

## Requirements

- Python 3.7+
- ANT+ USB adapter
- Compatible heart rate sensors
- Dependencies:
  - flask
  - flask-socketio
  - openant

## Installation

1. Clone the repository:
```bash
git clone https://github.com/MaxDukov/crossfit-heartrate-monitor.git
cd crossfit-heartrate-monitor
```

2. Create a virtual environment and activate it:
```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

3. Install dependencies:
```bash
pip install -r requirements.txt
```

## Usage

1. Connect your ANT+ USB adapter
2. Run the application:
```bash
python hr.py
```

3. Open your web browser and navigate to `http://[your-device-ip]:5000`

4. Add athletes through the "Athlete Management" page:
   - Enter first name, last name, and maximum heart rate
   - The maximum heart rate is used to calculate heart rate zones

5. Monitor heart rates:
   - The system automatically detects and connects to available ANT+ sensors
   - Bind sensors to athletes using the "Bind to Athlete" button
   - Real-time heart rate data is displayed with color coding:
     - Green: 0-60% of max heart rate
     - Yellow: 60-85% of max heart rate
     - Red: >85% of max heart rate

## Raspberry Pi Setup

The application is optimized for Raspberry Pi deployment:

1. Install Raspberry Pi OS (preferably Lite version)
2. Install required packages:
```bash
sudo apt-get update
sudo apt-get install python3-pip python3-venv
```

3. Follow the standard installation steps above
4. For automatic startup on boot, create a systemd service:
```bash
sudo nano /etc/systemd/system/hr-monitor.service
```

Add the following content:
```ini
[Unit]
Description=CrossFit Heart Rate Monitor
After=network.target

[Service]
User=pi
WorkingDirectory=/path/to/crossfit-hr-monitor
ExecStart=/path/to/crossfit-hr-monitor/venv/bin/python hr.py
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl enable hr-monitor
sudo systemctl start hr-monitor
```

## Project Structure

- `hr.py` - Main application file
- `templates/` - HTML templates
  - `index.html` - Main monitoring page
  - `athletes.html` - Athlete management page
- `static/` - Static files (CSS, JavaScript)
- `database.db` - SQLite database file

## ToDo list
- To create docker to make setup more comfortable. (in progress)
- To add customisation function (club logo)
- To add photo upload function (athlet photo)
- To add training history function, probably with export.

## Known bugs
- Backend didn't send "zero heartrate" when lost sensor. In progress
- WSGI server should be used in normal way. 

## License

GPL License