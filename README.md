# TwitchBot

Многозадачный Twitch‑бот с аналитикой чата на Go. Сохраняет сообщения в PostgreSQL (или в памяти), отвечает на команды и автоматически переподключается при обрывах связи.

## Возможности

- Подключение к Twitch IRC с автоматическим переподключением и keep‑alive
- Сохранение всех сообщений чата в PostgreSQL или в памяти (in‑memory)
- Потоковые ответы (reply‑теги) — бот отвечает напрямую на сообщение пользователя
- Команды чата: `!ping`, `!echo <текст>`, `!top`, `!randomclip <канал>`, `!help`
- Гибкое хранилище: `--store=postgres` (по умолчанию) или `--store=memory`
- Полная контейнеризация (Docker, docker‑compose)
- Миграции базы данных (golang‑migrate)
- Graceful shutdown и idle‑таймер переподключения

## Быстрый старт

- git clone https://github.com/dobro09/TwitchBot.git && cd TwitchBot
- cp .env.example .env   # заполните токены и настройки
- make compose-up         # запуск в Docker (бот + БД)
- make service-run        # с PostgreSQL
- make inmemory-runbot STORE_TYPE=memory  # с in‑memory хранилищем

## Команды бота

| Команда               | Пример использования         | Ответ бота |
|-----------------------|------------------------------|------------|
| `!ping`               | `!ping`                      | `Pong!` |
| `!echo <текст>`       | `!echo Всем привет!`         | `Всем привет!` |
| `!top`                | `!top`                       | `user1: 5 сообщений, user2: 3 сообщений, user3: 1 сообщений` |
| `!randomclip <канал>` | `!randomclip user`           | `https://clips.twitch.tv/...` |
| `!help`               | `!help`                      | Список доступных команд |

## Переменные окружения

| Переменная           | Обязательность | Описание |
|----------------------|----------------|----------|
| `TWITCH_TOKEN`       | Да             | OAuth‑токен аккаунта бота |
| `BOT_NAME`           | Да             | Логин бота |
| `CHANNEL`            | Да             | Канал для подключения |
| `DATABASE_URL`       | Нет (для in‑memory) | Строка подключения к PostgreSQL |
| `TWITCH_CLIENT_ID`   | Нет (для !randomclip) | Client ID приложения Twitch |
| `TWITCH_CLIENT_SECRET` | Нет (для !randomclip) | Client Secret приложения Twitch |
| `STORE_TYPE`         | Нет             | `postgres` (по умолчанию) или `memory` |

## Архитектура

Проект построен по принципам Clean Architecture и разделён на независимые слои:

- `cmd/twitchbot/` – точка входа (main.go)
- `internal/model/` – доменные модели (Message, Command, UserStat, RAWMessage)
- `internal/store/` – интерфейс MessageStore + реализации (PostgresStore, InMemoryStore)
- `internal/usecase/` – бизнес-логика (ChatUsecase, обработчики команд)
- `internal/delivery/botdelivery/` – IRC-транспорт (подключение, чтение/запись, пинг-понг)
- `internal/twitchapi/` – инфраструктура: клиент Twitch API (OAuth, клипы)
- `internal/utils/` – парсер IRC-сообщений (RawMessage)
- `db/migrations/` – SQL-миграции
