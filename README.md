# user-wallet-service

REST API микросервис для управления пользователями, ролями и финансовыми счетами на Go.

## Технологии

- **Go 1.26** + `net/http` (stdlib роутинг, Go 1.22+ method-based паттерны)
- **PostgreSQL 18** — СУБД
- **pgx/v5** + **sqlx** — драйвер и struct scanning
- **golang-migrate** — миграции
- **JWT** (`golang-jwt/jwt/v5`) — аутентификация (HS256)
- **bcrypt** (`golang.org/x/crypto`) — хеширование паролей
- **slog** — структурное JSON-логирование
- **Docker** + **Docker Compose** — контейнеризация

## Быстрый старт

### 1. Клонирование и настройка

```bash
git clone <repo-url>
cd user-wallet-service
cp deployments/.env.example deployments/.env
```

Отредактируйте `deployments/.env` — задайте пароль БД и JWT-секрет (минимум 32 символа):

```env
DB_PASSWORD=your-secure-password
JWT_SECRET=your-very-long-secret-key-min-32-chars
```

### 2. Запуск через Docker Compose

```bash
cd deployments
docker compose up --build
```

Сервер стартует на `http://localhost:8080`. PostgreSQL — на порту `5432`.

### 3. Запуск локально (без Docker)

Требуется PostgreSQL. Создайте базу данных:

```bash
createdb user_wallet_servicedb
```

Задайте переменные окружения (или используйте `.env`):

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your-password
export DB_NAME=user_wallet_servicedb
export JWT_SECRET=your-very-long-secret-key-min-32-chars
```

Миграции применяются автоматически при старте сервера.

Запуск:

```bash
go run ./cmd/app
```

### 4. Тесты

```bash
# Unit-тесты (без БД)
go test -short ./...

# Интеграционные тесты (требуют PostgreSQL)
go test -tags integration -v ./...
```

## Структура проекта

```
cmd/app/main.go                  Точка входа
internal/
  controller/                    HTTP-обработчики
    auth_controller.go           Регистрация, логин, refresh, logout
    user_controller.go           GET /me, PATCH /me/profile
    account_controller.go        Баланс, пополнение, списание, перевод
    transaction_controller.go    История транзакций
    admin_controller.go          Управление пользователями и ролями
    health.go                    Health check
    respond.go                   RFC 9457 ответы
  service/                       Бизнес-логика
    auth_service.go              Аутентификация
    user_service.go              Пользователи и профили
    account_service.go           Финансовые операции
    transaction_service.go       Фильтрация транзакций
    admin_service.go             Администрирование
  repository/                    SQL-запросы
    user_repository.go
    account_repository.go        (включая Transfer с SQL-транзакцией)
    transaction_repository.go
    session_repository.go
    role_repository.go
    profile_repository.go
  model/                         Сущности, DTO, ошибки
  infrastructure/                Низкоуровневые реализации
    config/                      Env-based конфигурация
    database/                    PostgreSQL (pgx + sqlx)
    httpserver/                  HTTP-сервер
    jwt/                         JWT менеджер
    logger/                      slog JSON
    migrate/                     golang-migrate
    password/                    bcrypt
  router/                        Роуты и middleware
migrations/                      SQL-миграции
deployments/                     Docker, docker-compose, .env
```

## API

Базовый URL: `http://localhost:8080`

### Middleware chain

```
Request → RequestID → RateLimit → Logger → Recovery → [Auth → RoleCheck] → Handler
```

### Авторизация

Защищённые эндпоинты требуют заголовок:

```
Authorization: Bearer <access_token>
```

### Эндпоинты

#### Auth (публичные)

| Метод | Путь | Описание | Тело |
|-------|------|----------|------|
| `POST` | `/auth/register` | Регистрация | `{"email":"...", "password":"..."}` |
| `POST` | `/auth/login` | Логин | `{"email":"...", "password":"..."}` |
| `POST` | `/auth/refresh` | Ротация токенов | `{"refresh_token":"..."}` |
| `POST` | `/auth/logout` | Выход | `{"refresh_token":"..."}` |

#### User (auth)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/me` | Текущий пользователь + профиль |
| `PATCH` | `/me/profile` | Обновление профиля |

Тело `PATCH /me/profile`:
```json
{
  "first_name": "Иван",
  "last_name": "Петров",
  "birth_date": "1990-05-15T00:00:00Z"
}
```

#### Account (auth)

| Метод | Путь | Описание | Тело |
|-------|------|----------|------|
| `GET` | `/account` | Баланс | — |
| `POST` | `/account/deposit` | Пополнение | `{"amount":"100.50"}` |
| `POST` | `/account/withdraw` | Списание | `{"amount":"50.00"}` |
| `POST` | `/account/transfer` | Перевод | `{"to_user_id":"uuid", "amount":"30.00"}` |

#### Transactions (auth)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/transactions` | История операций |

Query-параметры:
- `page` — номер страницы (по умолчанию 1)
- `limit` — элементов на странице (по умолчанию 20)
- `sort` — `asc`, `desc` (по умолчанию), `amount_asc`, `amount_desc`
- `type` — `deposit`, `withdraw`, `transfer_in`, `transfer_out`
- `status` — `active`, `inactive`
- `date_from`, `date_to` — ISO 8601
- `amount_min`, `amount_max` — строковые значения

#### Admin (auth + role=admin)

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/admin/users` | Список пользователей |
| `PATCH` | `/admin/users/{id}/block` | Блокировка |
| `PATCH` | `/admin/users/{id}/unblock` | Разблокировка |
| `PATCH` | `/admin/users/{id}/role` | Смена роли (`{"role_id": 2}`) |
| `GET` | `/admin/roles` | Список ролей |

#### Health

| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/health` | Проверка здоровья (включая БД) |

### Примеры запросов

**Регистрация:**

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepass123"}'
```

**Логин:**

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepass123"}'
```

Ответ:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiJ9...",
  "refresh_token": "a1b2c3d4e5f6..."
}
```

**Получение баланса:**

```bash
curl http://localhost:8080/account \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9..."
```

**Пополнение:**

```bash
curl -X POST http://localhost:8080/account/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9..." \
  -d '{"amount":"500.00"}'
```

**Перевод:**

```bash
curl -X POST http://localhost:8080/account/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9..." \
  -d '{"to_user_id":"550e8400-e29b-41d4-a716-446655440000","amount":"100.00"}'
```

## Формат ошибок

Все ошибки в формате RFC 9457 (`application/problem+json`):

```json
{
  "type": "https://example.com/problems/invalid-input",
  "title": "Validation failed",
  "status": 400,
  "detail": "validation failed",
  "instance": "/auth/register",
  "errors": {
    "email": "required",
    "password": "at least 8 characters"
  }
}
```

| Kind | HTTP Status | type URI |
|------|-------------|----------|
| `invalid` | 400 | `.../invalid-input` |
| `unauthorized` | 401 | `.../unauthorized` |
| `forbidden` | 403 | `.../forbidden` |
| `not_found` | 404 | `.../not-found` |
| `conflict` | 409 | `.../conflict` |
| `insufficient` | 422 | `.../insufficient-funds` |
| `rate_limit` | 429 | `.../rate-limit-exceeded` |
| `internal` | 500 | `.../internal-error` |

## Конфигурация

Все настройки через переменные окружения:

| Переменная | По умолчанию | Описание |
|-----------|--------------|----------|
| `HTTP_PORT` | `8080` | Порт HTTP-сервера |
| `DB_HOST` | `localhost` | Хост PostgreSQL |
| `DB_PORT` | `5432` | Порт PostgreSQL |
| `DB_USER` | `postgres` | Пользователь БД |
| `DB_PASSWORD` | `postgres` | Пароль БД |
| `DB_NAME` | `user_wallet_servicedb` | Имя базы данных |
| `DB_MAX_OPEN_CONNS` | `25` | Макс. открытых соединений |
| `DB_MAX_IDLE_CONNS` | `5` | Макс. idle-соединений |
| `DB_CONN_MAX_LIFETIME` | `5m` | Время жизни соединения |
| `JWT_SECRET` | **обязательно** | Секрет HS256 (мин. 32 символа) |
| `JWT_ACCESS_TTL` | `15m` | Время жизни access-токена |
| `JWT_REFRESH_TTL` | `168h` (7 дней) | Время жизни refresh-токена |
| `RATE_LIMIT_RPS` | `10` | Запросов в секунду на IP |
| `RATE_LIMIT_BURST` | `20` | Макс. burst запросов подряд |
| `RATE_LIMIT_TTL` | `10m` | Время хранения записей в кэше лимитера |

## Схема базы данных

```
roles ──< users >── profiles
                  >── accounts ──< transactions
                  >── sessions
```

- **roles** — `user` (1), `admin` (2)
- **users** — UUID PK, email (UNIQUE, CHECK), password_hash, role_id (FK), status
- **profiles** — user_id (FK, PK, CASCADE), first_name, last_name, age
- **accounts** — UUID PK, user_id (FK, UNIQUE), balance (NUMERIC(21,2), CHECK >= 0)
- **transactions** — UUID PK, account_id (FK), type_id (FK), amount, status
- **sessions** — UUID PK, user_id (FK, CASCADE), refresh_token_hash (SHA-256), expires_at

Миграции применяются автоматически при старте сервера.

## Postman-коллекция

Файл `postman_collection.json` в корне проекта. Импортируйте в Postman:

1. File → Import → Выберите файл
2. Создайте переменную окружения `base_url` = `http://localhost:8080`
3. После логина скопируйте `access_token` и `refresh_token` в переменные окружения
