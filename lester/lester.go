package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	pb "lester/proto/grpc/proto"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedLesterServiceServer
	offers            []*pb.MichaelResponse
	currentIndex      int32
	mutex             sync.Mutex
	rabbitConn        *amqp.Connection
	rabbitChannel     *amqp.Channel
	stopNotifications map[string]chan bool // Para detener las notificaciones por personaje
}

// Estructura para mensajes de estrellas
type StarMessage struct {
	Stars     int32  `json:"stars"`
	Character string `json:"character"`
	Timestamp int64  `json:"timestamp"`
}

func NewServer() *server {
	s := &server{
		currentIndex:      0,
		stopNotifications: make(map[string]chan bool),
	}
	s.loadOffersFromCSV("ofertas_pequeno.csv")
	s.connectRabbitMQ()
	return s
}

func (s *server) connectRabbitMQ() {
	var err error
	// Intentar conectar varias veces con retry
	for i := 0; i < 10; i++ {
		// Usar variables de entorno para la conexión
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
		log.Printf("No se pudo conectar a RabbitMQ después de varios intentos: %v", err)
		return
	}
	
	s.rabbitChannel, err = s.rabbitConn.Channel()
	if err != nil {
		log.Printf("Error al crear canal RabbitMQ: %v", err)
		return
	}
	
	// Declarar las colas para Franklin y Trevor
	queues := []string{"franklin_stars", "trevor_stars"}
	for _, q := range queues {
		_, err = s.rabbitChannel.QueueDeclare(
			q,     // name
			false, // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Printf("Error al declarar cola %s: %v", q, err)
		}
	}
	
	log.Println("Conexión a RabbitMQ establecida")
}

func (s *server) loadOffersFromCSV(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	for i, record := range records {
		if i == 0 {
			continue
		}

		if len(record) >= 4 {
			botin, _ := strconv.ParseUint(record[0], 10, 64)
			pFranklin, _ := strconv.ParseInt(record[1], 10, 32)
			pTrevor, _ := strconv.ParseInt(record[2], 10, 32)
			rPolicial, _ := strconv.ParseInt(record[3], 10, 32)

			offer := &pb.MichaelResponse{
				Botin:     botin,
				PFranklin: int32(pFranklin),
				PTrevor:   int32(pTrevor),
				RPolicial: int32(rPolicial),
			}
			s.offers = append(s.offers, offer)
		}
	}

	log.Printf("Loaded %d offers from CSV", len(s.offers))
}

func (s *server) MichaelOffer(ctx context.Context, in *pb.MichaelRequest) (*pb.MichaelResponse, error) {
	log.Printf("Recibida petición de: %v", in.GetOffer())

	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if in.GetOffer() == "aceptar" {
		log.Printf("Aceptación de oferta recibida")
		log.Printf("Enviando confirmación de fin de fase...")
		return &pb.MichaelResponse{
			POferta: "end_phase",
		}, nil
	}

	rand.Seed(time.Now().UnixNano())
	if rand.Float64() > 0.9 {
		log.Printf("No hay ofertas disponibles para Michael")
		return &pb.MichaelResponse{
			POferta: "rechazo",
		}, nil
	}

	if s.currentIndex == 3 {
		log.Printf("3 rechazos detectados, se esperan 10 segundos...")
		time.Sleep(10 * time.Second)
	}
	
	offer := s.offers[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % int32(len(s.offers))

	log.Printf("Returning offer %d of %d", s.currentIndex, len(s.offers))
	return offer, nil
}

// Nueva función para iniciar notificaciones de estrellas
func (s *server) IniciarNotificaciones(ctx context.Context, in *pb.NotificacionRequest) (*pb.NotificacionResponse, error) {
	personaje := in.GetPersonaje()
	riesgoPolicial := in.GetRiesgoPolicial()
	
	log.Printf("Iniciando notificaciones de estrellas para %s con riesgo policial %d", personaje, riesgoPolicial)
	
	// Detener notificaciones anteriores si existen
	if stopChan, exists := s.stopNotifications[personaje]; exists {
		close(stopChan)
		time.Sleep(100 * time.Millisecond) // Dar tiempo para que se detenga
	}
	
	// Crear nuevo canal de stop
	stopChan := make(chan bool)
	s.stopNotifications[personaje] = stopChan
	
	// Calcular frecuencia de estrellas
	frecuenciaEstrellas := 100 - riesgoPolicial
	if frecuenciaEstrellas <= 0 {
		frecuenciaEstrellas = 1
	}
	
	// Iniciar goroutine para enviar notificaciones
	go s.enviarNotificacionesEstrellas(personaje, frecuenciaEstrellas, stopChan)
	
	return &pb.NotificacionResponse{
		Iniciado: true,
	}, nil
}

func (s *server) enviarNotificacionesEstrellas(personaje string, frecuencia int32, stopChan chan bool) {
	estrellas := int32(0)
	queueName := fmt.Sprintf("%s_stars", personaje)
	
	// Simular turnos con un ticker
	ticker := time.NewTicker(time.Duration(frecuencia) * 100 * time.Millisecond) // Ajustar tiempo para pruebas
	defer ticker.Stop()
	
	log.Printf("Comenzando envío de estrellas para %s cada %d turnos", personaje, frecuencia)
	
	for {
		select {
		case <-stopChan:
			log.Printf("Deteniendo notificaciones para %s", personaje)
			return
		case <-ticker.C:
			estrellas++
			msg := StarMessage{
				Stars:     estrellas,
				Character: personaje,
				Timestamp: time.Now().Unix(),
			}
			
			body, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error al serializar mensaje: %v", err)
				continue
			}
			
			if s.rabbitChannel != nil {
				err = s.rabbitChannel.Publish(
					"",        // exchange
					queueName, // routing key
					false,     // mandatory
					false,     // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					})
				
				if err != nil {
					log.Printf("Error al publicar mensaje: %v", err)
				} else {
					log.Printf("Estrella %d enviada a %s", estrellas, personaje)
				}
			}
		}
	}
}

// Nueva función para detener notificaciones
func (s *server) DetenerNotificaciones(ctx context.Context, in *pb.DetenerRequest) (*pb.DetenerResponse, error) {
	personaje := in.GetPersonaje()
	
	log.Printf("Deteniendo notificaciones para %s", personaje)
	
	if stopChan, exists := s.stopNotifications[personaje]; exists {
		close(stopChan)
		delete(s.stopNotifications, personaje)
		return &pb.DetenerResponse{
			Detenido: true,
		}, nil
	}
	
	return &pb.DetenerResponse{
		Detenido: false,
	}, nil
}

// RecibirPago - Fase 4: Verificar el pago recibido
func (s *server) RecibirPago(ctx context.Context, in *pb.PagoRequest) (*pb.PagoResponse, error) {
	monto := in.GetMonto()
	concepto := in.GetConcepto()
	
	log.Printf("Lester recibió pago de $%d por concepto: %s", monto, concepto)
	
	// Lester siempre está satisfecho con cualquier pago que reciba
	if concepto == "resto" {
		log.Printf("Lester recibe el resto del botín: $%d", monto)
		return &pb.PagoResponse{
			Correcto: true,
			Mensaje:  fmt.Sprintf("Perfecto, $%d extra para mí. Un placer hacer negocios.", monto),
		}, nil
	}
	
	// Para el reparto normal
	log.Printf("Lester confirma recepción del pago")
	return &pb.PagoResponse{
		Correcto: true,
		Mensaje:  "Un placer hacer negocios.",
	}, nil
}

func main() {
	s := NewServer()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLesterServiceServer(grpcServer, s)

	log.Printf("Server starting on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}