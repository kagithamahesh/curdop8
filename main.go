package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Define a model for GORM
type User struct {
	gorm.Model
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email" gorm:"unique"`
}

var DB *gorm.DB

func main() {
	// Load environment variables from .env file for local development
	if os.Getenv("INSTANCE_CONNECTION_NAME") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// Use environment variables for Cloud SQL connection
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	instanceConnectionName := os.Getenv("INSTANCE_CONNECTION_NAME")

	var dsn string
	if instanceConnectionName != "" {
		// DSN for Cloud Run (using Unix socket)
		dsn = fmt.Sprintf("%s:%s@unix(/cloudsql/%s)/%s?parseTime=true", dbUser, dbPass, instanceConnectionName, dbName)
	} else {
		// DSN for local development (using TCP)
		dsn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser, dbPass, dbName)
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(&User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from Cloud Run!"})
	})

	router.GET("/users", listUsers)
	router.POST("/users", createUser)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func listUsers(c *gin.Context) {
	var users []User
	DB.Find(&users)
	c.JSON(200, users)
}

func createUser(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	DB.Create(&newUser)
	c.JSON(201, newUser)
}
