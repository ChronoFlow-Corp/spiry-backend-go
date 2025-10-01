FROM golang:1.24.5 as app

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CONFIG_PATH=config.yaml

RUN go build -o main cmd/spiry/main.go

EXPOSE 8080

CMD ["./main", "--steps", "4"]