# Subs Collector

REST-сервис, агрегирующий информацию о пользовательских подписках.

---

## Схема БД

- **`services`** — справочник доступных подписок (уникальное поле `name`).
- **`user_subscriptions`** — подписки пользователей, ссылается на `services(id)`, хранит зафиксированную цену на момент
  оформления.

### Миграции

- **`001_create_tables.sql`** — инициализация всех необходимых для работы таблиц.
- **`001_test_data.sql`** — тестовые данные для проверки работоспособности.

---

## Конфигурация

Переменные окружения:

- `DATABASE_URL` — строка подключения к PostgreSQL.
- `PORT` — порт HTTP (по умолчанию `8080`).

---

## Запуск в Docker

1. Поднимите сервисы:
   ```bash
   docker compose up -d
   ```

2. Примените миграции:
   ```bash
   docker exec -i subs-db psql -U postgres -d subs_collector < ./migrations/001_create_tables.sql
   ```

3. Запустите приложение:
   ```bash
   docker compose up app
   ```