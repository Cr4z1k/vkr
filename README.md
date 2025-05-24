# vkr

## Быстрый старт через Docker Compose

### 1. Клонируйте репозиторий

```powershell
git clone <адрес-репозитория>
cd vkr_back
```

### 2. Настройте переменные окружения

Скопируйте файл `.env.example` в `.env` и заполните значения:

- `DOCKER_NETWORK` — имя docker-сети (по умолчанию: pipeline_net)
- `KAFKA_BROKER` — адрес Kafka-брокера (например: kafka:9092)
- `ORCHESTRATOR_PORT=` - порт оркестратора (например: 8080)

### 3. Запустите сервисы

```powershell
docker-compose up --build
```

- Сервис orchestrator будет доступен на порту 8080.
- Kafka будет доступна внутри docker-сети как `kafka:9092`.

### 4. Остановка сервисов

```powershell
docker-compose down
```

### 5. Примечания
- Для взаимодействия с API используйте endpoint: `POST http://localhost:8080/setConfigs`
- Для корректной работы убедитесь, что порт 8080 свободен.
