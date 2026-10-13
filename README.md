# TripGo — репозиторий для лабораторных работ

## Структура
```
cmd/trip-service/        composition root, единственный main
internal/
  config/                конфигурация из переменных окружения
  database/              pgxpool, маршрутизация tx/пул, TransactionManager
  handler/               HTTP-хендлеры, валидация, problem+json
  queries/               SQL через squirrel ($N), типы строк, доменные ошибки
  server/                chi-роутер, http.Server с таймаутами
migrations/              миграции goose
api/generated/           код, сгенерированный из OpenAPI (не редактируется)
contracts/               контракты курса (не редактируются)
```


## Лабораторная работа 1:
три ручки (`createTrip`, `getTrip`, `finishTrip`),
`/health` и `/ready`, миграции, pgxpool, squirrel, собственный менеджер
транзакций.

### Быстрый старт

```bash
tripgoctl cluster start
tripgoctl environment start   
tripgoctl connect             
make migrate
make run
```

Проверка:

```bash
curl -sS localhost:8080/health
curl -sS localhost:8080/ready

curl -sS -X POST localhost:8080/api/v1/trips \
  -H 'content-type: application/json' \
  -d '{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c",
       "driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5",
       "start_point":{"latitude":59.9398,"longitude":30.3146},
       "end_point":{"latitude":59.9290,"longitude":30.3626},
       "price":1450}'
```

## Переменные окружения

| Переменная | Описание |
|---|---|
| `HTTP_ADDR` | адрес HTTP-сервера |
| `LOG_LEVEL` | уровень логирования (`debug`, `info`, `warn`, `error`) |
| `SHUTDOWN_TIMEOUT` | graceful shutdown |
| `HTTP_READ_TIMEOUT` | таймаут чтения запроса |
| `HTTP_READ_HEADER_TIMEOUT` | таймаут чтения заголовков |
| `HTTP_WRITE_TIMEOUT` | таймаут записи ответа |
| `HTTP_IDLE_TIMEOUT` | таймаут keep-alive соединений |
| `DATABASE_URL` | строка подключения к PostgreSQL |
| `DATABASE_MAX_CONNS` | максимум соединений в пуле |
| `DATABASE_MIN_CONNS` | минимум соединений в пуле |
| `DATABASE_MAX_CONN_LIFETIME` | время жизни соединения |
| `DATABASE_CONNECT_TIMEOUT` | таймаут подключения/ping |
| `DATABASE_QUERY_TIMEOUT` | таймаут каждого запроса к БД |

## Makefile

| Цель | Что делает |
|---|---|
| `make migrate` | накатывает миграции goose (`DATABASE_URL` берётся из `.env`) |
| `make run` | запускает сервис локально |
| `make generate` | генерирует код по OpenAPI |

## Docker

Многостадийная сборка - `deploy/Dockerfile`:
- в builder-стадии собирается статический бинарь (`CGO_ENABLED=0`)
- в итоговый образ на базе `gcr.io/distroless/static-debian12:nonroot` кладётся только бинарь

Build:
```bash
docker build -t trip-service:lab1 -f deploy/Dockerfile .
```

Запуск рядом с PostgreSQL:
```bash
docker run --rm --network host --env-file .env trip-service:lab1
curl -sS localhost:8080/ready
```

## Решения

### Уровень изоляции

Транзакции стартуют с `ReadCommitted` (пока меняется только при изменении кода - потом через конфиг может добавлю).

Почему: корректность обеспечивается через ограничения БД (индексы)
