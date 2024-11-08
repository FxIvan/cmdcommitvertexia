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

	prompt := `
	Genera un commit claro y detallado según la convención proporcionada. El formato de salida esperado es:
	[PREFIJO] #TICKET Descripción FECHA

	Usa los siguientes prefijos:
	- FIX: Corrección de errores
	- FEAT: Nueva funcionalidad
	- DOCS: Cambios en documentación
	- STYLE: Formato (tabulación, espacios, etc.) sin afectar la funcionalidad
	- REFACTOR: Cambios en el código sin afectar funcionalidad
	- TEST: Adición de pruebas
	- CHORE: Cambios en herramientas o bibliotecas auxiliares
	- PERF: Mejora de rendimiento
	- SECURITY: Mejora de seguridad
	- WIP: Trabajo en progreso
	- RELEASE: Nueva versión
	- REVERT: Revertir un commit anterior
	- BUILD: Cambios en sistema de construcción o dependencias externas
	- CI: Cambios en configuración de CI
	- DEPLOY: Cambios en scripts de despliegue
	- DEVOPS: Cambios en scripts de DevOps
	- DOCKER: Cambios en scripts de Docker
	- K8S: Cambios en configuración de Kubernetes
	- SWARM: Cambios en Docker Swarm
`

	// Crear la descripción del usuario con el ticket y la fecha actuales
	descriptionUser := fmt.Sprintf(`
	Crea un commit según las buenas prácticas anteriores.
	Ticket: %s
	Descripción: %s
	Fecha actual: %s
`, ticket, description, dateCurrent)

	// Formatear la solicitud JSON con los datos relevantes y configuración de generación
	jsonData := fmt.Sprintf(`{
	"contents": [
		{
			"role": "model",
			"parts": { "text": "%s" }
		},
		{
			"role": "user",
			"parts": { "text": "%s" }
		}
	],
	"safety_settings": {
		"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
		"threshold": "BLOCK_LOW_AND_ABOVE"
	},
	"generation_config": {
		"temperature": 0.2,
		"topP": 0.8,
		"topK": 40
	}
}`, prompt, descriptionUser)

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
	fmt.Print("Response -->", response)
	options := strings.Split(response, "\n")
	for _, option := range options {
		// if strings.HasPrefix(option, "[") {
		fmt.Println(option)
		// }
	}

	return nil
}

func main() {
	if err := MakeRequests(); err != nil {
		fmt.Println(err)
	}
}
