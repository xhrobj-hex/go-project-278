# Сокращатель ссылок

## Описание

Сокращатель ссылок — сервис, который позволяет превратить длинную URL-адресную строку в короткий код и использовать его для быстрого перехода.

## Функциональность

- Генерация короткой ссылки из длинной
- Переадресация при переходе по короткой ссылке
- Хранение данных в PostgreSQL
- Опционально: статистика переходов

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
