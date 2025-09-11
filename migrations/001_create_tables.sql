-- Таблица справочника сервисов
CREATE TABLE IF NOT EXISTS services
(
    id         SERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Таблица подписок пользователей, ссылается на services
CREATE TABLE IF NOT EXISTS user_subscriptions
(
    id         SERIAL PRIMARY KEY,
    service_id INT         NOT NULL REFERENCES services (id) ON DELETE RESTRICT,
    price      INTEGER     NOT NULL, -- фиксация цены на момент подписки
    user_id    UUID        NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_service_id ON user_subscriptions (service_id);