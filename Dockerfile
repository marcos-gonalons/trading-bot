FROM golang

ARG app_env
ENV APP_ENV $app_env

ARG app_root_folder
ENV APP_ROOT_FOLDER $app_root_folder

ARG broker_username
ENV BROKER_USERNAME $broker_username

ARG broker_password
ENV BROKER_PASSWORD $broker_password

COPY ./src /go/src/TradingBot/src
WORKDIR /go/src/TradingBot/src

RUN curl https://glide.sh/get | sh

RUN go get github.com/derekparker/delve/cmd/dlv
RUN go get ./

EXPOSE 8081 2345

CMD ["dlv", "debug", "TradingBot/src", "--headless", "--listen=:2345", "--api-version=2", "--log"]