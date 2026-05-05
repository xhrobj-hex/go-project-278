# Проверка CRUD на Render

## 0. Проверить, что сервис жив

```bash
curl -i https://go-project-278-3ycr.onrender.com/ping
```

Ожидаем:

- `200 OK`
- тело: `pong`

---

## 1. Получить пустой список ссылок

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links"
```

Ожидаем:

- `200 OK`
- тело: `[]` или JSON-массив со ссылками

---

## 2. Создать ссылку с явным `short_name`

```bash
curl -i -X POST "https://go-project-278-3ycr.onrender.com/api/links" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.irregularverbs.ru/table/","short_name":"ivtable"}'
```

Ожидаем:

- `201 Created`
- JSON-объект с полями:
  - `id`
  - `original_url`
  - `short_name`
  - `short_url`

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
- `short_url` собран из `BASE_URL` и `short_name`

---

## 4. Получить список ссылок ещё раз

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links"
```

Ожидаем:

- `200 OK`
- в массиве есть созданные ссылки

---

## 5. Получить ссылку по `id`

Подставить реальный `id`, полученный после создания, например `1`:

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links/1"
```

Ожидаем:

- `200 OK`
- JSON-объект ссылки

---

## 6. Проверить `404` для несуществующего `id`

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links/666"
```

Ожидаем:

- `404 Not Found`
- тело:

```json
{"error":"link not found"}
```

---

## 7. Проверить `400` для невалидного `id`

```bash
curl -i "https://go-project-278-3ycr.onrender.com/api/links/bad"
```

Ожидаем:

- `400 Bad Request`
- тело:

```json
{"error":"invalid id"}
```

---

## 8. Обновить ссылку

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

## 9. Проверить `400` для PUT без `original_url`

```bash
curl -i -X PUT "https://go-project-278-3ycr.onrender.com/api/links/1" \
  -H "Content-Type: application/json" \
  -d '{"short_name":"itest"}'
```

Ожидаем:

- `400 Bad Request`
- тело:

```json
{"error":"original_url is required"}
```

---

## 10. Проверить `404` для PUT по несуществующему `id`

```bash
curl -i -X PUT "https://go-project-278-3ycr.onrender.com/api/links/666" \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.irregularverbs.ru/test/","short_name":"itest"}'
```

Ожидаем:

- `404 Not Found`
- тело:

```json
{"error":"link not found"}
```

---

## 11. Проверить конфликт `short_name`

Потом попытаться создать ссылку с `short_name`, который уже был использован ранее:

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

## 12. Удалить ссылку

Подставить существующий `id`, например `1`:

```bash
curl -i -X DELETE "https://go-project-278-3ycr.onrender.com/api/links/1"
```

Ожидаем:

- `204 No Content`
- пустое тело

---

## 13. Проверить, что ссылка действительно удалена

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

## 14. Проверить `404` для DELETE по несуществующему `id`

```bash
curl -i -X DELETE "https://go-project-278-3ycr.onrender.com/api/links/666"
```

Ожидаем:

- `404 Not Found`
- тело:

```json
{"error":"link not found"}
```

---

## 15. Проверить `400` для DELETE с невалидным `id`

```bash
curl -i -X DELETE "https://go-project-278-3ycr.onrender.com/api/links/bad"
```

Ожидаем:

- `400 Bad Request`
- тело:

```json
{"error":"invalid id"}
```
