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