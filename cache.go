package main

import (
    "sync"
)

type OrderCache struct {
    mu     sync.RWMutex
    orders map[string]*Order
}

func NewOrderCache() *OrderCache {
    return &OrderCache{
        orders: make(map[string]*Order),
    }
}

func (c *OrderCache) Set(orderUID string, order *Order) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.orders[orderUID] = order
}

func (c *OrderCache) Get(orderUID string) (*Order, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    order, exists := c.orders[orderUID]
    return order, exists
}

func (c *OrderCache) Delete(orderUID string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.orders, orderUID)
}

func (c *OrderCache) GetAll() []*Order {
    c.mu.RLock()
    defer c.mu.RUnlock()
    orders := make([]*Order, 0, len(c.orders))
    for _, order := range c.orders {
        orders = append(orders, order)
    }
    return orders
}

func (c *OrderCache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.orders = make(map[string]*Order)
}