#Makefile для Concurrent API Fetcher

BINARY_NAME = workerpool
IMAGE_NAME = api-fetcher
CONTAINER_NAME = api-fetcher-container
PORT = 8080
COVERAGE_DIR = coverage

.PHONY: all build docker_build docker_run docker_stop docker_test docker_clean test test_sh coverage clean style help

all: build

build:
	go build -o $(BINARY_NAME) .

docker_build:
	docker build -t $(IMAGE_NAME) .

docker_run: docker_build
	docker run -d -p $(PORT):$(PORT) --name $(CONTAINER_NAME) $(IMAGE_NAME)
	@echo "Контейнер запущен. Для просмотра логов: make docker_logs"

docker_stop:
	docker stop $(CONTAINER_NAME) 2>/dev/null || true
	docker rm $(CONTAINER_NAME) 2>/dev/null || true
	@echo "Контейнер остановлен и удален"

docker_logs:
	docker logs -f $(CONTAINER_NAME)

#(запускает тесты внутри контейнера)
docker_test: docker_build
	@echo "Запуск интеграционных тестов в Docker..."
	docker run --rm -p $(PORT):$(PORT) --name $(CONTAINER_NAME)-test $(IMAGE_NAME) &
	@sleep 5
	@if curl -s http://localhost:$(PORT)/ > /dev/null; then \
		./test.sh; \
		docker stop $(CONTAINER_NAME)-test 2>/dev/null || true; \
	else \
		echo "Ошибка: контейнер не запустился"; \
		docker stop $(CONTAINER_NAME)-test 2>/dev/null || true; \
		exit 1; \
	fi

#запуск тестов на собранном бинарнике
test_local: build
	@echo "Запуск тестов на локальном бинарнике..."
	./$(BINARY_NAME) &
	@sleep 2
	./test.sh
	@pkill -f $(BINARY_NAME) 2>/dev/null || true

docker_clean: docker_stop
	docker rmi $(IMAGE_NAME) 2>/dev/null || true
	@echo "Docker образ удален"

test:
	go test -v -timeout=30s ./...

# (требует запущенного сервера)
test_sh:
	@if ! curl -s http://localhost:$(PORT)/ > /dev/null; then \
		echo "Ошибка: сервер не запущен на порту $(PORT)"; \
		echo "Запустите сервер: make docker_run или make build && ./$(BINARY_NAME)"; \
		exit 1; \
	fi
	./test.sh

coverage:
	mkdir -p $(COVERAGE_DIR)
	go clean -testcache
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/report.html
	@echo "Отчет о покрытии создан: $(COVERAGE_DIR)/report.html"

clean:
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_DIR)/coverage.out $(COVERAGE_DIR)/report.html 2>/dev/null || true
	rm -rf $(COVERAGE_DIR) 2>/dev/null || true
	rm -rf .vscode 2>/dev/null || true

style:
	gofmt -w *.go
	@echo "Код отформатирован"

help:
	@echo "Доступные команды:"
	@echo "  make all              - Собрать Go-бинарник (цель по умолчанию)"
	@echo "  make build            - Собрать Go-бинарник"
	@echo "  make docker_build     - Собрать Docker-образ"
	@echo "  make docker_run       - Запустить контейнер в фоне"
	@echo "  make docker_stop      - Остановить контейнер"
	@echo "  make docker_logs      - Просмотр логов контейнера"
	@echo "  make docker_test      - Запустить интеграционные тесты в Docker"
	@echo "  make test_local       - Запустить тесты на локальном бинарнике"
	@echo "  make docker_clean     - Полная очистка Docker (контейнер + образ)"
	@echo "  make test             - Запустить unit-тесты Go"
	@echo "  make test_sh          - Запустить интеграционные тесты (требует сервер)"
	@echo "  make coverage         - Запустить тестирование с покрытием"
	@echo "  make clean            - Удалить собранные файлы"
	@echo "  make style            - Форматирование кода"
	@echo "  make help             - Показать эту справку"