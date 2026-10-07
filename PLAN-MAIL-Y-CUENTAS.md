# Plan: Mail y cuentas

Épica E2 del tablero. Objetivo: que el cliente reciba un mail de confirmación de su reserva y que pueda registrarse con Google.

Este archivo va en la raíz del repo. Se trabaja **una fase por sesión**: la persona indica qué fase hacer.

## Estado

| Fase | Estado |
|------|--------|
| 1. Datos de contacto prolijos | Hecha (commit `4909fab`) |
| 2. Mail en reservas manuales | Descartada |
| 3. Mail de confirmación | Implementada, **sin commitear**. Falta: correr `go test ./...`, revisar `.env.example`, aplicar `009` y commitear |
| 4a. Google + "Mi perfil" (backend) | Pendiente |
| 4b. Google + "Mi perfil" (frontend) | Pendiente, después de 4a |

---

## Reglas para Claude Code

- Antes de tocar código, leé `CLAUDE.md` y `RESUMEN-PROYECTO.md`. Respetá la arquitectura: Repository → Service → Controller. Las interfaces las define el service, y los errores de dominio van en `service/errors.go`.
- Hacé **solo la fase pedida**. Si ves algo de otra fase o un bug ajeno, anotalo al final de tu respuesta en lugar de arreglarlo.
- Arrancá en modo plan: mostrá qué archivos vas a tocar y esperá el OK.
- Migraciones: archivo SQL nuevo con el número siguiente (la última hoy es `009`). No hay runner; se aplican a mano con `psql`. Tienen que funcionar sobre una base con datos.
- Tests: todo cambio en service lleva test. Seguí el estilo de los `*_test.go` que ya existen. Al terminar, `go test ./...` y `npm run build` tienen que pasar.
- Textos visibles en español rioplatense (vos), igual que el resto de la app. El frontend sigue el diseño actual (`DESIGN-SYSTEM.md`).
- Al terminar la fase: actualizá `RESUMEN-PROYECTO.md` (endpoints, migraciones, decisiones) y `CLAUDE.md` si hay variables de entorno nuevas. Proponé un mensaje de commit.
- Si algo de este plan choca con el código real, frená y preguntá.

---

## Lo que tiene que hacer Magdalena, no Claude Code

- **Antes de la Fase 3:** crear una cuenta en [Resend](https://resend.com), verificar un dominio (o usar el remitente de prueba `onboarding@resend.dev`, que solo manda a tu propio mail) y generar la API key.
- **Antes de la Fase 4b:** en Google Cloud Console, crear un OAuth Client ID de tipo *Web application*. Orígenes autorizados: `http://localhost:5173` y el dominio de producción.
- Aplicar cada migración nueva con `psql`, en local y en producción.

---

## Fase 1: Datos de contacto prolijos — HECHA

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

## Fase 2: Mail en reservas manuales — DESCARTADA

Descartada el 2026-10-02. En una reserva manual el cliente ya arregló todo con el recepcionista, casi siempre por WhatsApp o teléfono, así que un mail de confirmación no le suma nada. Pedirle el email haría más lenta la carga y agregaría errores de tipeo.

Si en el futuro hace falta (por ejemplo, para los recordatorios de "Avisos al cliente"), se agrega una columna `customer_email` en `bookings` que acepte `NULL`, sin tocar los datos existentes.

---

## Fase 3: Mail de confirmación — IMPLEMENTADA

**Objetivo:** que el cliente reciba **un solo** mail cuando su reserva queda confirmada.

### Cuándo se envía
- **Reserva online:** cuando el webhook de Mercado Pago confirma la seña (`HandleWebhook` pasa `deposit_status` a `paid` y la reserva a `confirmed`). **No** al crear la reserva, porque puede vencer en 20 minutos.
- **Reservas manuales, turnos fijos y bloqueos:** no se envía nada.
- **Seña marcada a mano por el staff** (`PATCH /bookings/:id/payment`) en una reserva online: hoy **no** se envía. Si se quiere, se agrega llamando al mismo notifier desde `MarkPaidManually`.
- **Usuarios registrados antes de la Fase 1** (no tienen email): no reciben mail. Si después lo cargan en "Mi perfil" (Fase 4), no se reenvía el de reservas ya pagadas, porque el envío solo lo dispara el webhook. Aceptado. Los registros nuevos siempre tienen email.
- **Riesgo aceptado:** si el proceso se cae entre el claim y el envío, ese mail se pierde (documentado en `notification_service.go`).

### Enviar una sola vez
Mercado Pago reenvía el mismo webhook varias veces. Para no mandar el mail repetido:
- **Migración `009`:** `confirmation_email_sent_at TIMESTAMP NULL` en `bookings`.
- Antes de enviar, reclamar el envío con un `UPDATE bookings SET confirmation_email_sent_at = now() WHERE id = $1 AND confirmation_email_sent_at IS NULL`. Solo se envía si afectó una fila.
- Si el envío falla, volver a poner el campo en `NULL` y loguear el error.

### Diseño
- Interfaz `EmailSender` (`Send(ctx, to, subject, html, text string) error`), definida en el service que la usa.
- Implementaciones en `pkg/email/`:
  - `ResendSender`: usa la API HTTP de Resend (`POST https://api.resend.com/emails`) con `net/http`, sin SDK.
  - `LogSender`: solo loguea el destinatario y el asunto, no el cuerpo. Se usa cuando `RESEND_API_KEY` está vacío, así el desarrollo local funciona sin cuenta.
- Un `NotificationService` (o método en un service existente, lo que quede más limpio) con `SendBookingConfirmation(bookingID)`. Busca el `BookingDetail`, resuelve el email del destinatario y arma el mail.
- **Nunca rompe la reserva:** se llama **después** del commit, en una goroutine con su propio `context.WithTimeout` de 15 segundos. No usa el contexto del request, que se cancela cuando termina la respuesta HTTP.
- Si no hay email (reserva sin usuario registrado, o usuario viejo sin mail), no hace nada.

### Contenido del mail
- Asunto: `Reserva confirmada — {espacio}, {fecha}`.
- Cuerpo en HTML simple, más una versión en texto plano:
  - nombre del espacio;
  - fecha en español (por ejemplo, "sábado 4 de octubre"), en horario `America/Argentina/Buenos_Aires`;
  - horario del turno;
  - seña pagada y saldo a pagar, en pesos con separador de miles;
  - link a "Mis reservas" (`FRONTEND_URL` + `/mis-reservas`).
- Plantillas con `html/template` y `text/template`, en archivos separados del código Go (`embed`).

### Variables de entorno
- Nuevas: `RESEND_API_KEY` (opcional), `EMAIL_FROM` (por ejemplo `H1 Canchas <reservas@dominio>`) y `FRONTEND_URL`.
- Agregarlas en `config.go`, en `CLAUDE.md` y en `.env.example`. Aprovechar para sumar a `.env.example` las de Mercado Pago que hoy faltan (`MP_ACCESS_TOKEN`, `MP_WEBHOOK_SECRET`, `MP_WEBHOOK_URL`), sin valores reales.

### Tests (con un `EmailSender` falso)
- El webhook que confirma la seña envía 1 mail. El mismo webhook repetido no envía otro.
- Si el sender devuelve error, la reserva queda `confirmed` igual y `confirmation_email_sent_at` vuelve a `NULL`.
- Reserva manual: no envía nada.
- Usuario sin email: no envía nada y no da error.

---

## Fase 4: Login con Google y "Mi perfil"

**Objetivo:** registrarse e iniciar sesión con Google, y que todo usuario pueda completar o corregir sus datos.

Se divide en **dos sesiones**: 4a (backend) y 4b (frontend). La 4b arranca recién cuando la 4a está commiteada y con tests en verde.

---

### Fase 4a: Backend

#### Migración `010_google_login.sql`
```sql
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;   -- el UNIQUE se mantiene; Postgres admite varios NULL
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_sub VARCHAR(255) NULL UNIQUE;
-- Toda cuenta necesita al menos una forma de entrar.
ALTER TABLE users ADD CONSTRAINT users_login_method_chk
    CHECK (password_hash IS NOT NULL OR google_sub IS NOT NULL);
```
Corre sobre la base con datos: todas las filas actuales tienen `password_hash`, así que el `CHECK` pasa.

#### 1. Dominio y repository (el cambio más delicado)
Hoy `domain.User` tiene `Phone string` y `PasswordHash string`. Con las columnas nullable, **escanear un `NULL` en un `string` falla en runtime**. Por eso:
- `domain.User`: `Phone *string`, `PasswordHash *string` y nuevo `GoogleSub *string` (`db:"google_sub"`).
- Actualizar **todos** los usos: `user_repository.go` (insert y selects; sumar `google_sub`), `user_service.go` (`Register`, `CreateStaff`, `Login`), `dto/user_dto.go`, y los mocks/tests que arman `domain.User`. `go build ./...` es la red: si compila, no quedó ninguno.
- `UserResponse.Phone` pasa a `*string` **sin** `omitempty`: el frontend recibe `null`, no la ausencia del campo.
- Métodos nuevos en el repo: `FindByGoogleSub`, `SetGoogleSub(id, sub)`, `UpdateProfile(id, name, email, phone)`. `FindByEmail` busca por `lower(email)` (índice de la `008`). Como el resto, devuelven `nil, nil` si no hay fila.
- `Register` y `CreateStaff` siguen exigiendo teléfono y contraseña; solo cambian los tipos.

#### 2. Login con teléfono
En `Login`, si `user.PasswordHash == nil` → `ErrInvalidCredentials` (el de siempre). No revelar que la cuenta es de Google.

#### 3. Verificador de Google (inyectable)
En el service:
```go
type GoogleIdentity struct {
    Sub           string
    Email         string
    EmailVerified bool
    Name          string
}

type GoogleTokenVerifier interface {
    Verify(ctx context.Context, credential string) (*GoogleIdentity, error)
}
```
Implementación en `pkg/google/` con `google.golang.org/api/idtoken`: `idtoken.Validate(ctx, credential, clientID)` (valida firma, `aud` y `exp`). Además, chequear a mano que `iss` sea `accounts.google.com` o `https://accounts.google.com`. `sub` sale de `payload.Subject`; `email`, `email_verified` (bool) y `name` de `payload.Claims`.

#### 4. `POST /auth/google` (público)
Body `{ "credential": "<ID token>" }`. Nuevo `UserService.LoginWithGoogle(ctx, credential)` que devuelve lo mismo que `Login` (`*domain.User, token, error`). En este orden:
1. `Verify`. Si falla, o si `EmailVerified` es false → `ErrInvalidGoogleToken` (**401**).
2. Normalizar el email con `normalizeEmail` (`service/contact.go`).
3. `FindByGoogleSub(sub)`. Si existe → paso 6.
4. `FindByEmail(email)`. Si existe:
   - rol distinto de `customer` → `ErrGoogleStaffAccount` (**403**, "Esta cuenta tiene que iniciar sesión con teléfono y contraseña"). No se vincula.
   - `customer` con otro `google_sub` ya cargado → `ErrGoogleAccountConflict` (**409**). No se pisa.
   - `customer` sin `google_sub` → `SetGoogleSub` y paso 6.
5. Si no existe por ninguno: crear `customer` con `email`, `google_sub`, `phone = NULL`, `password_hash = NULL`. Nombre: el de Google recortado a 100 caracteres (`VARCHAR(100)`); si viene vacío, lo que está antes de la `@`. Si el `INSERT` choca por unique (dos requests simultáneos) → `ErrEmailAlreadyExists` (409), sin reintentar.
6. Si `!user.IsActive` → `ErrUserInactive` (igual que `Login`). Generar el JWT igual que `Login` y devolver el mismo `AuthResponse`.

Variable `GOOGLE_CLIENT_ID` (opcional). Si está vacía, la ruta se registra igual pero responde **503** ("El inicio de sesión con Google no está configurado"), así la app arranca sin ella. Sumarla a `config.go`, `CLAUDE.md` y `.env.example`.

#### 5. `GET /me` y `PATCH /me` (cualquier usuario autenticado, staff incluido)
- `GET /me` → `UserResponse` del usuario del token.
- `PATCH /me` con `{ "name"?: string, "email"?: string, "phone"?: string }`. DTO con punteros: **campo ausente = no se toca**.
  - `name`: trim; vacío o más de 100 caracteres → 400.
  - `email`: `normalizeEmail`. Vacío → `ErrEmailRequired` (el email no se puede borrar). Inválido → `ErrInvalidEmail`.
  - `phone`: `phone.Normalize`. Vacío o inválido → `ErrInvalidPhone` (el teléfono no se puede borrar una vez cargado).
  - Email o teléfono que ya usa **otro** usuario → `ErrEmailAlreadyExists` / `ErrPhoneAlreadyExists` (**409**). Chequear antes y además mapear la violación de unique del `UPDATE`.
  - Cambiar el email de una cuenta con `google_sub` está permitido: el login con Google busca por `sub`.
  - No se cambia `role` ni contraseña acá (cambiar contraseña queda fuera de alcance).
  - Devuelve el `UserResponse` actualizado.

#### 6. Reservar exige teléfono
- **Ojo:** `ErrPhoneRequired` ya existe y significa "falta el teléfono en el formulario" (registro, reserva manual, turnos fijos), mapeado a 400. **No reusarlo.** Crear `ErrProfilePhoneMissing` ("Cargá tu teléfono en Mi perfil para poder reservar").
- `BookingService.Create`: antes de chequear disponibilidad, buscar al usuario (interfaz chica nueva en el service, p. ej. `BookingUserLookup { FindByID }`) y si `Phone == nil` → `ErrProfilePhoneMissing`.
- Controller: **422** con `{ "error": "<mensaje>", "code": "phone_required" }`. El `code` existe para que el frontend no dependa del texto.

#### Tests 4a
- Verificador falso que implementa `GoogleTokenVerifier`; nunca se llama a Google.
- Token inválido → 401. `email_verified = false` → 401.
- Usuario nuevo por Google: se crea `customer` sin teléfono ni contraseña.
- Mismo `sub` dos veces: la segunda es login, no crea otro usuario.
- Email de un customer existente: se vincula. Email de admin/receptionist: 403 y no se vincula. Customer ya vinculado a otro `sub`: 409.
- Usuario inactivo por Google: `ErrUserInactive`.
- Login con teléfono sobre cuenta sin contraseña: `ErrInvalidCredentials`.
- Reservar sin teléfono: `ErrProfilePhoneMissing`. Después de `PATCH /me` con teléfono: reserva bien.
- `PATCH /me`: campo ausente no cambia nada; email/teléfono de otro usuario → 409; email vacío → error; email en mayúsculas se guarda en minúsculas.
- Repository: leer un usuario con `phone` y `password_hash` en `NULL` no falla.

---

### Fase 4b: Frontend

#### Google
- `npm i @react-oauth/google`. Variable `VITE_GOOGLE_CLIENT_ID` (sumarla al `.env` del frontend).
- Envolver la app con `GoogleOAuthProvider` **solo si** la variable tiene valor. Sin ella, el botón no aparece y nada se rompe.
- En `LoginPage`, en los dos tabs: usar el **componente `<GoogleLogin onSuccess={...} />`**, que entrega `credential` (el ID token). **No usar el hook `useGoogleLogin`**: por defecto devuelve un *access token*, no un ID token, y el backend lo rechaza.
- `api/auth.ts`: `loginWithGoogle(credential)` → `POST /auth/google`. Guardar la sesión con **la misma función** que usa hoy el login en `useAuth` (no duplicar). Mostrar los 403/409 con el mensaje del backend.

#### "Mi perfil"
- `api/users.ts`: `getMe()` y `updateMe(partial)`.
- `useAuth`: agregar `updateUser(user)` para refrescar el usuario guardado después de editar; si no, el Navbar y el chequeo de teléfono quedan con datos viejos.
- Página `/mi-perfil` (ruta privada en `App.tsx`): nombre, email y teléfono, precargados con `getMe()`. Mandar solo los campos que cambiaron. Mostrar los 409 junto al campo correspondiente. Link "Mi perfil" en el `Navbar` para cualquier usuario logueado.

#### Teléfono obligatorio para reservar
- En `SpaceDetailPage`, al reservar: si `user.phone` es `null`, no llamar a `createBooking`; navegar a `/mi-perfil?next=<ruta actual>` con el aviso "Cargá tu teléfono para poder reservar".
- Si el backend igual responde 422 con `code: "phone_required"`, hacer lo mismo.
- En `/mi-perfil`, al guardar con éxito y si hay `next`, volver a esa ruta.

#### Tipos y pantallas existentes
- `types/index.ts`: `phone: string | null` en el usuario.
- Donde se muestre el teléfono de un usuario (lista de usuarios del admin, detalle de reservas del staff), mostrar "—" si es `null`. `npm run build` tiene que pasar sin `!` ni casts que escondan el `null`.

#### Prueba manual
- Google con cuenta nueva → sin teléfono → al reservar va a "Mi perfil" → carga teléfono → vuelve y reserva.
- Google con el mail de un admin → mensaje de error, no entra.

---

## Fuera de alcance (no hacer)

- **No convertir en usuarios** a los clientes cargados a mano. Siguen como nombre y teléfono en la reserva.
- **Mail en reservas manuales:** descartado (ver Fase 2).
- **No permitir "reclamar" reservas manuales por teléfono** hasta que exista verificación por SMS.
- **Riesgo aceptado (revisar más adelante): vinculación automática de Google por email.** El email del registro con teléfono no se verifica. Alguien puede registrarse con el email de otra persona y su propia contraseña; cuando la persona real entra con Google, la cuenta se vincula y el atacante, que sigue sabiendo la contraseña, ve sus reservas. Opciones para cerrarlo: (A) al vincular, borrar `password_hash`; (B) no vincular automáticamente y hacerlo desde "Mi perfil" con la sesión iniciada; (C) verificar el email al registrarse. Se acepta por ahora por el bajo valor de los datos expuestos.
- Recordatorios, mails de cancelación y de vencimiento: se hacen en la épica "Avisos al cliente", reusando el `EmailSender` de la Fase 3.
