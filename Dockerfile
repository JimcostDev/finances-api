# Etapa de compilación
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o finances-api

# Etapa de producción
FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /build/finances-api .

EXPOSE 3000
CMD ["./finances-api"]