package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config agrupa toda la configuración de la app.
// Se carga una vez al arrancar y se pasa a quien la necesite.
type Config struct {
	Port          string
	DSN           string
	JWTSecret     string
	AllowedOrigin string

	MPAccessToken                     string
	MPWebhookSecret                   string
	MPWebhookURL                      string
	BookingHoldTTLMinutes             int
	DepositPercentage                 float64
	BookingExpirySweepIntervalMinutes int

	ResendAPIKey string // vacío = los mails solo se loguean
	EmailFrom    string
	FrontendURL  string
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

	mpAccessToken := mustGetEnv("MP_ACCESS_TOKEN")
	if mpAccessToken == "" {
		return nil, fmt.Errorf("MP_ACCESS_TOKEN es requerido")
	}

	mpWebhookSecret := mustGetEnv("MP_WEBHOOK_SECRET")
	if mpWebhookSecret == "" {
		return nil, fmt.Errorf("MP_WEBHOOK_SECRET es requerido")
	}

	mpWebhookURL := mustGetEnv("MP_WEBHOOK_URL")
	if mpWebhookURL == "" {
		return nil, fmt.Errorf("MP_WEBHOOK_URL es requerido")
	}

	bookingHoldTTLMinutes := getEnvInt("BOOKING_HOLD_TTL_MINUTES", 20)
	depositPercentage := getEnvFloat("DEPOSIT_PERCENTAGE", 0.15)
	bookingExpirySweepIntervalMinutes := getEnvInt("BOOKING_EXPIRY_SWEEP_INTERVAL_MINUTES", 2)

	resendAPIKey := getEnv("RESEND_API_KEY", "")
	emailFrom := getEnv("EMAIL_FROM", "H1 Canchas <onboarding@resend.dev>")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:5173")

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

		MPAccessToken:                     mpAccessToken,
		MPWebhookSecret:                   mpWebhookSecret,
		MPWebhookURL:                      mpWebhookURL,
		BookingHoldTTLMinutes:             bookingHoldTTLMinutes,
		DepositPercentage:                 depositPercentage,
		BookingExpirySweepIntervalMinutes: bookingExpirySweepIntervalMinutes,

		ResendAPIKey: resendAPIKey,
		EmailFrom:    emailFrom,
		FrontendURL:  frontendURL,
	}, nil
}

// getEnv devuelve el valor de la variable o el fallback si no existe.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt devuelve el valor entero de la variable o el fallback si no
// existe o no se puede parsear.
func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// getEnvFloat devuelve el valor float de la variable o el fallback si no
// existe o no se puede parsear.
func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

// mustGetEnv devuelve el valor o string vacío.
// La validación real se hace arriba en Load.
func mustGetEnv(key string) string {
	return os.Getenv(key)
}
