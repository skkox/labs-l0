package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// DB представляет подключение к базе данных
type DB struct {
	conn *pgx.Conn
}

// NewDB создает новое подключение к базе данных
func NewDB(dbURL string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	db := &DB{conn: conn}
	
	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось проверить подключение к базе данных: %w", err)
	}

	log.Println("Подключение к PostgreSQL успешно!")
	return db, nil
}

// Ping проверяет подключение к базе данных
func (db *DB) Ping() error {
	return db.conn.Ping(context.Background())
}

// Close закрывает подключение к базе данных
func (db *DB) Close() error {
	return db.conn.Close(context.Background())
}

// SaveOrder сохраняет заказ в базе данных
func (db *DB) SaveOrder(order *Order) error {
	ctx := context.Background()
	
	// Начинаем транзакцию
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	// Сохраняем основной заказ
	query := `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			locale = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id = EXCLUDED.customer_id,
			delivery_service = EXCLUDED.delivery_service,
			shardkey = EXCLUDED.shardkey,
			sm_id = EXCLUDED.sm_id,
			date_created = EXCLUDED.date_created,
			oof_shard = EXCLUDED.oof_shard
	`
	
	_, err = tx.Exec(ctx, query,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SMID, order.DateCreated, order.OOFShard)
	if err != nil {
		return fmt.Errorf("не удалось сохранить заказ: %w", err)
	}

	// Сохраняем информацию о доставке
	deliveryQuery := `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO UPDATE SET
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			zip = EXCLUDED.zip,
			city = EXCLUDED.city,
			address = EXCLUDED.address,
			region = EXCLUDED.region,
			email = EXCLUDED.email
	`
	
	_, err = tx.Exec(ctx, deliveryQuery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("не удалось сохранить информацию о доставке: %w", err)
	}

	// Сохраняем информацию об оплате
	paymentQuery := `
		INSERT INTO payment (order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO UPDATE SET
			transaction = EXCLUDED.transaction,
			request_id = EXCLUDED.request_id,
			currency = EXCLUDED.currency,
			provider = EXCLUDED.provider,
			amount = EXCLUDED.amount,
			payment_dt = EXCLUDED.payment_dt,
			bank = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total = EXCLUDED.goods_total,
			custom_fee = EXCLUDED.custom_fee
	`
	
	_, err = tx.Exec(ctx, paymentQuery,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("не удалось сохранить информацию об оплате: %w", err)
	}

	// Удаляем старые товары
	_, err = tx.Exec(ctx, "DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return fmt.Errorf("не удалось удалить старые товары: %w", err)
	}

	// Сохраняем товары
	for _, item := range order.Items {
		itemQuery := `
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		
		_, err = tx.Exec(ctx, itemQuery,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NMID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("не удалось сохранить товар: %w", err)
		}
	}

	// Подтверждаем транзакцию
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("не удалось подтвердить транзакцию: %w", err)
	}

	log.Printf("Заказ %s успешно сохранен в базе данных", order.OrderUID)
	return nil
}

// GetOrder получает заказ по ID
func (db *DB) GetOrder(orderUID string) (*Order, error) {
	ctx := context.Background()
	
	// Получаем основной заказ
	orderQuery := `
		SELECT order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`
	
	var order Order
	err := db.conn.QueryRow(ctx, orderQuery, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SMID, &order.DateCreated, &order.OOFShard)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("заказ с ID %s не найден", orderUID)
		}
		return nil, fmt.Errorf("ошибка при получении заказа: %w", err)
	}

	// Получаем информацию о доставке
	deliveryQuery := `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1
	`
	
	err = db.conn.QueryRow(ctx, deliveryQuery, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
		&order.Delivery.Email)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении информации о доставке: %w", err)
	}

	// Получаем информацию об оплате
	paymentQuery := `
		SELECT transaction, request_id, currency, provider, amount, payment_dt,
			bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE order_uid = $1
	`
	
	err = db.conn.QueryRow(ctx, paymentQuery, orderUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении информации об оплате: %w", err)
	}

	// Получаем товары
	itemsQuery := `
		SELECT chrt_id, track_number, price, rid, name, sale, size,
			total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`
	
	rows, err := db.conn.Query(ctx, itemsQuery, orderUID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении товаров: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price,
			&item.RID, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NMID, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании товара: %w", err)
		}
		items = append(items, item)
	}
	
	order.Items = items

	return &order, nil
}

// GetAllOrders получает все заказы из базы данных
func (db *DB) GetAllOrders() ([]*Order, error) {
	ctx := context.Background()
	
	// Получаем все order_uid
	query := "SELECT order_uid FROM orders ORDER BY date_created DESC"
	rows, err := db.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка заказов: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании order_uid: %w", err)
		}
		
		order, err := db.GetOrder(orderUID)
		if err != nil {
			log.Printf("Ошибка при получении заказа %s: %v", orderUID, err)
			continue
		}
		
		orders = append(orders, order)
	}

	return orders, nil
}
