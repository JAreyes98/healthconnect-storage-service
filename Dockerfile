# Stage 1: Build
FROM golang:1.23-alpine AS build

# Instalamos git por si alguna dependencia de go mod lo requiere
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Run
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app/main .

RUN mkdir -p uploads && chmod 777 uploads

EXPOSE 8082

CMD ["./main"]