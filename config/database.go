package config

import (
	"fmt"
	"log"
	"os"

	"github.com/JAreyes98/healthconnect-storage-service/internal/model"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	err := godotenv.Overload()
	if err != nil {
		log.Println("Can not load .env file, trying relative route...")
		godotenv.Overload("../../.env")
	}

	// 2. Leer variables con fallback para evitar punteros a MySQL
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// 3. Construir DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	log.Printf("Connecting to  Postgres: %s:%s -> DB:%s", host, port, dbname)

	// 4. Conectar
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: false})
	if err != nil {
		log.Fatalf("Error crítico: No se pudo conectar a Postgres. DSN usado: %s. Error: %v", dsn, err)
	}

	// 5. Migraciones
	db.AutoMigrate(
		&model.App{},
		&model.Bucket{},
		&model.FileMetadata{},
		&model.ReplicationRule{},
	)

	return db
}
