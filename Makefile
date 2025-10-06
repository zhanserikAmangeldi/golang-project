.PHONY: help docker-up docker-down user-run user-deps test-user clean

help: ## Показать эту помощь
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

docker-up: ## Поднять все сервисы (PostgreSQL, Redis, RabbitMQ)
	docker-compose up -d
	@echo "Ждем пока базы поднимутся..."
	@sleep 5
	@echo "Сервисы запущены!"
	@echo "PostgreSQL: localhost:54321"
	@echo "Redis: localhost:6379"
	@echo "RabbitMQ Management: http://localhost:15672 (admin/admin123)"

docker-down: ## Остановить все сервисы
	docker-compose down

docker-logs: ## Показать логи
	docker-compose logs -f

docker-clean: ## Удалить все volumes
	docker-compose down -v

user-deps: ## Установить зависимости для user-service
	cd user-service && go mod download && go mod tidy

user-run: ## Запустить user-service локально
	cd user-service && go run cmd/api/main.go

test-health: ## Проверить health endpoint
	@curl -s http://localhost:8081/health | json_pp || echo "Сервис не запущен"