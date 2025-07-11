FROM golang:1.23.8-alpine

WORKDIR /app

COPY go.sum go.mod ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main", "-config", "config.yaml"]
