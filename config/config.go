package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config agrupa toda la configuración de la app.
// Se carga una vez al arrancar y se pasa a quien la necesite.
type Config struct {
	Port string
	DSN  string // connection string para PostgreSQL
	JWTSecret string
}

// Load lee el .env y arma la Config.
// Si no encuentra una variable obligatoria, devuelve error.
func Load() (*Config, error) {
	// godotenv.Load no falla si no existe .env (útil en producción
	// donde las vars vienen del sistema operativo directamente)
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")

	host := mustGetEnv("DB_HOST")
	dbPort := mustGetEnv("DB_PORT")
	user := mustGetEnv("DB_USER")
	password := mustGetEnv("DB_PASSWORD")
	dbName := mustGetEnv("DB_NAME")
	jwtSecret := mustGetEnv("JWT_SECRET")

	if host == "" || dbPort == "" || user == "" || password == "" || dbName == "" || jwtSecret == "" {
		return nil, fmt.Errorf("faltan variables de entorno de base de datos")
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, dbPort, user, password, dbName,
	)

	return &Config{
		Port: port,
		DSN:  dsn,
		JWTSecret: jwtSecret,
	}, nil
}

// getEnv devuelve el valor de la variable o el fallback si no existe.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// mustGetEnv devuelve el valor o string vacío.
// La validación real se hace arriba en Load.
func mustGetEnv(key string) string {
	return os.Getenv(key)
}
