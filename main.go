package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Запуск сервиса заказов L0...")

	// Параметры подключения
	dbURL := "postgres://orders_user:StrongPassword123@localhost:5432/orders_db"
	natsURL := "nats://localhost:4222"
	clusterID := "test-cluster"
	clientID := "orders-service"
	channel := "orders"

	// 1. Подключаемся к PostgreSQL
	db, err := NewDB(dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	// 2. Создаем кэш
	cache := NewOrderCache()
	log.Println("Кэш создан")

	// 3. Восстанавливаем кэш из базы данных
	log.Println("Восстановление кэша из базы данных...")
	orders, err := db.GetAllOrders()
	if err != nil {
		log.Printf("Предупреждение: не удалось восстановить кэш из БД: %v", err)
	} else {
		for _, order := range orders {
			cache.Set(order.OrderUID, order)
		}
		log.Printf("Кэш восстановлен: загружено %d заказов", len(orders))
	}

	// 4. Подключаемся к NATS Streaming
	natsClient, err := NewNATSClient(clusterID, clientID, natsURL, db, cache)
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS Streaming: %v", err)
	}
	defer natsClient.Close()

	// 5. Подписываемся на канал
	if err := natsClient.Subscribe(channel); err != nil {
		log.Fatalf("Ошибка подписки на канал '%s': %v", channel, err)
	}

	log.Println("Сервис успешно запущен и готов к работе!")
	log.Println("Ожидание сообщений из NATS Streaming...")

	// Ожидаем сигнал завершения (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Получен сигнал завершения, останавливаем сервис...")
}