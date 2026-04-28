FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["/app/server"]
