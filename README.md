# Сокращатель ссылок

## Описание

Сокращатель ссылок — сервис, который позволяет превратить длинную URL-адресную строку в короткий код и использовать его для быстрого перехода.

## Функциональность

- Генерация короткой ссылки из длинной
- Переадресация при переходе по короткой ссылке
- Хранение данных в PostgreSQL

## Демо

Приложение доступно по ссылке: [go-project-278 на Render](https://go-project-278-3ycr.onrender.com/ping).

## Переменные окружения для Render

Для запуска приложения должны быть указаны переменные окружения:

- `DATABASE_URL` - строка подключения к PostgreSQL
- `BASE_URL` — публичный URL приложения на Render (`BASE_URL=https://go-project-278-3ycr.onrender.com`)
- `SENTRY_DSN` - DSN для Sentry (опционально, если не указан - sentry просто не подключится)
- `PORT` - порт HTTP-сервера (опционально, локально дефолтный `8080`, а на Render это автоматический env)

## Мониторинг ошибок

Для мониторинга ошибок подключен [Sentry](https://sentry.io).

## Локальный запуск через Docker Compose

Docker Compose запускает весь локальный стек:

- PostgreSQL
- backend-приложение на Go
- frontend-приложение из пакета `@hexlet/project-url-shortener-frontend`

Запустить проект:

```bash
make compose-up
```

После запуска frontend будет доступен по адресу:

```text
http://localhost:5173/
```

Backend будет доступен по адресу:

```text
http://localhost:8080/ping
```

## Локальный запуск Backend через Docker

Поднять PostgreSQL:

```bash
make postgres-up
```

или для созданного ранее контейнера:

```bash
make postgres-start
```

Собрать Docker-образ приложения:

```bash
make docker-build
```

Запустить приложение в Docker-контейнере:

```bash
make docker-run
```

После запуска приложение будет доступно по адресу:

```text
http://localhost:8080/ping
```

## Локальный запуск Backend'а "руками"

Поднять PostgreSQL:

```bash
make postgres-up
```

или для созданного ранее контейнера:

```bash
make postgres-start
```

Запустить приложение локально без Docker:

```bash
make run
```

После запуска приложение будет доступно по адресу:

```text
http://localhost:8080/ping
```

## Миграции

Миграции хранятся в директории `migrations/` и автоматически применяются приложением при старте через `goose`.

Для локальной разработки в Makefile оставлена команда:

```bash
make migrate-status
```

## Документация

- [Быстрый чек-лист CRUD на Render](docs/render-crud-checklist_quick.md)
- [Проверка CRUD на Render](docs/render-crud-checklist.md)

## Пример API

Создание короткой ссылки:

```bash
curl -X POST https://your-app.onrender.com/api/links \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com/long-url"}'
```

Пример ответа:

```json
{
  "id": 1,
  "original_url": "https://example.com/long-url",
  "short_name": "exmpl",
  "short_url": "https://your-app.onrender.com/exmpl",
  "created_at": "2025-01-01T12:34:56Z"
}
```

## Переход по короткой ссылке

При переходе по короткой ссылке, например:

```text
https://your-app.onrender.com/exmpl
```

пользователь будет перенаправлен на исходный адрес:

```text
https://example.com/long-url
```

### Hexlet tests and linter status

[![Actions Status](https://github.com/xhrobj-hex/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/xhrobj-hex/go-project-278/actions)

## Project CI - Quality checks -> lint, build and test

[![(-_-) GO CI](https://github.com/xhrobj-hex/go-project-278/actions/workflows/go-ci.yml/badge.svg)](https://github.com/xhrobj-hex/go-project-278/actions/workflows/go-ci.yml)

## SonarQube statuses

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=xhrobj-hex_go-project-278&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=xhrobj-hex_go-project-278)

[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=xhrobj-hex_go-project-278&metric=coverage)](https://sonarcloud.io/summary/new_code?id=xhrobj-hex_go-project-278)
