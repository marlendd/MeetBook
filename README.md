[![Review Assignment Due Date](https://classroom.github.com/assets/deadline-readme-button-22041afd0340ce965d47ae6ef1cefeee28c7c493a6346c4f15d667ab976d596c.svg)](https://classroom.github.com/a/uvnTmvcw)

# Room Booking Service

Сервис бронирования переговорок для внутреннего использования. Администраторы создают переговорки и расписания, пользователи бронируют слоты.

## Запуск

```bash
docker-compose up --build
```

Сервис поднимается на `http://localhost:8080`. Переменные окружения имеют дефолтные значения, дополнительная настройка не нужна.

## Стек

- Go 1.26, стандартный `net/http`
- PostgreSQL 16
- pgx/v5
- golang-migrate
- golang-jwt/jwt
- testcontainers-go (для E2E)

## Про генерацию слотов

Задание предлагало два варианта: генерировать слоты заранее (скользящее окно) или при запросе. Я выбрал второй вариант.

При вызове `GET /rooms/{roomId}/slots/list?date=...` сервис вычисляет слоты по расписанию и сохраняет их через upsert (`ON CONFLICT DO NOTHING` по `(room_id, start)`). При повторном запросе на ту же дату слоты уже есть в БД — возвращаются те же UUID, что нужно для бронирования по `slotId`.

Скользящее окно не стал делать: оно требует фонового воркера, а при таких объёмах (до 50 переговорок) это лишняя сложность без реальной выгоды. Самый нагруженный эндпоинт покрыт индексом `idx_slots_room_start ON slots (room_id, start)`.

## Про уникальность брони

Ограничение "один слот — одна активная бронь" реализовано через partial unique index:

```sql
CREATE UNIQUE INDEX idx_bookings_slot_active ON bookings (slot_id) WHERE status = 'active';
```

## Про dummyLogin

Эндпоинт возвращает фиксированные UUID для каждой роли:
- admin → `00000000-0000-0000-0000-000000000001`
- user → `00000000-0000-0000-0000-000000000002`

Эти пользователи создаются миграцией `002_seed_users`, поэтому FK в таблице `bookings` не ломается при тестировании.
```

## Тесты

```bash
# юнит
go test ./tests/unit/... -coverprofile=coverage.out -coverpkg=./internal/...
go tool cover -func=coverage.out | tail -1
# coverage: 70.5%

# e2e
go test ./tests/e2e/...
```

E2E покрывает два сценария: создание переговорки → расписания → брони; отмену брони с проверкой идемпотентности.

## Makefile

```bash
make up        # поднять сервис со всеми зависимостями
make down      # остановить
make seed      # наполнить БД тестовыми переговорками и расписаниями
make test      # юнит-тесты с покрытием
make test-e2e  # E2E тесты
make swagger   # регенерировать Swagger-документацию (swaggo/swag)
```

## Swagger UI

Документация доступна по адресу `http://localhost:8080/swagger/index.html` после запуска сервиса. Генерируется из аннотаций в коде командой `make swagger`.

## Регистрация и авторизация по email/паролю

Реализованы дополнительные эндпоинты:

- `POST /register` — создать пользователя (email, password, role)
- `POST /login` — получить JWT по email и паролю

Пароль хранится в виде bcrypt-хеша. Эндпоинт `/dummyLogin` по-прежнему доступен и работает независимо.

## Линтер

Конфигурация в `.golangci.yaml`. Включены линтеры: `errcheck`, `govet`, `staticcheck`, `unused`, `misspell`, `noctx`, `bodyclose`, `exhaustive`, `godot`, `nilerr`, `prealloc`, `unconvert`, `unparam`, `whitespace`, `errorlint`, `copyloopvar`. Форматтеры: `gofmt`, `goimports`.

## Нагрузочное тестирование

Инструмент: [k6](https://k6.i

При создании брони можно передать `createConferenceLink: true`. Сервис обратится к `ConferenceClient` (в текущей реализации — мок `internal/conference.MockClient`) и сохранит ссылку в поле `conferenceLink` брони.

Принятые решения по сбоям:

- **Внешний сервис недоступен / вернул ошибку** — бронь уже создана в БД, откатывать её нецелесообразно (пользователь получил слот). Ссылка просто не сохраняется, `conferenceLink` в ответе будет `null`. Ошибка логируется как `WARN`.
- **Ошибка при сохранении ссылки в БД после успешного ответа внешнего сервиса** — аналогично: бронь остаётся активной, ссылка теряется, логируется `WARN`. 
- **`createConferenceLink: false` или поле не передано** — запрос к внешнему сервису не делается.

## Нагрузочное тестирование

Инструмент: [k6](https://k6.io/). Скрипт: `tests/load/load_test.js`.

Сценарий: плавный разгон до 100 VU за 30s → 1 минута под нагрузкой → остывание.
Тестируемый эндпоинт: `GET /rooms/{roomId}/slots/list` (самый нагруженный по условию задания).

Результаты тестирования:

```
THRESHOLDS 

    http_req_duration{endpoint:slots}
    ✓ 'p(95)<200' p(95)=19.23ms

    http_req_failed
    ✓ 'rate<0.001' rate=0.00%


  █ TOTAL RESULTS 

    checks_total.......: 61322   510.675307/s
    checks_succeeded...: 100.00% 61322 out of 61322
    checks_failed......: 0.00%   0 out of 61322

    ✓ slots 200

    HTTP
    http_req_duration..............: avg=8.13ms   min=125.08µs med=6.74ms   max=96.7ms   p(90)=16.84ms  p(95)=19.23ms 
      { endpoint:slots }...........: avg=8.13ms   min=426.2µs  med=6.74ms   max=96.7ms   p(90)=16.84ms  p(95)=19.23ms 
      { expected_response:true }...: avg=8.13ms   min=125.08µs med=6.74ms   max=96.7ms   p(90)=16.84ms  p(95)=19.23ms 
    http_req_failed................: 0.00%  0 out of 61326
    http_reqs......................: 61326  510.708618/s

    EXECUTION
    iteration_duration.............: avg=109.89ms min=100.5ms  med=108.42ms max=214.02ms p(90)=119.01ms p(95)=122.18ms
    iterations.....................: 61322  510.675307/s
    vus............................: 1      min=1          max=99 
    vus_max........................: 100    min=100        max=100

    NETWORK
    data_received..................: 179 MB 1.5 MB/s
    data_sent......................: 24 MB  196 kB/s

running (2m00.1s), 000/100 VUs, 61322 complete and 0 interrupted iterations
default ✓ [ 100% ] 000/100 VUs  2m0s
```

Тест прогнали на 100 виртуальных пользователях в течение 2 минут. За это время выполнили 61 322 итерации - примерно 510 запросов в секунду, что в 5 раз выше целевого RPS из задания. Ошибок ноль. Самый нагруженный эндпоинт GET /slots/list отвечал в среднем за 8ms, p95 — 19ms, что в 10 раз лучше требуемых 200ms. Сервис уверенно держит нагрузку с большим запасом.

Запуск - `make load-test` (сервис должен быть поднят через `make up`).
