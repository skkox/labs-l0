package main

import (
    "encoding/json"
    "log"
    "time"

    "github.com/nats-io/stan.go"
)

// Копируем структуры из основного пакета
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

type Delivery struct {
    Name    string `json:"name"`
    Phone   string `json:"phone"`
    Zip     string `json:"zip"`
    City    string `json:"city"`
    Address string `json:"address"`
    Region  string `json:"region"`
    Email   string `json:"email"`
}

type Payment struct {
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

func main() {
    sc, err := stan.Connect("test-cluster", "test-publisher", stan.NatsURL("nats://localhost:4222"))
    if err != nil {
        log.Fatalf("Ошибка подключения к NATS: %v", err)
    }
    defer sc.Close()

    log.Println("Подключено к NATS Streaming")

    // Создаем тестовый заказ
    testOrder := Order{
        OrderUID:    "b583feb7b2b84b6test",
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
        Locale:            "en",
        InternalSignature: "",
        CustomerID:        "test",
        DeliveryService:   "meest",
        Shardkey:          "9",
        SMID:              99,
        DateCreated:       time.Now(),
        OOFShard:          "1",
    }

    data, err := json.Marshal(testOrder)
    if err != nil {
        log.Fatalf("Ошибка сериализации JSON: %v", err)
    }

    err = sc.Publish("orders", data)
    if err != nil {
        log.Fatalf("Ошибка публикации сообщения: %v", err)
    }

    log.Printf("Заказ %s успешно отправлен в NATS", testOrder.OrderUID)
    
    time.Sleep(1 * time.Second)
}