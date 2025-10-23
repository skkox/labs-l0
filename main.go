package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5"


func main() {
    dbURL := "postgres://orders_user:StrongPassword123@localhost:5432/orders_db"

    conn, err := pgx.Connect(context.Background(), dbURL)
    if err != nil {
        log.Fatalf("Не удалось подключиться к базе: %v", err)
    }
    defer conn.Close(context.Background())

    fmt.Println("Подключение к базе успешно!")
}
