FROM golang:1.19-alpine

WORKDIR /server

COPY . .
RUN apk add --update gcc musl-dev
RUN go mod download

RUN go build -o /cloud4students

EXPOSE 3000
CMD [ "/cloud4students" ]
