# syntax = docker/dockerfile:1.5

########################
# Etapa de compilación
########################
FROM golang:1.24-alpine AS builder

# Instala certificados para la compilación (necesario si alguno de tus imports hace HTTPS)
RUN apk add --no-cache ca-certificates git

# Configuración para producir binarios estáticos
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

WORKDIR /build

# Copia go.mod y go.sum primero para cachear dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copia el resto del código fuente
COPY . .

# Compila el binario
RUN go build -o /app/finances-api

##############################
# Etapa de producción mínima
##############################
FROM alpine:3.20 AS runner

# Instala solo certificados (no necesitas Go ni git aquí)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copia el binario compilado
COPY --from=builder /app/finances-api .

# Expone el puerto de tu API (ajústalo si es distinto)
EXPOSE 3000

# Arranca tu aplicación
CMD ["./finances-api"]
