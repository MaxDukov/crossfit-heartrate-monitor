# Используем Python 3.11 как базовый образ
FROM python:3.11-slim

# Установка необходимых системных зависимостей
RUN apt-get update && apt-get install -y \
    libusb-1.0-0 \
    libusb-1.0-0-dev \
    && rm -rf /var/lib/apt/lists/*

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY requirements.txt .

# Устанавливаем зависимости Python
RUN pip install --no-cache-dir -r requirements.txt

# Копируем исходный код приложения
COPY . .

# Создаем директорию для логов
RUN mkdir -p /app/logs

# Создаем директорию для базы данных
RUN mkdir -p /app/data

# Устанавливаем переменные окружения
ENV PYTHONUNBUFFERED=1
ENV FLASK_APP=hr.py
ENV FLASK_ENV=production

# Открываем порт для веб-интерфейса
EXPOSE 5000

# Запускаем приложение
CMD ["python", "hr.py"] 
