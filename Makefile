build:
	docker build ./ -t trading-bot \
		--build-arg username=${username} \
		--build-arg password=${password} \
		--build-arg account_id=${account_id}

dependencies: build

up:
	docker-compose up

down:
	docker-compose down

restart: down up

ssh:
	docker-compose exec trading-bot bash