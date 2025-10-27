package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/stan.go"
)

func main() {
	log.Println("Запуск сервиса заказов L0...")

	dbURL := "postgres://orders_user:StrongPassword123@localhost:5432/orders_db"
	natsURL := "nats://localhost:4222"
	clusterID := "test-cluster"
	clientID := "orders-service"
	channel := "orders"

	db, err := NewDB(dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	cache := NewOrderCache()
	log.Println("Кэш создан")

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

	natsClient, err := NewNATSClient(clusterID, clientID, natsURL, db, cache)
	if err != nil {
		log.Fatalf("Ошибка подключения к NATS Streaming: %v", err)
	}
	defer natsClient.Close()

	if err := natsClient.Subscribe(channel); err != nil {
		log.Fatalf("Ошибка подписки на канал '%s': %v", channel, err)
	}

	server := NewServer(cache)
	go func() {
		if err := server.Start(":8080"); err != nil {
			log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
		}
	}()

	log.Println("Сервис успешно запущен и готов к работе!")
	log.Println("HTTP-сервер: http://localhost:8080")
	log.Println("Ожидание сообщений из NATS Streaming...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Получен сигнал завершения, останавливаем сервис...")
}

func ConnectNATS() (stan.Conn, error) {
	clientID := fmt.Sprintf("order-service-%d", time.Now().UnixNano())

	sc, err := stan.Connect(
		"test-cluster",
		clientID,
		stan.NatsURL("nats://localhost:4222"),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("NATS Streaming connection lost: %v", reason)
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к NATS Streaming: %w", err)
	}

	return sc, nil
}