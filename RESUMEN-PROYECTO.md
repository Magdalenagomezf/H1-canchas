# H1-canchas — Full Stack

Sistema de reservas de canchas deportivas (pádel, fútbol) y quinchos/salones. Clientes reservan turnos online; recepcionistas gestionan reservas manualmente desde el local.

---

## Stack

### Backend
- **Go 1.25** + **Gin** (HTTP)
- **PostgreSQL** + **sqlx** (acceso a datos)
- **JWT** (golang-jwt/jwt v5) + **bcrypt** (auth)
- **godotenv** (.env)

### Frontend
- **React 19** + **TypeScript** + **Vite**
- **MUI v9** (Material UI) — tema verde/marrón personalizado
- **TanStack Query v5** — fetching y cache
- **React Router v7**
- **Framer Motion** — animaciones
- **Axios** — cliente HTTP

---

## Estructura del repositorio

```
H1-canchas/
├── backend/          # Go REST API
│   ├── cmd/
│   ├── config/
│   ├── internal/     # controllers, service, repository, domain, middleware, dto
│   ├── migrations/
│   ├── pkg/
│   ├── go.mod
│   └── .env          # gitignored
├── frontend/         # React + TypeScript
│   └── src/
│       ├── api/
│       ├── components/
│       ├── hooks/
│       ├── pages/
│       ├── theme/
│       └── types/
└── docker-compose.yml  # solo PostgreSQL
```

---

## Arquitectura backend

```
Repository → Service → Controller
```

Inyección de dependencias manual en `cmd/main.go`. Las interfaces las define el **service**, no el repository (Interface Segregation). Por ejemplo, `BookingService` usa `SpaceRepoForBooking` (solo `GetByID`) en vez de la interfaz completa de spaces.

---

## Entidades

### Users
- Roles: `customer`, `receptionist`, `admin`
- Login con teléfono + password. Registro público siempre crea `customer`. Staff lo crea un admin via `POST /admin/users`.

### Spaces
- Tipos: `cancha_padel`, `cancha_futbol`, `quincho`
- Se desactivan lógicamente (`is_active = false`), nunca se borran.
- El tipo no es editable — cambiar tipo con reservas históricas genera inconsistencias.

### SpaceSlots
- **Plantillas reutilizables**, no instancias por día. Se crean una vez y aplican a todos los días de la semana.
- Una reserva combina slot + fecha específica.
- Sin day_of_week: las canchas abren los mismos horarios todos los días (decisión intencional para este negocio).

### Bookings
- Une: espacio + slot + fecha + cliente.
- Cliente puede ser registrado (`customer_user_id`) o anónimo (nombre y teléfono, cargado por recepcionista).
- Estados: `pending`, `confirmed`, `cancelled`, `completed`.
- El precio se copia del espacio al crear la reserva — no cambia si el precio del espacio cambia después.

---

## Endpoints implementados

### Autenticación (públicos)
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| POST | `/auth/register` | Público | Crea usuario `customer`, devuelve `{ token, user }` |
| POST | `/auth/login` | Público | Login con teléfono + password, devuelve `{ token, user }` |

### Spaces
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| GET | `/spaces` | Público | Lista espacios activos |
| GET | `/spaces/:id` | Público | Detalle de un espacio |
| GET | `/spaces/:id/slots` | Público | Slots activos del espacio (sin disponibilidad por fecha) |
| POST | `/spaces` | Admin / Receptionist | Crear espacio |
| PUT | `/spaces/:id` | Admin / Receptionist | Editar nombre, descripción y precio |
| POST | `/spaces/:id/slots` | Admin | Agregar slot a un espacio |
| DELETE | `/spaces/:id` | Admin | Desactivar espacio |

### Bookings
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| GET | `/bookings` | Customer (propias) / Staff (todas) | Listado. Staff acepta `?date=YYYY-MM-DD` |
| GET | `/bookings/:id` | Owner / Staff | Detalle con espacio, slot y cliente |
| POST | `/bookings` | Customer | Crear reserva (sale en `pending`) |
| POST | `/bookings/manual` | Receptionist / Admin | Reserva manual sin cuenta (sale en `confirmed`) |
| PATCH | `/bookings/:id/cancel` | Owner / Staff | Cancelar reserva |

### Administración
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| POST | `/admin/users` | Admin | Crear usuario `receptionist` o `admin` |

---

## Flujo de estados de una reserva

```
pending ──► confirmed ──► completed
   │              │
   └──────────────┴──► cancelled
```

- **pending:** creada por el cliente desde la web.
- **confirmed:** reserva manual del recepcionista (directo). También destino del endpoint `confirm` (diferido).
- **completed:** turno finalizado (endpoint diferido).
- **cancelled:** cancelada por el cliente o el staff.

---

## Protección contra doble reserva

1. **Chequeo en el service** (`ExistsActiveBooking`) antes del INSERT.
2. **Índice único en PostgreSQL** (`unique_active_booking` sobre `space_id + slot_id + booking_date` para status `pending`/`confirmed`) como defensa final.

Si dos requests llegan al mismo tiempo, el primero inserta y el segundo falla en la constraint (409).

---

## Decisiones de diseño

- El tipo de espacio no es editable post-creación.
- Slots son plantillas sin day_of_week — las canchas abren igual todos los días.
- Registro público siempre es `customer`.
- Reservas manuales van directo a `confirmed`.
- `GET /bookings`: customer ve solo las suyas, staff ve todas (con filtro opcional por fecha).
- Todas las respuestas de reservas devuelven `BookingDetail` (JOIN con spaces, slots y users) — evita múltiples requests desde el frontend.
- Auth devuelve `{ token, user }` — el frontend no necesita decodificar el JWT para mostrar nombre y rol.

---

## Estado del frontend

### Páginas listas
| Ruta | Página | Estado |
|---|---|---|
| `/` | HomePage | ✅ Hero, grilla de espacios, "Cómo funciona", CTA |
| `/login` | LoginPage | ✅ Login + Registro en tabs, manejo de errores |
| `/canchas` | SpacesPage | ✅ Listado con filtros por tipo |
| `/canchas/:id` | SpaceDetailPage | ✅ Detalle, selector de fecha, grilla de slots, reserva |

### Páginas pendientes
| Ruta | Página | Notas |
|---|---|---|
| `/mis-reservas` | MyBookingsPage | Lista de reservas del usuario + cancelar |
| `/panel` | StaffPanel | Solo receptionist/admin: reservas del día, crear reserva manual |

### Componentes compartidos
- `Navbar` — con estado de auth, responsive (drawer en mobile)
- `SpaceCard` + `SpaceCardSkeleton` — usado en HomePage y SpacesPage

### Hooks
- `useAuth` — contexto global de autenticación, persiste en localStorage

---

## Bugs corregidos

1. Se podía reservar un espacio desactivado → el service verifica `space.IsActive` antes de crear.
2. Se podía reservar en fecha pasada → se valida que la fecha no sea anterior a hoy.
3. Desactivar un espacio ya inactivo devolvía 500 → el service verifica `IsActive` antes de llamar al repo.
4. CORS bloqueaba todas las requests del frontend → middleware CORS agregado en `main.go`.
5. Tipos del frontend no coincidían con respuestas del backend → `types/index.ts` corregido para usar estructura anidada (`space`, `slot`, `customer`).

---

## Cambios al backend durante fase frontend

- CORS middleware en `main.go` (permite `http://localhost:5173`).
- `/auth/login` y `/auth/register` ahora devuelven `{ token, user: { id, name, phone, role } }` además del JWT.
- `go.mod`, `go.sum`, `.env`, `.env.example` movidos a `backend/` (estructura monorepo correcta).

---

## Endpoints diferidos (esperan decisiones de negocio)

- `PATCH /bookings/:id/confirm` — pending → confirmed
- `PATCH /bookings/:id/complete` — confirmed → completed

Preguntas a resolver con el dueño del negocio antes de implementarlos:
- ¿Quién confirma? ¿El recepcionista manualmente o MercadoPago automáticamente al recibir el pago?
- ¿Se puede cancelar una reserva ya confirmada (y pagada)? ¿Hay reembolso?

---

## Pendientes técnicos

### Backend
1. **`GET /spaces/:id/slots?date=YYYY-MM-DD`** — actualmente ignora la fecha y devuelve todos los slots sin indicar disponibilidad. Cuando se implemente, devolver `is_available: bool` por slot cruzando contra reservas existentes.

2. **Tabla `space_blocks`** — bloquear una cancha en fecha puntual (feriado, mantenimiento). Campos: `space_id`, `slot_id` (nullable = bloquea el día completo), `block_date`, `reason`, `created_by`.

3. **Campo `payment_status` en bookings** — `no_pagada / pendiente_pago / pagada / reembolsada`. Se agrega al integrar MercadoPago.

4. **Paginación** en `GET /bookings` y `GET /spaces`.

### Frontend
1. **Rutas protegidas** — redirigir a `/login` si el usuario accede a `/mis-reservas` sin sesión.
2. **Redirigir si ya está logueado** — si un usuario autenticado va a `/login`, mandarlo a `/`.
3. **`/mis-reservas`** — lista de reservas del usuario con opción de cancelar.
4. **`/panel`** — panel de staff: ver todas las reservas del día, crear reserva manual.

---

## Limitaciones conocidas (post-MVP)

- **Disponibilidad por día de semana:** No hay soporte para "este slot no está disponible los domingos". Requeriría agregar `day_of_week` a slots o una tabla de schedules.
- **SELECT FOR UPDATE:** El chequeo de disponibilidad ocurre fuera de la transacción. Para alta concurrencia se debería hacer dentro del `tx`. Para el volumen actual del MVP es suficiente el índice único.
- **JWT no revocable:** Si se desactiva un usuario, su token sigue válido hasta que expire (24hs).
- **Sin refresh token:** El usuario queda deslogueado al expirar el JWT sin aviso.
