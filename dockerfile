# syntax = docker/dockerfile:1.5
########################
# Etapa de compilación
########################
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates

# Necesario para compilar binario estático
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

WORKDIR /build

# Primero copia solo mod para acelerar la cache
COPY go.mod go.sum ./
RUN go mod download

# Luego el resto del código
COPY . .

# Compila con nombre binario finances-api
RUN go build -o /app/finances-api

##############################
# Etapa de producción mínima
##############################
FROM scratch AS runner
WORKDIR /app

COPY --from=builder /app/finances-api ./

# Expone el puerto que usa tu API, presumiblemente 3000
EXPOSE 3000

CMD ["./finances-api"]

