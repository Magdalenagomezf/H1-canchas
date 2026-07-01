package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config agrupa toda la configuración de la app.
// Se carga una vez al arrancar y se pasa a quien la necesite.
type Config struct {
	Port          string
	DSN           string
	JWTSecret     string
	AllowedOrigin string
}

// Load lee el .env y arma la Config.
// Si no encuentra una variable obligatoria, devuelve error.
func Load() (*Config, error) {
	// godotenv.Load falla si no existe .env (útil en producción
	// donde las vars vienen del sistema operativo directamente)
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	allowedOrigin := getEnv("ALLOWED_ORIGIN", "http://localhost:5173")
	jwtSecret := mustGetEnv("JWT_SECRET")

	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET es requerido")
	}

	// Railway provee DATABASE_URL completa; en local se arma desde variables separadas.
	var dsn string
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	} else {
		host := mustGetEnv("DB_HOST")
		dbPort := mustGetEnv("DB_PORT")
		user := mustGetEnv("DB_USER")
		password := mustGetEnv("DB_PASSWORD")
		dbName := mustGetEnv("DB_NAME")
		if host == "" || dbPort == "" || user == "" || password == "" || dbName == "" {
			return nil, fmt.Errorf("faltan variables de entorno: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME")
		}
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, dbPort, user, password, dbName,
		)
	}

	return &Config{
		Port:          port,
		DSN:           dsn,
		JWTSecret:     jwtSecret,
		AllowedOrigin: allowedOrigin,
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
