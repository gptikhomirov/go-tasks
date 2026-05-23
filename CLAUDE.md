# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## О проекте

Учебный проект: REST API на Go, который ведётся от нуля к продакшен-практикам.
Цель — не просто рабочий код, а демонстрация того, как код пишут в продакшене
(структура, обработка ошибок, слои, миграции и т.д.). Со временем проект будет
усложняться. При изменениях предпочитай объясняющий, идиоматичный Go-код и
указывай на продакшен-практики, а не только на «лишь бы работало».

Домен на текущем этапе — CRUD задач (`tasks`).

Имя Go-модуля — `go-rest`; внутренние импорты идут через префикс `go-rest/...`.

## Команды

```bash
# Поднять PostgreSQL (создаёт БД и применяет sql/init.sql при первом старте)
docker compose up -d

# Запустить API (по умолчанию порт 8080, БД на localhost:5432)
go run ./cmd/api

# Сборка
go build ./cmd/api

# Тесты (тестов пока нет; команда — на будущее)
go test ./...
go test ./internal/handlers -run TestXxx   # один тест

# Форматирование и статический анализ
go fmt ./...
go vet ./...
```

Переменные окружения, которые читает `cmd/api/main.go`:
- `SERVER_PORT` — порт сервера (по умолчанию `8080`).
- `DATABASE_URL` — строка подключения Postgres (по умолчанию указывает на
  локальный docker-compose: `tasks_user:password@localhost:5432/tasks_db`).

## Архитектура

Слоистая структура с однонаправленной зависимостью
`main → handlers → database → models`:

- **cmd/api/main.go** — точка входа. Читает env, подключается к БД, собирает
  зависимости (store → handler), регистрирует маршруты в `http.ServeMux` и
  оборачивает его в `loggingMiddleware`. Маршрутизация использует
  method-aware паттерны Go 1.22+ (`"GET /tasks/{id}"`), без сторонних роутеров.

- **internal/handlers** — HTTP-слой. `Handler` держит `*database.TaskStore`.
  Парсит запрос, валидирует вход, вызывает store, формирует JSON-ответ.
  Общие хелперы — `respondWithJSON` / `respondWithError`; `{id}` достаётся из
  пути через `r.PathValue("id")`. Маппинг ошибок в HTTP-статусы делается здесь
  через `errors.Is(err, database.ErrTaskNotFound)` → 404, иначе 500.

- **internal/database** — слой доступа к данным (`TaskStore`) на `sqlx` +
  драйвере `lib/pq`. Сырой SQL, без ORM. Сигнальная ошибка `ErrTaskNotFound`
  объявлена здесь и оборачивается через `fmt.Errorf("...: %w", ErrTaskNotFound)` —
  именно её распознаёт слой handlers. `Update` сначала читает запись
  (read-modify-write), применяя поля из `UpdateTaskInput`.

- **internal/models** — структуры данных с тегами `json` и `db`. Ключевой
  паттерн разделения DTO:
  - `Task` — полная сущность (ответ на чтение/запись одной задачи).
  - `TaskListItem` — урезанная проекция для списка (`GetAll`, без `id`/`updated_at`).
  - `CreateTaskInput` — поля для создания (значимые типы).
  - `UpdateTaskInput` — поля-указатели (`*string`, `*bool`) для частичного
    обновления: `nil` означает «поле не передано, не менять».

## База данных

Схема и сид-данные — в `sql/init.sql`; применяются автоматически при первом
запуске docker-compose (примонтированы в `/docker-entrypoint-initdb.d/`).
Миграционного инструмента пока нет — изменение схемы означает правку
`init.sql` и пересоздание volume (`docker compose down -v`).

## API и тестирование запросов

Эндпоинты: `GET/POST /tasks`, `GET/PUT/DELETE /tasks/{id}`.
Коллекция Postman для ручной проверки лежит в `postman/collections/Tasks API/`.

## Прочее

- `reviews/` — заметки код-ревью по датам (например, `reviews/2026-05-19-review.md`),
  фиксируют решения и замечания по ходу обучения.
- `cmd/api/main.go` содержит TODO-маркеры на ближайшие шаги (вынос env-конфига,
  CORS-middleware).
