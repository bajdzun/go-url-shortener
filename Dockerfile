FROM golang:1.21-alpine

RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/api/main.go"]
