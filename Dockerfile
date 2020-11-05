FROM golang

ARG app_env
ENV APP_ENV $app_env

ARG app_root_folder
ENV APP_ROOT_FOLDER $app_root_folder

ARG ibroker_api_url
ENV IBROKER_API_URL $ibroker_api_url

COPY ./src /go/src/TradingBot/src
WORKDIR /go/src/TradingBot/src

RUN curl https://glide.sh/get | sh

RUN go get github.com/derekparker/delve/cmd/dlv
RUN go get ./

EXPOSE 8081 2345

CMD ["dlv", "debug", "TradingBot/src", "--headless", "--listen=:2345", "--api-version=2", "--log"]