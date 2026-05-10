# Сокращатель ссылок

## Описание

Сокращатель ссылок — сервис, который позволяет превратить длинную URL-адресную строку в короткий код и использовать его для быстрого перехода.

## Функциональность

- Генерация короткой ссылки из длинной
- Переадресация при переходе по короткой ссылке
- Сбор и просмотр статистики посещений
- Хранение данных в PostgreSQL

## Демо

Приложение доступно по ссылке: [go-project-278 на Render](https://go-project-278-3ycr.onrender.com).

## Переменные окружения для Render

Для запуска приложения должны быть указаны переменные окружения:

- `DATABASE_URL` - строка подключения к PostgreSQL
- `BASE_URL` — публичный URL приложения на Render (`BASE_URL=https://go-project-278-3ycr.onrender.com`)
- `SENTRY_DSN` - DSN для Sentry (опционально, если не указан - Sentry не подключается)
- `PORT` - внешний порт контейнера для Caddy (`PORT=80`)

## Мониторинг ошибок

Для мониторинга ошибок подключен [Sentry](https://sentry.io).

## Локальный запуск через Docker Compose

Docker Compose запускает весь локальный стек:

- PostgreSQL
- приложение в одном контейнере:
  - Caddy отдаёт UI и проксирует API-запросы
  - Go-приложение обрабатывает backend-запросы

Запустить проект:

```bash
make compose-up
```

После запуска веб-интерфейс будет доступен по адресу:

```text
http://localhost:8080/
```

Backend-маршрут для проверки:

```text
http://localhost:8080/ping
```

API также доступен через тот же origin:

```text
http://localhost:8080/api/links
```

## Локальный запуск через Docker

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

После запуска веб-интерфейс будет доступен по адресу:

```text
http://localhost:8080
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

В этом режиме UI через Caddy не запускается.

## Миграции

Миграции хранятся в директории `migrations/` и автоматически применяются приложением при старте через `goose`.

Для локальной разработки в Makefile оставлена команда:

```bash
make migrate-status
```

### Hexlet tests and linter status

[![Actions Status](https://github.com/xhrobj-hex/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/xhrobj-hex/go-project-278/actions)

## Project CI - Quality checks -> lint, build and test

[![(-_-) GO CI](https://github.com/xhrobj-hex/go-project-278/actions/workflows/go-ci.yml/badge.svg)](https://github.com/xhrobj-hex/go-project-278/actions/workflows/go-ci.yml)

## SonarQube statuses

[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=xhrobj-hex_go-project-278&metric=coverage)](https://sonarcloud.io/summary/new_code?id=xhrobj-hex_go-project-278)
