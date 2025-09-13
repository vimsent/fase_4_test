package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "michael/proto/grpc/proto"

	"google.golang.org/grpc"
)

const (
	address_lester   = "lester:50051"   // Usar nombres de servicio Docker
	address_trevor   = "trevor:50052"
	address_franklin = "franklin:50053"
)

// Estructura para guardar información del atraco
type HeistInfo struct {
	Botin      uint64
	PFranklin  int32
	PTrevor    int32
	RPolicial  int32
	Fase2      string // Quien hizo la distracción
	Fase3      string // Quien hizo el golpe
	BotinExtra int64
	Exito      bool
	MotivoFallo string
	Fase       int
}

func communicateWithDistractionService(ctx context.Context, address string, message string, exito int32, isTrevor bool) (string, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if isTrevor {
		client := pb.NewTrevorServiceClient(conn)
		response, err := client.Distraccion(ctx, &pb.TrevorRequest{Notificar: message, PTrevor: exito})
		if err != nil {
			return "", err
		}
		return response.Resultado, nil
	} else {
		client := pb.NewFranklinServiceClient(conn)
		response, err := client.Distraccion(ctx, &pb.FranklinRequest{Notificar: message, PFranklin: exito})
		if err != nil {
			return "", err
		}
		return response.Resultado, nil
	}
}

// Nueva función para comunicar con el servicio de Golpe
func communicateWithGolpeService(ctx context.Context, address string, probabilidad int32, riesgoPolicial int32, isTrevor bool) (*pb.GolpeResponse, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	personaje := "franklin"
	if isTrevor {
		personaje = "trevor"
	}

	golpeReq := &pb.GolpeRequest{
		Personaje:      personaje,
		Probabilidad:   probabilidad,
		RiesgoPolicial: riesgoPolicial,
	}

	if isTrevor {
		client := pb.NewTrevorServiceClient(conn)
		return client.IniciarGolpe(ctx, golpeReq)
	} else {
		client := pb.NewFranklinServiceClient(conn)
		return client.IniciarGolpe(ctx, golpeReq)
	}
}

// Función para obtener el botín del personaje que completó el golpe
func obtenerBotinTotal(ctx context.Context, address string, personaje string, isTrevor bool) (int64, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	botinReq := &pb.BotinRequest{
		Personaje: personaje,
	}

	if isTrevor {
		client := pb.NewTrevorServiceClient(conn)
		resp, err := client.ObtenerBotin(ctx, botinReq)
		if err != nil {
			return 0, err
		}
		return resp.BotinTotal, nil
	} else {
		client := pb.NewFranklinServiceClient(conn)
		resp, err := client.ObtenerBotin(ctx, botinReq)
		if err != nil {
			return 0, err
		}
		return resp.BotinTotal, nil
	}
}

// Función para generar el reporte final
func generarReporte(info HeistInfo) {
	file, err := os.Create("Reporte.txt")
	if err != nil {
		log.Printf("Error al crear archivo de reporte: %v", err)
		return
	}
	defer file.Close()

	file.WriteString("=========================================================\n")
	file.WriteString("==              REPORTE FINAL DE LA MISION            ==\n")
	file.WriteString("=========================================================\n")
	
	if info.Exito {
		file.WriteString(fmt.Sprintf("Mision: Asalto al Banco #%d\n", time.Now().Unix()%10000))
		file.WriteString("Resultado Global: MISION COMPLETADA CON EXITO!\n\n")
		file.WriteString("--- DETALLES DEL ATRACO ---\n")
		file.WriteString(fmt.Sprintf("Fase 2 (Distracción): %s - EXITOSA\n", info.Fase2))
		file.WriteString(fmt.Sprintf("Fase 3 (Golpe): %s - EXITOSA\n", info.Fase3))
		file.WriteString("\n--- REPARTO DEL BOTIN ---\n")
		file.WriteString(fmt.Sprintf("Botin Base: $%d\n", info.Botin))
		file.WriteString(fmt.Sprintf("Botin Extra (Habilidad de Chop): $%d\n", info.BotinExtra))
		
		botinTotal := int64(info.Botin) + info.BotinExtra
		file.WriteString(fmt.Sprintf("Botin Total: $%d\n", botinTotal))
		
		// TODO: Agregar lógica de reparto en Fase 4
		file.WriteString("\n---------------------------------------------------------\n")
		file.WriteString("Nota: Implementar Fase 4 para el reparto del botín\n")
	} else {
		file.WriteString(fmt.Sprintf("Mision: Asalto al Banco #%d\n", time.Now().Unix()%10000))
		file.WriteString("Resultado Global: MISION FALLIDA\n\n")
		file.WriteString("--- DETALLES DEL FRACASO ---\n")
		file.WriteString(fmt.Sprintf("Fase del fracaso: %d\n", info.Fase))
		
		if info.Fase == 2 {
			file.WriteString(fmt.Sprintf("Personaje: %s\n", info.Fase2))
		} else if info.Fase == 3 {
			file.WriteString(fmt.Sprintf("Personaje: %s\n", info.Fase3))
		}
		
		file.WriteString(fmt.Sprintf("Motivo: %s\n", info.MotivoFallo))
		file.WriteString(fmt.Sprintf("Botin perdido: $%d\n", info.Botin))
		if info.BotinExtra > 0 {
			file.WriteString(fmt.Sprintf("Botin extra perdido: $%d\n", info.BotinExtra))
		}
	}
	
	file.WriteString("=========================================================\n")
	log.Println("Reporte generado: Reporte.txt")
}

func main() {
	log.Println("Esperando 15 segundos para iniciar el programa...")
	time.Sleep(15 * time.Second)

	conn, err := grpc.Dial(address_lester, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor: %v", err)
	}
	defer conn.Close()

	client := pb.NewLesterServiceClient(conn)
	ctx := context.Background()
	
	var heistInfo HeistInfo

	//------------------------------------FASE 1------------------------------------
	operation := "solicitar_trabajo"
	log.Println("========== INICIO FASE 1: NEGOCIACIÓN ==========")
	count := 0
	
	for {
		count++
		log.Printf("Intento %d", count)
		resp, err := client.MichaelOffer(ctx, &pb.MichaelRequest{Offer: operation})
		if err != nil {
			log.Fatalf("Error al enviar la operación: %v", err)
		}
		
		if resp.POferta == "rechazo" {
			log.Printf("Lester no tiene oferta (10%% probabilidad)")
			log.Printf("Se vuelve a solicitar trabajo...")
			continue
		} else {
			log.Printf("Oferta de Lester:")
			log.Printf("  Botín: $%d", resp.Botin)
			log.Printf("  Éxito Franklin: %d%%", resp.PFranklin)
			log.Printf("  Éxito Trevor: %d%%", resp.PTrevor)
			log.Printf("  Riesgo Policial: %d%%", resp.RPolicial)
			
			if (resp.PFranklin > 50 || resp.PTrevor > 50) && resp.RPolicial < 80 {
				log.Printf("✓ Trabajo aceptado")
				heistInfo.Botin = resp.Botin
				heistInfo.PFranklin = resp.PFranklin
				heistInfo.PTrevor = resp.PTrevor
				heistInfo.RPolicial = resp.RPolicial
				
				resp2, err := client.MichaelOffer(ctx, &pb.MichaelRequest{Offer: "aceptar"})
				if err != nil {
					log.Fatalf("Error al enviar la operación: %v", err)
				}
				if resp2.POferta == "end_phase" {
					log.Printf("Lester confirma fin de fase 1")
					break
				}
			} else {
				log.Printf("✗ Trabajo no cumple requisitos mínimos")
				log.Printf("Consultando a Lester por nueva oferta...")
			}
			time.Sleep(2 * time.Second)
		}
	}

	//------------------------------------FASE 2------------------------------------
	log.Println("\n========== INICIO FASE 2: DISTRACCIÓN ==========")
	
	var chosenPartner string
	var useTrevor bool
	var exito int32
	var partnerAddress string

	if heistInfo.PTrevor > heistInfo.PFranklin {
		chosenPartner = "Trevor"
		useTrevor = true
		exito = heistInfo.PTrevor
		partnerAddress = address_trevor
		heistInfo.Fase2 = "Trevor"
	} else {
		chosenPartner = "Franklin"
		useTrevor = false
		exito = heistInfo.PFranklin
		partnerAddress = address_franklin
		heistInfo.Fase2 = "Franklin"
	}

	log.Printf("Contactando a %s (mayor probabilidad de éxito: %d%%)", chosenPartner, exito)

	message := "iniciar distracción"
	result, err := communicateWithDistractionService(ctx, partnerAddress, message, exito, useTrevor)
	if err != nil {
		log.Fatalf("Error al comunicarse con %s: %v", chosenPartner, err)
	}

	log.Printf("Respuesta de %s: %s", chosenPartner, result)
	
	if result != "exito" {
		heistInfo.Exito = false
		heistInfo.Fase = 2
		heistInfo.MotivoFallo = fmt.Sprintf("%s fracasó en la distracción", chosenPartner)
		generarReporte(heistInfo)
		log.Println("MISIÓN FALLIDA - Fase 2")
		return
	}

	//------------------------------------FASE 3------------------------------------
	log.Println("\n========== INICIO FASE 3: EL GOLPE ==========")
	
	// Determinar quién hace el golpe (el que NO hizo la distracción)
	var golpePartner string
	var golpeAddress string
	var golpeProbabilidad int32
	var useGolpeTrevor bool
	
	if useTrevor {
		// Si Trevor hizo la distracción, Franklin hace el golpe
		golpePartner = "Franklin"
		golpeAddress = address_franklin
		golpeProbabilidad = heistInfo.PFranklin
		useGolpeTrevor = false
		heistInfo.Fase3 = "Franklin"
	} else {
		// Si Franklin hizo la distracción, Trevor hace el golpe
		golpePartner = "Trevor"
		golpeAddress = address_trevor
		golpeProbabilidad = heistInfo.PTrevor
		useGolpeTrevor = true
		heistInfo.Fase3 = "Trevor"
	}
	
	log.Printf("Enviando a %s para el golpe principal", golpePartner)
	
	// Notificar a Lester para que inicie las notificaciones de estrellas
	log.Printf("Notificando a Lester para iniciar alertas de estrellas...")
	notifResp, err := client.IniciarNotificaciones(ctx, &pb.NotificacionRequest{
		Personaje:      golpePartner,
		RiesgoPolicial: heistInfo.RPolicial,
	})
	if err != nil {
		log.Printf("Error al iniciar notificaciones: %v", err)
	} else if notifResp.Iniciado {
		log.Printf("Lester comenzó a enviar notificaciones de estrellas a %s", golpePartner)
	}
	
	// Dar tiempo para que se establezca la comunicación
	time.Sleep(1 * time.Second)
	
	// Iniciar el golpe
	log.Printf("%s iniciando el golpe...", golpePartner)
	golpeResp, err := communicateWithGolpeService(ctx, golpeAddress, golpeProbabilidad, heistInfo.RPolicial, useGolpeTrevor)
	if err != nil {
		log.Fatalf("Error al comunicarse con %s para el golpe: %v", golpePartner, err)
	}
	
	// Detener notificaciones de Lester
	detenerResp, err := client.DetenerNotificaciones(ctx, &pb.DetenerRequest{
		Personaje: golpePartner,
	})
	if err != nil {
		log.Printf("Error al detener notificaciones: %v", err)
	} else if detenerResp.Detenido {
		log.Printf("Notificaciones de estrellas detenidas")
	}
	
	// Evaluar resultado del golpe
	if !golpeResp.Exito {
		log.Printf("✗ %s fracasó en el golpe: %s", golpePartner, golpeResp.MotivoFallo)
		heistInfo.Exito = false
		heistInfo.Fase = 3
		heistInfo.MotivoFallo = golpeResp.MotivoFallo
		heistInfo.BotinExtra = golpeResp.BotinExtra
		generarReporte(heistInfo)
		log.Println("MISIÓN FALLIDA - Fase 3")
		return
	}
	
	log.Printf("✓ %s completó el golpe con éxito!", golpePartner)
	log.Printf("  Estrellas finales: %d", golpeResp.EstrellasFinales)
	if golpeResp.BotinExtra > 0 {
		log.Printf("  Botín extra (Chop): $%d", golpeResp.BotinExtra)
	}
	
	heistInfo.BotinExtra = golpeResp.BotinExtra
	heistInfo.Exito = true
	
	// TODO: Implementar Fase 4 - Reparto del botín
	log.Println("\n========== FASE 4: REPARTO DEL BOTÍN ==========")
	log.Println("TODO: Implementar la fase 4 de reparto del botín")
	
	// Generar reporte final
	generarReporte(heistInfo)
	log.Println("\n========== MISIÓN COMPLETADA CON ÉXITO ==========")
}