
# Пошаговая инструкция по созданию gRPC-клиента и сервера на Go

Ниже приведён максимально подробный **пошаговый гайд** по созданию gRPC-клиента и сервера на Go. Я расскажу, что нужно установить, как оформить `.proto`, как сгенерировать код, как реализовать сервер, как написать клиента и как запустить всё вместе.

## Содержание

1. [Предварительные требования](#1-предварительные-требования)  
2. [Установка protoc и Go-плагинов](#2-установка-protoc-и-go-плагинов)  
3. [Структура проекта](#3-структура-проекта)  
4. [Написание файла `.proto`](#4-написание-файла-proto)  
5. [Генерация Go-кода из `.proto`](#5-генерация-go-кода-из-proto)  
6. [Реализация gRPC-сервера на Go](#6-реализация-grpc-сервера-на-go)  
7. [Реализация gRPC-клиента на Go](#7-реализация-grpc-клиента-на-go)  
8. [Запуск и проверка](#8-запуск-и-проверка)  
9. [Расширение: добавление нового метода](#9-расширение-добавление-нового-метода)  
10. [Дополнительно: TLS-аутентификация (опционально)](#10-дополнительно-tls-аутентификация-опционально)  
11. [Советы по отладке и свежие «подводные камни»](#11-советы-по-отладке-и-свежие-подводные-камни)  

---

## 1. Предварительные требования

1. **Go ≥ 1.18**  
   Убедитесь, что Go установлен и доступен в `$PATH`.  

   ```bash
   go version
   # -> go version go1.20.4 darwin/amd64 (пример)
   ```

2. **protoc (Protocol Buffers compiler) ≥ 3.x**  
   `protoc` — это компилятор `.proto` в разные языки (Go, Python, Java и т.д.).  
   - **macOS (Homebrew):**  

     ```bash
     brew install protobuf
     ```

   - **Ubuntu/Debian:**  

     ```bash
     sudo apt update
     sudo apt install -y protobuf-compiler
     ```

   - **Windows (choco):**  

     ```powershell
     choco install protoc
     ```

3. **Плагины для Go**  
   - `protoc-gen-go` — генерирует Go-структуры (message, enum) из `.proto`.  
   - `protoc-gen-go-grpc` — генерирует Go-интерфейсы (gRPC-стабы) для сервисов (rpc).  

   Установим их через `go install`:  

   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

   После этого в `$GOPATH/bin` или `$GOBIN` (часто это `~/go/bin`) появятся бинарники:

   ```
   protoc-gen-go
   protoc-gen-go-grpc
   ```

   **Важно:** убедитесь, что `$(go env GOPATH)/bin` есть в вашем `$PATH`, иначе `protoc` не найдёт эти плагины.

4. **gRPC-библиотека для Go**  
   В проекте понадобятся пакеты:

   ```bash
   go get google.golang.org/grpc
   go get google.golang.org/protobuf
   ```

   (Обычно это прописывается автоматически, когда вы будете импортировать пакеты в коде.)

---

## 2. Установка protoc и Go-плагинов

Убедимся, что у вас в системе установлены и доступны:

- `protoc` (команда должна вернуть версию).  

  ```bash
  protoc --version
  # -> libprotoc 3.xx.x
  ```

- `protoc-gen-go` и `protoc-gen-go-grpc`.  

  ```bash
  which protoc-gen-go
  # -> /Users/you/go/bin/protoc-gen-go

  which protoc-gen-go-grpc
  # -> /Users/you/go/bin/protoc-gen-go-grpc
  ```

Если какие-то из команд не находятся, нужно добавить `~/go/bin` (или другой `$GOBIN`) в переменную окружения `PATH`, например:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## 3. Структура проекта

Допустим, наш репозиторий называется `grpc_example`. Создадим базовую структуру:

```
grpc_example/
│
├── balance.proto
├── go.mod
├── go.sum
│
├── server/
│   └── main.go         # код gRPC-сервера
│
└── client/
    └── main.go         # код gRPC-клиента
```

- **`balance.proto`** — общий `.proto`-файл, из которого будут генерироваться и серверные, и клиентские Go-файлы.
- **`server/main.go`** — реализация gRPC-сервера.
- **`client/main.go`** — реализация простого gRPC-клиента для тестирования.

> В этом примере мы разделили код на папки `server/` и `client/`. Поэтому в Go-модуле мы укажем модуль `grpc_example` (или полный путь `github.com/ваш_ник/grpc_example`).

---

### 3.1 Инициализация Go-модуля

В корне `grpc_example/` запускаем:

```bash
go mod init github.com/your_username/grpc_example
```

(`github.com/your_username/grpc_example` замените на свой актуальный путь, если собираетесь пушить в GitHub.)

После этого в `grpc_example/go.mod` появится:

```go
module github.com/your_username/grpc_example

go 1.20

require (
    google.golang.org/grpc v1.x.x
    google.golang.org/protobuf v1.x.x
)
```

---

## 4. Написание файла `balance.proto`

Создайте файл `grpc_example/balance.proto`. Мы опишем очень простой сервис `BalanceService` с единственным методом `GetBalance`.

```proto
syntax = "proto3";

// Опционально: описание пакета и правила генерации Go-кода
package balance;

// Указываем, где генерировать Go-пакет. Для нашего случая:
option go_package = "github.com/your_username/grpc_example/balancepb;balancepb";

// Сервис BalanceService: метод GetBalance
service BalanceService {
  // Запрос баланса: передаём имя пользователя
  rpc GetBalance (GetBalanceRequest) returns (GetBalanceResponse);
}

// Структура запроса
message GetBalanceRequest {
  string username = 1;
}

// Структура ответа
message GetBalanceResponse {
  string username = 1;
  double balance  = 2;
}
```

Обратите внимание:

1. `package balance;` — логический пакет для protobuf.
2. `option go_package = "github.com/your_username/grpc_example/balancepb;balancepb";`  
   - Первая часть (`github.com/your_username/grpc_example/balancepb`) — это **Go-модуль**, куда будет записан сгенерированный `.go`-файл.  
   - Вторая часть (`balancepb`) — **имя пакета** в Go-коде.  
   Итого: при генерации появится каталог `grpc_example/balancepb`, внутри которого будет код с `package balancepb`.

---

## 5. Генерация Go-кода из `balance.proto`

Перейдём в корень проекта `grpc_example/` и выполним команду:

```bash
protoc \
  --go_out=. \
  --go-grpc_out=. \
  balance.proto
```

Что произойдёт:

- Плагин `--go_out=.` создаст файл `balancepb/balance.pb.go` с Go-структурами сообщений (`GetBalanceRequest`, `GetBalanceResponse`).
- Плагин `--go-grpc_out=.` создаст файл `balancepb/balance_grpc.pb.go` с интерфейсом сервера `BalanceServiceServer` и клиентским интерфейсом `BalanceServiceClient`.

После этого в проекте появится папка:

```
grpc_example/
├── balancepb/
│   ├── balance.pb.go
│   └── balance_grpc.pb.go
...
```

Внутри сгенерированных файлов будет примерно следующее (сокращённо):

```go
// balancepb/balance.pb.go
package balancepb

type GetBalanceRequest struct {
    Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
    // ...
}
type GetBalanceResponse struct {
    Username string  `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
    Balance  float64 `protobuf:"fixed64,2,opt,name=balance,proto3" json:"balance,omitempty"`
    // ...
}

// balancepb/balance_grpc.pb.go
package balancepb

type BalanceServiceClient interface {
    GetBalance(ctx context.Context, in *GetBalanceRequest, opts ...grpc.CallOption) (*GetBalanceResponse, error)
}

type BalanceServiceServer interface {
    GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceResponse, error)
}

func RegisterBalanceServiceServer(s *grpc.Server, srv BalanceServiceServer) { ... }
```

Эти два сгенерированных файла — **единственный источник правды** для сервера и клиента. Вы **не правите** их вручную, а вносите изменения только в `balance.proto` и потом пересоздаёте.

---

## 6. Реализация gRPC-сервера на Go

Перейдите в директорию `grpc_example/server/` и создайте файл `main.go`. Полный пример сервера:

```go
// server/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net"

    "google.golang.org/grpc"
    "github.com/your_username/grpc_example/balancepb"
)

// serverImpl реализует интерфейс balancepb.BalanceServiceServer
type serverImpl struct {
    balancepb.UnimplementedBalanceServiceServer
}

// Реализация метода GetBalance
func (s *serverImpl) GetBalance(ctx context.Context, req *balancepb.GetBalanceRequest) (*balancepb.GetBalanceResponse, error) {
    username := req.GetUsername()
    // Здесь могла бы быть логика обращения к БД или API
    // Для примера вернём баланс = 123.45
    balance := 123.45

    response := &balancepb.GetBalanceResponse{
        Username: username,
        Balance:  balance,
    }
    fmt.Printf("Получен запрос GetBalance для %q, отдаем баланс %.2f\n", username, balance)
    return response, nil
}

func main() {
    // 1. Создаём TCP-листенер
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Не удалось слушать порт :50051: %v", err)
    }

    // 2. Создаём gRPC-сервер
    grpcServer := grpc.NewServer()

    // 3. Регистрируем нашу реализацию в gRPC-сервере
    balancepb.RegisterBalanceServiceServer(grpcServer, &serverImpl{})

    fmt.Println("gRPC-сервер запущен, слушает :50051")
    // 4. Запускаем блокирующий процесс: принимаем входящие соединения
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Ошибка при запуске gRPC-сервера: %v", err)
    }
}
```

### Пояснения

1. **Импортируем `balancepb`** — папку с сгенерированным кодом из `.proto`.  

   ```go
   import "github.com/your_username/grpc_example/balancepb"
   ```

2. **Тип `serverImpl`**  
   Встраиваем в себя `balancepb.UnimplementedBalanceServiceServer`. Это нужно, чтобы не реализовывать все методы, если их несколько.  
   Достаточно дописать только `GetBalance(...)`.

3. **Метод `GetBalance`**  
   - Принимает `context.Context` и указатель `*balancepb.GetBalanceRequest`.  
   - Извлекает поле `username := req.GetUsername()`.  
   - Формирует ответ `balancepb.GetBalanceResponse` с полями `Username` и `Balance`.  
   - Возвращает `(response, nil)`.

4. **main**  
   - `net.Listen("tcp", ":50051")` — создаёт TCP-сервер, прослушивающий порт 50051.  
   - `grpc.NewServer()` — инициализирует пустой gRPC-сервер.  
   - `balancepb.RegisterBalanceServiceServer(grpcServer, &serverImpl{})` — регистрирует наш `serverImpl` как обработчик сервиса `BalanceService`.  
   - `grpcServer.Serve(lis)` — блокирует выполнение и начинает принимать входящие gRPC-подключения.

> Как только вы скомпилируете и запустите `server/main.go`, сервер будет висеть в фоне и ожидать запросы на порту `50051`.

---

## 7. Реализация gRPC-клиента на Go

Теперь создадим `grpc_example/client/main.go`. Клиент отправит запрос и получит ответ от сервера.

```go
// client/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "google.golang.org/grpc"
    "github.com/your_username/grpc_example/balancepb"
)

func main() {
    // 1. Подключаемся к gRPC-серверу
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("Не удалось подключиться к серверу: %v", err)
    }
    defer conn.Close()

    // 2. Создаём клиента BalanceService
    client := balancepb.NewBalanceServiceClient(conn)

    // 3. Делаем контекст с таймаутом (чтобы не ждать вечно)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 4. Формируем запрос
    req := &balancepb.GetBalanceRequest{
        Username: "alice",
    }

    // 5. Вызываем метод GetBalance
    res, err := client.GetBalance(ctx, req)
    if err != nil {
        log.Fatalf("Ошибка при вызове GetBalance: %v", err)
    }

    // 6. Обрабатываем и выводим ответ
    fmt.Printf("Баланс пользователя %q: %.2f\n", res.GetUsername(), res.GetBalance())
}
```

### Пояснения

1. **`grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())`**  
   - `WithInsecure()` — означает, что **нет TLS** (для тестов/локального запуска).  
   - `WithBlock()` — заставляет `Dial` ждать реального установления соединения, вместо того чтобы возвращать сразу.  

2. **`balancepb.NewBalanceServiceClient(conn)`**  
   - Возвращает объект клиента, реализующий интерфейс `BalanceServiceClient`.  
   В этом интерфейсе есть метод `GetBalance(ctx, request)`.

3. **Контекст с таймаутом `context.WithTimeout(...)`**  
   - Мы оборачиваем `GetBalance` в контекст, чтобы запрос автоматически отменился, если сервер не ответит за 5 секунд.

4. **Вызов `client.GetBalance(ctx, req)`**  
   - Сервер получит `username = "alice"`, выполнит метод (см. выше) и вернёт `GetBalanceResponse{Username:"alice", Balance:123.45}`.

5. **Вывод результата в консоль**  

   ```go
   fmt.Printf("Баланс пользователя %q: %.2f\n", res.GetUsername(), res.GetBalance())
   ```

---

## 8. Запуск и проверка

1. **Открываем два терминала (или вкладки).**

2. **В одном терминале запускаем сервер:**  

   ```bash
   cd grpc_example/server
   go run main.go
   ```

   Вы должны увидеть:  

   ```
   gRPC-сервер запущен, слушает :50051
   ```

3. **В другом терминале запускаем клиента:**  

   ```bash
   cd grpc_example/client
   go run main.go
   ```

   Ожидаемый вывод:  

   ```
   Баланс пользователя "alice": 123.45
   ```

4. **Если клиент пишет об ошибке (например, «connection refused»),**  
   проверьте, что сервер действительно запущен и слушает порт `50051`, что нет брандмауэра и что `grpc.WithInsecure()` подходит для вашего случая.  

---

## 9. Расширение: добавление нового метода

Допустим, вы захотели добавить в сервис метод `GetHistory` (например, список прошлых транзакций). Как это сделать “согласно нормам”?

### Шаг 1. Редактируем `balance.proto`

```proto
syntax = "proto3";
package balance;
option go_package = "github.com/your_username/grpc_example/balancepb;balancepb";

service BalanceService {
  rpc GetBalance (GetBalanceRequest)  returns (GetBalanceResponse);
  rpc GetHistory (GetHistoryRequest)  returns (GetHistoryResponse);
}

message GetBalanceRequest  { string username = 1; }
message GetBalanceResponse { string username = 1; double balance = 2; }

message GetHistoryRequest { string username = 1; int32 limit = 2; }
message GetHistoryResponse {
  repeated string transactions = 1;
}
```

### Шаг 2. Перегенерировать Go-код

В корне `grpc_example/`:

```bash
protoc --go_out=. --go-grpc_out=. balance.proto
```

Теперь в папке `balancepb/` появятся новые файлы/дополнения с типами `GetHistoryRequest`, `GetHistoryResponse` и интерфейсом `GetHistory(...)`.

### Шаг 3. Реализовать метод в сервере

В `server/main.go`:

```go
func (s *serverImpl) GetHistory(ctx context.Context, req *balancepb.GetHistoryRequest) (*balancepb.GetHistoryResponse, error) {
    username := req.GetUsername()
    limit := req.GetLimit()
    // Тут могла бы быть логика для получения истории из БД.
    // Для примера вернём простой срез строк:
    transactions := []string{
        fmt.Sprintf("User %s transaction 1", username),
        fmt.Sprintf("User %s transaction 2", username),
    }
    if int32(len(transactions)) > limit {
        transactions = transactions[:limit]
    }
    return &balancepb.GetHistoryResponse{Transactions: transactions}, nil
}
```

### Шаг 4. Обновить клиент

В `client/main.go`:

```go
    // предыдущий код ...
    // ----

    // Пример вызова нового метода:
    histReq := &balancepb.GetHistoryRequest{Username: "alice", Limit: 5}
    histRes, err := client.GetHistory(ctx, histReq)
    if err != nil {
        log.Fatalf("GetHistory error: %v", err)
    }
    fmt.Println("Последние транзакции:", histRes.GetTransactions())
```

### Шаг 5. Перезапуск

1. Останавливаем текущий сервер (Ctrl+C) и запускаем заново:

   ```bash
   cd grpc_example/server
   go run main.go
   ```

2. Запускаем клиента, проверяем оба метода.

Готово! Вы без проблем расширили API.

---

## 10. Дополнительно: TLS-аутентификация (опционально)

#### Зачем это нужно?  

Чтобы ваши gRPC-соединения были зашифрованы (например, в продакшене или при удалённом развёртывании).

### Шаг 10.1. Сгенерировать самоподписанные сертификаты (пример)

```bash
# Создаём приватный ключ
openssl genrsa -out server.key 2048

# Создаём CSR (запрос на сертификат) — можно оставить дефолтные значения
openssl req -new -key server.key -out server.csr

# Подписываем CSR собственным CA (самоподписанный)
openssl x509 -req -in server.csr -signkey server.key -out server.crt -days 365
```

В результате:  

- `server.crt` — публичный сертификат  
- `server.key` — приватный ключ  

### Шаг 10.2. Запуск gRPC-сервера с TLS

В `server/main.go` нужно прописать:

```go
import (
    "google.golang.org/grpc/credentials"
    // ...
)

func main() {
    certFile := "server.crt"
    keyFile := "server.key"

    // Загружаем сертификат
    creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
    if err != nil {
        log.Fatalf("Не удалось загрузить сертификаты: %v", err)
    }

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Не удалось слушать: %v", err)
    }

    grpcServer := grpc.NewServer(grpc.Creds(creds))
    balancepb.RegisterBalanceServiceServer(grpcServer, &serverImpl{})

    log.Println("gRPC-сервер с TLS запущен на :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Ошибка сервера: %v", err)
    }
}
```

### Шаг 10.3. Запуск gRPC-клиента с TLS

В `client/main.go`:

```go
import (
    "google.golang.org/grpc/credentials"
    // ...
)

func main() {
    // Загружаем сертификат CA (в данном случае — самоподписанный)
    creds, err := credentials.NewClientTLSFromFile("server.crt", "")
    if err != nil {
        log.Fatalf("Не удалось загрузить CA-сертификат: %v", err)
    }

    conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Fatalf("Не удалось подключиться с TLS: %v", err)
    }
    defer conn.Close()

    client := balancepb.NewBalanceServiceClient(conn)
    // ... дальше, как раньше
}
```

Теперь соединение будет шифроваться по TLS. Обратите внимание, что клиент **доверяет** сертификату `server.crt`. В продакшене обычно используют сертификаты от доверенных CA, а не самоподписанные.

---

## 11. Советы по отладке и свежие «подводные камни»

1. **Проверка версии protoc-плагинов**  
   Если вы получаете ошибку вида “`protoc-gen-go` not found”, вероятно, бинарник не в `$PATH`.  
   Проверьте:

   ```bash
   which protoc-gen-go
   which protoc-gen-go-grpc
   ```

2. **Синтаксис команды protoc**  
   - Путь к `.proto` относительно текущей директории!  
   - Обязательно указывайте `--go_out=.` и `--go-grpc_out=.` в одной и той же папке, чтобы файлы лежали там, где нужно.

3. **Пакеты и `go_package`**  
   - Важно корректно прописать `option go_package = "github.com/your_username/grpc_example/balancepb;balancepb";`  
     иначе сгенерированные файлы могут иметь неправильный `package`-директорию, и при импорте у вас будут ошибки.

4. **Проблемы с портами**  
   - Часто порт `:50051` уже занят. Если при старте сервера вылетает `bind: address already in use`, выберите свободный порт (например, `:50052`).

5. **`WithInsecure()` в gRPC-DIAL**  
   - В новых версиях gRPC Go метод `grpc.WithInsecure()` уже **deprecated**.  
   - Вместо него можно использовать `grpc.WithTransportCredentials(insecure.NewCredentials())`.  
   Пример:

   ```go
   import "google.golang.org/grpc/credentials/insecure"

   conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
   ```

6. **Горутинный panic в сервере**  
   - Если в методе `GetBalance` вы вызываете какую-нибудь “паниковую” операцию, а метод не возвращает ошибку, gRPC зафиксирует panic и закроет соединение.  
   - Рекомендуется всю бизнес-логику оборачивать в защитный блок:

     ```go
     func (s *serverImpl) GetBalance(ctx context.Context, req *balancepb.GetBalanceRequest) (*balancepb.GetBalanceResponse, error) {
         defer func() {
             if r := recover(); r != nil {
                 log.Printf("Recovered from panic in GetBalance: %v", r)
             }
         }()
         // ... остальной код ...
     }
     ```

7. **Понимание контекста (context.Context)**  
   - gRPC-методы всегда получают `context.Context`.  
   - При вызове `client.GetBalance(...)` вы можете передавать контекст с таймаутом или дедлайном — так запрос отменится, если сервер не ответит вовремя.  
   - В серверной части `ctx` содержит метаданные вызова (заголовки, аутентификацию) и сигнал отмены, если клиент прервал соединение.

8. **Совместимость версий**  
   - Если у вас **несовпадают версии** `google.golang.org/grpc` и `google.golang.org/protobuf`, могут вылезти редкие ошибки компиляции.  
   - Всегда делайте `go mod tidy` после установки новых пакетов.

---

## Заключение

В этой инструкции мы:

1. Установили необходимые инструменты (`protoc`, плагины, библиотеки`).  
2. Описали `.proto`-файл и сгенерировали Go-код.  
3. Реализовали gRPC-сервер (`server/main.go`).  
4. Реализовали gRPC-клиента (`client/main.go`).  
5. Провели запуск и проверку работоспособности.  
6. Показали, как расширить API через добавление новых RPC-методов.  
7. Объяснили, как включить TLS-шифрование (при необходимости).  
8. Привели советы по отладке и распространённые ошибки.

Теперь у вас есть **полностью рабочая связка клиент/сервер на gRPC с Go**, и вы можете использовать эту схему для своих микросервисов, бэкенд-сервисов или других задач, где нужен быстрый и типобезопасный RPC-протокол.

Удачи в разработке! Если появятся дополнительные вопросы — спрашивайте.
