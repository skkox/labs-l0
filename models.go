package main

import "time"

// Order представляет структуру заказа согласно заданию
type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SMID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OOFShard          string    `json:"oof_shard"`
}

// Delivery представляет информацию о доставке
type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

// Payment представляет информацию об оплате
type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

// Item представляет товар в заказе
type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NMID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

// OrderCache представляет кэш заказов в памяти
type OrderCache struct {
	orders map[string]*Order
}

// NewOrderCache создает новый кэш заказов
func NewOrderCache() *OrderCache {
	return &OrderCache{
		orders: make(map[string]*Order),
	}
}

// Set добавляет заказ в кэш
func (c *OrderCache) Set(orderUID string, order *Order) {
	c.orders[orderUID] = order
}

// Get получает заказ из кэша
func (c *OrderCache) Get(orderUID string) (*Order, bool) {
	order, exists := c.orders[orderUID]
	return order, exists
}

// GetAll возвращает все заказы из кэша
func (c *OrderCache) GetAll() map[string]*Order {
	return c.orders
}
