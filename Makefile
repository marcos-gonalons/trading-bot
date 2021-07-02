build:
	docker build ./ -t trading-bot \
		--build-arg username=${username} \
		--build-arg password=${password} \
		--build-arg account_id=${account_id} \
		--build-arg api_url=${api_url}

up:
	docker-compose up

down:
	docker-compose down

restart: down up

ssh:
	docker-compose exec trading-bot bash

test:
	docker-compose exec trading-bot go test ./src/...