package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/stan.go"
)

// NATSClient представляет клиент NATS Streaming
type NATSClient struct {
	conn  stan.Conn
	sub   stan.Subscription
	db    *DB
	cache *OrderCache
}

// NewNATSClient создает новое подключение к NATS Streaming
func NewNATSClient(clusterID, clientID, natsURL string, db *DB, cache *OrderCache) (*NATSClient, error) {
	// Подключаемся к NATS Streaming
	conn, err := stan.Connect(
		clusterID,
		clientID,
		stan.NatsURL(natsURL),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Соединение с NATS потеряно: %v", reason)
		}),
	)
	if err != nil {
		return nil, err
	}

	log.Println("Подключение к NATS Streaming успешно!")

	return &NATSClient{
		conn:  conn,
		db:    db,
		cache: cache,
	}, nil
}

// Subscribe подписывается на канал и обрабатывает сообщения
func (nc *NATSClient) Subscribe(channel string) error {
	// Подписываемся на канал с опциями
	sub, err := nc.conn.Subscribe(channel, nc.handleMessage,
		stan.SetManualAckMode(),
		stan.DurableName("orders-service"), 
		stan.StartWithLastReceived(),
	)
	if err != nil {
		return err
	}

	nc.sub = sub
	log.Printf("Подписка на канал '%s' успешна!", channel)
	return nil
}

// handleMessage обрабатывает полученное сообщение
func (nc *NATSClient) handleMessage(msg *stan.Msg) {
	log.Printf("Получено сообщение из NATS (Sequence: %d)", msg.Sequence)

	// Валидация: проверяем, что это валидный JSON
	var order Order
	if err := json.Unmarshal(msg.Data, &order); err != nil {
		log.Printf("Ошибка парсинга JSON: %v. Данные: %s", err, string(msg.Data))
		msg.Ack()
		return
	}

	// Валидация: проверяем обязательные поля
	if order.OrderUID == "" {
		log.Printf("Ошибка валидации: отсутствует order_uid. Данные: %s", string(msg.Data))
		msg.Ack()
		return
	}

	// Сохраняем в базу данных
	if err := nc.db.SaveOrder(&order); err != nil {
		log.Printf("Ошибка сохранения заказа в БД: %v", err)
		// НЕ подтверждаем сообщение, чтобы попробовать обработать его снова
		return
	}

	// Сохраняем в кэш
	nc.cache.Set(order.OrderUID, &order)
	log.Printf("Заказ %s успешно обработан и сохранен", order.OrderUID)

	// Подтверждаем обработку сообщения
	msg.Ack()
}

// Close закрывает подключение к NATS Streaming
func (nc *NATSClient) Close() error {
	if nc.sub != nil {
		if err := nc.sub.Unsubscribe(); err != nil {
			log.Printf("Ошибка при отписке: %v", err)
		}
	}

	if nc.conn != nil {
		// Даем время на обработку последних сообщений
		time.Sleep(100 * time.Millisecond)
		return nc.conn.Close()
	}

	return nil
}
