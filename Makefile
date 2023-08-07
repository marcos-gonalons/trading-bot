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

debug-command:
	docker-compose run --rm -p 127.0.0.1:2345:2345/tcp trading-bot dlv debug ./src/commands/${command} --headless --listen=:2345 --api-version=2 --log -- ${args}

run-command:
	docker-compose run --rm trading-bot go run ./src/commands/${command}/. ${args}

down:
	docker-compose down
	
ssh:
	docker-compose exec trading-bot bash

test: 
	docker-compose run --rm trading-bot go test -coverprofile .tests-coverage/raw.txt ./src/... || true
	docker-compose run --rm trading-bot go tool cover -html=.tests-coverage/raw.txt -o .tests-coverage/report.html

######

single:
	docker-compose run --rm trading-bot go run ./src/commands/marketReplay/. EUR/USD single ranges longs
	
combo:
	docker-compose run --rm trading-bot go run ./src/commands/marketReplay/. EUR/USD combo ranges longs

