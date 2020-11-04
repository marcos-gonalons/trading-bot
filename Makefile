build:
	docker build ./ -t trading-bot-image \
		--build-arg app_env=dev \
		--build-arg app_root_folder=/go/src/TradingBot/src/ \
		--build-arg broker_username=${broker_username} \
		--build-arg broker_password=${broker_password}

chmod:
	docker-compose exec trading-bot-container bash -c "chmod -R 777 /go/src/TradingBot/"

glide-install:
	docker-compose exec trading-bot-container bash -c "cd /go/src/TradingBot && glide install"

glide-update:
	docker-compose exec trading-bot-container bash -c "cd /go/src/TradingBot && glide update"

glide-get:
	docker-compose exec trading-bot-container bash -c "cd /go/src/TradingBot && glide get ${package}"

restart: stop up
 
ssh:
	docker-compose exec trading-bot-container bash

stop:
	docker-compose stop

up:
	docker-compose up
