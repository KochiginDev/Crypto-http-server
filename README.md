# crypthttp

HTTP-сервис на Go для отслеживания курсов криптовалют. Пользователь регистрируется, добавляет монеты по тикеру (BTC, ETH, ...), а сервис берёт актуальную цену у [CoinGecko](https://www.coingecko.com/), сохраняет её в PostgreSQL и ведёт историю цен со статистикой.

## Возможности

- Регистрация и вход с JWT-аутентификацией (пароли хранятся как bcrypt-хэши)
- Добавление и удаление криптовалют по тикеру
- Получение списка отслеживаемых монет и информации по конкретной монете
- Обновление цены монеты по запросу
- История цен и статистика (минимум, максимум, среднее, изменение цены)
- Единый формат ошибок и корректные HTTP-статусы

## Стек

- Go, роутер [chi](https://github.com/go-chi/chi)
- PostgreSQL 16, драйвер [pgx](https://github.com/jackc/pgx)
- [golang-jwt](https://github.com/golang-jwt/jwt) и `golang.org/x/crypto/bcrypt`
- Внешний API: CoinGecko
- Тесты: Python (`tests/tests.py`), запускаются через `make test`

## Структура проекта

```
.
├── cryptoserver.go          # точка входа, сборка зависимостей
├── api/http/httpapi/        # HTTP-слой: хендлеры, роутер, middleware, ответы и ошибки
├── usecases/                # интерфейсы бизнес-логики (Auth, Crypto)
│   └── service/             # реализация бизнес-логики
├── repository/              # интерфейсы хранилища
│   └── pg/                  # реализация на PostgreSQL
├── provider/                # интерфейсы провайдера цен
│   └── coingecko/           # клиент CoinGecko
├── domain/                  # доменные сущности и ошибки
├── pkg/http/                # запуск HTTP-сервера
├── init/init.sql            # схема БД
├── docker-compose.yml       # PostgreSQL для разработки
├── compile.sh / execute.sh  # сборка и запуск (используются тестами)
└── tests/tests.py           # интеграционные тесты API
```

Зависимости направлены внутрь: HTTP-слой знает про `usecases`, а `usecases/service` работает через интерфейсы `repository` и `provider`. Конкретные реализации подставляются в `cryptoserver.go`.

## Быстрый старт

### Требования

- Go (версия указана в `go.mod`)
- Docker и Docker Compose
- Python 3 (только для запуска тестов)

### 1. Запустить PostgreSQL

```bash
docker compose up -d
```

Контейнер создаёт базу `productdb` и при первом запуске выполняет `init/init.sql` (таблицы `users`, `cryptos`, `price_history`).

### 2. (Необязательно) Задать ключ CoinGecko

```bash
export COINGECKO_KEY=your_demo_api_key
```

Без ключа сервис тоже работает, но публичный API CoinGecko имеет жёсткие лимиты запросов. Ключ передаётся в заголовке `x-cg-demo-api-key`.

### 3. Собрать и запустить сервер

```bash
./compile.sh
./execute.sh
```

или напрямую:

```bash
go run cryptoserver.go
```

Сервер слушает порт `8080`.

## Конфигурация

| Параметр | Значение | Где задаётся |
|---|---|---|
| Порт | `8080` | `cryptoserver.go` |
| DSN базы данных | `postgres://postgres:mypassword@localhost:5432/productdb` | `cryptoserver.go` |
| Ключ CoinGecko | переменная окружения `COINGECKO_KEY` | окружение |
| Время жизни токена | 24 часа | `cryptoserver.go` |
| Секрет JWT | генерируется случайно при каждом старте | `cryptoserver.go` |

> **Важно:** секрет JWT создаётся заново при каждом запуске сервера, поэтому после перезапуска все ранее выданные токены перестают действовать, и нужно снова выполнить вход.

## API

Все запросы и ответы в формате JSON. Все эндпоинты `/crypto*` требуют заголовок:

```
Authorization: Bearer <token>
```

### Аутентификация

| Метод | Путь | Описание |
|---|---|---|
| `POST` | `/auth/register` | Регистрация, возвращает токен |
| `POST` | `/auth/login` | Вход, возвращает токен |

Тело запроса:

```json
{ "username": "alice", "password": "secret123" }
```

Ограничения: имя пользователя от 1 до 50 символов, пароль от 6 до 50 символов.

Ответ (`201` при регистрации, `200` при входе):

```json
{ "token": "eyJhbGciOi..." }
```

### Криптовалюты

| Метод | Путь | Описание | Успех |
|---|---|---|---|
| `POST` | `/crypto` | Добавить монету | `201` |
| `GET` | `/crypto` | Список монет | `200` |
| `GET` | `/crypto/{symbol}` | Информация о монете | `200` |
| `PUT` | `/crypto/{symbol}/refresh` | Обновить цену | `200` |
| `GET` | `/crypto/{symbol}/history` | История цен | `200` |
| `GET` | `/crypto/{symbol}/stats` | Статистика | `200` |
| `DELETE` | `/crypto/{symbol}` | Удалить монету | `200` |

Тикер нечувствителен к регистру: `btc`, `BTC` и ` btc ` считаются одним и тем же.

#### `POST /crypto`

```json
{ "symbol": "BTC" }
```

Ответ `201`:

```json
{
  "crypto": {
    "symbol": "BTC",
    "name": "Bitcoin",
    "current_price": 59526,
    "last_updated": "2026-11-25T10:15:30.123456Z"
  }
}
```

#### `GET /crypto`

```json
{
  "cryptos": [
    { "symbol": "BTC", "name": "Bitcoin", "current_price": 59526, "last_updated": "..." }
  ]
}
```

Если монет нет, возвращается `{"cryptos": []}`.

#### `GET /crypto/{symbol}`

Возвращает саму монету без обёртки:

```json
{ "symbol": "BTC", "name": "Bitcoin", "current_price": 59526, "last_updated": "..." }
```

#### `PUT /crypto/{symbol}/refresh`

Запрашивает свежую цену у CoinGecko, сохраняет её и добавляет запись в историю.

```json
{ "crypto": { "symbol": "BTC", "name": "Bitcoin", "current_price": 59530, "last_updated": "..." } }
```

#### `GET /crypto/{symbol}/history`

```json
{
  "symbol": "BTC",
  "history": [
    { "price": 59526, "timestamp": "2026-11-25T10:15:30Z" },
    { "price": 59530, "timestamp": "2026-11-25T10:20:11Z" }
  ]
}
```

#### `GET /crypto/{symbol}/stats`

```json
{
  "symbol": "BTC",
  "current_price": 59530,
  "stats": {
    "min_price": 59526,
    "max_price": 59530,
    "avg_price": 59528,
    "price_change": 4,
    "price_change_percent": 0.0067,
    "records_count": 2
  }
}
```

#### `DELETE /crypto/{symbol}`

Удаляет монету вместе с её историей (`ON DELETE CASCADE`). Ответ: `200` и пустой объект `{}`.

### Ошибки

Ошибки возвращаются в одном формате:

```json
{ "error": "Not Found" }
```

| Статус | Сообщение | Когда |
|---|---|---|
| `400` | `Invalid Input` | некорректное тело запроса, пустой тикер, невалидные имя или пароль |
| `400` | `Unknown Symbol` | такого тикера нет в CoinGecko |
| `401` | `Invalid Credentials` | неверный логин или пароль, нет токена или токен недействителен |
| `404` | `Not Found` | монета не отслеживается |
| `409` | `Already Exists` | пользователь или монета уже существует |
| `500` | `Internal Error` | внутренняя ошибка, в том числе недоступность провайдера |

### Пример сессии

```bash
# регистрация
curl -s -X POST localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# сохранить токен
TOKEN=$(curl -s -X POST localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}' | jq -r .token)

# добавить BTC
curl -s -X POST localhost:8080/crypto \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"symbol":"BTC"}'

# посмотреть статистику
curl -s localhost:8080/crypto/BTC/stats -H "Authorization: Bearer $TOKEN"
```

## База данных

Три таблицы (полная схема в `init/init.sql`):

- `users`: пользователи и хэши паролей
- `cryptos`: отслеживаемые монеты (тикер, название, текущая цена, id в CoinGecko)
- `price_history`: история цен, связана с `cryptos` через внешний ключ с каскадным удалением

Список монет общий для всех пользователей: аутентификация защищает доступ к API, но монеты не привязаны к конкретному пользователю.

## Тесты

Интеграционные тесты запускают собранный сервер и проверяют все эндпоинты по HTTP.

```bash
make test
```

`make install` (вызывается автоматически) ставит Python-зависимости из `requirements.txt`.

Тесты требуют доступ к интернету (запросы к CoinGecko) и запущенный PostgreSQL.

> **Перед каждым запуском очищайте таблицу монет.** Тест добавляет BTC и не удаляет его, поэтому при повторном запуске на той же базе `POST /crypto` вернёт `409`:
>
> ```bash
> docker exec -it my_go_postgres psql -U postgres -d productdb -c "DELETE FROM cryptos;"
> ```
>
> или полностью пересоздайте базу: `docker compose down -v && docker compose up -d`.

## Очистка

```bash
make clean            # удалить собранный бинарник и временные файлы
docker compose down   # остановить PostgreSQL (данные сохранятся в томе)
docker compose down -v  # остановить и удалить данные
```
