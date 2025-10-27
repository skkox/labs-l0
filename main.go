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
	// Создаем уникальный ID клиента на основе времени
	clientID := fmt.Sprintf("order-service-%d", time.Now().UnixNano())

	// Пробуем подключиться несколько раз
	var sc stan.Conn
	var err error
	for i := 0; i < 3; i++ {
		sc, err = stan.Connect(
			"test-cluster",
			clientID,
			stan.NatsURL("nats://localhost:4222"),
		)
		if err == nil {
			log.Printf("NATS: успешное подключение с ID: %s", clientID)
			break
		}
		log.Printf("NATS: попытка %d неудачна: %v", i+1, err)
		time.Sleep(time.Second)
	}

	return sc, err
}