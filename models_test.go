package main

import (
	"testing"
	"time"
)

func TestOrderCache(t *testing.T) {
	cache := NewOrderCache()
	
	// Создаем тестовый заказ
	testOrder := &Order{
		OrderUID:    "test-order-123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NMID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:             "en",
		InternalSignature: "",
		CustomerID:         "test",
		DeliveryService:    "meest",
		Shardkey:           "9",
		SMID:               99,
		DateCreated:        time.Now(),
		OOFShard:           "1",
	}
	
	// Тест добавления заказа в кэш
	cache.Set(testOrder.OrderUID, testOrder)
	
	// Тест получения заказа из кэша
	retrievedOrder, exists := cache.Get(testOrder.OrderUID)
	if !exists {
		t.Error("Заказ должен существовать в кэше")
	}
	
	if retrievedOrder.OrderUID != testOrder.OrderUID {
		t.Errorf("Ожидался OrderUID %s, получен %s", testOrder.OrderUID, retrievedOrder.OrderUID)
	}
	
	if retrievedOrder.Delivery.Name != testOrder.Delivery.Name {
		t.Errorf("Ожидалось имя %s, получено %s", testOrder.Delivery.Name, retrievedOrder.Delivery.Name)
	}
	
	if len(retrievedOrder.Items) != len(testOrder.Items) {
		t.Errorf("Ожидалось %d товаров, получено %d", len(testOrder.Items), len(retrievedOrder.Items))
	}
	
	// Тест получения несуществующего заказа
	_, exists = cache.Get("non-existent-order")
	if exists {
		t.Error("Несуществующий заказ не должен быть найден в кэше")
	}
	
	// Тест получения всех заказов
	allOrders := cache.GetAll()
	if len(allOrders) != 1 {
		t.Errorf("Ожидался 1 заказ в кэше, получено %d", len(allOrders))
	}
}
