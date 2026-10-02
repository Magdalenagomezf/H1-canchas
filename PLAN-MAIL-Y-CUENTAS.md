# Plan: Mail y cuentas

Épica E2 del tablero. Objetivo: que el cliente reciba un mail de confirmación de su reserva y que pueda registrarse con Google.

Este archivo va en la raíz del repo. Se trabaja **una fase por sesión**: la persona indica qué fase hacer.

---

## Reglas para Claude Code

- Antes de tocar código, leé `CLAUDE.md` y `RESUMEN-PROYECTO.md`. Respetá la arquitectura: Repository → Service → Controller. Las interfaces las define el service, y los errores de dominio van en `service/errors.go`.
- Hacé **solo la fase pedida**. Si ves algo de otra fase o un bug ajeno, anotalo al final de tu respuesta en lugar de arreglarlo.
- Arrancá en modo plan: mostrá qué archivos vas a tocar y esperá el OK.
- Migraciones: archivo SQL nuevo con el número siguiente (la última hoy es `007`). No hay runner; se aplican a mano con `psql`. Tienen que funcionar sobre una base con datos.
- Tests: todo cambio en service lleva test. Seguí el estilo de los `*_test.go` que ya existen. Al terminar, `go test ./...` y `npm run build` tienen que pasar.
- Textos visibles en español rioplatense (vos), igual que el resto de la app. El frontend sigue el diseño actual (`DESIGN-SYSTEM.md`).
- Al terminar la fase: actualizá `RESUMEN-PROYECTO.md` (endpoints, migraciones, decisiones) y `CLAUDE.md` si hay variables de entorno nuevas. Proponé un mensaje de commit.
- Si algo de este plan choca con el código real, frená y preguntá.

---

## Lo que tiene que hacer Magdalena, no Claude Code

- **Antes de la Fase 3:** crear una cuenta en [Resend](https://resend.com), verificar un dominio (o usar el remitente de prueba `onboarding@resend.dev`, que solo manda a tu propio mail) y generar la API key.
- **Antes de la Fase 4:** en Google Cloud Console, crear un OAuth Client ID de tipo *Web application*. Orígenes autorizados: `http://localhost:5173` y el dominio de producción.
- Aplicar cada migración nueva con `psql`, en local y en producción.

---

## Fase 1: Datos de contacto prolijos

**Objetivo:** que email y teléfono se guarden de una sola forma, para poder buscar por ellos y compararlos.

### Backend
1. **Normalizar el email:** `strings.TrimSpace` + `strings.ToLower`. Un email vacío se guarda como `NULL`, nunca como `""`. Validar el formato con `net/mail.ParseAddress`. Aplicarlo en `Register` y `CreateStaff`.
2. **Normalizar el teléfono** con una función `NormalizePhone` en un lugar compartido (por ejemplo, `pkg/phone`). Usar `github.com/nyaruka/phonenumbers` con región por defecto `AR` y guardar en formato E.164 (`+549…`). Si no es válido, devolver un error de dominio nuevo (`ErrInvalidPhone`). Aplicarla en:
   - `Register`, `CreateStaff` y **`Login`**, porque el usuario escribe el número como quiera y hay que normalizarlo antes de `FindByPhone`;
   - `CreateManual`, en `customer_phone`.
3. **Migración `008`:**
   - Cambiar el `UNIQUE` de `email` por un índice único sobre `lower(email)`.
   - Pasar a `lower(trim())` los emails existentes, y convertir los `''` en `NULL`.
   - **No** normalizar teléfonos existentes en SQL, porque es muy propenso a errores. En su lugar, escribir un comando chico `backend/cmd/normalize-phones/main.go` que lo haga con la misma función de Go, **muestre primero** qué cambiaría y qué duplicados aparecerían, y solo aplique los cambios con `--apply`.
4. **Exponer el email** en `UserResponse` (`email` como `*string`, con `omitempty`).
5. **Email obligatorio en el registro nuevo:** `RegisterRequest.Email` pasa a `binding:"required"`. En `CreateStaff` sigue siendo opcional. Los usuarios que ya existen quedan sin mail; lo pueden cargar en la Fase 4 desde "Mi perfil".

### Frontend
- Agregar el campo Email al tab de registro de `LoginPage`. Mostrar los errores del backend (formato inválido, email ya usado).
- Actualizar `types/index.ts` con `email` en el usuario.

### Tests
- `NormalizePhone`: `3834123456`, `0383 15-412-3456`, `+54 9 383 412 3456` dan el mismo resultado; `abc` da error.
- Registro con email en mayúsculas: se guarda en minúsculas. Un segundo registro con el mismo email en otra combinación de mayúsculas devuelve conflicto.
- Login con el teléfono escrito en otro formato: funciona.

---

## Fase 2: Mail en reservas manuales

**Objetivo:** que el recepcionista pueda cargar un mail al crear una reserva a mano, para que ese cliente también reciba la confirmación.

1. **Migración `009`:** `customer_email VARCHAR(150) NULL` en `bookings`.
2. `CreateManualBookingRequest.CustomerEmail`: opcional, normalizado igual que en la Fase 1.
3. `BookingService.CreateManual` lo recibe y lo guarda. El repository lo inserta y lo devuelve en `BookingDetail`.
4. En el detalle de la reserva, el email del cliente sale de `users.email` si es un usuario registrado, o de `customer_email` si es una reserva manual. Exponerlo como `customer.email` en la respuesta.
5. Frontend: campo "Email (opcional)" en el formulario de reserva manual de `StaffPanelPage`, en los dos lugares donde hoy se arma `customer_phone`.

### Tests
- Reserva manual con email y sin email.
- Un email inválido devuelve error de validación.

---

## Fase 3: Mail de confirmación

**Objetivo:** que el cliente reciba **un solo** mail cuando su reserva queda confirmada.

### Cuándo se envía
- **Reserva online:** cuando el webhook de Mercado Pago confirma la seña (`HandleWebhook` pasa `deposit_status` a `paid` y la reserva a `confirmed`). **No** al crear la reserva, porque puede vencer en 20 minutos.
- **Reserva manual:** al crearla, si tiene email.
- **Turnos fijos y bloqueos:** no se envía nada.

### Enviar una sola vez
Mercado Pago reenvía el mismo webhook varias veces. Para no mandar el mail repetido:
- **Migración `010`:** `confirmation_email_sent_at TIMESTAMP NULL` en `bookings`.
- Antes de enviar, reclamar el envío con un `UPDATE bookings SET confirmation_email_sent_at = now() WHERE id = $1 AND confirmation_email_sent_at IS NULL`. Solo se envía si afectó una fila.
- Si el envío falla, volver a poner el campo en `NULL` y loguear el error.

### Diseño
- Interfaz `EmailSender` (`Send(ctx, to, subject, html, text string) error`), definida en el service que la usa.
- Implementaciones en `pkg/email/`:
  - `ResendSender`: usa la API HTTP de Resend (`POST https://api.resend.com/emails`) con `net/http`, sin SDK.
  - `LogSender`: solo loguea el destinatario y el asunto, no el cuerpo. Se usa cuando `RESEND_API_KEY` está vacío, así el desarrollo local funciona sin cuenta.
- Un `NotificationService` (o método en un service existente, lo que quede más limpio) con `SendBookingConfirmation(bookingID)`. Busca el `BookingDetail`, resuelve el email del destinatario y arma el mail.
- **Nunca rompe la reserva:** se llama **después** del commit, en una goroutine con su propio `context.WithTimeout` de 15 segundos. No usa el contexto del request, que se cancela cuando termina la respuesta HTTP.
- Si no hay email (usuario viejo o reserva manual sin mail), no hace nada.

### Contenido del mail
- Asunto: `Reserva confirmada — {espacio}, {fecha}`.
- Cuerpo en HTML simple, más una versión en texto plano:
  - nombre del espacio;
  - fecha en español (por ejemplo, "sábado 4 de octubre"), en horario `America/Argentina/Buenos_Aires`;
  - horario del turno;
  - seña pagada y saldo a pagar, en pesos con separador de miles;
  - link a "Mis reservas" (`FRONTEND_URL` + `/mis-reservas`), solo si es un usuario registrado.
- Plantillas con `html/template` y `text/template`, en archivos separados del código Go (`embed`).

### Variables de entorno
- Nuevas: `RESEND_API_KEY` (opcional), `EMAIL_FROM` (por ejemplo `H1 Canchas <reservas@dominio>`) y `FRONTEND_URL`.
- Agregarlas en `config.go`, en `CLAUDE.md` y en `.env.example`. Aprovechar para sumar a `.env.example` las de Mercado Pago que hoy faltan (`MP_ACCESS_TOKEN`, `MP_WEBHOOK_SECRET`, `MP_WEBHOOK_URL`), sin valores reales.

### Tests (con un `EmailSender` falso)
- El webhook que confirma la seña envía 1 mail. El mismo webhook repetido no envía otro.
- Si el sender devuelve error, la reserva queda `confirmed` igual y `confirmation_email_sent_at` vuelve a `NULL`.
- Reserva manual con email envía el mail; sin email, no.
- Usuario sin email: no envía nada y no da error.

---

## Fase 4: Login con Google y "Mi perfil"

**Objetivo:** registrarse e iniciar sesión con Google, y que todo usuario pueda completar o corregir sus datos.

### Migración `011`
- `password_hash` pasa a ser `NULL`.
- `phone` pasa a ser `NULL`. El `UNIQUE` se mantiene, porque Postgres permite varios `NULL`.
- `google_sub VARCHAR(255) NULL UNIQUE`.

### Backend
1. **`POST /auth/google`** con `{ "credential": "<ID token>" }`:
   - Verificar el token con `google.golang.org/api/idtoken`, usando `aud = GOOGLE_CLIENT_ID`. Exigir `email_verified = true`.
   - Buscar por `google_sub`. Si existe, es un login.
   - Si no, buscar por email (normalizado):
     - si es un `customer`, vincular (guardar `google_sub`) y hacer login;
     - si es staff (`receptionist` o `admin`), **no vincular** y devolver un error ("Esta cuenta tiene que iniciar sesión con teléfono y contraseña").
   - Si no existe por ninguno de los dos, crear un `customer` con nombre y email de Google, sin teléfono y sin contraseña.
   - Devolver el mismo `AuthResponse` que `/auth/login`.
2. **Login con teléfono:** si el usuario no tiene `password_hash`, devolver el mismo error de credenciales inválidas que ya existe. No revelar que la cuenta es de Google.
3. **`GET /me`** y **`PATCH /me`** (`name`, `email`, `phone`, todos opcionales). Normalizar con las funciones de la Fase 1 y validar que no choquen con otro usuario.
4. **Reservar exige teléfono:** `BookingService.Create` devuelve `ErrPhoneRequired` si el usuario no tiene teléfono. Mapearlo a un código HTTP claro en el controller.
5. `UserResponse.Phone` pasa a `*string`.
6. Variable nueva: `GOOGLE_CLIENT_ID`.

### Frontend
- Librería `@react-oauth/google`, con `VITE_GOOGLE_CLIENT_ID`.
- Botón "Continuar con Google" en `LoginPage`, en los dos tabs. Manda el `credential` a `/auth/google` y guarda la sesión igual que el login actual (`useAuth`).
- **Página `/mi-perfil`** (ruta privada) para editar nombre, email y teléfono. Link desde el Navbar.
- Si el usuario logueado no tiene teléfono: redirigir a `/mi-perfil` con un aviso ("Cargá tu teléfono para poder reservar") al entrar a reservar, y también si el backend devuelve `ErrPhoneRequired`.
- Actualizar `types/index.ts`: `phone` puede ser `null`.

### Tests
- Token inválido o de otra `aud`: 401. Usar un verificador inyectable (interfaz) para no llamar a Google en los tests.
- Usuario nuevo por Google: se crea sin teléfono. Reservar da `ErrPhoneRequired`. Con `PATCH /me` + teléfono, reserva bien.
- Email existente de un customer: se vincula. Email existente de un admin: no se vincula y da error.
- Login con teléfono sobre una cuenta sin contraseña: credenciales inválidas.

---

## Fuera de alcance (no hacer)

- **No convertir en usuarios** a los clientes cargados a mano. Siguen como nombre, teléfono y email en la reserva.
- **No permitir "reclamar" reservas manuales por teléfono** hasta que exista verificación por SMS.
- Recordatorios, mails de cancelación y de vencimiento: se hacen en la épica "Avisos al cliente", reusando el `EmailSender` de la Fase 3.
