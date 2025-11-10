# telegram-normalizer

## Что делает

1. Подписывается на Kafka-топик обновлений `TOPIC_NAME_TELEGRAM_UPDATES`, десериализует Telegram `Update` и выносит в `contract.NormalizedMessage`.
2. Если в сообщении есть медиа (фото или документ), скачивает файл через Telegram API, определяет MIME по расширению и загружает в хранилище S3 (`internal/s3`).
3. Сформированное `NormalizedMessage` дополняется ссылками на S3 и публикуется в Kafka-топик `TOPIC_NAME_NORMALIZED_MSG`.
4. Логи идут через `internal/logger`, ошибки всегда возвращаются и логируются в Kafka consumer.

## Запуск

1. Подготовьте `.env` с нужными переменными (см. ниже) и загрузите его `export $(cat .env | xargs)` или `set -a && source .env && set +a`.
2. Соберите и запустите вручную:
   ```bash
   go run ./cmd/telegram-normalizer
   ```
3. Либо соберите и запустите докер-образ:
   ```bash
   docker build -t telegram-normalizer .
   docker run --rm --env-file .env telegram-normalizer
   ```

## Переменные окружения

Все переменные обязательны, кроме `SASL_*`, если Kafka не требует авторизации.

- `TOKEN` — токен Telegram-бота.
- `BOOTSTRAP_SERVERS_VALUE` — `host:port[,host:port]` для подключения к Kafka.
- `GROUP_ID_TELEGRAM_NORMALIZER` — consumer group для обработки обновлений.
- `TOPIC_NAME_TELEGRAM_UPDATES` — входящий топик Telegram-обновлений.
- `TOPIC_NAME_NORMALIZED_MSG` — топик для нормализованных сообщений (Kafka producer отправляет сюда).
- `CLIENT_ID_TELEGRAM_NORMALIZER` — идентификатор Kafka-клиента (продюсер и консьюмер).
- `SASL_USERNAME` и `SASL_PASSWORD` — если используется SASL/PLAIN авторизация.
- `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` — параметры доступа к S3-совместимому хранилищу.
- `S3_USE_SSL` — `true/false`, включает TLS при обращении к S3.

## Примечания

- Telegram-парсер (`internal/telegram/parser.go`) извлекает текст, метаданные и только самый большой фоторазмер.
- Скачивание файлов и загрузка в S3 (`internal/telegram/downloader.go` + `internal/s3/uploader.go`) выполняются перед публикацией в Kafka.
- `contract.NormalizedMessage` содержит оригинальный `Update` в `OriginalUpdate` для последующего потребления.
- Kafka producer/consumer используют SCRAM-SHA512 (`internal/messaging`), поэтому даже пустые SASL-поля валидны, пока брокер разрешает.
