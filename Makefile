.PHONY: all clean server-openwrt client-windows client-linux client-macos client-macos-universal deps test help

# Версия проекта
VERSION := 1.0.0
LDFLAGS := -s -w -X main.version=$(VERSION)

# Директории
BIN_DIR := bin
CMD_SERVER := cmd/server
CMD_CLIENT := cmd/client

# Цвета для вывода
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

all: deps server-openwrt server-linux server-windows client-windows client-linux client-macos
	@echo "$(GREEN)✓ Все бинарники скомпилированы успешно!$(NC)"

help:
	@echo "OpenWRT Clipboard - Makefile"
	@echo ""
	@echo "Доступные команды:"
	@echo "  make all                  - Собрать все бинарники"
	@echo "  make deps                 - Установить зависимости"
	@echo "  make server-openwrt       - Собрать сервер для OpenWRT (ARM64)"
	@echo "  make server-linux         - Собрать сервер для Linux (AMD64)"
	@echo "  make server-windows       - Собрать сервер для Windows (AMD64)"
	@echo "  make client-windows       - Собрать клиент для Windows"
	@echo "  make client-linux         - Собрать клиент для Linux"
	@echo "  make client-macos         - Собрать клиент для macOS (ARM64)"
	@echo "  make client-macos-universal - Собрать универсальный клиент для macOS"
	@echo "  make clean                - Удалить собранные бинарники"
	@echo "  make test                 - Запустить тесты"
	@echo ""

deps:
	@echo "$(YELLOW)Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Зависимости установлены$(NC)"

# Сервер для OpenWRT (ARM64)
server-openwrt: deps
	@echo "$(YELLOW)Компиляция сервера для OpenWRT (ARM64)...$(NC)"
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-server-openwrt \
		./$(CMD_SERVER)
	@ls -lh $(BIN_DIR)/clipboard-server-openwrt
	@echo "$(GREEN)✓ Сервер для OpenWRT собран: $(BIN_DIR)/clipboard-server-openwrt$(NC)"

# Сервер для Linux (AMD64)
server-linux: deps
	@echo "$(YELLOW)Компиляция сервера для Linux (AMD64)...$(NC)"
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-server-linux \
		./$(CMD_SERVER)
	@ls -lh $(BIN_DIR)/clipboard-server-linux
	@echo "$(GREEN)✓ Сервер для Linux собран: $(BIN_DIR)/clipboard-server-linux$(NC)"

# Сервер для Windows (AMD64)
server-windows: deps
	@echo "$(YELLOW)Компиляция сервера для Windows (AMD64)...$(NC)"
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-server-windows.exe \
		./$(CMD_SERVER)
	@ls -lh $(BIN_DIR)/clipboard-server-windows.exe
	@echo "$(GREEN)✓ Сервер для Windows собран: $(BIN_DIR)/clipboard-server-windows.exe$(NC)"

# Клиент для Windows (AMD64)
client-windows: deps
	@echo "$(YELLOW)Компиляция клиента для Windows (x64)...$(NC)"
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS) -H=windowsgui" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-client-windows.exe \
		./$(CMD_CLIENT)
	@ls -lh $(BIN_DIR)/clipboard-client-windows.exe
	@echo "$(GREEN)✓ Клиент для Windows собран: $(BIN_DIR)/clipboard-client-windows.exe$(NC)"

# Клиент для Linux (AMD64)
client-linux: deps
	@echo "$(YELLOW)Компиляция клиента для Linux (x64)...$(NC)"
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-client-linux \
		./$(CMD_CLIENT)
	@ls -lh $(BIN_DIR)/clipboard-client-linux
	@echo "$(GREEN)✓ Клиент для Linux собран: $(BIN_DIR)/clipboard-client-linux$(NC)"

# Клиент для macOS (ARM64 - Apple Silicon)
client-macos: deps
	@echo "$(YELLOW)Компиляция клиента для macOS (ARM64)...$(NC)"
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-client-macos \
		./$(CMD_CLIENT)
	@ls -lh $(BIN_DIR)/clipboard-client-macos
	@echo "$(GREEN)✓ Клиент для macOS собран: $(BIN_DIR)/clipboard-client-macos$(NC)"

# Универсальный клиент для macOS (ARM64 + AMD64)
client-macos-universal: deps
	@echo "$(YELLOW)Компиляция универсального клиента для macOS...$(NC)"
	@mkdir -p $(BIN_DIR)
	
	@echo "  - Сборка ARM64..."
	@GOOS=darwin GOARCH=arm64 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-client-macos-arm64 \
		./$(CMD_CLIENT)
	
	@echo "  - Сборка AMD64..."
	@GOOS=darwin GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-trimpath \
		-o $(BIN_DIR)/clipboard-client-macos-amd64 \
		./$(CMD_CLIENT)
	
	@echo "  - Создание универсального бинарника..."
	@lipo -create -output $(BIN_DIR)/clipboard-client-macos-universal \
		$(BIN_DIR)/clipboard-client-macos-arm64 \
		$(BIN_DIR)/clipboard-client-macos-amd64
	
	@rm $(BIN_DIR)/clipboard-client-macos-arm64 $(BIN_DIR)/clipboard-client-macos-amd64
	@ls -lh $(BIN_DIR)/clipboard-client-macos-universal
	@echo "$(GREEN)✓ Универсальный клиент для macOS собран: $(BIN_DIR)/clipboard-client-macos-universal$(NC)"

# Быстрая локальная сборка (для текущей платформы)
local-server:
	@echo "$(YELLOW)Компиляция сервера для текущей платформы...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/clipboard-server ./$(CMD_SERVER)
	@echo "$(GREEN)✓ Сервер собран: $(BIN_DIR)/clipboard-server$(NC)"

local-client:
	@echo "$(YELLOW)Компиляция клиента для текущей платформы...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/clipboard-client ./$(CMD_CLIENT)
	@echo "$(GREEN)✓ Клиент собран: $(BIN_DIR)/clipboard-client$(NC)"

# Тестирование
test:
	@echo "$(YELLOW)Запуск тестов...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)✓ Тесты пройдены$(NC)"

# Очистка
clean:
	@echo "$(YELLOW)Удаление бинарников...$(NC)"
	@rm -rf $(BIN_DIR)
	@echo "$(GREEN)✓ Бинарники удалены$(NC)"

# Информация о размерах бинарников
sizes:
	@echo "Размеры скомпилированных бинарников:"
	@ls -lh $(BIN_DIR)/ 2>/dev/null || echo "Нет скомпилированных бинарников"

# Загрузка сервера на роутер (требует настройки переменных)
ROUTER_IP ?= 192.168.1.1
ROUTER_USER ?= root

deploy-server: server-openwrt
	@echo "$(YELLOW)Загрузка сервера на роутер $(ROUTER_IP)...$(NC)"
	@scp $(BIN_DIR)/clipboard-server-openwrt $(ROUTER_USER)@$(ROUTER_IP):/tmp/clipboard-server
	@ssh $(ROUTER_USER)@$(ROUTER_IP) "chmod +x /tmp/clipboard-server"
	@echo "$(GREEN)✓ Сервер загружен на роутер$(NC)"
	@echo "Для запуска: ssh $(ROUTER_USER)@$(ROUTER_IP) '/tmp/clipboard-server'"

# Запуск сервера локально для тестирования
run-server: local-server
	@echo "$(YELLOW)Запуск сервера локально...$(NC)"
	@$(BIN_DIR)/clipboard-server

# Запуск клиента локально для тестирования
run-client: local-client
	@echo "$(YELLOW)Запуск клиента локально...$(NC)"
	@$(BIN_DIR)/clipboard-client -server ws://localhost:9090/ws
