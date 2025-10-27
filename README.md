# Проект Labs L0

Микросервис для обработки заказов через NATS Streaming с сохранением в PostgreSQL.

---

## 1. Запуск NATS Streaming Server

1. Открой PowerShell №1.
2. Перейди в папку с сервером:

```powershell
cd "C:\Users\Виктория\Downloads\nats-streaming-server-v0.25.6-windows-amd64\nats-streaming-server-v0.25.6-windows-amd64"
````

3. Запусти сервер:

```powershell
.\nats-streaming-server.exe -cid test-cluster
```

 Ожидаемый результат:

```
STREAM: Streaming Server is ready
```

 Не закрывай это окно — сервер работает в фоне.

---

## 2. Запуск основного сервиса (сервер)

1. Открой новое окно PowerShell.
2. Перейди в папку с проектом:

```powershell
cd "C:\Users\Виктория\OneDrive\Документы\Projects\labs-l0"
```

3. Запусти сервер:

```powershell
go run .
```

 Если всё правильно, появится:

```
HTTP-сервер запущен на :8080
```

 Сервис доступен по адресу: [http://localhost:8080](http://localhost:8080)

---

## 3. Отправка тестового заказа

1. Открой PowerShell №3.
2. Перейди в папку `publisher`:

```powershell
cd "C:\Users\Виктория\OneDrive\Документы\Projects\labs-l0\publisher"
```

3. Запусти скрипт для отправки тестового заказа:

```powershell
go run test_publisher.go
```

 Если всё верно, увидишь:

```
 Сообщение успешно отправлено в канал orders!
```

---

## 4. Проверка на сайте

1. Открой браузер.
2. Введи ID заказа для поиска: b583feb7b2b84b6test

 Ты увидишь детали отправленного заказа.
