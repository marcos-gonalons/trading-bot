build:
	docker build ./ -t trading-bot --build-arg api_name=simulator

build-prod:
	docker build ./ -t trading-bot \
		--build-arg username=${username} \
		--build-arg password=${password} \
		--build-arg account_id=${account_id} \
		--build-arg api_name=${api_name}

up: down
	docker-compose up -d
	docker-compose exec --no-TTY trading-bot go run src/main.go 2> error_output.txt &

up-debug: down
	docker-compose up -d
	docker-compose exec trading-bot dlv debug ./src --headless --listen=:2345 --api-version=2 --log

down:
	docker-compose down
	
sanitize-csv:
	docker-compose exec trading-bot go run ./src/commands/csvSanitizer/main.go ./.candles-csv/

ssh:
	docker-compose exec trading-bot bash

test:
	docker-compose exec trading-bot go test ./src/...