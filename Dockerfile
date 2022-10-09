FROM golang

ARG username
ENV USERNAME $username

ARG password
ENV PASSWORD $password

ARG account_id
ENV ACCOUNT_ID $account_id

ARG api_name
ENV API_NAME $api_name

COPY ./ /TradingBot
WORKDIR /TradingBot

RUN go install github.com/go-delve/delve/cmd/dlv@latest

RUN go get ./src

EXPOSE 2345
