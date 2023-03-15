FROM golang:1.19-alpine

WORKDIR /server

COPY . .

RUN apk add --update gcc musl-dev
RUN go mod tidy

CMD  go run main.go

EXPOSE 3000