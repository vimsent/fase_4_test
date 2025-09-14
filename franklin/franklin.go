package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	pb "franklin/proto/grpc/proto"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFranklinServiceServer
	rabbitConn     *amqp.Connection
	rabbitChannel  *amqp.Channel
	estrellas      int32
	estrellasLock  sync.RWMutex
	botinExtra     int64
	botinBase      uint64
	chopActivado   bool
}

type StarMessage struct {
	Stars     int32  `json:"stars"`
	Character string `json:"character"`
	Timestamp int64  `json:"timestamp"`
}

func NewServer() *server {
	s := &server{
		estrellas:    0,
		botinExtra:   0,
		chopActivado: false,
	}
	s.connectRabbitMQ()
	s.startConsumer()
	return s
}

func (s *server) connectRabbitMQ() {
	var err error
	for i := 0; i < 10; i++ {
		rabbitHost := os.Getenv("RABBITMQ_HOST")
		if rabbitHost == "" {
			rabbitHost = "rabbitmq"
		}
		
		connStr := fmt.Sprintf("amqp://user:pass@%s:5672/", rabbitHost)
		s.rabbitConn, err = amqp.Dial(connStr)
		if err == nil {
			break
		}
		log.Printf("Intento %d de conectar a RabbitMQ falló: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		log.Printf("No se pudo conectar a RabbitMQ: %v", err)
		return
	}
	
	s.rabbitChannel, err = s.rabbitConn.Channel()
	if err != nil {
		log.Printf("Error al crear canal RabbitMQ: %v", err)
		return
	}
	
	// Declarar cola para Franklin
	_, err = s.rabbitChannel.QueueDeclare(
		"franklin_stars", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		log.Printf("Error al declarar cola: %v", err)
	}
	
	log.Println("Conexión a RabbitMQ establecida para Franklin")
}

func (s *server) startConsumer() {
	if s.rabbitChannel == nil {
		log.Println("No hay canal RabbitMQ disponible")
		return
	}
	
	msgs, err := s.rabbitChannel.Consume(
		"franklin_stars", // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	if err != nil {
		log.Printf("Error al consumir mensajes: %v", err)
		return
	}
	
	go func() {
		for msg := range msgs {
			var starMsg StarMessage
			err := json.Unmarshal(msg.Body, &starMsg)
			if err != nil {
				log.Printf("Error al deserializar mensaje: %v", err)
				continue
			}
			
			s.estrellasLock.Lock()
			s.estrellas = starMsg.Stars
			
			// Activar habilidad de Chop a las 3 estrellas
			if s.estrellas >= 3 && !s.chopActivado {
				s.chopActivado = true
				log.Printf("¡Chop activado! Franklin comenzará a obtener botín extra")
			}
			
			s.estrellasLock.Unlock()
			
			log.Printf("Franklin recibió notificación: %d estrellas", starMsg.Stars)
		}
	}()
}

// Fase 2: Distracción (sin cambios)
func (s *server) Distraccion(ctx context.Context, in *pb.FranklinRequest) (*pb.FranklinResponse, error) {
	log.Printf("Recibida petición de distracción de: %v", in.GetNotificar())
	turnosNecesarios := 200 - in.GetPFranklin()
	rand.Seed(time.Now().UnixNano())
	log.Printf("Turnos necesarios: %d", turnosNecesarios)
	turnos := int32(0)
	imprevisto := turnosNecesarios / 2
	
	for turnos < turnosNecesarios {
		if turnos == imprevisto {
			if rand.Float64() > 0.9 {
				log.Printf("Chop ladró y distrajo a Franklin, el trabajo ha fracasado...")
				return &pb.FranklinResponse{
					Resultado: "fracaso",
				}, nil
			}
		}
		log.Printf("Creando distracción... Turno %d", turnos)
		turnos++
		time.Sleep(50 * time.Millisecond) // Simular trabajo
	}
	
	log.Printf("Franklin termina distracción con éxito")
	return &pb.FranklinResponse{
		Resultado: "exito",
	}, nil
}

// Fase 3: El Golpe
func (s *server) IniciarGolpe(ctx context.Context, in *pb.GolpeRequest) (*pb.GolpeResponse, error) {
	log.Printf("Franklin iniciando el golpe principal")
	
	// Reset de variables
	s.estrellas = 0
	s.botinExtra = 0
	s.chopActivado = false
	
	s.botinBase = 100000 

	probabilidad := in.GetProbabilidad()
	turnosNecesarios := 200 - probabilidad
	
	log.Printf("Turnos necesarios para el golpe: %d", turnosNecesarios)
    log.Printf("Botín base establecido: $%d", s.botinBase)


	turnos := int32(0)
	consultaIntervalo := turnosNecesarios / 5 // Consultar cada 20% del progreso
	if consultaIntervalo == 0 {
		consultaIntervalo = 1
	}
	
	for turnos < turnosNecesarios {
		// Consultar estrellas periódicamente
		if turnos%consultaIntervalo == 0 {
			s.estrellasLock.RLock()
			estrellasActuales := s.estrellas
			chopActivo := s.chopActivado
			s.estrellasLock.RUnlock()
			
			log.Printf("Turno %d/%d - Estrellas: %d - Chop activo: %v", 
				turnos, turnosNecesarios, estrellasActuales, chopActivo)
			
			// Verificar límite de fracaso (5 estrellas)
			if estrellasActuales >= 5 {
				log.Printf("Franklin alcanzó 5 estrellas - Misión fallida")
				return &pb.GolpeResponse{
					Exito:           false,
					MotivoFallo:     "Alcanzó 5 estrellas de búsqueda",
					BotinExtra:      s.botinExtra,
					EstrellasFinales: estrellasActuales,
				}, nil
			}
			
			// Si Chop está activo, ganar $1,000 por turno
			if chopActivo {
				s.botinExtra += 1000
			}
		}
		
		turnos++
		time.Sleep(100 * time.Millisecond) // Simular trabajo
	}
	
	s.estrellasLock.RLock()
	estrellasFinales := s.estrellas
	s.estrellasLock.RUnlock()
	
	log.Printf("Franklin completó el golpe con éxito - Botín extra: $%d", s.botinExtra)
	
	return &pb.GolpeResponse{
		Exito:           true,
		MotivoFallo:     "",
		BotinExtra:      s.botinExtra,
		EstrellasFinales: estrellasFinales,
	}, nil
}

// Consultar estrellas actuales
func (s *server) ConsultarEstrellas(ctx context.Context, in *pb.EstrellasRequest) (*pb.EstrellasResponse, error) {
	s.estrellasLock.RLock()
	defer s.estrellasLock.RUnlock()
	
	return &pb.EstrellasResponse{
		Estrellas: s.estrellas,
	}, nil
}

// Obtener botín total
func (s *server) ObtenerBotin(ctx context.Context, in *pb.BotinRequest) (*pb.BotinResponse, error) {
	botinTotal := int64(s.botinBase) + s.botinExtra
	
	log.Printf("Franklin entregando botín total: $%d (base: %d, extra: %d)", 
		botinTotal, s.botinBase, s.botinExtra)
	
	return &pb.BotinResponse{
		BotinTotal: botinTotal,
	}, nil
}

// RecibirPago - Fase 4: Verificar el pago recibido
func (s *server) RecibirPago(ctx context.Context, in *pb.PagoRequest) (*pb.PagoResponse, error) {
	monto := in.GetMonto()
	concepto := in.GetConcepto()
	
	log.Printf("Franklin recibió pago de $%d por concepto: %s", monto, concepto)
	
	// Calcular el pago esperado
	botinTotal := int64(s.botinBase) + s.botinExtra
	pagoEsperado := botinTotal / 4
	
	// Verificar si el pago es correcto
	if monto == pagoEsperado {
		log.Printf("Franklin confirma: Pago correcto")
		return &pb.PagoResponse{
			Correcto: true,
			Mensaje:  "¡Excelente! El pago es correcto.",
		}, nil
	} else if monto < pagoEsperado {
		log.Printf("Franklin reclama: Pago insuficiente (esperaba $%d)", pagoEsperado)
		return &pb.PagoResponse{
			Correcto: false,
			Mensaje:  fmt.Sprintf("Hey, falta dinero. Esperaba $%d pero recibí $%d", pagoEsperado, monto),
		}, nil
	} else {
		log.Printf("Franklin sorprendido: Pago mayor al esperado")
		return &pb.PagoResponse{
			Correcto: true,
			Mensaje:  "Más de lo esperado, ¡gracias!",
		}, nil
	}
}

func main() {
	s := NewServer()
	
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFranklinServiceServer(grpcServer, s)
	
	log.Printf("Franklin server starting on port 50053...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}