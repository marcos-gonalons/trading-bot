FROM golang

ARG username
ENV USERNAME $username

ARG password
ENV PASSWORD $password

ARG account_id
ENV ACCOUNT_ID $account_id

ARG api_url
ENV API_URL $api_url

COPY ./ /TradingBot
WORKDIR /TradingBot

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go get ./src

EXPOSE 2345

CMD ["dlv", "debug", "./src", "--headless", "--listen=:2345", "--api-version=2", "--log"]
