FROM golang:1.19-alpine

WORKDIR /server

COPY . .

RUN apk add --no-cache make
RUN apk add --update gcc musl-dev
RUN go mod tidy

CMD make run
EXPOSE 3000
