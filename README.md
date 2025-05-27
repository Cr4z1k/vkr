# vkr

## Быстрый старт через Docker Compose

### 1. Клонируйте репозиторий

```powershell
git clone https://github.com/Cr4z1k/vkr_back
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

---

## Примеры работы с Kafka и Benthos

### Kafka produce example (Linux/macOS)
```sh
echo '{"user":"alice","links":["a","b","c"]}' \
  | docker-compose exec -T kafka \
      /opt/bitnami/kafka/bin/kafka-console-producer.sh \
        --bootstrap-server kafka:9092 \
        --topic foo
```

### Kafka produce example (Windows)
```powershell
'{"user":"alice","links":["a","b","c"]}' |
  docker-compose exec -T kafka `
    /opt/bitnami/kafka/bin/kafka-console-producer.sh `
      --bootstrap-server kafka:9092 `
      --topic foo
```

### Kafka consume example (Linux/macOS)
```sh
docker-compose exec -T kafka \
  /opt/bitnami/kafka/bin/kafka-console-consumer.sh \
    --bootstrap-server kafka:9092 \
    --topic foo_out \
    --from-beginning \
    --max-messages 1
```

### Kafka consume example (Windows)
```powershell
docker-compose exec -T kafka `
  /opt/bitnami/kafka/bin/kafka-console-consumer.sh `
    --bootstrap-server kafka:9092 `
    --topic foo_out `
    --from-beginning `
    --max-messages 1
```

### Benthos input example
```yaml
input:
  kafka:
    addresses:       ["kafka:9092"]
    topics:          ["foo"]
    consumer_group:  "test_group"
```

### Benthos pipeline example
```yaml
- mapping: |
    root = this
```

### Benthos output example
```yaml
output:
  kafka:
    addresses: ["kafka:9092"]
    topic:     "foo_out"
```
