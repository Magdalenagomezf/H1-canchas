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
- **@fontsource-variable/archivo** + **@fontsource/jetbrains-mono** — tipografías de la landing y el login
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
- Se desactivan lógicamente (`is_active = false`). El borrado físico (`DELETE /admin/spaces/:id`) solo lo puede hacer un admin y **solo si el espacio no tiene ninguna reserva**; si tiene historial, únicamente se puede desactivar.
- Solo el **admin** puede crear espacios. El recepcionista puede editarlos (nombre, descripción, precio) pero no crearlos.
- El tipo no es editable — cambiar tipo con reservas históricas genera inconsistencias.

### SpaceSlots
- **Plantillas reutilizables**, no instancias por día. Se crean una vez y aplican a todos los días de la semana.
- Una reserva combina slot + fecha específica.
- Sin day_of_week: las canchas abren los mismos horarios todos los días (decisión intencional para este negocio).

### Bookings
- Une: espacio + slot + fecha + cliente.
- Cliente puede ser registrado (`customer_user_id`) o anónimo (nombre y teléfono, cargado por recepcionista).
- Estados: `pending`, `confirmed`, `cancelled`, `completed`, `expired`.
- El precio se copia del espacio al crear la reserva — no cambia si el precio del espacio cambia después.
- Cada reserva tiene dos tramos de pago independientes: `deposit_status` (seña) y `balance_status` (saldo). Ver "Integración de pagos".
- Puede pertenecer a un batch (`batch_id`) si fue generada por un turno fijo o un bloqueo de mantenimiento.

### BookingBatches
- Agrupan reservas creadas en bloque por el staff. Dos tipos:
  - `recurring_teacher` — **turno fijo semanal** (ej. un profesor todos los martes 18–19 hs).
  - `maintenance` — **bloqueo de mantenimiento** de un turno o de todos los turnos de un espacio durante un rango de días.
- Estados: `active`, `cancelled`.
- Rango máximo: 366 días. La fecha de inicio no puede ser anterior a hoy (horario de Argentina).

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
| POST | `/spaces` | Admin | Crear espacio |
| PUT | `/spaces/:id` | Admin / Receptionist | Editar nombre, descripción y precio |
| POST | `/spaces/:id/slots` | Admin | Agregar slot a un espacio |
| DELETE | `/spaces/:id` | Admin | Desactivar espacio (borrado lógico) |
| DELETE | `/admin/spaces/:id` | Admin | Borrado físico — falla si el espacio tiene reservas |

### Bookings
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| GET | `/bookings` | Cualquier usuario autenticado | Reservas propias del usuario |
| GET | `/admin/bookings` | Receptionist / Admin | Todas las reservas. Acepta `?date=YYYY-MM-DD` |
| GET | `/bookings/:id` | Owner / Staff | Detalle con espacio, slot y cliente |
| POST | `/bookings` | Customer | Crear reserva (sale en `pending` con hold de 20 min) |
| POST | `/bookings/manual` | Receptionist / Admin | Reserva manual sin cuenta (sale en `confirmed`) |
| PATCH | `/bookings/:id/cancel` | Owner / Staff | Cancelar reserva (el cliente solo mientras la seña no esté pagada) |

### Turnos fijos y bloqueos (booking batches)
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| POST | `/bookings/recurring` | Receptionist / Admin | Turno fijo semanal: crea una reserva `confirmed` por cada fecha del rango que cae en el día de la semana indicado |
| POST | `/bookings/block` | Receptionist / Admin | Bloqueo de mantenimiento: crea reservas `confirmed` a nombre "Mantenimiento" para un turno (o todos si no se indica) en cada día del rango. Motivo obligatorio |
| GET | `/bookings/batches` | Receptionist / Admin | Lista turnos fijos y bloqueos, filtrable por tipo |
| PATCH | `/bookings/batches/:id/cancel` | Receptionist / Admin | Cancela las reservas del batch **desde hoy en adelante**; las pasadas quedan como historial |

En `recurring` y `block`, las fechas que ya estaban ocupadas **se saltean** y se informan en la respuesta; no abortan el resto del batch.

### Pagos
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| POST | `/bookings/:id/payments` | Owner / Staff | Genera el link de Mercado Pago (Checkout Pro) para la seña o el saldo |
| PATCH | `/bookings/:id/payment` | Receptionist / Admin | Marca manualmente un tramo como pagado (efectivo / transferencia / posnet) |
| POST | `/payments/webhook` | Público (Mercado Pago) | Notificación de pago. Reconsulta el pago a la API de MP; un `x-signature` inválido solo loguea un WARN |

### Sistema
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| GET | `/health` | Público | Health check |

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
   │  │           │
   │  └───────────┴──► cancelled
   │
   └──► expired
```

- **pending:** creada por el cliente desde la web. Tiene un hold de 20 minutos para pagar la seña.
- **confirmed:** la seña se acreditó (vía webhook de Mercado Pago), o es una reserva manual del staff / generada por un batch (directo).
- **expired:** la seña no se pagó dentro del hold. El slot se libera. Un sweep en background (`main.go`) hace la transición periódicamente, y además el service expira las reservas vencidas de un slot antes de crear una nueva en ese mismo slot.
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
- `GET /bookings` devuelve siempre las reservas propias del usuario; el staff usa `GET /admin/bookings` para ver todas (con filtro opcional por fecha).
- Todas las respuestas de reservas devuelven `BookingDetail` (JOIN con spaces, slots y users) — evita múltiples requests desde el frontend.
- Auth devuelve `{ token, user }` — el frontend no necesita decodificar el JWT para mostrar nombre y rol.
- Solo el admin crea espacios (decisión de negocio: el recepcionista no puede dar de alta canchas).
- Turnos fijos y bloqueos **materializan reservas reales** en vez de guardar una regla recurrente evaluada en tiempo real. Así la protección contra doble reserva (índice único) funciona sin cambios y no hace falta un cron. El costo es el límite de 366 días por batch.

---

## Estado del frontend

### Páginas
| Ruta | Página | Estado |
|---|---|---|
| `/` | HomePage | ✅ Rediseñada (ver "Diseño de landing y login"): hero con línea LED, 01 El lugar (slideshow), 02 Espacios (índice tipográfico con datos de `GET /spaces`), 03 El complejo, 04 Cómo reservar, CTA final y footer |
| `/login` | LoginPage | ✅ Login + Registro en tabs, manejo de errores. Rediseñada con el estilo de la landing (la lógica no cambió) |
| `/canchas` | SpacesPage | ✅ Listado con filtros por tipo |
| `/canchas/:id` | SpaceDetailPage | ✅ Detalle, selector de fecha, grilla de slots disponibles, reserva, modal con resumen de la seña antes de ir a Mercado Pago |
| `/mis-reservas` | MyBookingsPage | ✅ Lista de reservas (activas / historial), cancelar con dialog de confirmación |
| `/pago/resultado/:bookingId` | PaymentResultPage | ✅ Página de retorno desde Mercado Pago con el estado del pago |
| `/panel` | StaffPanelPage | ✅ Panel receptionist/admin — reservas del día con navegación, reserva manual, turnos fijos y bloqueos, gestión de canchas y usuarios (admin) |

### Rutas protegidas
- `PrivateRoute` — redirige a `/login` si no hay sesión (usado en `/mis-reservas`)
- `PublicOnlyRoute` — redirige a `/` si ya está autenticado (usado en `/login`)
- `StaffRoute` — redirige a `/login` sin sesión, o a `/` si el rol es `customer` (usado en `/panel`)

### Componentes compartidos
- `Navbar` — con estado de auth, responsive (drawer en mobile). Se oculta en `/` y `/login`
- `components/landing/` — un componente por sección de la landing + `LandingNav` (nav propio de `/` y `/login`, con Mis reservas, Panel para staff y Salir)
- `SpaceCard` + `SpaceCardSkeleton` — usado en HomePage y SpacesPage

### Hooks
- `useAuth` — contexto global de autenticación, persiste en sessionStorage (aislamiento por tab)

### Diseño de landing y login
- Estilo inspirado en la arquitectura del complejo (Grupo Mazzucco): fondo oscuro (`night`), línea de luz cálida tipo LED como firma, señalética numerada `01/`, radius 0, sin sombras ni cajas.
- Tokens en `@theme` de `src/index.css` con nombres nuevos (`night`, `graphite`, `concrete`, `paper`, `court`, `dusk`, `light`). Radius 0 y fuentes aplicados solo dentro de `.landing`, así el resto de las rutas no cambia.
- Botones de la landing en azul `dusk` (variantes `court` y `line` en `button.tsx`).
- Textos, imágenes y placeholders editables en `src/components/landing/content.ts`.
- Gotcha: las animaciones de revelado con `clip-path` observan el marco exterior con `useInView`; un elemento 100% recortado nunca cuenta como visible y la imagen no aparece.

---

## Bugs corregidos

1. Se podía reservar un espacio desactivado → el service verifica `space.IsActive` antes de crear.
2. Se podía reservar en fecha pasada → se valida que la fecha no sea anterior a hoy.
3. Desactivar un espacio ya inactivo devolvía 500 → el service verifica `IsActive` antes de llamar al repo.
4. CORS bloqueaba todas las requests del frontend → middleware CORS agregado en `main.go`.
5. Tipos del frontend no coincidían con respuestas del backend → `types/index.ts` corregido para usar estructura anidada (`space`, `slot`, `customer`).
6. Logs de depuración imprimían partes del webhook secret de Mercado Pago y datos de la firma → eliminados. Se mantiene solo un `WARN` sin datos sensibles cuando la firma no coincide.

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
- **Webhook**: nunca se confía en el payload de la notificación — siempre se vuelve a consultar el pago contra la API de Mercado Pago, y esa reconsulta es la barrera de seguridad real. El header `x-signature` (HMAC-SHA256) se verifica, pero si no coincide solo se loguea un WARN y el procesamiento sigue: Checkout Pro envía notificaciones IPN legacy cuya firma nunca valida. El manejo es idempotente (Mercado Pago puede reenviar la misma notificación más de una vez).
- **Ambiente**: se arranca con credenciales sandbox de Mercado Pago; pasar a producción es solo cambiar variables de entorno, sin cambios de código. Testing del webhook en desarrollo vía túnel (ngrok/Cloudflare) apuntando al backend local, porque Mercado Pago no puede llamar a `localhost`.

---

## Pendientes técnicos

### Backend
1. **Paginación** en `GET /bookings`, `GET /admin/bookings` y `GET /spaces`.

Resueltos: bloqueo de canchas por fecha (implementado como booking batches de tipo `maintenance`, no como tabla `space_blocks`) y estado de pago (`deposit_status` / `balance_status`).

### Frontend (landing y login)
1. **Fotos faltantes**: `public/images/fachada-noche.jpg` (hero) y `public/images/locales.jpg` (El complejo). Sin ellas se ven bloques oscuros.
2. **Placeholders en `content.ts`**: número de WhatsApp (hoy falso), link de Google Maps, dirección y cantidades de `FACTS` (canchas y quinchos).
3. **Verificar mobile** (375px y 768px) y un **login/registro de punta a punta** con el diseño nuevo.
4. Menores de la revisión de código:
   - Desde `/login`, los links a secciones de la home (`/#espacios`, etc.) recargan la página completa. Se podría usar `<Link>` con `hash`.
   - `/login/` con barra final muestra los dos navs (el viejo y el nuevo). Normalizar el pathname en `App.tsx`.
   - En el menú mobile de la landing (Sheet en portal), el foco con teclado sale con el outline verde global en vez del claro.
   - Las tabs del login no tienen navegación con flechas del teclado (patrón ARIA tabs).

### Migraciones
Se aplican manualmente con `psql`, en orden: `001` → `006` (`005_add_payments`, `006_add_booking_batches`).

---

## Limitaciones conocidas (post-MVP)

- **Disponibilidad por día de semana:** No hay soporte para "este slot no está disponible los domingos". Requeriría agregar `day_of_week` a slots o una tabla de schedules.
- **SELECT FOR UPDATE:** El chequeo de disponibilidad ocurre fuera de la transacción. Para alta concurrencia se debería hacer dentro del `tx`. Para el volumen actual del MVP es suficiente el índice único.
- **JWT no revocable:** Si se desactiva un usuario, su token sigue válido hasta que expire (24hs).
- **Sin refresh token:** El usuario queda deslogueado al expirar el JWT sin aviso.
- **Creación de batches no transaccional:** Turnos fijos y bloqueos crean sus reservas una por una. Si la base de datos falla a mitad de camino, quedan creadas las reservas que alcanzaron a insertarse y el batch queda incompleto. Se puede cancelar el batch y volver a crearlo. Para hacerlo atómico habría que envolver la creación en una única transacción.
