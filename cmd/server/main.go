package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denisuvarov/openwrt-clipboard/internal/server"
)

var (
	addr    = flag.String("addr", ":8080", "HTTP server address")
	version = "dev" // Будет заменено при сборке через -ldflags
)

func main() {
	flag.Parse()

	log.Printf("OpenWRT Clipboard Server %s", version)
	log.Printf("Starting server on %s", *addr)

	// Создаем Hub
	hub := server.NewHub()
	go hub.Run()

	// Настраиваем HTTP роуты
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub, w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","clients":%d,"version":"%s"}`, hub.ClientCount(), version)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>OpenWRT Clipboard Server</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { color: #333; margin-top: 0; }
        .status { color: #28a745; font-weight: bold; }
        .info { 
            background: #e3f2fd; 
            padding: 15px; 
            border-radius: 5px;
            margin: 20px 0;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        .stat-card {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            border-left: 4px solid #007bff;
        }
        .stat-label { font-size: 12px; color: #666; }
        .stat-value { font-size: 24px; font-weight: bold; color: #333; }
        code {
            background: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔄 OpenWRT Clipboard Server</h1>
        <p>Статус: <span class="status">РАБОТАЕТ</span></p>
        
        <div class="info">
            <strong>ℹ️ Информация:</strong><br>
            Сервер синхронизации буфера обмена для локальной сети.<br>
            WebSocket эндпоинт: <code>ws://%s/ws</code>
        </div>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-label">Подключенных клиентов</div>
                <div class="stat-value" id="clients">%d</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Версия</div>
                <div class="stat-value">%s</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Время работы</div>
                <div class="stat-value" id="uptime">-</div>
            </div>
        </div>

        <h3>Endpoints:</h3>
        <ul>
            <li><code>/ws</code> - WebSocket endpoint для клиентов</li>
            <li><code>/health</code> - Health check (JSON)</li>
            <li><code>/</code> - Эта страница</li>
        </ul>
    </div>
    
    <script>
        const startTime = Date.now();
        
        function updateUptime() {
            const uptime = Math.floor((Date.now() - startTime) / 1000);
            const hours = Math.floor(uptime / 3600);
            const minutes = Math.floor((uptime %% 3600) / 60);
            const seconds = uptime %% 60;
            document.getElementById('uptime').textContent = 
                hours.toString().padStart(2, '0') + ':' +
                minutes.toString().padStart(2, '0') + ':' +
                seconds.toString().padStart(2, '0');
        }
        
        function updateClients() {
            fetch('/health')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('clients').textContent = data.clients;
                })
                .catch(e => console.error(e));
        }
        
        setInterval(updateUptime, 1000);
        setInterval(updateClients, 5000);
        updateUptime();
        updateClients();
    </script>
</body>
</html>`, r.Host, hub.ClientCount(), version)
	})

	// HTTP сервер
	httpServer := &http.Server{
		Addr:         *addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		if err := httpServer.Close(); err != nil {
			log.Printf("HTTP server close error: %v", err)
		}
	}()

	// Запускаем сервер
	log.Printf("Server is ready. Open http://%s in browser", *addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}

	log.Println("Server stopped")
}
