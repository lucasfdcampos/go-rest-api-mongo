.PHONY: help run build test clean docker-up docker-down deps

help: ## Mostra esta mensagem de ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Instala as dependências do projeto
	go mod download
	go mod tidy

run: ## Executa a aplicação
	go run cmd/main.go

build: ## Compila a aplicação
	go build -o bin/api cmd/main.go

test: ## Executa os testes
	go test -v ./...

clean: ## Remove arquivos compilados
	rm -rf bin/

docker-up: ## Inicia os containers (MongoDB e Kafka)
	docker compose up -d

docker-down: ## Para os containers
	docker compose down

docker-logs: ## Mostra os logs dos containers
	docker compose logs -f

docker-restart: ## Reinicia os containers
	docker compose restart

create-topics: ## Cria os tópicos Kafka necessários
	docker exec -it go-api-kafka kafka-topics --create --topic user-registration --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 || true
	docker exec -it go-api-kafka kafka-topics --create --topic user-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1 || true

list-topics: ## Lista os tópicos Kafka
	docker exec -it go-api-kafka kafka-topics --list --bootstrap-server localhost:9092
