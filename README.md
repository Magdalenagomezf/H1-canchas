# H1-canchas

Backend REST API en Go para la gestión de reservas de canchas deportivas (pádel, fútbol) y quinchos/salones.

## Stack

- Go 1.25 + Gin
- PostgreSQL + sqlx
- JWT (HS256) + bcrypt
- Docker (solo para la base de datos)

## Requisitos

- Go 1.25+
- Docker Desktop
- psql (para aplicar migraciones)

## Instalación

```bash
# 1. Clonar el repositorio
git clone https://github.com/Magdalenagomezf/H1-canchas.git
cd H1-canchas

# 2. Configurar variables de entorno
cp .env.example .env
# Editar .env con los valores correspondientes

# 3. Levantar la base de datos
docker-compose up -d

# 4. Aplicar migraciones (en orden)
Get-Content .\migrations\001_create_users.sql      | docker exec -i h1_canchas_postgres psql -U postgres -d h1_canchas
Get-Content .\migrations\002_create_spaces.sql     | docker exec -i h1_canchas_postgres psql -U postgres -d h1_canchas
Get-Content .\migrations\003_create_space_slots.sql | docker exec -i h1_canchas_postgres psql -U postgres -d h1_canchas
Get-Content .\migrations\004_create_bookings.sql   | docker exec -i h1_canchas_postgres psql -U postgres -d h1_canchas

# 5. Correr el servidor
go run ./cmd/main.go
```

El servidor queda disponible en `http://localhost:8080`.

## Variables de entorno

Copiar `.env.example` a `.env` y completar:

| Variable | Descripción |
|---|---|
| `DB_HOST` | Host de PostgreSQL (ej: `localhost`) |
| `DB_PORT` | Puerto (ej: `5432`) |
| `DB_USER` | Usuario de PostgreSQL |
| `DB_PASSWORD` | Contraseña de PostgreSQL |
| `DB_NAME` | Nombre de la base de datos |
| `JWT_SECRET` | Clave secreta para firmar tokens JWT |
| `PORT` | Puerto del servidor (default: `8080`) |

## Endpoints principales

| Método | Ruta | Acceso |
|---|---|---|
| POST | `/auth/register` | Público |
| POST | `/auth/login` | Público |
| GET | `/spaces` | Público |
| GET | `/spaces/:id/slots` | Público |
| POST | `/bookings` | Customer |
| GET | `/bookings` | Customer / Staff |
| PATCH | `/bookings/:id/cancel` | Owner / Staff |
| POST | `/bookings/manual` | Receptionist / Admin |
| POST | `/spaces` | Receptionist / Admin |
| PUT | `/spaces/:id` | Receptionist / Admin |
| POST | `/spaces/:id/slots` | Admin |
| DELETE | `/spaces/:id` | Admin |
| POST | `/admin/users` | Admin |

## Arquitectura

El proyecto sigue una arquitectura de 3 capas con inyección de dependencias manual:

```
Controller → Service → Repository → PostgreSQL
```

Las interfaces las define el **service** (no el repository), lo que permite testear la lógica de negocio de forma aislada. Los errores de dominio están centralizados en `internal/service/errors.go`.

## Estructura del proyecto

```
cmd/            # Entry point (main.go)
config/         # Lectura de variables de entorno
internal/
  domain/       # Structs del negocio (User, Space, Booking, etc.)
  repository/   # Acceso a PostgreSQL (SQL puro con sqlx)
  service/      # Lógica de negocio y validaciones
  controllers/  # Handlers HTTP (Gin)
  dto/          # Structs de request/response
  middleware/   # Auth JWT y control de roles
migrations/     # Archivos SQL versionados (001 → 004)
pkg/
  database/     # Conexión a PostgreSQL
  jwt.go        # Generación y validación de tokens
```
