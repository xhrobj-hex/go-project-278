# Быстрая проверка CRUD на Render

## 0. Проверить, что сервис жив

```bash
curl -i https://go-project-278-3ycr.onrender.com/ping
```

Ожидаем:

- `200 OK`
- тело: `pong`

---

## 1. Получить список ссылок

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links"
```

Ожидаем:

- `200 OK`
- JSON-массив

---

## 2. Создать ссылку с явным `short_name`

```bash
curl -i -X POST "https://go-project-278-3ycr.onrender.com/api/links" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.irregularverbs.ru/table/","short_name":"ivtable"}'
```

Ожидаем:

- `201 Created`
- JSON-объект с полями `id`, `original_url`, `short_name`, `short_url`

---

## 3. Создать ссылку без `short_name`

```bash
curl -i -X POST "https://go-project-278-3ycr.onrender.com/api/links" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.irregularverbs.ru/test/"}'
```

Ожидаем:

- `201 Created`
- `short_name` сгенерирован сервером

## 4. Снова получить список ссылок

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links"
```

Ожидаем:

- `200 OK`
- JSON-массив должен стать больше

---

## 5. Получить ссылку по `id`

Подставить реальный `id`, например `1`:

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links/1"
```

Ожидаем:

- `200 OK`
- JSON-объект ссылки

---

## 6. Обновить ссылку

Подставить существующий `id`, например `1`:

```bash
curl -i -X PUT "https://go-project-278-3ycr.onrender.com/api/links/1" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.irregularverbs.ru/test/","short_name":"itest"}'
```

Ожидаем:

- `200 OK`
- обновлённый JSON-объект ссылки

---

## 7. Проверить конфликт `short_name`

```bash
curl -i -X POST "https://go-project-278-3ycr.onrender.com/api/links" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.irregularverbs.ru/test/","short_name":"itest"}'
```

Ожидаем:

- `409 Conflict`
- тело:

```json
{"error":"short_name already exists"}
```

---

## 8. Удалить ссылку

Подставить существующий `id`, например `1`:

```bash
curl -i -X DELETE "https://go-project-278-3ycr.onrender.com/api/links/1"
```

Ожидаем:

- `204 No Content`

---

## 9. Проверить, что ссылка действительно удалена

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links/1"
```

Ожидаем:

- `404 Not Found`
- тело:

```json
{"error":"link not found"}
```

---

## 10. Проверить `400` для невалидного `id`

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links/bad"
```

Ожидаем:

- `400 Bad Request`
- тело:

```json
{"error":"invalid id"}
```
