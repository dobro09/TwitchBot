include .env
export

service-run:
	go run ./cmd/twitchbotapp

migrate-up:
	migrate -path db/migrations -database ${DATABASE_URL} up

migrate-down:
	migrate -path db/migrations -database ${DATABASE_URL} down

compose-up:
	docker compose up -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f bot

compose-build:
	docker compose build

onlydb-up:
	docker compose up -d db

onlydb-down:
	docker compose down db

updaterunbot:
	docker compose build --no-cache bot
	docker compose up -d --force-recreate

inmemory-runbot:
	go run ./cmd/twitchbotapp
# 	для инмемори запуска: make inmemory-runbot STORE_TYPE=memory