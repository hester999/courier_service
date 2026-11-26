# Courier Service

Сервис для управления курьерами, реализующий REST API для создания, обновления и получения данных.  
Проект разработан на Go и использует PostgreSQL в качестве хранилища данных.  
Код организован по принципам **Clean Architecture** и сопровождается тестовой Postman-коллекцией.

---

## Описание проекта

Сервис для управления курьерами, заказами и доставками. Реализует автоматическое назначение курьеров на заказы, отслеживание просроченных доставок и оптимизированную обработку статусов курьеров.

### Основная функциональность:

**Управление курьерами:**
- создание курьера с проверкой корректности телефона (RU/AM) и статуса;
- обновление данных курьера (PUT, с валидацией);
- получение данных по ID;
- получение списка курьеров с пагинацией;

**Управление заказами:**
- создание заказов;
- получение списка заказов;

**Управление доставками:**
- автоматическое назначение курьера на заказ (выбор курьера с минимальным количеством заказов);
- отмена назначения курьера на заказ;
- автоматическая обработка просроченных заказов;

**Фоновые задачи:**
- автоматическая проверка просроченных заказов с накоплением в кеше;
- обновление статусов курьеров при истечении TTL;

**Системные эндпоинты:**
- `/ping` и `/healthcheck` для проверки работы сервиса.

---

## Технологии

| Компонент | Назначение |
|------------|------------|
| PostgreSQL | база данных |
| sqlx | работа с SQL |
| Squirrel | генерация SQL-запросов |
| Gorilla Mux | HTTP-роутер |
| Goose | миграции базы данных |
| Phonenumbers | валидация номеров RU/AM |
| Docker Compose | контейнеризация |
| Postman | тестирование API |

---

## Переменные окружения

Создайте файл `.env` в корне проекта:

```env
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=avito_db
POSTGRES_PORT=5432
POSTGRES_HOST=localhost
MAX_CONNECTIONS=10
PORT=8080
MAX_ATTEMPTS=5
DELAY=3
PERIOD_CHECK=10
BATCH_SIZE=3
CACHE_TTL=60
CACHE_SIZE=1
```

| Переменная | Описание |
|-------------|-----------|
| `POSTGRES_USER` | имя пользователя PostgreSQL |
| `POSTGRES_PASSWORD` | пароль пользователя |
| `POSTGRES_DB` | имя базы данных |
| `POSTGRES_PORT` | порт базы данных |
| `POSTGRES_HOST` | хост БД (`localhost` или `db` при docker-compose) |
| `MAX_CONNECTIONS` | максимальное количество соединений с БД |
| `PORT` | порт HTTP-сервера |
| `MAX_ATTEMPTS` | количество попыток подключения к БД |
| `DELAY` | задержка между попытками подключения к БД (секунды) |
| `PERIOD_CHECK` | интервал проверки просроченных заказов (секунды, по умолчанию 10) |
| `BATCH_SIZE` | размер батча для обработки курьеров в кеше (по умолчанию 3) |
| `CACHE_TTL` | время жизни записи в кеше до обработки (секунды, по умолчанию 60) |
| `CACHE_SIZE` | максимальное количество ключей в LRU-кеше (по умолчанию 1) |

---

## Структура проекта

```
cmd/
│── server.go          # Точка входа, запуск приложения
internal/
├── apperr/            # Обработка ошибок AppErr
├── cli/               # CLI-флаги
├── config/            # Загрузка .env
├── handler/           # HTTP-контроллеры
│   ├── courier/       # CRUD-эндпоинты для курьеров
│   ├── delivery/      # Эндпоинты для управления доставками
│   ├── order/         # Эндпоинты для управления заказами
│   └── service/       # Healthcheck и Ping
├── infra/             # Инфраструктурный слой
│   ├── background/    # Фоновые задачи (checker статусов)
│   ├── cache/         # LRU-кеш для батчинга
│   └── db/            # Подключение к БД и транзакции
├── model/             # Доменные сущности
├── repo/              # Доступ к БД (sqlx, Squirrel)
│   ├── courier/       # Репозиторий курьеров
│   ├── delivery/      # Репозиторий доставок
│   └── order/         # Репозиторий заказов
├── router/            # Настройка маршрутов (gorilla/mux)
├── usecases/          # Бизнес-логика
│   ├── courier/       # Use cases для курьеров
│   │   └── status_check/  # Проверка статусов и обработка evicted
│   ├── delivery/      # Use cases для доставок
│   └── order/         # Use cases для заказов
├── utils/             # Вспомогательные функции (JSON, валидация)
└── pkg/               # Общие библиотеки (логгер)
migration/             # SQL миграции (goose)
test/                  # Тесты конфигурации
build/                 # Скомпилированные бинарники
report/                # Отчеты о покрытии тестами
```

---

## Сборка и запуск

### Вариант 1. Локально (через Makefile)

```bash
make build
make run
```

**Основные команды Makefile:**

| Команда | Описание |
|---------|----------|
| `make build` | Сборка бинарного файла |
| `make run` | Запуск приложения |
| `make test` | Запуск unit-тестов |
| `make test-cover` | Запуск тестов с генерацией HTML отчета о покрытии |
| `make integration-test-courier` | Запуск интеграционных тестов для курьеров |
| `make integration-test-delivery` | Запуск интеграционных тестов для доставок |
| `make migrate-up` | Применение миграций |
| `make migrate-down` | Откат миграций |
| `make migrate-status` | Проверка статуса миграций |
| `make clean` | Удаление скомпилированных файлов и отчетов |

---

### Вариант 2. С флагами

Приложение поддерживает передачу порта через флаг `-port`, который переопределяет значение из `.env`.

Пример:
```bash
go run cmd/server.go -port=9090
```

Если флаг не задан, порт берётся из переменной окружения `PORT`.

---

### Вариант 3. Через Docker Compose

Docker Compose настраивает полное окружение для разработки и тестирования.

**Сервисы:**

| Сервис | Порт | Описание |
|--------|------|----------|
| `courier_service` | 8080 | Основное приложение Go |
| `postgres` | 5432 | База данных для разработки |
| `postgres-test` | 9999 | База данных для интеграционных тестов (автоматически запускается при тестах) |

**Запуск:**

1. Соберите и запустите контейнеры:
```bash
docker-compose up --build
```

Или в фоновом режиме:
```bash
docker-compose up -d --build
```

2. Применение миграций (в отдельном терминале, после запуска контейнеров):
```bash
make migrate-up
```

3. После запуска сервис будет доступен по адресу:
```
http://localhost:8080
```

4. Проверка доступности:
```bash
curl http://localhost:8080/ping
```

Ожидаемый ответ:
```json
{"message": "pong"}
```

**Управление контейнерами:**

Просмотр логов:
```bash
docker-compose logs -f courier_service
docker-compose logs -f postgres
```

Остановка контейнеров:
```bash
docker-compose down
```

Остановка с удалением volumes (удалит все данные БД):
```bash
docker-compose down -v
```

Перезапуск конкретного сервиса:
```bash
docker-compose restart courier_service
```

---

## API Эндпоинты

### Сервисные эндпоинты

| Метод | URL | Описание | Коды ответов |
|--------|-----|-----------|--------------|
| GET | `/ping` | Проверка доступности сервера | 200 |
| HEAD | `/healthcheck` | Проверка состояния без тела | 204 |

### Курьеры

| Метод | URL | Описание | Коды ответов |
|--------|-----|-----------|--------------|
| POST | `/api/v1/courier` | Создание курьера | 201, 400, 409 |
| PUT | `/api/v1/courier` | Обновление данных курьера | 200, 400, 404 |
| GET | `/api/v1/courier/{id}` | Получение курьера по ID | 200, 400, 404 |
| GET | `/api/v1/couriers?limit=10&offset=0` | Получение списка курьеров | 200, 204, 400 |

### Заказы

| Метод | URL | Описание | Коды ответов |
|--------|-----|-----------|--------------|
| POST | `/api/v1/order` | Создание заказа | 200, 400, 500 |
| GET | `/api/v1/orders` | Получение списка заказов | 200, 204, 400, 500 |

### Доставки

| Метод | URL | Описание | Коды ответов |
|--------|-----|-----------|--------------|
| POST | `/api/v1/delivery/assign` | Назначение курьера на заказ | 200, 400, 404, 409, 500 |
| POST | `/api/v1/delivery/unassign` | Отмена назначения курьера на заказ | 200, 400, 404, 500 |

---

## Примеры запросов

### Создание курьера
```bash
curl -X POST http://localhost:8080/api/v1/courier \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Антон",
    "phone": "+37444111222",
    "status": "available",
    "transport_type": "car"
  }'
```

Ответ 201:
```json
{
  "id": 1,
  "name": "Антон",
  "phone": "+37444111222",
  "status": "available",
  "transport_type": "car"
}
```

---

### Обновление курьера
```bash
curl -X PUT http://localhost:8080/api/v1/courier \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "name": "Антон Иванов",
    "phone": "+79031234567",
    "status": "paused"
  }'
```

---

### Создание заказа
```bash
curl -X POST http://localhost:8080/api/v1/order \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Заказ продуктов",
    "count": 5,
    "description": "Доставить продукты по адресу"
  }'
```

---

### Назначение курьера на заказ
```bash
curl -X POST http://localhost:8080/api/v1/delivery/assign \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "259293324-85a5-454b-b58b-7d7386f74785"
  }'
```

---

### Отмена назначения курьера
```bash
curl -X POST http://localhost:8080/api/v1/delivery/unassign \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "259293324-85a5-454b-b58b-7d7386f74785"
  }'
```

---

## Тестирование через Postman

Файл коллекции:  
`courier_service.postman_collection.json`

### Содержимое коллекции
- **create/** — создание курьеров (валидные и невалидные данные);
- **update/** — обновление (ошибки и успешные кейсы);
- **get/** — выборка по ID и списком;
- **service_handler/** — проверки `/ping` и `/healthcheck`.

---

## Ошибки

Ошибки возвращаются в формате JSON:
```json
{"error": "courierRepo not found"}
```

| Код | Описание |
|-----|-----------|
| 400 | Некорректные данные |
| 404 | Курьер не найден |
| 409 | Дубликат записи |
| 500 | Внутренняя ошибка сервера |

---

## SQL-структура таблицы `couriers`

```sql

CREATE TABLE IF NOT EXISTS couriers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

DROP TABLE IF EXISTS couriers;

```

---



## Тестирование

Проект включает unit-тесты и интеграционные тесты. Все тесты запускаются через Makefile.

### Unit-тесты

Запуск всех unit-тестов:
```bash
make test
```

Запуск тестов конкретного пакета:
```bash
go test ./internal/handler/courier/...
go test ./internal/usecases/delivery/...
```

### Coverage (покрытие кода)

Генерация HTML отчета о покрытии кода:
```bash
make test-cover
```

Результаты сохраняются в папке `report/`:
- `coverage.out` - сырой файл покрытия
- `coverage.html` - HTML отчет (автоматически открывается в браузере)

Пример вывода:
```
ok      avito/internal/config   coverage: 64.7% of statements
ok      avito/internal/handler/courier  coverage: 85.6% of statements
ok      avito/internal/repo/courier     coverage: 93.8% of statements
```

### Интеграционные тесты

Интеграционные тесты требуют запущенную PostgreSQL базу данных. Для этого используется Docker Compose.

**Запуск интеграционных тестов для репозитория курьеров:**
```bash
make integration-test-courier
```

**Запуск интеграционных тестов для репозитория доставок:**
```bash
make integration-test-delivery
```

Что происходит при запуске:
1. Запускается контейнер `postgres-test` на порту 9999
2. Применяются миграции к тестовой базе
3. Запускаются тесты с тегом `integration`
4. Результаты выводятся в консоль

**Ручной запуск интеграционных тестов:**
```bash
# Запустить тестовую БД
docker compose up -d postgres-test

# Подождать готовности БД
sleep 3

# Применить миграции
goose -dir ./migration postgres "host=localhost user=test password=test dbname=test port=9999 sslmode=disable" up

# Запустить тесты
go test -tags=integration ./internal/repo/courier -v
```

### Структура тестов

- `internal/handler/*/..._test.go` - unit-тесты для HTTP handlers (с моками)
- `internal/usecases/*/..._test.go` - unit-тесты для бизнес-логики
- `internal/repo/*/..._integration_test.go` - интеграционные тесты для репозиториев
- `test/` - тесты для конфигурации и утилит

---

## Архитектура и особенности реализации

### Управление транзакциями

Вся бизнес-логика обернута в менеджер транзакций от Avito (`go-transaction-manager`), что обеспечивает ACID-совместимость операций с базой данных.

### Назначение курьера на заказ (AssignCourier)

**Процесс:**
1. Проверка существования заказа в БД
2. Получение курьеров с минимальным количеством заказов и статусом `available`
3. Случайный выбор курьера из доступных
4. Вычисление дедлайна заказа на основе типа транспорта курьера
5. Обновление таблиц `delivery` и `couriers` в одной транзакции

**Особенности:**
- Курьер выбирается среди тех, у кого меньше всего выполненных заказов (равномерное распределение нагрузки)
- Дедлайн рассчитывается автоматически на основе транспорта (car, scooter, on_foot)

### Отмена назначения курьера (UnassignCourier)

**Процесс:**
1. Проверка существования заказа в БД
2. Обновление статуса курьера на `available`
3. Удаление записи из таблицы `delivery`
4. Увеличение счетчика выполненных заказов у курьера

**Бизнес-логика:**
- Нажатие на кнопку "завершить заказ" означает успешное выполнение
- Связь курьера с заказом удаляется из таблицы `delivery`

### Автоматическая проверка просроченных заказов (Checker)

**Механизм работы:**

1. **Периодическая проверка** (каждые `PERIOD_CHECK` секунд):
   - Поиск курьеров со статусом `busy`, у которых заказ просрочен (`deadline < NOW()`)

2. **Батчинг через LRU-кеш**:
   - Если найдено меньше записей, чем `BATCH_SIZE` - данные кешируются
   - При каждом обращении TTL ключа обновляется
   - При достижении критической массы (`BATCH_SIZE`) - все записи обрабатываются батчем

3. **Обработка evicted записей**:
   - При истечении TTL в кеше вызывается `ProcessEvicted`
   - Автоматически обновляются статусы курьеров на `available`
   - Удаляются соответствующие записи из `delivery`

**Преимущества:**
- Снижение нагрузки на БД (батчевая обработка)
- Потокобезопасность (LRU-кеш с TTL)
- Гарантированная обработка даже при истечении TTL

**Проектные решения:**

1. **Unique constraint на `order_id` в `delivery`**:
   - Один заказ может быть назначен только одному курьеру
   - Позволяет одному курьеру иметь несколько заказов (многоадресная доставка)

2. **Определение успешного заказа**:
   - Для клиента - заказ успешен, если вызван `unassign` (нажата кнопка)
   - Для системы - заказ просрочен, если дедлайн истек, а курьер все еще `busy`
   - При просрочке заказ автоматически снимается с курьера через checker

---

**Результаты тестов**

```
hester@DESKTOP-AN5LJ4T:/mnt/d/projects/GO/course-go-avito-hester999$ make test
go test -count 1 -race -parallel 20 -p 20 ./...
?       avito/cli       [no test files]
?       avito/cmd/server        [no test files]
?       avito/internal/apperr   [no test files]
ok      avito/internal/config   1.024s
ok      avito/internal/handler/courier  1.037s
ok      avito/internal/handler/courier/dto      1.033s
?       avito/internal/handler/courier/mocks    [no test files]
ok      avito/internal/handler/delivery 1.060s
ok      avito/internal/handler/delivery/dto     1.065s
?       avito/internal/handler/delivery/mocks   [no test files]
?       avito/internal/handler/service  [no test files]
?       avito/internal/infra/background/status_checker  [no test files]
?       avito/internal/infra/cache      [no test files]
?       avito/internal/infra/db [no test files]
?       avito/internal/model    [no test files]
ok      avito/internal/repo/courier     1.053s
ok      avito/internal/repo/delivery    1.027s
?       avito/internal/router   [no test files]
ok      avito/internal/usecases/courier 1.074s
ok      avito/internal/usecases/courier/mapper  1.024s
?       avito/internal/usecases/courier/mocks   [no test files]
ok      avito/internal/usecases/courier/status_check    1.010s
?       avito/internal/usecases/courier/status_check/mocks      [no test files]
ok      avito/internal/usecases/delivery        1.017s
ok      avito/internal/usecases/delivery/factory        1.013s
?       avito/internal/usecases/delivery/mocks  [no test files]
ok      avito/internal/utils    1.011s
?       avito/pkg       [no test files]
```

## Swagger документация

API документация доступна в файле `swagger.yaml`. Используйте Swagger UI для интерактивного просмотра и тестирования эндпоинтов.

Для просмотра можно использовать:
- [Swagger Editor](https://editor.swagger.io/) - загрузить файл `swagger.yaml`
- Локальный Swagger UI сервер  

### Чуть больее подробно про логику
Я представил абстрактную службу доставки, где один курьер ведёт несколько заказов по разным адресам. 
В таком сценарии: Заказы должны быть назначены только одному курьеру, что оправдывает UNIQUE (order_id). 
Курьер может иметь несколько заказов одновременно. Логика успешного заказа — это метрика для системы, а не для клиента: клиент может считать заказ успешным, даже если дедлайн по внутренним правилам был нарушен; система считает заказ успешным только если он завершён курьером вручную (через unassign); если дедлайн прошёл, а курьер всё ещё busy, система может теоретически снять заказ с курьера и передать другому. При этом снятие полностью удаляет запись из delivery, разрывая связь между курьером и заказом.






# ДЗ 7,8,9

### usecases/delivery/AssignCouriersOnOrdersFromGateWay

Добавлена логика межсервисного взаимодействия по grpc. Хотелось бы узнать как делать ретраи на попытки подключения клиентом, потому что я не нашел что-то типа метода ping, по этому сейчас там не очень красивый код


### Кафка: usecases/event и  usecases/bootstrap

Не могу сказать, что до конца осознал работу с кафкой.
1) Сделал отдельный main и **service**  в ***docker-compose*** 
2) Не удалось потестить **cancel**, так как у меня не приходит статус от **order-service**, такая же проблема и с **create**, подробнее писал в чатике.

#### Синхронные запросы
Тут не совсем до конца понял их смысл. Из того что удалось понять, это то, что статус может поменяться до того, как он придет в кафку, и тогда случится что-то типа
1) created - первый раз
2) canceled - пользователь отменил 
3) created - пришел в кафку

Заказ улетел на создание.

Для проверки использовал тот же **gateway**  что и в 7 задании, просто расширил его методами, которые описаны в proto. 

Так же появился вопрос в момент выполнения задания, что делать если у нас рассинхрон в статусах? 
Пример: 
* Создали заказ -> кафка упала -> статус поменялся на тот, который мы не обрабатываем (cooking) -> подняли кафку ->  он прочитал невалидный статус и скрипнул его->потеряли заказ.

Пока решил эту проблему синхронной загрузкой, она назначает курьеров на заказы, статус которых не **cancel**  и не **completed**.
Из потенциальных проблем вижу большую нагрузку на сервис, если выгружать все сразу(мб использовать бакеты). 

Так же из-за этого возникла проблема с синхронным запросом из 7 задания. Так как в случае ошибки я не двигаю курсор, он всегда выкидывает ошибку 
```
AssignCouriers: ERROR: duplicate key value violates unique constraint "delivery_order_id_key" (SQLSTATE 23505)
```
Если будут подсказки по решению этой проблемы, то буду очень признателен.

### Мониторинг

Тут пока все очень тяжело пошло. Дашборды было достаточно проблемно строить, из-за того, что первый раз работал в графане.  
Из проблем которые тут возникли:
1) Не до конца понял, как архитектурно правильно положить метрики и методы их загрузки.
2) Не получилось предзагрузить коды ответов, по этому они будут автоматически добавляться, как только первый раз пройдут через ***middleware*** 

Из метрик которые удалось отобразить:
1) RPM(Request per minute) - ```rate(http_requests_total[60s]) *60```
2) Общее количество запросов по ручке - ```sum by (path) (http_requests_total)```
3) Общее количество запросов по статусу - ```sum by(status) (http_requests_total)```
4) Общее количество звпросов - ```sum(http_requests_total)```

### Логгер

Сделал обертку над стандартным логгером и покрасил его цветами.


## Что добавить 

1) Пропустил тесты, так как не успевал до первого ревью

### Чтобы еще хотелось попробовать в рамках заданий 

1) Больше работы с кафкой
2) В будущем если будет возможность, добавить в курс раздел по сборке, так как у меня возникли трудности, чтобы собрать все это

# Новый env

```
#POSTGRES_USER=admin
#POSTGRES_PASSWORD=admin
#POSTGRES_DB=avito_db
#POSTGRES_PORT=5432
#POSTGRES_HOST=localhost
#MAX_CONNECTIONS=10
#PORT=8080
#COURIER_PORT=8082
#MAX_ATTEMPTS=10
#DELAY=3
#PERIOD_CHECK=20
#BATCH_SIZE=3
#CACHE_TTL=60
#CACHE_SIZE=1
#ORDER_SERVICE_HOST=localhost:50051
#POLLING_PERIOD=3
#RETRY_CONNECT_TO_ORDER_SERVICE=10
#CONNECT_DELAY=2
#KAFKA_PORT=9093
#KAFKA_HOST=localhost
#KAFKA_UI_PORT=5050
#KAFKA_TOPIC=order.status.changed
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin
POSTGRES_DB=avito_db
POSTGRES_PORT=5432
POSTGRES_HOST=postgres
MAX_CONNECTIONS=10
PORT=8080
COURIER_PORT=8082
MAX_ATTEMPTS=10
DELAY=3
PERIOD_CHECK=20
BATCH_SIZE=3
CACHE_TTL=60
CACHE_SIZE=1
ORDER_SERVICE_HOST=service-order:50051
POLLING_PERIOD=3
RETRY_CONNECT_TO_ORDER_SERVICE=10
CONNECT_DELAY=2
KAFKA_PORT=9092
KAFKA_UI_PORT=5050
KAFKA_TOPIC=order.status.changed

```
Для запуска вне докера, раскомментировать комментарии и закомментировать то, что сейчас в нем не кмменты




# RateLimiter
Для написания rate limiter использовал 
```
https://dev.to/vivekalhat/rate-limiting-for-beginners-what-it-is-and-how-to-build-one-in-go-955
```

# Backoff 

Для backoff использовал готовую библиотеку + интерцептор в качестве middleware для grpc.

Для теста отправлял первые 2 запроса от service order как 409. 


# CI/CD

От себя добавил порог покрытия, при котором ci не пройдет. Он получился на уровне 30%, но только из-за того, что в coverage попадают например protobuf файлы.  
Есть ли какая-то возможность игнорировать их, потому, что если запускать unit тесты, то покрытие в среднем 80+. 

# Профилирование

Возник вопрос, почему 6000 считается небезопасным? Когда запускал на нем, html не отдавался. 

Репорты по explain analize, лежат в папке sql_report.

