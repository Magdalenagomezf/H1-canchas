package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // driver PostgreSQL, se registra solo al importar
)

// New abre la conexión a PostgreSQL y verifica que esté viva.
// Recibe el DSN armado por config. Devuelve *sqlx.DB o error.
func New(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error abriendo conexión: %w", err)
	}

	// Configuración del pool de conexiones.
	// MaxOpenConns: máximo de conexiones simultáneas a la BD.
	// MaxIdleConns: cuántas se mantienen abiertas cuando no hay actividad.
	// ConnMaxLifetime: cada cuánto se recicla una conexión.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping verifica que la BD esté realmente accesible.
	// sqlx.Open no conecta todavía, solo valida el DSN.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	return db, nil
}
