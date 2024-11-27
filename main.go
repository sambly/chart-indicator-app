package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var entryJS string

func main() {

	godotenv.Load(".env")
	mode, _ := os.LookupEnv("ENVIRONMENT")
	if mode == "" {
		mode = "production"
	}

	r := gin.Default()
	r.LoadHTMLGlob("frontend/templates/*.tmpl")

	if mode == "production" {
		// Production: Загружаем путь к JS-файлу из манифеста
		manifest, err := readManifest()
		if err != nil {
			panic(fmt.Sprintf("Ошибка при загрузке манифеста: %v", err))
		}

		entry, ok := manifest["src/main.ts"].(map[string]interface{})
		if !ok {
			panic("Ошибка: 'src/main.ts' не найден в манифесте")
		}

		entryJS, ok = entry["file"].(string)
		if !ok {
			panic("Ошибка: ключ 'file' отсутствует в записи 'src/main.ts'")
		}

	} else {
		// Development: Используем Vite Dev Server
		entryJS = "http://localhost:5173/src/main.ts" // Vite dev server
	}

	if mode == "production" {
		r.Static("/assets", "./frontend/dist/assets")
	}

	r.GET("/", getChart)

	r.Run(":8080")

}

func readManifest() (map[string]interface{}, error) {
	manifestPath := "./frontend/dist/.vite/manifest.json"
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть %s: %w", manifestPath, err)
	}
	defer file.Close()

	var manifest map[string]interface{}
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("ошибка при разборе manifest.json: %w", err)
	}
	return manifest, nil
}
