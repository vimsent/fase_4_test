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
	address_lester   = "lester:50051"
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
	BotinTotal int64
	Exito      bool
	MotivoFallo string
	Fase       int
	// Para fase 4
	PagoFranklin int64
	PagoTrevor   int64
	PagoLester   int64
	PagoMichael  int64
	Resto        int64
	RespuestaFranklin string
	RespuestaTrevor   string
	RespuestaLester   string
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

// Nueva función para enviar pago a un personaje
func enviarPago(ctx context.Context, address string, monto int64, concepto string) (*pb.PagoResponse, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	pagoReq := &pb.PagoRequest{
		Monto:    monto,
		Concepto: concepto,
	}

	// Determinar qué cliente usar basado en la dirección
	switch address {
	case address_lester:
		client := pb.NewLesterServiceClient(conn)
		return client.RecibirPago(ctx, pagoReq)
	case address_trevor:
		client := pb.NewTrevorServiceClient(conn)
		return client.RecibirPago(ctx, pagoReq)
	case address_franklin:
		client := pb.NewFranklinServiceClient(conn)
		return client.RecibirPago(ctx, pagoReq)
	default:
		return nil, fmt.Errorf("dirección desconocida: %s", address)
	}
}

func generarReporte(info HeistInfo) {
    // CORRECCIÓN: Crear directorio si no existe y manejar errores mejor
    file, err := os.Create("Reporte.txt")
    if err != nil {
        log.Printf("Error al crear archivo de reporte: %v", err)
        // Intentar crear en directorio temporal
        file, err = os.Create("/tmp/Reporte.txt")
        if err != nil {
            log.Printf("Error al crear archivo en /tmp: %v", err)
            return
        }
        log.Printf("Archivo creado en /tmp/Reporte.txt")
    }
    defer file.Close()

    file.WriteString("=========================================================\n")
    file.WriteString("==              REPORTE FINAL DE LA MISION            ==\n")
    file.WriteString("=========================================================\n")
    
    if info.Exito {
        file.WriteString(fmt.Sprintf("Mision: Asalto al Banco #%d\n", time.Now().Unix()%10000))
        file.WriteString("Resultado Global: MISION COMPLETADA CON EXITO!\n\n")
        
        file.WriteString("--- DETALLES DE LA MISION ---\n")
        file.WriteString(fmt.Sprintf("Fase 2 - Distraccion: %s\n", info.Fase2))
        file.WriteString(fmt.Sprintf("Fase 3 - Golpe Principal: %s\n", info.Fase3))
        file.WriteString("\n")
        
        file.WriteString("--- REPARTO DEL BOTIN ---\n")
        file.WriteString(fmt.Sprintf("Botin Base: $%d\n", info.Botin))
        file.WriteString(fmt.Sprintf("Botin Extra (Habilidad de Chop): $%d\n", info.BotinExtra))
        file.WriteString(fmt.Sprintf("Botin Total: $%d\n", info.BotinTotal))
        file.WriteString("\n")
        
        // CORRECCIÓN: Verificar que los valores no sean 0 antes de mostrar
        if info.BotinTotal > 0 {
            file.WriteString("---------------------------------------------------------\n")
            file.WriteString(fmt.Sprintf("Pago a Franklin: $%d\n", info.PagoFranklin))
            if info.RespuestaFranklin != "" {
                file.WriteString(fmt.Sprintf("Respuesta de Franklin: \"%s\"\n", info.RespuestaFranklin))
            }
            file.WriteString(fmt.Sprintf("Pago a Trevor: $%d\n", info.PagoTrevor))
            if info.RespuestaTrevor != "" {
                file.WriteString(fmt.Sprintf("Respuesta de Trevor: \"%s\"\n", info.RespuestaTrevor))
            }
            file.WriteString(fmt.Sprintf("Pago a Lester: $%d", info.PagoLester))
            if info.Resto > 0 {
                file.WriteString(fmt.Sprintf(" + $%d (resto)", info.Resto))
            }
            file.WriteString("\n")
            if info.RespuestaLester != "" {
                file.WriteString(fmt.Sprintf("Respuesta de Lester: \"%s\"\n", info.RespuestaLester))
            }
            file.WriteString(fmt.Sprintf("Pago a Michael: $%d\n", info.PagoMichael))
            file.WriteString("---------------------------------------------------------\n")
            
            // Verificación de suma
            totalPagado := info.PagoFranklin + info.PagoTrevor + info.PagoLester + info.PagoMichael + info.Resto
            file.WriteString(fmt.Sprintf("Verificacion - Total Pagado: $%d\n", totalPagado))
            file.WriteString(fmt.Sprintf("Verificacion - Botin Original: $%d\n", info.BotinTotal))
            
            if totalPagado == info.BotinTotal {
                file.WriteString("✓ Los números cuadran perfectamente\n")
            } else {
                file.WriteString(fmt.Sprintf("⚠ DISCREPANCIA: Diferencia de $%d\n", info.BotinTotal - totalPagado))
            }
        } else {
            file.WriteString("ERROR: No se pudo calcular el reparto (Botin Total = 0)\n")
        }
    } else {
        file.WriteString(fmt.Sprintf("Mision: Asalto al Banco #%d\n", time.Now().Unix()%10000))
        file.WriteString("Resultado Global: MISION FALLIDA\n\n")
        file.WriteString("--- DETALLES DEL FRACASO ---\n")
        file.WriteString(fmt.Sprintf("Fase del fracaso: %d\n", info.Fase))
        
        if info.Fase == 2 {
            file.WriteString(fmt.Sprintf("Personaje que fallo: %s\n", info.Fase2))
        } else if info.Fase == 3 {
            file.WriteString(fmt.Sprintf("Personaje que fallo: %s\n", info.Fase3))
        }
        
        file.WriteString(fmt.Sprintf("Motivo: %s\n", info.MotivoFallo))
        file.WriteString(fmt.Sprintf("Botin perdido: $%d\n", info.Botin))
        if info.BotinExtra > 0 {
            file.WriteString(fmt.Sprintf("Botin extra perdido: $%d\n", info.BotinExtra))
        }
    }
    
    file.WriteString("=========================================================\n")
    file.WriteString(fmt.Sprintf("Reporte generado: %s\n", time.Now().Format("2006-01-02 15:04:05")))
    file.WriteString("=========================================================\n")
    
    log.Println("Reporte generado exitosamente: Reporte.txt")
}

func main() {
	log.Println("Esperando 30 segundos para iniciar el programa...")
	time.Sleep(30 * time.Second)

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
	
	var golpePartner string
	var golpeAddress string
	var golpeProbabilidad int32
	var useGolpeTrevor bool
	
	if useTrevor {
		golpePartner = "Franklin"
		golpeAddress = address_franklin
		golpeProbabilidad = heistInfo.PFranklin
		useGolpeTrevor = false
		heistInfo.Fase3 = "Franklin"
	} else {
		golpePartner = "Trevor"
		golpeAddress = address_trevor
		golpeProbabilidad = heistInfo.PTrevor
		useGolpeTrevor = true
		heistInfo.Fase3 = "Trevor"
	}
	
	log.Printf("Enviando a %s para el golpe principal", golpePartner)
	
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
	
	time.Sleep(1 * time.Second)
	
	log.Printf("%s iniciando el golpe...", golpePartner)
	golpeResp, err := communicateWithGolpeService(ctx, golpeAddress, golpeProbabilidad, heistInfo.RPolicial, useGolpeTrevor)
	if err != nil {
		log.Fatalf("Error al comunicarse con %s para el golpe: %v", golpePartner, err)
	}
	
	detenerResp, err := client.DetenerNotificaciones(ctx, &pb.DetenerRequest{
		Personaje: golpePartner,
	})
	if err != nil {
		log.Printf("Error al detener notificaciones: %v", err)
	} else if detenerResp.Detenido {
		log.Printf("Notificaciones de estrellas detenidas")
	}
	
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
	
	//------------------------------------FASE 4------------------------------------
log.Println("\n========== FASE 4: REPARTO DEL BOTÍN ==========")

// CORRECCIÓN 1: Calcular el botín total correctamente
var botinTotal int64

// Intentar obtener el botín del personaje que completó el golpe
botinObtenido, err := obtenerBotinTotal(ctx, golpeAddress, golpePartner, useGolpeTrevor)
if err != nil {
    log.Printf("Error al obtener el botín del personaje: %v", err)
    // Usar el botín base + extra como fallback
    botinTotal = int64(heistInfo.Botin) + heistInfo.BotinExtra
    log.Printf("Usando botín calculado como fallback: $%d", botinTotal)
} else if botinObtenido == 0 {
    // CORRECCIÓN 2: Si el personaje devuelve 0, usar el botín calculado
    botinTotal = int64(heistInfo.Botin) + heistInfo.BotinExtra
    log.Printf("El personaje devolvió 0, usando botín calculado: $%d", botinTotal)
} else {
    botinTotal = botinObtenido
    log.Printf("Botín obtenido del personaje: $%d", botinTotal)
}

// CORRECCIÓN 3: Verificar que el botín total sea mayor a 0
if botinTotal <= 0 {
    log.Printf("ERROR: Botín total es 0 o negativo. Abortando reparto.")
    heistInfo.BotinTotal = 0
    heistInfo.Exito = false
    heistInfo.Fase = 4
    heistInfo.MotivoFallo = "Error en el cálculo del botín total"
    generarReporte(heistInfo)
    return
}

heistInfo.BotinTotal = botinTotal
log.Printf("Botín total confirmado: $%d", botinTotal)

// Calcular reparto
pagoPorPersona := botinTotal / 4
resto := botinTotal % 4

log.Printf("Reparto calculado: $%d por persona", pagoPorPersona)
if resto > 0 {
    log.Printf("Resto para Lester: $%d", resto)
}

heistInfo.PagoFranklin = pagoPorPersona
heistInfo.PagoTrevor = pagoPorPersona
heistInfo.PagoLester = pagoPorPersona
heistInfo.PagoMichael = pagoPorPersona
heistInfo.Resto = resto

// CORRECCIÓN 4: Verificar que los pagos sean válidos antes de enviar
if pagoPorPersona <= 0 {
    log.Printf("ERROR: Pago por persona es 0 o negativo: $%d", pagoPorPersona)
    heistInfo.RespuestaFranklin = "Pago inválido"
    heistInfo.RespuestaTrevor = "Pago inválido"  
    heistInfo.RespuestaLester = "Pago inválido"
} else {
    // Pagar a Franklin
    log.Printf("Pagando a Franklin: $%d", pagoPorPersona)
    respFranklin, err := enviarPago(ctx, address_franklin, pagoPorPersona, "reparto")
    if err != nil {
        log.Printf("Error al pagar a Franklin: %v", err)
        heistInfo.RespuestaFranklin = "Error en el pago"
    } else {
        heistInfo.RespuestaFranklin = respFranklin.Mensaje
        log.Printf("Franklin responde: %s", respFranklin.Mensaje)
    }
    
    // Pagar a Trevor
    log.Printf("Pagando a Trevor: $%d", pagoPorPersona)
    respTrevor, err := enviarPago(ctx, address_trevor, pagoPorPersona, "reparto")
    if err != nil {
        log.Printf("Error al pagar a Trevor: %v", err)
        heistInfo.RespuestaTrevor = "Error en el pago"
    } else {
        heistInfo.RespuestaTrevor = respTrevor.Mensaje
        log.Printf("Trevor responde: %s", respTrevor.Mensaje)
    }
    
    // Pagar a Lester (reparto + resto)
    log.Printf("Pagando a Lester: $%d (reparto)", pagoPorPersona)
    
    // Primero el reparto normal
    respLester, err := enviarPago(ctx, address_lester, pagoPorPersona, "reparto")
    if err != nil {
        log.Printf("Error al pagar reparto a Lester: %v", err)
        heistInfo.RespuestaLester = "Error en el pago del reparto"
    } else {
        heistInfo.RespuestaLester = respLester.Mensaje
        log.Printf("Lester responde por el reparto: %s", respLester.Mensaje)
    }
    
    // Luego el resto si existe
    if resto > 0 {
        log.Printf("Pagando resto a Lester: $%d", resto)
        respLesterResto, err := enviarPago(ctx, address_lester, resto, "resto")
        if err != nil {
            log.Printf("Error al pagar resto a Lester: %v", err)
            heistInfo.RespuestaLester += " | Error en el resto"
        } else {
            heistInfo.RespuestaLester = respLesterResto.Mensaje
            log.Printf("Lester responde por el resto: %s", respLesterResto.Mensaje)
        }
    }
}

// Michael se queda con su parte
log.Printf("Michael se queda con: $%d", pagoPorPersona)

// CORRECCIÓN 5: Asegurar que la generación del reporte siempre funcione
log.Printf("Generando reporte con los siguientes datos:")
log.Printf("  - Éxito: %v", heistInfo.Exito)
log.Printf("  - Botín Total: $%d", heistInfo.BotinTotal)
log.Printf("  - Pagos: Franklin=$%d, Trevor=$%d, Lester=$%d, Michael=$%d", 
    heistInfo.PagoFranklin, heistInfo.PagoTrevor, heistInfo.PagoLester, heistInfo.PagoMichael)

// Generar reporte final
generarReporte(heistInfo)
log.Println("\n========== MISIÓN COMPLETADA CON ÉXITO ==========")
}