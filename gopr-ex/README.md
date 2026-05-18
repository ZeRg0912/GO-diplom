# REST API для блога на Go

Дипломный проект по курсу Go. В проекте реализован простой REST API для блог-платформы.

В API есть регистрация и авторизация пользователей, создание и получение постов, комментарии к постам, хранение данных в JSON-файлах и отложенное логирование действий через канал и отдельную горутину.

Проект сделан на основе шаблона `gopr-temp-ex`. В качестве хранилища используется не PostgreSQL, а JSON-файлы, так как по заданию разрешено хранить данные в файлах или в базе данных.

## Что реализовано

- `POST /register` — регистрация пользователя;
- `POST /login` — вход пользователя и выдача токена;
- `POST /posts` — создание поста, только для авторизованного пользователя;
- `GET /posts` — получение всех постов;
- `GET /posts/{id}` — получение одного поста;
- `POST /posts/{id}/comments` — создание комментария, только для авторизованного пользователя;
- `GET /posts/{id}/comments` — получение комментариев к посту;
- `GET /health` — проверка состояния сервера.

Также реализовано:

- проверка уникальности `email` и `username`;
- хеширование паролей через `bcrypt`;
- авторизация через простой HMAC-токен;
- сохранение пользователей, постов и комментариев в JSON-файлы;
- ответы и ошибки в формате JSON;
- логирование создания постов и комментариев через канал и горутину;
- запуск локально и через Docker Compose.

## Используемые технологии

- Go 1.21+
- `net/http`
- `encoding/json`
- `golang.org/x/crypto/bcrypt`
- JSON-файлы
- Docker / Docker Compose

## Структура проекта

```text
.
├── main.go
├── api/
│   └── main.go
├── internal/
│   ├── app/
│   ├── errors/apperrors/
│   ├── handler/
│   ├── logger/
│   ├── middleware/
│   ├── model/
│   ├── service/
│   └── storage/
├── pkg/
│   └── auth/
├── data/
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md
```

Кратко по основным папкам:

- `internal/handler` — HTTP-обработчики;
- `internal/service` — основная логика приложения;
- `internal/storage` — работа с JSON-файлами;
- `internal/logger` — логирование через канал и горутину;
- `internal/middleware` — middleware для авторизации, CORS, логирования запросов;
- `internal/model` — модели данных;
- `pkg/auth` — работа с паролями и токенами.

## Переменные окружения

Пример файла `.env`:

```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
TOKEN_SECRET=change-me-in-production
DATA_DIR=data
LOG_FILE=log.txt
```

Для локальной проверки можно создать `.env` из примера:

```powershell
Copy-Item .env.example .env
```

Если `.env` не создать, приложение всё равно использует значения по умолчанию.

## Запуск локально

Команды ниже приведены для Windows 11 и PowerShell.

Сначала нужно находиться в корне проекта, где лежит `go.mod`.

Проверить версию Go:

```powershell
go version
```

Установить зависимости:

```powershell
go mod tidy
```

Проверить форматирование, тестовую сборку и сборку проекта:

```powershell
gofmt -w .
go test ./...
go build ./...
```

Если при `go test ./...` выводится `[no test files]`, это нормально. В проекте нет отдельных unit-тестов, эта команда используется для проверки компиляции пакетов.

Запуск сервера:

```powershell
go run main.go
```

Также можно запустить через точку входа из шаблона:

```powershell
go run ./api
```

После запуска сервер доступен по адресу:

```text
http://localhost:8080
```

Остановить сервер можно через `Ctrl+C`.

## Проверка API в PowerShell

Для POST-запросов в PowerShell удобнее использовать `Invoke-RestMethod`, потому что с `curl.exe` иногда возникают проблемы с передачей JSON-строк.

### 1. Health check

```powershell
$base = "http://localhost:8080"

Invoke-RestMethod -Method Get -Uri "$base/health"
```

Ожидаемый ответ:

```text
status
------
ok
```

### 2. Регистрация

```powershell
$registerBody = @{
    username = "user1"
    email    = "user1@example.com"
    password = "password123"
} | ConvertTo-Json -Compress

$registerResponse = Invoke-RestMethod `
    -Method Post `
    -Uri "$base/register" `
    -ContentType "application/json" `
    -Body $registerBody

$registerResponse
$token = $registerResponse.token
$token
```

После регистрации пользователь сохраняется в файл:

```text
data/users.json
```

Пароль сохраняется не открытым текстом, а в поле `password_hash`.

### 3. Повторная регистрация

Повторная регистрация того же пользователя должна вернуть ошибку `409`, потому что `email` и `username` должны быть уникальными.

```powershell
try {
    Invoke-RestMethod `
        -Method Post `
        -Uri "$base/register" `
        -ContentType "application/json" `
        -Body $registerBody
} catch {
    $_.Exception.Response.StatusCode.value__
    $_.ErrorDetails.Message
}
```

Ожидаемый статус:

```text
409
```

### 4. Логин

```powershell
$loginBody = @{
    email    = "user1@example.com"
    password = "password123"
} | ConvertTo-Json -Compress

$loginResponse = Invoke-RestMethod `
    -Method Post `
    -Uri "$base/login" `
    -ContentType "application/json" `
    -Body $loginBody

$token = $loginResponse.token
$token
```

После логина возвращается токен. Его нужно использовать для создания постов и комментариев.

### 5. Создание поста без токена

```powershell
$postBody = @{
    title   = "First post"
    content = "First post content"
} | ConvertTo-Json -Compress

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri "$base/posts" `
        -ContentType "application/json" `
        -Body $postBody
} catch {
    $_.Exception.Response.StatusCode.value__
    $_.ErrorDetails.Message
}
```

Ожидаемый статус:

```text
401
```

### 6. Создание поста с токеном

```powershell
$postResponse = Invoke-RestMethod `
    -Method Post `
    -Uri "$base/posts" `
    -ContentType "application/json" `
    -Headers @{ Authorization = "Bearer $token" } `
    -Body $postBody

$postResponse
$postID = $postResponse.id
```

После создания пост сохраняется в файл:

```text
data/posts.json
```

### 7. Получение всех постов

```powershell
Invoke-RestMethod -Method Get -Uri "$base/posts"
```

### 8. Получение поста по ID

```powershell
Invoke-RestMethod -Method Get -Uri "$base/posts/$postID"
```

### 9. Создание комментария без токена

```powershell
$commentBody = @{
    text = "First comment"
} | ConvertTo-Json -Compress

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri "$base/posts/$postID/comments" `
        -ContentType "application/json" `
        -Body $commentBody
} catch {
    $_.Exception.Response.StatusCode.value__
    $_.ErrorDetails.Message
}
```

Ожидаемый статус:

```text
401
```

### 10. Создание комментария с токеном

```powershell
$commentResponse = Invoke-RestMethod `
    -Method Post `
    -Uri "$base/posts/$postID/comments" `
    -ContentType "application/json" `
    -Headers @{ Authorization = "Bearer $token" } `
    -Body $commentBody

$commentResponse
```

После создания комментарий сохраняется в файл:

```text
data/comments.json
```

### 11. Получение комментариев к посту

```powershell
Invoke-RestMethod -Method Get -Uri "$base/posts/$postID/comments"
```

### 12. Проверка файлов

```powershell
Get-Content .\data\users.json
Get-Content .\data\posts.json
Get-Content .\data\comments.json
```

### 13. Проверка логирования

После создания поста и комментария нужно подождать 1-2 секунды и проверить файл:

```powershell
Get-Content .\log.txt
```

Пример записей:

```text
2026-05-18T16:36:26Z user 1 created post 1
2026-05-18T16:36:48Z user 1 created comment 1 on post 1
```

## Запуск через Docker Compose

Перед запуском нужно убедиться, что Docker Desktop запущен (Windows).

Собрать и запустить контейнер:

```powershell
docker compose up --build
```

Или запустить в фоне:

```powershell
docker compose up -d --build
```

Проверить API:

```powershell
curl.exe -i http://localhost:8080/health
```

Посмотреть логи контейнера:

```powershell
docker compose logs -f
```

Остановить контейнеры:

```powershell
docker compose down
```

В `docker-compose.yml` подключена папка `data`, поэтому данные сохраняются на компьютере:

```yaml
volumes:
  - ./data:/app/data
```

## API

Все ответы возвращаются в формате JSON.

### `GET /health`

Проверка работы сервера.

Ответ:

```json
{"status":"ok"}
```

### `POST /register`

Регистрация пользователя.

Пример тела запроса:

```json
{
  "username": "user1",
  "email": "user1@example.com",
  "password": "password123"
}
```

Успешный ответ: `201 Created`.

### `POST /login`

Вход пользователя.

Пример тела запроса:

```json
{
  "email": "user1@example.com",
  "password": "password123"
}
```

Успешный ответ: `200 OK`.

### `POST /posts`

Создание поста. Нужен заголовок авторизации.

```http
Authorization: Bearer <token>
```

Пример тела запроса:

```json
{
  "title": "First post",
  "content": "First post content"
}
```

Успешный ответ: `201 Created`.

### `GET /posts`

Получение всех постов.

Успешный ответ: `200 OK`.

### `GET /posts/{id}`

Получение одного поста по ID.

Успешный ответ: `200 OK`.

### `POST /posts/{id}/comments`

Создание комментария к посту. Нужен заголовок авторизации.

```http
Authorization: Bearer <token>
```

Пример тела запроса:

```json
{
  "text": "First comment"
}
```

Успешный ответ: `201 Created`.

### `GET /posts/{id}/comments`

Получение комментариев к посту.

Успешный ответ: `200 OK`.

## Хранение данных

Данные хранятся в JSON-файлах:

```text
data/users.json
data/posts.json
data/comments.json
```

Файлы создаются автоматически при запуске приложения, если их нет.

Данные не очищаются после остановки сервера. Это сделано специально, чтобы пользователи, посты и комментарии сохранялись между запусками.

Для чистой проверки можно удалить файлы:

```powershell
Remove-Item .\data\users.json -ErrorAction SilentlyContinue
Remove-Item .\data\posts.json -ErrorAction SilentlyContinue
Remove-Item .\data\comments.json -ErrorAction SilentlyContinue
Remove-Item .\log.txt -ErrorAction SilentlyContinue
```

## Логирование

При создании поста или комментария формируется строка события, например:

```text
user 1 created post 1
user 1 created comment 1 on post 1
```

Событие отправляется в канал. Отдельная горутина читает канал, ждёт 1 секунду и записывает событие в `log.txt`. Также событие выводится в консоль.

При остановке сервера через `Ctrl+C` канал закрывается, а оставшиеся события дописываются в файл.