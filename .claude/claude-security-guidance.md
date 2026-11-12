# Reglas de seguridad — H1 Canchas

Contexto: backend en Go + Gin + sqlx + PostgreSQL (Supabase), frontend en React + Vite + TypeScript (Vercel), backend en Render.
Maneja datos personales (nombre, teléfono, email; Ley 25.326) y pagos de seña con Mercado Pago.
Arquitectura: controller → service → repository. Las reglas de negocio y de permisos viven en el **service**.

Las reglas marcadas **(pendiente, E8)** describen cómo tiene que quedar, no cómo está hoy: no asumir que ya se cumplen.

## 1. Control de acceso (lo más importante)

- Toda ruta que reciba un `:id` (reserva, pago, usuario) tiene que verificar que el recurso **pertenece al usuario del JWT**, salvo que el rol sea `receptionist` o `admin`. Esconder el botón en el frontend no cuenta.
  - Ejemplo: `GET /bookings/:id` y `PATCH /bookings/:id/cancel` → un `customer` solo puede ver y cancelar reservas donde `customer_user_id` sea su `user_id`. Si no es suyo, responder 404 (no 403) para no confirmar que existe.
- El `user_id` y el `role` salen **siempre del JWT** (contexto de Gin), nunca del body, query o headers que mande el cliente.
- Las rutas de staff usan `RequireRole(...)`. Una ruta nueva sin middleware de auth es un error, salvo que sea pública a propósito (listar espacios, disponibilidad, `/health`, webhook, `/auth/*`).
- El registro público y el login con Google siempre crean `customer`. Ningún endpoint público puede aceptar un campo `role`.
- Un cliente no puede cancelar una reserva con la seña paga: eso lo hace el staff (`ErrCancelRequiresStaffAfterPayment`).

## 2. Base de datos

- Solo queries parametrizadas (`$1`, `$2`, ...). Nunca armar SQL concatenando strings ni con `fmt.Sprintf` con datos del usuario.
- Si hace falta un `ORDER BY` o un filtro dinámico, usar una lista blanca de columnas permitidas.
- Las reservas no se borran: se cancelan o expiran cambiando el estado. Los espacios se desactivan (`is_active = false`); el borrado físico (`DELETE /admin/spaces/:id`) solo se permite si el espacio nunca tuvo reservas (`ErrSpaceHasBookings`). No agregar borrados físicos de datos con historial.
- La protección contra doble reserva depende del índice único `unique_active_booking`. No eliminarlo ni cambiar los estados que cubre sin revisar la lógica.
- **Supabase:** el backend no usa Supabase Auth ni su SDK, así que las políticas con `auth.uid()` no aplican; el control de acceso es el del service. Pero Supabase expone igual su Data API (PostgREST) sobre el esquema `public`. Cada tabla tiene que tener **RLS activado sin políticas** (eso bloquea los roles `anon` y `authenticated`) o la Data API apagada. Las tablas nuevas activan RLS en su migración (`ALTER TABLE ... ENABLE ROW LEVEL SECURITY`). Antes de activarlo en una base existente, confirmar que el backend se conecta con el rol dueño de las tablas y probarlo primero en QA. Revisar el Security Advisor de Supabase en QA y PROD. **(pendiente, E8)**

## 3. Pagos con Mercado Pago

- El webhook **nunca confía en el body**: vuelve a consultar el pago a la API de Mercado Pago antes de cambiar estados. Esa consulta es la barrera de seguridad real.
- La firma `x-signature` se compara con `hmac.Equal`, pero un mismatch **solo se loguea como WARN, a propósito**: Checkout Pro manda notificaciones IPN viejas cuya firma nunca valida. No convertir ese WARN en un rechazo: se romperían todos los pagos.
- El monto de la seña se calcula en el backend (`total_price` × `DEPOSIT_PERCENTAGE`). Nunca usar un monto que venga del frontend.
- El precio de la reserva se copia al crearla y no se modifica después.
- El webhook es idempotente: si llega dos veces la misma notificación, no se duplican pagos ni cambios de estado.
- En producción se usa `init_point`, no `sandbox_init_point`. Las credenciales de prueba y las reales no se mezclan.

## 4. Autenticación

- Contraseñas con bcrypt (costo 12). Nunca guardarlas, loguearlas ni devolverlas en una respuesta, ni siquiera hasheadas.
- Contraseña mínima de 8 caracteres (hoy son 6). **(pendiente, E8)**
- El JWT vence a las 24 h y se firma con `JWT_SECRET` desde variables de entorno. Solo se acepta HS256 (`pkg/jwt.go` rechaza otros algoritmos): no aflojar ese chequeo.
- Login con Google: el backend verifica la firma, la audiencia, el emisor y `email_verified` del token de Google. Ojo: el email de una cuenta registrada con contraseña **no está verificado**, así que vincular Google por email es un riesgo aceptado (D7 en `ROADMAP.md`). No sumar otras vinculaciones automáticas por email. Las cuentas de staff nunca se vinculan con Google.
- Límite de intentos en login, registro, login con Google y restablecer contraseña. Hoy no hay ninguno. **(pendiente, E8)**
- Restablecer contraseña (E5): token aleatorio de un solo uso (`crypto/rand`), con vencimiento corto, guardado hasheado en la base y comparado en tiempo constante. La respuesta es la misma exista o no la cuenta, para no revelar qué emails están registrados. Al usarlo, invalidar los demás tokens pendientes de ese usuario.

## 5. Secretos y configuración

- Ningún secreto en el código: `JWT_SECRET`, credenciales de la base, tokens de Mercado Pago, claves de Resend y de Google van en variables de entorno.
- Los `.env` no se suben a git. `.env.example` solo lleva nombres de variables con valores de ejemplo.
- En el frontend, todo lo que empieza con `VITE_` termina en el bundle público: ahí nunca va un secreto (`VITE_GOOGLE_CLIENT_ID` es público por diseño).
- En GitHub Actions, los secrets de PROD solo existen en el environment `production`.
- Ningún test, script ni auditoría corre contra PROD ni contra datos reales de clientes: se usa QA o un entorno descartable.

## 6. Respuestas, errores y logs

- Los errores al cliente son genéricos ("credenciales inválidas", "no encontrado"). Nada de mensajes de la base, stack traces ni detalles internos.
- Los logs no guardan contraseñas, tokens JWT, tokens de restablecimiento, headers `Authorization` ni datos completos de pago.
- Las respuestas no exponen `password_hash`, `google_sub` ni datos de otros clientes.

## 7. HTTP: CORS y headers

- CORS: `Access-Control-Allow-Origin` es exactamente `ALLOWED_ORIGIN` (en PROD, el dominio HTTPS del frontend). Nunca `*` y nunca copiar el header `Origin` del pedido. Métodos permitidos: los que usa la API (`GET, POST, PUT, PATCH, DELETE, OPTIONS`); headers: `Content-Type` y `Authorization`. El preflight `OPTIONS` responde 2xx.
- Headers de seguridad **(pendiente, E8)**:
  - Frontend (en `frontend/vercel.json`): `Strict-Transport-Security: max-age=31536000; includeSubDomains`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy: geolocation=(), microphone=(), camera=()` y una `Content-Security-Policy` con `default-src 'self'` y `object-src 'none'`. La CSP tiene que permitir lo que la app usa de verdad: la URL del backend en `connect-src`, el login de Google (`https://accounts.google.com`) y Google Fonts. Sin `unsafe-eval`.
  - Backend (middleware de Gin): `X-Content-Type-Options: nosniff` y `Strict-Transport-Security`.
  - Verificar en QA con securityheaders.com y mirando la consola del navegador (errores de CSP) antes de pasar a PROD.
- Validar y limpiar la entrada en el service (`TrimSpace`, largo máximo, formato de teléfono E.164 y de fecha `YYYY-MM-DD`).
- Rechazar fechas pasadas y espacios desactivados al reservar.

## 8. Frontend

- No usar `dangerouslySetInnerHTML` con datos que vengan del usuario o de la API.
- No usar `eval` ni `new Function`.
- El token no se loguea en consola ni se manda a servicios externos.
- Los permisos del frontend son solo para la interfaz: la verdad la decide el backend.

## 9. Cómo se arregla un hallazgo de seguridad

- De a uno: primero un test que falle y demuestre el problema, después el arreglo más chico posible, después confirmar que el test pasa y que el resto de la suite sigue verde.
- Un hallazgo está confirmado solo si tiene un camino de exposición, un control que falla y un impacto concreto. Si falta alguno, queda "a validar".
- No agregar dependencias nuevas para arreglar un hallazgo sin decidirlo antes.
