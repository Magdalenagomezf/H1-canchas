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
- **Tailwind CSS v4** + utilidades shadcn/ui (clsx, tailwind-merge)
- **TanStack Query v5** — fetching y cache
- **React Router v7**
- **Framer Motion** — animaciones
- **lucide-react** — íconos
- **sonner** — toasts
- **vaul** — drawers
- **react-day-picker** + **date-fns** — selector de fechas
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
| GET | `/admin/users` | Admin | Listar todos los usuarios |
| PATCH | `/admin/users/:id/role` | Admin | Cambiar rol de un usuario |
| DELETE | `/admin/users/:id` | Admin | Eliminar usuario (no puede borrarse a sí mismo; falla si tiene reservas) |

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

### Páginas
| Ruta | Página | Estado |
|---|---|---|
| `/` | HomePage | ✅ Hero, grilla de espacios, "Cómo funciona", CTA |
| `/login` | LoginPage | ✅ Login + Registro en tabs, manejo de errores |
| `/canchas` | SpacesPage | ✅ Listado con filtros por tipo |
| `/canchas/:id` | SpaceDetailPage | ✅ Detalle, selector de fecha, grilla de slots disponibles, reserva |
| `/mis-reservas` | MyBookingsPage | ✅ Lista de reservas (activas / historial), cancelar con dialog de confirmación |
| `/panel` | StaffPanelPage | ✅ Panel receptionist/admin — 3 tabs (admin): reservas del día con navegación, reserva manual, gestión de canchas y usuarios |

### Rutas protegidas
- `PrivateRoute` — redirige a `/login` si no hay sesión (usado en `/mis-reservas`)
- `PublicOnlyRoute` — redirige a `/` si ya está autenticado (usado en `/login`)
- `StaffRoute` — redirige a `/login` sin sesión, o a `/` si el rol es `customer` (usado en `/panel`)

### Componentes compartidos
- `Navbar` — con estado de auth, responsive (drawer en mobile)
- `SpaceCard` + `SpaceCardSkeleton` — usado en HomePage y SpacesPage

### Hooks
- `useAuth` — contexto global de autenticación, persiste en sessionStorage (aislamiento por tab)

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

- `PATCH /bookings/:id/complete` — confirmed → completed (sigue diferido)

`PATCH /bookings/:id/confirm` ya no aplica tal cual: la confirmación de una reserva online pasa a ser automática cuando el webhook de Mercado Pago confirma el pago de la seña (ver sección siguiente). La pregunta de cancelación/reembolso ya se resolvió — ver "Integración de pagos".

---

## Integración de pagos — Mercado Pago (decisiones de negocio)

Decisiones tomadas con el dueño del negocio al planificar la integración de Checkout Pro. Documentadas acá para poder revisarlas/cambiarlas más adelante sin tener que reconstruir el contexto.

- **Modelo de pago en dos partes**: toda reserva se paga en dos tramos, nunca de una. Una **seña (`deposit`)** del **15% fijo** de `total_price` (por ahora fijo vía env var, no por espacio) que se cobra para confirmar la reserva, y un **saldo (`balance`)** con el 85% restante, que se cobra después. Ambos tramos son pagables **por Mercado Pago (link) o manualmente** (efectivo/transferencia/posnet) — no hay un tramo "solo online" y otro "solo manual". Cada booking tiene entonces `deposit_status` y `balance_status` independientes (no un único `payment_status`).
- **Hold de disponibilidad**: al crear una reserva online, el slot queda `pending` con un **hold de 20 minutos**. Si la seña no se paga en ese lapso, el slot se libera automáticamente (transición a estado `expired`, no revive el índice de disponibilidad).
- **Checkout**: redirect a Mercado Pago (Checkout Pro clásico con `init_point`), no Bricks embebido.
- **Reservas manuales del staff** (`CreateManual`, por teléfono o presencial): siguen confirmándose directo, **sin pasar por Mercado Pago automáticamente** — se asume que el staff ya coordinó con el cliente. El staff puede después, para esa misma reserva: (a) generar un link de pago de Mercado Pago (de la seña o del saldo), o (b) marcar manualmente cualquiera de los dos tramos como pagado, registrando el método (efectivo/transferencia/posnet).
- **Cancelación**: el cliente puede cancelar su propia reserva libremente mientras la **seña** no esté pagada (aunque el pago esté en curso/`pending`). Una vez que la seña se acreditó, **solo el staff puede cancelar** esa reserva. El reembolso, si corresponde, se gestiona **manualmente fuera del sistema** (dashboard de MP o devolución directa) — la v1 no llama a la API de reembolsos de Mercado Pago.
- **Caso límite — pago tardío**: si el pago de la seña se acredita después de que el slot ya expiró (y posiblemente fue tomado por otra persona), el sistema **no revive la reserva vencida**, pero sí registra el pago como recibido para que el staff lo resuelva a mano (reembolso o reubicación del cliente).
- **Webhook**: nunca se confía en el payload de la notificación — siempre se vuelve a consultar el pago contra la API de Mercado Pago, y se valida el header `x-signature` (HMAC-SHA256). El manejo es idempotente (Mercado Pago puede reenviar la misma notificación más de una vez).
- **Ambiente**: se arranca con credenciales sandbox de Mercado Pago; pasar a producción es solo cambiar variables de entorno, sin cambios de código. Testing del webhook en desarrollo vía túnel (ngrok/Cloudflare) apuntando al backend local, porque Mercado Pago no puede llamar a `localhost`.

---

## Pendientes técnicos

### Backend
1. **Tabla `space_blocks`** — bloquear una cancha en fecha puntual (feriado, mantenimiento). Campos: `space_id`, `slot_id` (nullable = bloquea el día completo), `block_date`, `reason`, `created_by`.

2. ~~Campo `payment_status` en bookings~~ — en implementación como parte de la integración de Mercado Pago (ver sección "Integración de pagos"), con `deposit_status`/`balance_status` en vez de un único campo.

3. **Paginación** en `GET /bookings` y `GET /spaces`.

---

## Limitaciones conocidas (post-MVP)

- **Disponibilidad por día de semana:** No hay soporte para "este slot no está disponible los domingos". Requeriría agregar `day_of_week` a slots o una tabla de schedules.
- **SELECT FOR UPDATE:** El chequeo de disponibilidad ocurre fuera de la transacción. Para alta concurrencia se debería hacer dentro del `tx`. Para el volumen actual del MVP es suficiente el índice único.
- **JWT no revocable:** Si se desactiva un usuario, su token sigue válido hasta que expire (24hs).
- **Sin refresh token:** El usuario queda deslogueado al expirar el JWT sin aviso.
