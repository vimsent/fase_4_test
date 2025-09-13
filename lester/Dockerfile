# Etapa de build
FROM golang:1.24.2 AS builder

WORKDIR /app

# Copiar dependencias y descargarlas
COPY go.mod go.sum ./
RUN go mod download

# Copiar el archivo CSV de ofertas
COPY ofertas_pequeno.csv .

# Copiar el código fuente
COPY . .

# Compilar todo el paquete principal
RUN go build -o lester lester.go

# Imagen final (más liviana)
FROM golang:1.24.2

WORKDIR /root/

# Copiar binario desde la etapa anterior
COPY --from=builder /app/lester .

# Copiar el CSV desde la etapa de build
COPY --from=builder /app/ofertas_pequeno.csv .

# Cambiar el entrypoint según la entidad
CMD ["./lester"]

