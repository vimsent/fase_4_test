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

	pb "trevor/proto/grpc/proto"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTrevorServiceServer
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
	estrellas     int32
	estrellasLock sync.RWMutex
	botinBase     uint64
	furiaActivada bool
	limiteFracaso int32
}

type StarMessage struct {
	Stars     int32  `json:"stars"`
	Character string `json:"character"`
	Timestamp int64  `json:"timestamp"`
}

func NewServer() *server {
	s := &server{
		estrellas:     0,
		furiaActivada: false,
		limiteFracaso: 5, // Límite inicial
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
	
	// Declarar cola para Trevor
	_, err = s.rabbitChannel.QueueDeclare(
		"trevor_stars", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Printf("Error al declarar cola: %v", err)
	}
	
	log.Println("Conexión a RabbitMQ establecida para Trevor")
}

func (s *server) startConsumer() {
	if s.rabbitChannel == nil {
		log.Println("No hay canal RabbitMQ disponible")
		return
	}
	
	msgs, err := s.rabbitChannel.Consume(
		"trevor_stars", // queue
		"",             // consumer
		true,           // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
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
			
			// Activar furia a las 5 estrellas
			if s.estrellas >= 5 && !s.furiaActivada {
				s.furiaActivada = true
				s.limiteFracaso = 7
				log.Printf("¡FURIA ACTIVADA! Trevor ahora puede soportar hasta 7 estrellas")
			}
			
			s.estrellasLock.Unlock()
			
			log.Printf("Trevor recibió notificación: %d estrellas", starMsg.Stars)
		}
	}()
}

// Fase 2: Distracción (sin cambios)
func (s *server) Distraccion(ctx context.Context, in *pb.TrevorRequest) (*pb.TrevorResponse, error) {
	log.Printf("Recibida petición de distracción de: %v", in.GetNotificar())
	turnosNecesarios := 200 - in.GetPTrevor()
	rand.Seed(time.Now().UnixNano())
	log.Printf("Turnos necesarios: %d", turnosNecesarios)
	turnos := int32(0)
	imprevisto := turnosNecesarios / 2
	
	for turnos < turnosNecesarios {
		if turnos == imprevisto {
			if rand.Float64() > 0.9 {
				log.Printf("Trevor está borracho, el trabajo ha fracasado...")
				return &pb.TrevorResponse{
					Resultado: "fracaso",
				}, nil
			}
		}
		log.Printf("Creando distracción... Turno %d", turnos)
		turnos++
		time.Sleep(50 * time.Millisecond) // Simular trabajo
	}
	
	log.Printf("Trevor termina distracción con éxito")
	return &pb.TrevorResponse{
		Resultado: "exito",
	}, nil
}

// Fase 3: El Golpe
func (s *server) IniciarGolpe(ctx context.Context, in *pb.GolpeRequest) (*pb.GolpeResponse, error) {
	log.Printf("Trevor iniciando el golpe principal")
	
	// Reset de variables
	s.estrellas = 0
	s.furiaActivada = false
	s.limiteFracaso = 5

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
			limiteActual := s.limiteFracaso
			s.estrellasLock.RUnlock()
			
			log.Printf("Turno %d/%d - Estrellas: %d/%d", 
				turnos, turnosNecesarios, estrellasActuales, limiteActual)
			
			// Verificar límite de fracaso
			if estrellasActuales >= limiteActual {
				log.Printf("Trevor alcanzó el límite de %d estrellas - Misión fallida", limiteActual)
				return &pb.GolpeResponse{
					Exito:           false,
					MotivoFallo:     fmt.Sprintf("Alcanzó %d estrellas de búsqueda", limiteActual),
					BotinExtra:      0,
					EstrellasFinales: estrellasActuales,
				}, nil
			}
		}
		
		turnos++
		time.Sleep(100 * time.Millisecond) // Simular trabajo
	}
	
	s.estrellasLock.RLock()
	estrellasFinales := s.estrellas
	s.estrellasLock.RUnlock()
	
	log.Printf("Trevor completó el golpe con éxito")
	
	return &pb.GolpeResponse{
		Exito:           true,
		MotivoFallo:     "",
		BotinExtra:      0, // Trevor no genera botín extra
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
	botinTotal := int64(s.botinBase)
	
	log.Printf("Trevor entregando botín total: $%d", botinTotal)
	
	return &pb.BotinResponse{
		BotinTotal: botinTotal,
	}, nil
}

// RecibirPago - Fase 4: Verificar el pago recibido
func (s *server) RecibirPago(ctx context.Context, in *pb.PagoRequest) (*pb.PagoResponse, error) {
	monto := in.GetMonto()
	concepto := in.GetConcepto()
	
	log.Printf("Trevor recibió pago de $%d por concepto: %s", monto, concepto)
	
	// Calcular el pago esperado
	pagoEsperado := int64(s.botinBase) / 4
	
	// Verificar si el pago es correcto
	if monto == pagoEsperado {
		log.Printf("Trevor confirma: Pago correcto")
		return &pb.PagoResponse{
			Correcto: true,
			Mensaje:  "¡Justo lo que esperaba!",
		}, nil
	} else if monto < pagoEsperado {
		log.Printf("Trevor enfadado: Pago insuficiente (esperaba $%d)", pagoEsperado)
		return &pb.PagoResponse{
			Correcto: false,
			Mensaje:  fmt.Sprintf("¡¿QUÉ?! ¡Me estás robando! Esperaba $%d y me das $%d", pagoEsperado, monto),
		}, nil
	} else {
		log.Printf("Trevor contento: Pago mayor al esperado")
		return &pb.PagoResponse{
			Correcto: true,
			Mensaje:  "¡Ja! Más dinero para mí, perfecto.",
		}, nil
	}
}

func main() {
	s := NewServer()
	
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTrevorServiceServer(grpcServer, s)
	
	log.Printf("Trevor server starting on port 50052...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}