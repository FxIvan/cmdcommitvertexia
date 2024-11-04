package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/fxivan/commitnamegen_ia/util"
)

func formatJSON(data []byte) string {
	// Convertir `data` a string para manejar líneas individualmente
	dataStr := string(data)

	var strBuilder strings.Builder
	for _, line := range strings.Split(dataStr, "\n") {
		if strings.HasPrefix(line, "data: ") {

			var dat map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &dat); err != nil {
				panic(err)
			}

			byt := []byte(strings.TrimPrefix(line, "data: "))

			if err := json.Unmarshal(byt, &dat); err != nil {
				panic(err)
			}

			if candidates, ok := dat["candidates"].([]interface{}); ok {
				if len(candidates) > 0 {
					if content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{}); ok {
						if parts, ok := content["parts"].([]interface{}); ok {
							if len(parts) > 0 {
								if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
									strBuilder.WriteString(text)
								} else {
									fmt.Println("Campo 'text' no encontrado o no es una cadena")
								}
							} else {
								fmt.Println("Lista 'parts' vacía o no es una lista")
							}
						} else {
							fmt.Println("Campo 'parts' no encontrado o no es una lista")
						}
					} else {
						fmt.Println("Campo 'content' no encontrado o no es un mapa")
					}
				} else {
					fmt.Println("Lista 'candidates' vacía o no es una lista")
				}
			} else {
				fmt.Println("Campo 'candidates' no encontrado o no es una lista")
			}

		}
	}

	return strBuilder.String()
}

func MakeRequests() error {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	url := os.Getenv("URL_PROYECT_GCP")

	if len(os.Args) < 3 {
		fmt.Println("Uso: go run main.go <ticket> <description>")
		return nil
	}

	ticket := os.Args[1]
	description := os.Args[2]
	fmt.Println("Ticket -->", ticket)
	fmt.Println("Description -->", description)
	dateCurrent := time.Now().Format("2006-01-02")

	prompt := fmt.Sprintf(`
		Tu salida que me tiene que dar es un commit que se entienda con lo que yo te pase, la salida que espero es:
		[PREFIJO] #TICKET DescripcionCommit FECHA

		Prefijo:
		- FIX: Para correcciones de errores
		- FEAT: Para nuevas funcionalidades
		- DOCS: Para cambios en la documentación
		- STYLE: Cambios de formato, tabulaciones, espacios o puntos y coma, etc; no afectan al usuario.
		- REFACTOR: Para un cambio en el código que no corrige un error ni agrega una función
		- TEST: Para agregar pruebas que faltaban
		- CHORE: Para cambios en el proceso de compilación o herramientas auxiliares y bibliotecas como la generación de documentación
		- PERF : Para mejoras de rendimiento
		- SECURITY: Para mejoras de seguridad
		- WIP: Para trabajo en progreso
		- RELEASE: Para versiones
		- REVERT: Para revertir a un commit anterior
		- BUILD: Para cambios que afectan el sistema de construcción o dependencias externas
		- CI: Para cambios en archivos y scripts de configuración de CI
		- DEPLOY: Para cambios en scripts y configuración de despliegue
		- DEVOPS: Para cambios en scripts y configuración de DevOps
		- DOCKER: Para cambios en scripts y configuración de Docker
		- K8S: Para cambios en scripts y configuración de Kubernetes
		- SWARM: Para cambios en scripts y configuración de Docker Swarm

		Este es una breve descripcion del commit y quiero que mejores ese comentario para que sea entendible y tenga sentido con lo que te pase.
		Descripcion: %s

		Necesito que sea una sola salida, no quiero que me des varias salidas, solo una salida con el prefijo correcto y la descripcion correcta.}
		Ejemplo:
		- [PREFIJO] #%s %s %s
		Quiero que me des 4 opciones de salida, si no me das las 4 opciones de salida, no se considera la tarea como completada.
	`, description, ticket, description, dateCurrent)

	jsonData := fmt.Sprintf(`{
		"contents": {
			"role": "user",
			"parts": {
				"text": "%s"
			}
		},
		"safety_settings": {
			"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
			"threshold": "BLOCK_LOW_AND_ABOVE"
		},
		"generation_config": {
			"temperature": 0.2,
			"topP": 0.8,
			"topK": 40
		}
	}`, prompt)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return fmt.Errorf("error creando la solicitud: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+strings.TrimSpace(util.GenerateToken()))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando la solicitud: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error leyendo el cuerpo de la respuesta: %v", err)
	}

	response := formatJSON(body)

	options := strings.Split(response, "\n")
	for _, option := range options {
		if strings.HasPrefix(option, "[") {
			fmt.Println(option)
		}
	}

	return nil
}

func main() {
	if err := MakeRequests(); err != nil {
		fmt.Println(err)
	}
}
