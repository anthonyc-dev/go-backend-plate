.PHONY: build up down logs clean re

build:
	docker build -t rest-api:latest .

up:
	docker-compose up -d

down:
	docker-compose down

# logs:
# 	docker-compose logs -f

logs-app:
	docker-compose logs -f app

clean:
	docker-compose down -v
	docker rmi rest-api-dev:latest || true

re: down clean up

ps:
	docker-compose ps

shell:
	docker exec -it rest-api sh

dev-up:
	docker compose -f docker-compose.dev.yml up

dev-build:
	docker compose -f docker-compose.dev.yml up --build

dev-down:
	docker compose -f docker-compose.dev.yml down

prod:
	docker compose -f docker-compose.prod.yml up --build -d

prod-down:
	docker compose -f docker-compose.prod.yml down

logs:
	docker compose logs -f

restart:
	docker compose -f docker-compose.dev.yml up && docker compose -f docker-compose.dev.yml up

# Safe cleanup: prune all unused Docker objects and build cache
cleanup:
	docker system prune -a -f && docker builder prune -a -f
