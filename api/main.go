package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LicaSterian/storage/api/db"
	"github.com/LicaSterian/storage/api/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	host    string
	port    string
	dbname  string
	apiPort string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host = os.Getenv("POSTGRES_HOST")
	port = os.Getenv("POSTGRES_PORT")
	dbname = os.Getenv("POSTGRES_DB")

	apiPort = fmt.Sprintf(":%s", os.Getenv("API_PORT"))

	psqlInfo := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=disable",
		host, port, dbname)
	sqlDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Panic(err)
	}
	defer sqlDB.Close()
	err = sqlDB.Ping()
	if err != nil {
		log.Panic(err)
	}

	storage := db.NewStorage(sqlDB)
	handlers := handlers.New(sqlDB, storage)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// return origin == "https://github.com"
		// },
		MaxAge: 12 * time.Hour,
	}))

	r.POST("/file/upload", handlers.PostUploadFile)
	r.POST("/files", handlers.GetAllFiles)
	r.GET("/file/:id", handlers.GetFile)
	r.DELETE("/file/:id", handlers.DeleteFile)

	log.Println("Listening on port:", apiPort)
	log.Panic(r.Run(apiPort))
}
