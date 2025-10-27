package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	cache  *OrderCache
	router *mux.Router
}

func NewServer(cache *OrderCache) *Server {
	s := &Server{
		cache:  cache,
		router: mux.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleIndex()).Methods("GET")
	s.router.HandleFunc("/api/order/{id}", s.handleGetOrder()).Methods("GET")
}

func (s *Server) handleIndex() http.HandlerFunc {
	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Сервис заказов L0</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #302620 0%, #1a1410 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: #B9AB99;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.6);
            padding: 40px;
            max-width: 800px;
            width: 100%;
        }
        h1 {
            color: #6D2A22;
            margin-bottom: 10px;
            font-size: 2.5em;
            font-weight: bold;
        }
        .subtitle {
            color: #302620;
            margin-bottom: 30px;
            font-size: 1.1em;
        }
        .search-box {
            display: flex;
            gap: 10px;
            margin-bottom: 30px;
        }
        input[type="text"] {
            flex: 1;
            padding: 15px;
            border: 2px solid #775F4E;
            border-radius: 10px;
            font-size: 16px;
            background: white;
            transition: border-color 0.3s;
        }
        input[type="text"]:focus {
            outline: none;
            border-color: #6D2A22;
        }
        button {
            padding: 15px 30px;
            background: linear-gradient(135deg, #6D2A22 0%, #302620 100%);
            color: #B9AB99;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: bold;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(109, 42, 34, 0.5);
        }
        button:active {
            transform: translateY(0);
        }
        #result {
            margin-top: 20px;
        }
        .order-card {
            background: white;
            border-radius: 10px;
            padding: 20px;
            margin-top: 20px;
            border: 2px solid #775F4E;
        }
        .order-section {
            margin-bottom: 20px;
        }
        .order-section h3 {
            color: #6D2A22;
            margin-bottom: 10px;
            font-size: 1.3em;
            padding-bottom: 8px;
            border-bottom: 2px solid #B9AB99;
        }
        .order-field {
            display: grid;
            grid-template-columns: 200px 1fr;
            gap: 10px;
            padding: 8px 0;
            border-bottom: 1px solid #948C7A;
        }
        .order-field:last-child {
            border-bottom: none;
        }
        .field-label {
            font-weight: bold;
            color: #775F4E;
        }
        .field-value {
            color: #302620;
        }
        .error {
            background: #6D2A22;
            color: #B9AB99;
            padding: 15px;
            border-radius: 10px;
            border-left: 4px solid #302620;
        }
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
        }
        .items-table th {
            background: #6D2A22;
            color: #B9AB99;
            padding: 10px;
            text-align: left;
        }
        .items-table td {
            padding: 10px;
            border-bottom: 1px solid #948C7A;
            color: #302620;
        }
        .items-table tr:last-child td {
            border-bottom: none;
        }
        .items-table tr:hover {
            background: #B9AB99;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Сервис заказов L0</h1>
        <p class="subtitle">Поиск заказа по ID</p>
        
        <div class="search-box">
            <input type="text" id="orderId" placeholder="Введите ID заказа (например: b563feb7b2b84b6test)">
            <button onclick="searchOrder()">Найти</button>
        </div>
        
        <div id="result"></div>
    </div>

    <script>
        function searchOrder() {
            const orderId = document.getElementById('orderId').value.trim();
            if (!orderId) {
                document.getElementById('result').innerHTML = '<div class="error">Пожалуйста, введите ID заказа</div>';
                return;
            }

            fetch('/api/order/' + encodeURIComponent(orderId))
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Заказ не найден');
                    }
                    return response.json();
                })
                .then(order => {
                    displayOrder(order);
                })
                .catch(error => {
                    document.getElementById('result').innerHTML = '<div class="error">' + error.message + '</div>';
                });
        }

        function displayOrder(order) {
            let html = '<div class="order-card">';
            
            // Основная информация
            html += '<div class="order-section">';
            html += '<h3>Основная информация</h3>';
            html += '<div class="order-field"><span class="field-label">ID заказа:</span><span class="field-value">' + order.order_uid + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Трек-номер:</span><span class="field-value">' + order.track_number + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Дата создания:</span><span class="field-value">' + new Date(order.date_created).toLocaleString('ru-RU') + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Служба доставки:</span><span class="field-value">' + order.delivery_service + '</span></div>';
            html += '</div>';

            // Доставка
            html += '<div class="order-section">';
            html += '<h3>Доставка</h3>';
            html += '<div class="order-field"><span class="field-label">Получатель:</span><span class="field-value">' + order.delivery.name + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Телефон:</span><span class="field-value">' + order.delivery.phone + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Email:</span><span class="field-value">' + order.delivery.email + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Адрес:</span><span class="field-value">' + order.delivery.address + ', ' + order.delivery.city + ', ' + order.delivery.zip + '</span></div>';
            html += '</div>';

            // Оплата
            html += '<div class="order-section">';
            html += '<h3>Оплата</h3>';
            html += '<div class="order-field"><span class="field-label">Транзакция:</span><span class="field-value">' + order.payment.transaction + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Сумма:</span><span class="field-value">' + order.payment.amount + ' ' + order.payment.currency + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Провайдер:</span><span class="field-value">' + order.payment.provider + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Банк:</span><span class="field-value">' + order.payment.bank + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Стоимость доставки:</span><span class="field-value">' + order.payment.delivery_cost + ' ' + order.payment.currency + '</span></div>';
            html += '<div class="order-field"><span class="field-label">Стоимость товаров:</span><span class="field-value">' + order.payment.goods_total + ' ' + order.payment.currency + '</span></div>';
            html += '</div>';

            // Товары
            if (order.items && order.items.length > 0) {
                html += '<div class="order-section">';
                html += '<h3>Товары</h3>';
                html += '<table class="items-table">';
                html += '<tr><th>Название</th><th>Бренд</th><th>Размер</th><th>Цена</th><th>Скидка</th><th>Итого</th></tr>';
                order.items.forEach(item => {
                    html += '<tr>';
                    html += '<td>' + item.name + '</td>';
                    html += '<td>' + item.brand + '</td>';
                    html += '<td>' + item.size + '</td>';
                    html += '<td>' + item.price + '</td>';
                    html += '<td>' + item.sale + '%</td>';
                    html += '<td>' + item.total_price + '</td>';
                    html += '</tr>';
                });
                html += '</table>';
                html += '</div>';
            }

            html += '</div>';
            document.getElementById('result').innerHTML = html;
        }

        // Поиск по Enter
        document.getElementById('orderId').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                searchOrder();
            }
        });
    </script>
</body>
</html>
`
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(tmpl))
	}
}

func (s *Server) handleGetOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		orderID := vars["id"]

		order, exists := s.cache.Get(orderID)
		if !exists {
			http.Error(w, "Заказ не найден", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}

func (s *Server) Start(addr string) error {
	log.Printf("HTTP-сервер запущен на %s", addr)
	return http.ListenAndServe(addr, s.router)
}
