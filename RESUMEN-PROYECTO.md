# H1-canchas — Full Stack

Sistema de reservas de canchas deportivas (pádel, fútbol, padbol, beach vóley) y quinchos/salones. Clientes reservan turnos online; recepcionistas gestionan reservas manualmente desde el local.

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
- **Datos de contacto normalizados (migración 008):**
  - Teléfono: se guarda en E.164 (`+549…`) vía `pkg/phone.Normalize` (región por defecto `AR`, lib `nyaruka/phonenumbers`). `3834123456`, `0383 15-412-3456` y `+54 9 383 412 3456` dan el mismo `+5493834123456`. La lib no agrega el `9` de móvil en formato local, así que `Normalize` lo agrega siempre para números argentinos. Se aplica en `Register`, `CreateStaff`, `Login` (un teléfono que no se puede normalizar da `credenciales incorrectas`, no se revela) y en `customer_phone` de las reservas manuales y turnos fijos. Los bloqueos de mantenimiento siguen guardando `-` como teléfono (no es un cliente real).
  - Email: `trim` + `lower`, vacío = `NULL`, formato validado con `net/mail`. Único sin distinguir mayúsculas (índice `users_email_lower_key` sobre `lower(email)`).
  - **Email obligatorio en el registro público.** En `POST /admin/users` (staff) sigue siendo opcional. Los usuarios existentes quedan sin email.
  - `UserResponse` (respuesta de auth y `GET /admin/users`) incluye `email` (se omite si es `null`).
  - Errores nuevos: `ErrInvalidPhone` (400), `ErrInvalidEmail` (400), `ErrEmailRequired` (400), `ErrEmailAlreadyExists` (409). `POST /auth/register` ahora responde 400 a los errores de validación (antes devolvía 500).

### Spaces
- Tipos: `cancha_padel`, `cancha_futbol`, `cancha_padbol`, `cancha_beach_voley`, `quincho`
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

- **Login con Google y "Mi perfil" (Fase 4a, migración 010):**
  - `users.phone`, `users.password_hash` y `users.google_sub` (`VARCHAR(255) UNIQUE`) son nullables; `domain.User` usa `*string` para los tres. `CHECK users_login_method_chk`: toda cuenta tiene contraseña o `google_sub`.
  - `UserResponse.phone` es `*string` **sin** `omitempty`: el frontend recibe `null`, no la ausencia del campo.
  - `POST /auth/google` (orden): verificar token (firma, `aud`, `exp` con `idtoken.Validate` + chequeo manual de `iss`) y `email_verified` → `FindByGoogleSub` → si no existe, `FindByEmail`: rol distinto de `customer` = 403 (no se vincula), `customer` con otro `google_sub` = 409 (no se pisa), `customer` sin `google_sub` = se vincula → si no existe, se crea `customer` con `phone` y `password_hash` en `NULL` (nombre de Google cortado a 100 caracteres, o lo anterior a la `@`). Un INSERT concurrente que choca por unique da `ErrEmailAlreadyExists` (409) sin reintentar. Al final se chequea `is_active` y se devuelve el mismo `AuthResponse` que `Login`.
  - Mensajes de `POST /auth/google` pensados para el usuario final: token rechazado = "No pudimos iniciar sesión con Google. Recargá la página y probá de nuevo." (la causa técnica se loguea como `WARN GoogleLogin rejected token`, nunca se muestra); `email_verified=false` = "Tu cuenta de Google no tiene el email verificado. Verificalo en Google o registrate con tu teléfono." (`ErrGoogleEmailNotVerified`, 401).
  - `Login` con teléfono sobre una cuenta sin contraseña da `credenciales incorrectas` (no revela que es de Google).
  - `PATCH /me`: el email y el teléfono no se pueden borrar; se normalizan igual que en el registro; se chequea que no los use otro usuario y además se mapea la violación de unique del `UPDATE`. Cambiar el email de una cuenta de Google está permitido (el login busca por `sub`). No cambia `role` ni contraseña.
  - **Reservar online exige teléfono:** `BookingService.Create` devuelve `ErrProfilePhoneMissing` ("Cargá tu teléfono en Mi perfil para poder reservar") y el controller responde **422** `{ "error": ..., "code": "phone_required" }`. Es un error distinto de `ErrPhoneRequired` (400, "falta en el formulario").
  - `GOOGLE_CLIENT_ID` es opcional: si está vacío, `POST /auth/google` responde 503 ("El inicio de sesión con Google no está configurado") y el resto de la app funciona.
  - `cmd/normalize-phones` ahora ignora usuarios con `phone IS NULL`.

## Endpoints implementados

### Autenticación (públicos)
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| POST | `/auth/register` | Público | Crea usuario `customer` (email obligatorio), devuelve `{ token, user }` |
| POST | `/auth/login` | Público | Login con teléfono + password, devuelve `{ token, user }` |
| POST | `/auth/password-reset` | Público | Body `{ email }`. Siempre 202 con el mismo mensaje, exista o no la cuenta; manda el mail con el link en segundo plano |
| POST | `/auth/password-reset/confirm` | Público | Body `{ token, password }`. 204 si cambió la contraseña; 400 con `code: "invalid_reset_token"` si el link no sirve (inexistente, usado, vencido o usuario inactivo); 400 si la contraseña no cumple la regla |
| POST | `/auth/password-reset/check` | Público | Body `{ token }`. 204 si el link todavía sirve; 400 con `code: "invalid_reset_token"` si no. No gasta el token. La pantalla lo llama al abrir, así un link usado o vencido muestra el aviso antes del formulario |
| POST | `/auth/google` | Público | Body `{ credential }` (ID token de Google). Login/registro/vinculación, devuelve `{ token, user }`. 401 token inválido o `email_verified=false` o usuario inactivo, 403 cuenta de staff, 409 conflicto de cuenta, 503 si falta `GOOGLE_CLIENT_ID` |

### Mi perfil (cualquier usuario autenticado, staff incluido)
| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| GET | `/me` | Autenticado | `UserResponse` del usuario del token |
| PATCH | `/me` | Autenticado | Body `{ name?, email?, phone? }`; campo ausente = no se toca. Devuelve el `UserResponse` actualizado. 400 validación, 409 email/teléfono de otro usuario |

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

### Restablecer contraseña (E5, 2026-11-13)
- **El link lleva un token de un solo uso**: 32 bytes de `crypto/rand`. La tabla `password_reset_tokens` guarda solo su SHA-256, así que si se filtra la base no hay links que sirvan. Vence en **1 hora** y cada usuario recibe como mucho uno cada **2 minutos**.
- **La respuesta no revela si el email está registrado**: siempre 202 con el mismo mensaje, y el mail sale en segundo plano para que tampoco lo delate el tiempo de respuesta.
- **El token viaja en el fragmento** (`/restablecer-contrasena#token=...`): el navegador no lo manda al servidor, así que no queda en logs ni en el header Referer. La página lo saca de la barra de direcciones apenas lo lee.
- **La pantalla comprueba el link apenas abre** (`/auth/password-reset/check`), con las mismas condiciones que el canje. Un link ya usado o vencido muestra el aviso enseguida, sin hacer llenar el formulario para nada (lo encontró la usuaria probando, 2026-11-13).
- **Al usarlo se invalidan todos los tokens pendientes** del usuario, en la misma transacción que cambia la contraseña.
- Funciona también para cuentas creadas solo con Google: con esto pasan a tener contraseña (para entrar con teléfono igual necesitan cargar el teléfono).
- **Limitación aceptada**: las sesiones abiertas (JWT) siguen valiendo hasta que vencen (24 h).
- Pendiente para E8: el límite de 2 minutos no es atómico ante dos pedidos simultáneos, y si falla el envío del mail el límite corre igual.
- Contraseña: mínimo 6 caracteres y **máximo 72 bytes** (límite de bcrypt; antes un valor más largo daba 500 en el registro). La misma regla vale para registro, staff y restablecer.

### Épica "Mail y cuentas" (cerrada el 2026-11-09)
- **Sin mail en reservas manuales** (era la Fase 2; se descartó el 2026-10-02). En una reserva manual el cliente ya arregló todo con el recepcionista, casi siempre por WhatsApp o teléfono, así que un mail de confirmación no le suma nada. Pedirle el email haría más lenta la carga y agregaría errores de tipeo. Si en el futuro hace falta (por ejemplo, para los recordatorios de "Avisos al cliente"), se agrega una columna `customer_email` en `bookings` que acepte `NULL`, sin tocar los datos existentes.
- **Los clientes cargados a mano no se convierten en usuarios.** Siguen como nombre y teléfono en la reserva.
- **No se permite "reclamar" reservas manuales por teléfono** hasta que exista verificación por SMS.
- **Riesgo aceptado (revisar más adelante): vinculación automática de Google por email.** El email del registro con teléfono no se verifica. Alguien puede registrarse con el email de otra persona y su propia contraseña; cuando la persona real entra con Google, la cuenta se vincula y el atacante, que sigue sabiendo la contraseña, ve sus reservas. Opciones para cerrarlo: (A) al vincular, borrar `password_hash`; (B) no vincular automáticamente y hacerlo desde "Mi perfil" con la sesión iniciada; (C) verificar el email al registrarse. Se acepta por ahora por el bajo valor de los datos expuestos.

---

## Estado del frontend

### Páginas
| Ruta | Página | Estado |
|---|---|---|
| `/` | HomePage | ✅ Rediseñada (ver "Diseño de landing y login"): hero con línea LED, 01 El lugar (frase + slideshow de 3 fotos cada 3,5 s), 02 Espacios (índice tipográfico con datos de `GET /spaces`), 03 El complejo, 04 Cómo reservar, CTA final y footer |
| `/login` | LoginPage | ✅ Login + Registro en tabs, manejo de errores. Rediseñada con el estilo de la landing (la lógica no cambió) |
| `/recuperar-contrasena` | ForgotPasswordPage | ✅ Pide el email y muestra siempre la misma confirmación. Se llega desde "¿Olvidaste tu contraseña?" en el login |
| `/restablecer-contrasena` | ResetPasswordPage | ✅ Abre desde el link del mail: lee el token del fragmento, pide la contraseña nueva dos veces y muestra el éxito con link al login, o el error de link inválido con link para pedir uno nuevo |
| `/canchas` | SpacesPage | ✅ Rediseñada. Filtros por tipo (TODOS + 5 tipos). En TODOS muestra una grilla compacta de 3 columnas (`SpacesGrid`); con un filtro activo muestra el índice tipográfico grande (`SpacesIndex`) |
| `/canchas/:id` | SpaceDetailPage | ✅ Detalle, selector de fecha, grilla de slots disponibles, reserva, modal con resumen de la seña antes de ir a Mercado Pago. Rediseñada: galería, pasos "01/ Fecha" y "02/ Turno", resumen fijo (barra inferior en mobile). Lógica de reserva y pago sin cambios |
| `/mis-reservas` | MyBookingsPage | ✅ Lista de reservas (activas / historial), cancelar con dialog de confirmación. Rediseñada: filas con hairlines y estados en mono |
| `/pago/resultado/:bookingId` | PaymentResultPage | ✅ Página de retorno desde Mercado Pago con el estado del pago. Rediseñada: título grande según el estado, resumen de la reserva y CTAs |
| `/panel` | StaffPanelPage | ✅ Panel receptionist/admin — reservas del día con navegación, reserva manual, turnos fijos y bloqueos, gestión de canchas y usuarios (admin). Rediseñada: tabs en mono, tablas sin bordes con scroll horizontal propio en mobile, modales con el estilo nuevo |

### Rutas protegidas
- `PrivateRoute` — redirige a `/login` si no hay sesión (usado en `/mis-reservas`)
- `PublicOnlyRoute` — redirige a `/` si ya está autenticado (usado en `/login`)
- `StaffRoute` — redirige a `/login` sin sesión, o a `/` si el rol es `customer` (usado en `/panel`)

### Componentes compartidos
- `Navbar` — el nav viejo. Ya no se muestra en ninguna ruta (todas usan `LandingNav`); solo aparecería en una URL inexistente. Candidato a eliminar
- `components/landing/` — un componente por sección de la landing, `LandingNav` (nav de todas las rutas, con Mis reservas, Panel para staff y Salir), `SpacesIndex` (índice tipográfico compartido por la home y `/canchas`), `SpacesGrid` (grilla compacta de `/canchas`) y `spaceHelpers.ts`
- `App.tsx` — `LANDING_PATHS` + `matchPath` decide qué rutas ocultan el `Navbar` viejo
- `SpaceCard` + `SpaceCardSkeleton` — usado en HomePage y SpacesPage

### Hooks
- `useAuth` — contexto global de autenticación, persiste en sessionStorage (aislamiento por tab)

### Diseño del frontend (todas las páginas)
- Estilo inspirado en la arquitectura del complejo (Grupo Mazzucco): fondo oscuro (`night`), línea de luz cálida tipo LED como firma, señalética numerada `01/`, radius 0, sin sombras ni cajas.
- Tokens en `@theme` de `src/index.css` con nombres nuevos (`night`, `graphite`, `concrete`, `paper`, `court`, `dusk`, `light`). Radius 0 y fuentes aplicados solo dentro de `.landing`, así el resto de las rutas no cambia.
- Botones de la landing en azul `dusk` (variantes `court` y `line` en `button.tsx`).
- Textos, imágenes y placeholders editables en `src/components/landing/content.ts`.
- Cada página nueva usa el mismo shell: `<div className="landing">` + wrapper `on-dark`, `LandingNav` (modo `solid` fuera de la home), línea LED arriba y `LandingFooter`.
- Gotcha: `ModalPortal` renderiza en `document.body`, fuera de `.landing`. El contenido de cada modal necesita su propio wrapper `landing on-dark` para heredar fuentes, radius 0 y foco.
- Gotcha: nunca poner `animate-fade-up` ni un `transform` permanente en contenedores de página: rompe el `position: fixed` del nav y de los modales.
- Gotcha: el botón de pago de Mercado Pago tiene que seguir siendo un `<a href>` real (deep link en mobile).
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
- `/auth/login` y `/auth/register` ahora devuelven `{ token, user: { id, name, phone, email?, role } }` además del JWT.
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

## Mail de confirmación de reserva (Fase 3)

- **Cuándo se envía**: solo cuando el webhook de Mercado Pago acredita la **seña** (`deposit`) y `ConfirmAfterPayment` devolvió `true` (la reserva no estaba cancelada/expirada). Un pago de saldo, un pago rechazado o un commit fallido no envían nada.
- **Después del commit y sin bloquear**: `PaymentService` llama a `NotifyBookingConfirmed` recién tras el commit, y este corre en una goroutine con su propio contexto (timeout 15 s), nunca el del request. Si el mail falla, la reserva sigue confirmada y el webhook responde OK; solo se loguea un WARN.
- **Se envía una sola vez (claim)**: `ClaimConfirmationEmail` hace `UPDATE ... SET confirmation_email_sent_at = now() WHERE id = $1 AND confirmation_email_sent_at IS NULL` y solo quien afecta 1 fila envía. Los reintentos del webhook de MP no duplican el mail. Si el envío falla se libera el claim (`ReleaseConfirmationEmail`) para poder reintentar. También se libera si el propio claim devuelve error (el `UPDATE` pudo aplicarse aunque se perdió la respuesta).
- **Idempotency key en Resend**: cada envío lleva `Idempotency-Key: booking-confirmation-<id>`. Si el request vence después de que Resend aceptó el mail, el claim se libera y el próximo reintento del webhook no genera un duplicado.
- **Riesgo aceptado**: si el proceso muere (deploy, crash) entre el claim y el envío, ese mail se pierde. Para este volumen se acepta; resolverlo requiere un estado intermedio o una cola.
- **Sin destinatario no se reclama**: reservas manuales del staff y usuarios sin email no envían nada ni consumen el claim.
- **El staff que marca la seña como pagada a mano NO dispara el mail**: `MarkPaidManually` no pasa por el webhook; es una decisión de esta fase.
- **Proveedor**: Resend (`POST https://api.resend.com/emails`, `pkg/email`). Sin `RESEND_API_KEY` se usa `LogSender`, que solo loguea destinatario y asunto (nunca el cuerpo). Variables: `RESEND_API_KEY`, `EMAIL_FROM`, `FRONTEND_URL` (link a `/mis-reservas`).
- **Fecha en español sin conversión de zona**: `booking_date` es un `DATE` que se escanea como medianoche UTC; convertirla a hora de Argentina la corre al día anterior. Se arma el texto con Year/Month/Day/Weekday del propio valor y tablas de nombres en español.
- **Templates**: `internal/service/templates/booking_confirmation.{html,txt}`, embebidos con `go:embed`.

---

## Pendientes técnicos

### Backend
1. **Paginación** en `GET /bookings`, `GET /admin/bookings` y `GET /spaces`.

Resueltos: bloqueo de canchas por fecha (implementado como booking batches de tipo `maintenance`, no como tabla `space_blocks`) y estado de pago (`deposit_status` / `balance_status`).

### Frontend (rediseño)
1. **Fotos faltantes**: `public/images/fachada-noche.jpg` (hero) y `public/images/locales.jpg` (El complejo). Sin ellas se ven bloques oscuros.
2. **Placeholders en `content.ts`**: número de WhatsApp (hoy falso), link de Google Maps, dirección y cantidades de `FACTS` (canchas y quinchos).
3. **Verificar mobile** (375px y 768px) y un **login/registro de punta a punta** con el diseño nuevo.
3b. **Probar el panel a mano** (reserva manual, turno fijo, bloqueo, cancelar) y el flujo de reserva hasta el modal de pago: el rediseño de esas páginas no se revisó visualmente.
3c. **Mercado Pago en producción**: `SpaceDetailPage` usa `sandbox_init_point` (URL de pruebas). En producción debería usar `init_point`; es un cambio de lógica de pago, pendiente de decidir.
3d. **Eliminar el `Navbar` viejo** (`components/Navbar.tsx`) o usarlo como fallback de una página 404.
3e. **Más fotos** para padbol y beach vóley (hoy 1 cada uno en `public/spaces/`).
4. Menores de la revisión de código:
   - Desde `/login`, los links a secciones de la home (`/#espacios`, etc.) recargan la página completa. Se podría usar `<Link>` con `hash`.
   - `/login/` con barra final muestra los dos navs (el viejo y el nuevo). Normalizar el pathname en `App.tsx`.
   - En el menú mobile de la landing (Sheet en portal), el foco con teclado sale con el outline verde global en vez del claro.
   - Las tabs del login no tienen navegación con flechas del teclado (patrón ARIA tabs).

### Próxima épica: "Avisos al cliente"
Recordatorios de la reserva y mails de cancelación y de vencimiento del hold. Reusa el `EmailSender` de `pkg/email` (el mismo del mail de confirmación).

### Migraciones
Se aplican manualmente con `psql`, en orden: `001` → `011` (`005_add_payments`, `006_add_booking_batches`, `007_add_padbol_beach_voley_types`, `008_normalize_user_email`, `009_add_confirmation_email_sent_at`, `010_google_login`, `011_password_reset_tokens`).

`011` crea la tabla `password_reset_tokens` (aditiva, no toca datos existentes). Hay que aplicarla **antes** de deployar el backend de restablecer contraseña: sin la tabla, pedir el link falla (el pedido responde 202 igual, pero el mail no sale) y confirmar responde 500.

`010` (Fase 4a) hace nullables `users.password_hash` y `users.phone`, agrega `users.google_sub VARCHAR(255) NULL UNIQUE` y el `CHECK users_login_method_chk` (contraseña o `google_sub`). Funciona con datos existentes (todas las filas tienen `password_hash`). Va en una transacción; correrla con `psql -v ON_ERROR_STOP=1 -f backend/migrations/010_google_login.sql`. Hay que aplicarla **antes** de deployar el backend de la 4a: sin la columna `google_sub`, todos los SELECT de usuarios fallan.

`009` agrega `bookings.confirmation_email_sent_at TIMESTAMP NULL` (aditiva, funciona con datos existentes). Hay que aplicarla **antes** de deployar el backend de la Fase 3: sin la columna, el claim del mail falla (se loguea el error, pero el webhook y la confirmación de la reserva no se ven afectados).

`008` normaliza los emails existentes (`lower(trim())`, `''` → `NULL`) y cambia el `UNIQUE` de `email` por un índice único sobre `lower(email)`. Si dos usuarios chocan al pasar a minúsculas, aborta con la lista de emails duplicados y no cambia nada: hay que corregirlos a mano y volver a correrla. Correrla con `psql -v ON_ERROR_STOP=1 -f backend/migrations/008_normalize_user_email.sql`.

### Deploy de la Fase 1 (orden obligatorio)
Los teléfonos ya guardados están en el formato en que se cargaron y `Login` ahora busca por el número normalizado. Si no se migran, **los usuarios existentes no pueden iniciar sesión**. Orden:

1. Aplicar la migración `008`.
2. Deployar el backend nuevo.
3. Inmediatamente después, desde `backend/`: `go run ./cmd/normalize-phones` (dry run: muestra qué cambiaría, los teléfonos inválidos y los usuarios cuyo número normalizado chocaría con otro).
4. Revisar la salida y correr `go run ./cmd/normalize-phones --apply`.

El comando usa la misma config/DB que el servidor. Los teléfonos que no se pueden normalizar y los usuarios que chocarían quedan sin tocar y hay que resolverlos a mano. Ignora el `-` de los bloqueos de mantenimiento.

### Deploy de "Mail y cuentas" (mail de confirmación + login con Google)
Todo funciona en local; en producción falta:

1. **Migraciones** `009` y `010`, en orden, **antes** de deployar el backend (ver "Migraciones").
2. **Teléfonos:** si la Fase 1 todavía no se deployó, seguir el orden de "Deploy de la Fase 1", incluido `normalize-phones --apply`.
3. **Variables del backend:** `RESEND_API_KEY`, `EMAIL_FROM`, `FRONTEND_URL` y `GOOGLE_CLIENT_ID`.
4. **Variable del frontend:** `VITE_GOOGLE_CLIENT_ID`, con el mismo valor que `GOOGLE_CLIENT_ID`. Vite la toma en el build.
5. **Resend:** verificar un dominio propio y usarlo en `EMAIL_FROM`. Con el remitente de prueba `onboarding@resend.dev` los mails caen en spam y solo llegan al dueño de la cuenta.
6. **Google Cloud** (OAuth Client ID de tipo *Web application*): agregar el dominio de producción en "Orígenes de JavaScript autorizados" y **publicar la app**. Mientras esté en modo *Prueba*, solo pueden entrar los usuarios de prueba cargados en la consola.

---

## Limitaciones conocidas (post-MVP)

- **Disponibilidad por día de semana:** No hay soporte para "este slot no está disponible los domingos". Requeriría agregar `day_of_week` a slots o una tabla de schedules.
- **SELECT FOR UPDATE:** El chequeo de disponibilidad ocurre fuera de la transacción. Para alta concurrencia se debería hacer dentro del `tx`. Para el volumen actual del MVP es suficiente el índice único.
- **JWT no revocable:** Si se desactiva un usuario, su token sigue válido hasta que expire (24hs).
- **Sin refresh token:** El usuario queda deslogueado al expirar el JWT sin aviso.
- **Creación de batches no transaccional:** Turnos fijos y bloqueos crean sus reservas una por una. Si la base de datos falla a mitad de camino, quedan creadas las reservas que alcanzaron a insertarse y el batch queda incompleto. Se puede cancelar el batch y volver a crearlo. Para hacerlo atómico habría que envolver la creación en una única transacción.
