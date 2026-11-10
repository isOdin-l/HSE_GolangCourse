# TripGo — репозиторий для лабораторных работ

Учебный сервис поездок для курса «Разработка микросервисов на Go». Работы
с номером NN сдаются веткой `homework/NN` → PR в `main` своего форка
(подробности в [`grading.md`](https://github.com/course-go-autumn-2026/course/blob/main/homework/docs/grading.md)).

Что сделано к текущему моменту:

- **ЛР1** — три ручки (`createTrip`, `getTrip`, `finishTrip`), `/health` и
  `/ready`, миграции, pgxpool, squirrel, собственный менеджер транзакций.
- **ЛР2** — телеметрия: трейсы и метрики через OpenTelemetry, корреляция логов
  с трейсами, экспорт в коллектор.

## Быстрый старт

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

## Архитектура: слои и зависимости

Зависимости идут строго внутрь: транспорт → usecase → domain, домен не
импортирует ничего из репозитория и инфраструктуры.

- **`cmd/trip-service`** (main) — composition root: читает конфиг из env,
  собирает контейнер, запуск, ожидание сигнала (SIGTERM/SIGINT), graceful
  shutdown.
- **`internal/container`** — единственное место, где «сходятся» все
  зависимости. Наружу отдаёт только собранный App и Shutdown.
- **`internal/server`** — сборка HTTP-сервера: chi, цепочка middleware,
  ChiServerOptions.
- **`internal/http`** — транспорт: хендлеры, маппинг ошибок в Problem
  (RFC 9457).
- **`internal/usecase`** — бизнес-логика; зависит только от собственных узких
  интерфейсов потребителя, а не от всего репозитория. Интерфейсы объявлены на
  стороне usecase, а не в репозитории.
- **`internal/repository`** — слой репозитория. Зависит только от собственных
  неэкспортируемых интерфейсов.
- **`internal/queries`** — SQL-слой: построение запросов через squirrel,
  маппинг pgx-ошибок в доменные ошибки.
- **`internal/domain/trip`** — доменные правила, статусы и переходы, ошибки
  домена. Не знает про БД, HTTP и телеметрию.
- **`internal/infra`** — postgres (пул, транзакции, otelpgx), telemetry
  (OTel SDK, логгер, метрики).
- **`api/generated`** — код, сгенерированный из OpenAPI-контракта. Не
  редактируется.

Поток вызовов:

```
Бизнес-логика ставит атомарные задачи
Репозиторий выбирает доступные инструменты
Queries      строит и выполняет SQL
Infra        даёт доступ к БД, кешам, брокерам и т.д.
```

Это позволяет заменять слои независимо: добавить кеш или сменить библиотеку/БД
— правится только `queries`, слой репозитория не трогается.

Порядок шатдауна — обратный порядку сборки: `httpServer.Shutdown(ctx)` →
закрытие пула БД → `telemetry.Shutdown` (flush), всё под `SHUTDOWN_TIMEOUT`.

## Переменные окружения

Полный список с значениями по умолчанию — в [`.env.example`](.env.example);
`.env` генерирует `tripgoctl environment start`.

| Переменная | Описание |
|---|---|
| `HTTP_ADDR` | адрес HTTP-сервера |
| `LOG_LEVEL` | уровень логирования (`debug`, `info`, `warn`, `error`) |
| `SHUTDOWN_TIMEOUT` | бюджет graceful shutdown |
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
| `OTEL_SERVICE_NAME` | имя сервиса в телеметрии |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | адрес OTLP-коллектора |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | протокол экспорта (`http/protobuf`) |
| `OTEL_RESOURCE_ATTRIBUTES` | доп. атрибуты ресурса (`service.namespace`, `deployment.environment`) |
| `OTEL_TRACES_SAMPLER` | сэмплирование трейсов |
| `OTEL_TRACES_SAMPLER_ARG` | аргумент сэмплера |
| `TELEMETRY_FLUSH_TIMEOUT` | бюджет финального flush телеметрии при шатдауне |

## Makefile

| Цель | Что делает |
|---|---|
| `make migrate` | накатывает миграции goose (`DATABASE_URL` берётся из `.env`) |
| `make run` | запускает сервис локально |
| `make generate` | генерирует код по OpenAPI и моки |
| `make lint` | golangci-lint (правила слоёв по depguard) |
| `make test-unit` | юнит-тесты с `-race`, покрытие пакетов `internal/...` |
| `make test-integration` | интеграционные тесты на реальном PostgreSQL |
| `make test` | всё вместе + сводка покрытия |

## Docker

Многостадийная сборка — `deploy/Dockerfile`:

- в builder-стадии собирается статический бинарь (`CGO_ENABLED=0`);
- в итоговый образ на базе `gcr.io/distroless/static-debian12:nonroot` кладётся
  только бинарь.

```bash
docker build -t trip-service:lab1 -f deploy/Dockerfile .
```

Размер image 18.7MB.

Запуск рядом с PostgreSQL:

```bash
docker run --rm --network host --env-file .env trip-service:lab1
curl -sS localhost:8080/ready
```

## Решения и отступления

- Транзакции стартуют с `ReadCommitted` (меняется только кодом — если
  понадобится, вынесу в конфиг). Корректность обеспечивается ограничениями БД
  (индексы).
- В хендлерах линтер ругался, что error всегда nil, но убрать его нельзя: это
  часть сигнатуры `StrictServerInterface`; на int32 — конвертация в структуру,
  сгенерированную по контракту (`api.Problem.Status`).
