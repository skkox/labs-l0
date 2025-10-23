package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5"
)
import (
    "encoding/json"
    "github.com/nats-io/stan.go"
)

func main() {
    dbURL := "postgres://orders_user:StrongPassword123@localhost:5432/orders_db"
    conn, err := pgx.Connect(context.Background(), dbURL)
    if err != nil {
        log.Fatalf("Не удалось подключиться к базе: %v", err)
    }
    defer conn.Close(context.Background())
    fmt.Println("Подключение к базе успешно!")

    sc, err := stan.Connect("test-cluster", "orders-service-client")
    if err != nil {
        log.Fatalf("Не удалось подключиться к NATS Streaming: %v", err)
    }
    defer sc.Close()
    fmt.Println("Подключение к NATS Streaming успешно!")

    _, err = sc.Subscribe("orders", func(m *stan.Msg) {
        fmt.Println("Получено сообщение:", string(m.Data))

        var order map[string]interface{}
        if err := json.Unmarshal(m.Data, &order); err != nil {
            log.Println("Ошибка при разборе JSON:", err)
            return
        }

        fmt.Println("Заказ распознан:", order["order_uid"])
    }, stan.DurableName("orders-durable"))
    if err != nil {
        log.Fatalf("Ошибка при подписке на канал: %v", err)
    }

    select {}
}