package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	conn *pgx.Conn
}

func NewDB(dbURL string) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	db := &DB{conn: conn}
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось проверить подключение к базе данных: %w", err)
	}

	log.Println("Подключение к PostgreSQL успешно!")
	return db, nil
}

func (db *DB) Ping() error {
	return db.conn.Ping(context.Background())
}

func (db *DB) Close() error {
	return db.conn.Close(context.Background())
}

func (db *DB) SaveOrder(order *Order) error {
    ctx := context.Background()
    
    // Начинаем транзакцию
    tx, err := db.conn.Begin(ctx)
    if err != nil {
        return fmt.Errorf("ошибка начала транзакции: %w", err)
    }
    defer tx.Rollback(ctx)

    log.Printf("Сохраняем заказ %s...", order.OrderUID)

    _, err = tx.Exec(ctx, `
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            track_number = $2, entry = $3, locale = $4, internal_signature = $5,
            customer_id = $6, delivery_service = $7, shardkey = $8, sm_id = $9,
            date_created = $10, oof_shard = $11`,
        order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
        order.InternalSignature, order.CustomerID, order.DeliveryService,
        order.Shardkey, order.SMID, order.DateCreated, order.OOFShard)
    if err != nil {
        return fmt.Errorf("ошибка сохранения заказа: %w", err)
    }
    log.Printf("Основная информация о заказе сохранена")

    _, err = tx.Exec(ctx, `
        INSERT INTO delivery (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO UPDATE SET
            name = $2, phone = $3, zip = $4, city = $5,
            address = $6, region = $7, email = $8`,
        order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
        order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
        order.Delivery.Region, order.Delivery.Email)
    if err != nil {
        return fmt.Errorf("ошибка сохранения информации о доставке: %w", err)
    }
    log.Printf("Информация о доставке сохранена")

    _, err = tx.Exec(ctx, `
        INSERT INTO payment (
            order_uid, request_id, currency, provider, amount,
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (order_uid) DO UPDATE SET
            request_id = $2, currency = $3, provider = $4, amount = $5,
            payment_dt = $6, bank = $7, delivery_cost = $8, goods_total = $9,
            custom_fee = $10`,
        order.OrderUID, order.Payment.RequestID, order.Payment.Currency,
        order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt,
        order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal,
        order.Payment.CustomFee)
    if err != nil {
        return fmt.Errorf("ошибка сохранения информации об оплате: %w", err)
    }
    log.Printf("Информация об оплате сохранена")

    _, err = tx.Exec(ctx, `DELETE FROM items WHERE order_uid = $1`, order.OrderUID)
    if err != nil {
        return fmt.Errorf("ошибка удаления старых товаров: %w", err)
    }

    for _, item := range order.Items {
        _, err = tx.Exec(ctx, `
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid,
                name, sale, size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
            order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
            item.RID, item.Name, item.Sale, item.Size, item.TotalPrice,
            item.NMID, item.Brand, item.Status)
        if err != nil {
            return fmt.Errorf("ошибка сохранения товара: %w", err)
        }
    }
    log.Printf("Сохранено %d товаров", len(order.Items))

    // Подтверждаем транзакцию
    if err = tx.Commit(ctx); err != nil {
        return fmt.Errorf("ошибка подтверждения транзакции: %w", err)
    }
    log.Printf("Заказ %s успешно сохранен в БД", order.OrderUID)
    return nil
}

func (db *DB) GetOrder(orderUID string) (*Order, error) {
    ctx := context.Background()
    
    var exists bool
    err := db.conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)", orderUID).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("ошибка проверки существования заказа: %w", err)
    }
    
    if !exists {
        log.Printf("Заказ с ID %s не существует в таблице orders", orderUID)
        return nil, fmt.Errorf("заказ с ID %s не найден", orderUID)
    }
    
    log.Printf("Заказ %s существует в базе, продолжаем загрузку...", orderUID)

    orderQuery := `
        SELECT order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1
    `
    
    var order Order
    err = db.conn.QueryRow(ctx, orderQuery, orderUID).Scan(
        &order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
        &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
        &order.Shardkey, &order.SMID, &order.DateCreated, &order.OOFShard)
    if err != nil {
        return nil, fmt.Errorf("ошибка при получении заказа: %w", err)
    }

    var paymentExists bool
    err = db.conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM payment WHERE order_uid = $1)", orderUID).Scan(&paymentExists)
    if err != nil {
        return nil, fmt.Errorf("ошибка проверки payment: %w", err)
    }
    log.Printf("Запись в payment для заказа %s: %v", orderUID, paymentExists)

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

    paymentQuery := `
        SELECT request_id, currency, provider, amount, payment_dt,
            bank, delivery_cost, goods_total, custom_fee
        FROM payment WHERE order_uid = $1
    `
    
    err = db.conn.QueryRow(ctx, paymentQuery, orderUID).Scan(
        &order.Payment.RequestID, &order.Payment.Currency,
        &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
        &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
        &order.Payment.CustomFee)
    if err != nil {
        if err == pgx.ErrNoRows {
            log.Printf("Информация об оплате для заказа %s не найдена, используем пустые значения", orderUID)
        } else {
            return nil, fmt.Errorf("ошибка при получении информации об оплате: %w", err)
        }
    }

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

func (db *DB) GetAllOrders() ([]*Order, error) {
	ctx := context.Background()
	
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
