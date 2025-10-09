package main

import (
	"fmt"
	"log"
	"main/internal/app"
	"main/internal/feeder"
	indicatorrsi "main/internal/indicator/rsi"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			panic(fmt.Sprintf("Error loading .env file: %v", err))
		}
		log.Println("No .env file found, using system environment variables")
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Vite dev server
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	feederType := strings.ToLower(os.Getenv("FEEDER"))
	if feederType == "" {
		feederType = "api"
	}

	var f feeder.Feeder

	if feederType == "api" {
		f = feeder.NewFeederApiCoinbase()
	} else if feederType == "json" {
		f = feeder.NewFeederJSONFile()
	} else {
		log.Fatalf("Unknown feeder type: %s", feederType)
	}

	app := app.NewApp(f)

	trendRSI, err := indicatorrsi.New(app)
	if err != nil {
		panic(fmt.Sprintf("trendRSI error %v", err))
	}
	trendRSI.Register(r)

	// Запуск сервера
	fmt.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
