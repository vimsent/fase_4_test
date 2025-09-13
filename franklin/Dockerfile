# Etapa de build
FROM golang:1.24.2 AS builder

WORKDIR /app

# Copiar dependencias y descargarlas
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar todo el paquete principal
RUN go build -o franklin franklin.go

# Imagen final (más liviana)
FROM golang:1.24.2

WORKDIR /root/

# Copiar binario desde la etapa anterior
COPY --from=builder /app/franklin .

# Cambiar el entrypoint según la entidad
CMD ["./franklin"]
