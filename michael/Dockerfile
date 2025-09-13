# Imagen base con Go
FROM golang:1.24.2 AS builder

# Crear directorio de trabajo
WORKDIR /app

# Copiar go.mod y go.sum primero para cachear dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar Michael
RUN go build -o michael michael.go

# Imagen final
FROM golang:1.24.2

WORKDIR /root/

# Copiar el binario desde la etapa anterior
COPY --from=builder /app/michael .

# Puerto que usará Michael como cliente (no expone, solo si hiciera de server)
# EXPOSE 50051

# Ejecutar el binario
CMD ["./michael"]