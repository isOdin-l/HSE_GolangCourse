# Лабораторная работа 2

## Слои и зависимости

Зависимости идут строго внутрь: транспорт → usecase → domain, домен не импортирует ничего из репозитория и инфраструктуры.

- **`cmd/trip-service`** (main) — читает конфиг из env, собирает контейнер, запуск, ожидание сигнала (SIGTERM/SIGINT), graceful shutdown.
- **`internal/container`** — единственное место, где «сходятся» все зависимости. Наружу отдаёт только собранный App и Shutdown.
- **`internal/server`** — сборка HTTP-сервера: chi, цепочка middleware, ChiServerOptions.
- **`internal/http`** — транспорт: хендлеры, маппинг ошибок в Problem (RFC 9457).
- **`internal/usecase`** — бизнес-логика; зависит только от собственных узких интерфейсов потребителя, а не от всего репозитория. Интерфейсы объявлены на стороне usecase, а не в репозитории.
- **`internal/repository`** — слой репозитория. Зависит только от собственных неэкспортируемых интерфейсов.
- **`internal/queries`** — SQL-слой: построение запросов через squirrel, маппинг pgx-ошибок в доменные ошибки.
- **`internal/domain/trip`** — доменные правила, статусы и переходы, ошибки домена. Не знает про БД, HTTP и телеметрию.
- **`internal/infra`** — postgres (пул, транзакции, otelpgx), telemetry (OTel SDK, логгер, метрики).
- **`api/generated`** — код, сгенерированный по OpenAPI-контракту: types + strict-server + chi-server. Руками не правится.

Итого по потоку вызовов:

```
Бизнес-логика ставит атомарные задачи
Репозиторий выбирает доступные инструменты
Queries      строит и выполняет SQL
Infra        даёт доступ к БД, кешам, брокерам и т.д.
```

Это позволяет заменять слои независимо: добавить кеш или сменить библиотеку/БД — правится только `queries`, слой репозитория не трогается.

`internal/queries` — SQL-слой, ещё одна абстракция между слоем репозитория и инфраструктурой, на которой выполняются запросы. Repository слой зависит только от собственных неэкспортируемых интерфейсов.

## Telemetry

- Инициализация OTel SDK — только через конструкторские функции (`internal/infra/telemetry`), глобального состояния нет: TracerProvider/MeterProvider идут зависимостью до нужного места.
- HTTP — otelhttp middleware: создаёт корневой спан с именем `{method} {route}` по semconv. Лейбл http.route одинаков для метрик и спанов: chi v5 заполняет `http.Request.Pattern` только после матчинга маршрута, поэтому middleware вне роутера этого поля не видят, а внутри роутера (chi Middlewares в ChiServerOptions) — видят.
- Метрики:
  - `http_server_request_duration_seconds` — otelhttp;
  - `http_server_active_requests` — свой up-down-counter на уровне маршрута. Осознанное отступление от цепочки «Recover→Tracing→Logging→Metrics→роутер»: без матчинга маршрута лейбл http.route не наполнить, а карточность по полному пути недопустима; для 404/405 метрику не пишем вовсе.
  - `tripgo_trips_created_total`, `tripgo_trips_completed_total` — бизнес-счётчики в usecase.
  - `db_client_operation_duration_seconds` — otelpgx на пуле БД, а не в каждом методе репозитория вручную: иначе половину запросов забудут обернуть, а репозиторий начнёт зависеть от телеметрии.
- Логи: `log/slog`, JSON на stdout, уровень из LOG_LEVEL. Middleware TracingHandler добавляет trace_id/span_id ко всем записям, сделанным внутри спана. В usecase — узкие интерфейсы (`tripMetrics`, `eventLogger`), чтобы бизнес-слой не тянул весь пакет otel/metric.
- Экспорт по OTLP (http/protobuf) в коллектор: за ним Jaeger (UI :22686) и Prometheus (:22909). При недоступности коллектора сервис работает штатно — экспортер скидывает данные и повторяет отправку.
- `TELEMETRY_FLUSH_TIMEOUT` — бюджет на финальный flush трейсов/метрик при шатдауне, если основной SHUTDOWN_TIMEOUT уже израсходован: тогда flush идёт отдельным контекстом.

Порядок шатдауна — обратный порядку сборки: `httpServer.Shutdown(ctx)` → закрытие пула БД → `telemetry.Shutdown` (flush), всё под SHUTDOWN_TIMEOUT.

## Отступления от целевой структуры

| Отступление | Почему |
|---|---|
| `internal/queries` вне целевой схемы | Отдельный SQL-слой между repository и infra: в задании больше слоёв, чем в итоговой схеме; иначе репозиторий тянет pgx/squirrel |
| `internal/server` вне целевой схемы | Сборка HTTP-сервера и chi-роутера в одном месте; в `internal/http` остаются только хендлеры |
| `api/generated` вместо `internal/generated` | oapi-codegen пишет в `api/` — конвенция курса разрешает такую структуру |
| Нет `internal/model` | Доменные структуры живут в `internal/domain/trip`; DTO HTTP-запросов генерируются по контракту, дублировать не пришлось |
| active_requests middleware на уровне маршрута | см. раздел Telemetry: без http.route лейбл не наполнить |

## nolint

- В хендлерах линтер ругался, что error всегда nil, но убрать его я не могу, потому что это часть сигнатуры StrictServerInterface.
- Также были проблемы с int32, но там идёт конвертация в структуру, сгенерированную по контракту (api.Problem.Status).

## Что проверить руками (ЛР2)

1. `make run` — сервис поднимается, `/ready` отдаёт 200.
2. `POST /api/v1/trips` по trip.json из README2 — 201, логи `trip created` и `request handled` с одинаковым trace_id.
3. Jaeger — трейс с корневым спаном `POST /api/v1/trips` и детьми `pool.acquire`, `BEGIN`, `INSERT trip`, `trip status history`.
4. Grafana/Prometheus — `http_server_request_duration_seconds` (лейблы http.route/http.request.method), `http_server_active_requests`, `tripgo_trips_created_total`.
5. Останов по SIGTERM — лишних горутин нет, последние трейсы доехали до Jaeger.
6. Останов коллектора: сервис не падает и не теряет запросы при недоступном коллекторе.
