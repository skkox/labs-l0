package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Сервис заказов L0")
	fmt.Println("                   ")
	fmt.Println("План выполнения:")
	fmt.Println("1. Настройка PostgreSQL")
	fmt.Println("2. Создание репозитория")
	fmt.Println("3. Создание Go-проекта")
	fmt.Println("4. Структуры данных")
	fmt.Println("5. Подключение к PostgreSQL")
	fmt.Println("6. Подключение к NATS Streaming")
	fmt.Println("7. Сохранение данных")
	fmt.Println("8. HTTP-сервер")
	fmt.Println("9. Тестирование")
	fmt.Println("")
	
	// Тестируем подключение к PostgreSQL
	fmt.Println("Тестирование подключения к PostgreSQL...")
	dbURL := "postgres://orders_user:StrongPassword123@localhost:5432/orders_db"
	
	db, err := NewDB(dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()
	
	fmt.Println("Подключение к PostgreSQL успешно!")
	fmt.Println("Готов к следующему шагу - NATS Streaming")
}
