# Roadmap H1 Canchas

Objetivo: inaugurar las canchas en 1 o 2 meses (estimado al 2026-11-09). La app ya está publicada en servidores gratuitos.

Cómo se usa este archivo:
- **Decisiones pendientes:** lo que hay que definir antes de poder implementar. Cuando se decide, se anota la decisión y se pasa a la épica.
- **Épicas:** grupos de tareas. Se separan en "antes de la inauguración" y "después".
- **Ideas sueltas:** para anotar rápido sin clasificar. Se ordenan más tarde.
- El estado técnico (endpoints, decisiones ya tomadas, bugs) vive en `RESUMEN-PROYECTO.md`.

---

## Decisiones pendientes

| # | Decisión | Afecta a |
|---|----------|----------|
| D1 | **Servidores para producción:** seguir en planes gratuitos o pasar a pagos. Los gratuitos suelen "dormirse" sin uso (la primera carga tarda y puede demorar el aviso de pago de Mercado Pago) y tienen límites de base de datos y backups. **Se decide 2 o 3 semanas antes de inaugurar** (2026-11-09): CI, QA y el pipeline se arman en los planes gratuitos; lo que conviene pagar es PROD. | E3, E4 |
| D3 | **Cuenta de Mercado Pago del negocio** y **política de cancelación y devolución de la seña.** Propuesta (2026-11-09), **a validar con el cliente**: para inaugurar, nivel 1 (abajo). Opción a ofrecerle: devolución si se cancela con más de 24 h. Revisar si aplica la normativa de defensa del consumidor (p. ej. "botón de arrepentimiento") a una reserva con fecha fija. | E4 |
| D9 | **Precios por horario y días especiales** (a hablar con el cliente): ¿cobran distinto de noche (por la luz) o los fines de semana? ¿Hay feriados o días cerrados? Hoy cada cancha tiene un solo precio y todos los turnos abren todos los días. Si aplica, se crea una épica. | — |
| D10 | **Porcentaje de la seña** (a hablar con el cliente): hoy es 15% (`DEPOSIT_PERCENTAGE`), un valor puesto por defecto. | E4 |
| D11 | **Facturación** (a hablar con el cliente): ¿cómo factura el complejo? Si tiene que emitir comprobantes electrónicos (ARCA, ex AFIP), puede ser un tema grande. Primero saber si aplica; no implica hacerlo ya. | — |

---

## Antes de la inauguración

### E3. Pipeline y entornos QA y producción
Orden decidido (2026-11-09): primero CI, después QA y PROD. Modelo: Continuous Delivery, como en el TP6 de la materia: QA se actualiza solo y a PROD solo pasa lo que se aprobó, en el horario que se elija.

**CI (no depende de qué servidores se elijan):**
- [x] GitHub Actions en cada PR: `go test` con un Postgres de servicio (así los tests de integración del repository dejan de saltearse) + build del frontend + lint.
- [x] Protección de rama en `main`: no se mergea con el CI en rojo. Hecho el 2026-11-09 (PR #1): ruleset sin bypass, todo cambio entra por PR, 0 aprobaciones, checks `Backend (vet + tests)` y `Frontend (lint + build)` obligatorios.

**Entornos** (D2 decidido el 2026-11-09: pipeline en GitHub Actions, como en Mallki):
- [ ] Decidir D1.
- [ ] Entorno QA: base propia (otro proyecto de base), backend, frontend, credenciales de **prueba** de Mercado Pago, origen de QA en Google Cloud.
- [ ] Entorno PROD: base propia, credenciales reales.
- [ ] Apagar el auto-deploy de los proveedores: solo el pipeline publica, y siempre el commit que se verificó.
- [ ] Pipeline: merge a `main` → CI verde → deploy a QA → smoke → E2E contra QA → **pausa esperando aprobación** (environment `production` de GitHub, con los secrets de PROD solo ahí) → deploy a PROD → smoke.
- [ ] Si fallan el smoke o los E2E de QA, PROD queda salteado y ni pide aprobación.
- [x] `/health` devuelve el commit que está corriendo (y ahora sí consulta la base: 503 si no responde). Hecho el 2026-11-09.
- [ ] Smoke test después de cada deploy (script `scripts/smoke.sh` hecho el 2026-11-09 y probado en el CI contra un build local; falta engancharlo al deploy): la app responde y con el commit esperado (no solo "está viva"). Con reintentos por el cold start.
- [ ] Documentar cómo se pasa un cambio de QA a PROD y cómo se vuelve atrás (rollback).

Si se cambia de proveedor, el CI, los tests y el proceso sobreviven enteros; se rehace solo el paso de deploy, las variables y la migración de la base (`pg_dump`/`pg_restore`). Con dominio propio, las URLs de Google, Mercado Pago y CORS no cambian.

### E4. Pagos reales con Mercado Pago
- [ ] Decidir D3 con el cliente.
- [ ] Mostrar la política de cancelación **antes de pagar** la seña.

Niveles de devolución de la seña (de menor a mayor trabajo). Hoy, con la seña paga, el cliente no puede cancelar desde la app: lo hace el staff, y cualquier devolución es manual desde Mercado Pago.
1. **Lo actual + política visible.** El staff decide cada caso y devuelve a mano. Sin código, salvo el texto. ← propuesto para inaugurar.
2. **Regla de 24 h, devolución manual.** El cliente cancela solo; con más de 24 h, la reserva queda en "devolución pendiente" en el panel y el recepcionista la devuelve desde Mercado Pago y la marca como hecha.
3. **Regla de 24 h, devolución automática** por la API de Mercado Pago. Lo más caro: necesita manejar devoluciones rechazadas o sin respuesta (estado intermedio, reintentos, aviso al staff). Solo si el volumen lo justifica.
- [ ] Pasar de `sandbox_init_point` a `init_point` en producción (hoy `SpaceDetailPage` usa siempre la URL de pruebas).
- [ ] Credenciales reales y URL del webhook de producción.
- [ ] Probar un pago real chico de punta a punta.

### E5. Cuentas en producción
- [ ] Deploy de "Mail y cuentas" (ver la sección del mismo nombre en `RESUMEN-PROYECTO.md`): migraciones `009` y `010`, variables, dominio propio en Resend.
- [ ] Google Cloud: agregar el dominio de producción y publicar la app para que entren clientes reales (hoy solo entran los usuarios de prueba).
- [ ] Restablecer contraseña por mail.
- [ ] **Cerrar el riesgo de vinculación de Google por email** (D7, decidido el 2026-11-09: opción A del riesgo aceptado en `RESUMEN-PROYECTO.md`). Al vincular Google a una cuenta existente, se borra su contraseña: quien se registró con un email ajeno pierde el acceso. Si la cuenta era legítima, el dueño sigue entrando con Google y puede volver a tener contraseña con "restablecer contraseña", así que conviene hacerlo **después** de esa tarea. Avisarle al usuario al vincular que su contraseña anterior dejó de funcionar. Al terminar, actualizar la sección del riesgo en `RESUMEN-PROYECTO.md`.

### E6. Cobros en el complejo: saldo de la cancha y ventas extra
**Hueco encontrado (2026-11-09):** el backend ya permite cobrar el saldo (link de Mercado Pago con `kind: "balance"`, o marcarlo pagado en efectivo, transferencia o posnet con `PATCH /bookings/:id/payment`), pero **el panel no tiene ninguna de las dos opciones**. Hoy no se puede registrar desde la app que alguien pagó la cancha completa. Es la base del cierre de caja (E7).

Diseño decidido (2026-11-09, D4):
- **Catálogo:** productos con nombre, precio y tipo (alquiler o venta). **Solo el admin** crea productos y cambia precios; el recepcionista solo los usa al cobrar (un precio mal cargado afecta lo que se cobra y el cierre de caja).
- **Sin control de stock** para inaugurar (queda en E16).
- **Consumos:** producto + cantidad + **precio del momento** (si el precio cambia, las cuentas viejas no). Pueden ir colgados de una reserva o sueltos (venta de mostrador).
- **Una cuenta por reserva:** saldo de la cancha + consumos, cobrados juntos con un solo botón "Cobrar": efectivo, transferencia, posnet o un solo link de Mercado Pago por el total.
- **Panel del recepcionista:** en cada reserva, botones rápidos ("+ Raqueta", "+ Pelotas", "+ Coca") y "Cobrar". Botón aparte "Venta de mostrador" para compras sin reserva, cobradas en el momento.
- Ojo: si se genera el link de Mercado Pago y después se agrega un consumo, el link queda con el monto viejo; hay que regenerarlo o avisar.

Tareas:
- [ ] "Cobrar" el saldo desde el panel (usa lo que ya existe en el backend).
- [ ] Catálogo de productos.
- [ ] Consumos en una reserva y venta de mostrador.
- [ ] Cobro conjunto (saldo + consumos), con un solo link de Mercado Pago por el total.

### E7. Cierre de caja y reportes
Decidido (2026-11-09, D5):
- **Por día de cobro** (lo que entró ese día), no por día de la reserva: así cuadra con el cajón y con Mercado Pago. Ejemplo: una seña pagada el lunes para el sábado cuenta el lunes.
- **Separado por medio de pago:** efectivo (lo que tiene que haber en el cajón), transferencia, posnet y Mercado Pago.
- **El recepcionista cierra y el admin revisa.**
- **Correcciones:** al cerrar, el día queda bloqueado para el recepcionista (no puede cargar ni modificar cobros de esa fecha). Solo el admin corrige, y cada corrección queda registrada (qué cambió, quién, cuándo y por qué); el cierre muestra que fue corregido. Si al recepcionista se le pasó algo, le avisa al admin.
- La vista "por día de reserva" (cuánto facturó cada día de canchas) va en el resumen mensual, como información para el negocio.

**Hueco encontrado (2026-11-09):** los pagos manuales no guardan fecha ni hora. Cuando el staff marca un pago en efectivo, `bookings` guarda estado, medio y quién (`*_status`, `*_method`, `*_recorded_by`), pero no cuándo. Solo los pagos de Mercado Pago quedan con fecha en `payments`. Sin eso no hay cierre por día de cobro.

Tareas:
- [ ] **Registro de cobros:** una fila por cada plata que entra (seña, saldo, consumo o venta de mostrador), con monto, medio, fecha y hora, y quién cobró. Se comparte con E6. Cambia cómo se guardan los pagos: migración propia y cuidado con los datos existentes.
- [ ] Cierre del día: lo esperado por medio de pago, el efectivo contado por el recepcionista y la diferencia.
- [ ] Revisión del cierre por el admin.
- [ ] Bloqueo del día cerrado y correcciones del admin con registro.
- [ ] Exportar el cierre a PDF.
- [ ] Resumen del mes (incluye la vista por día de reserva).

### E8. Legal y seguridad
D6 decidido (2026-11-09): las dos cosas, lo legal y lo técnico. Checklist armado a partir del estado del código a esa fecha; se puede completar con un test de seguridad externo.

**Lo que ya está bien:**
- Contraseñas guardadas con bcrypt (costo 12), nunca en texto plano.
- CORS limitado a un solo origen (`ALLOWED_ORIGIN`).
- El webhook de Mercado Pago no confía en lo que le mandan: vuelve a consultar el pago a la API de Mercado Pago.
- Los datos de tarjeta nunca pasan por la app: los maneja Mercado Pago.
- Login con Google verificado del lado del backend (firma, audiencia, emisor y email verificado).
- Permisos por rol en el backend (`RequireRole`), no solo escondiendo botones en el frontend.
- Los `.env` no se suben a git.

**Imprescindible antes de inaugurar:**
- [ ] **Términos y condiciones** y **Política de privacidad**: Google los puede pedir al publicar la app, y la app guarda nombre, teléfono y email, alcanzados por la Ley 25.326 de protección de datos personales. Incluir la política de cancelación (D3). Conviene que lo revise alguien con conocimiento legal.
- [ ] **Límite de intentos** en login, registro, login con Google y restablecer contraseña. Hoy no hay ninguno: alguien podría probar contraseñas sin fin contra un teléfono.
- [ ] **Backups de la base de PROD**, y probar al menos una vez que se pueden restaurar. Revisar qué incluye el plan del proveedor elegido (D1).
- [ ] **Contraseña mínima de 8 caracteres** (hoy son 6) y contraseñas fuertes para las cuentas de staff.
- [ ] **Secretos de PROD:** solo en las variables del proveedor y en el environment `production` de GitHub. Rotar cualquiera que se haya compartido o pegado en un chat.
- [ ] **Dependencias con vulnerabilidades conocidas:** `govulncheck` (Go) y `npm audit` (frontend) en el CI, y Dependabot en GitHub.
- [ ] Cerrar el riesgo de vinculación de Google por email (ver E5).

**Conveniente:**
- [ ] **Sesiones de staff dado de baja:** hoy un token sigue valiendo hasta 24 h aunque se desactive el usuario. Para el staff conviene que el backend chequee en cada pedido que el usuario siga activo.
- [ ] **Headers de seguridad** (HSTS, `X-Content-Type-Options`, `X-Frame-Options` o CSP), en Vercel y en el backend.
- [ ] **Escaneo de seguridad automático** contra **QA**, nunca contra PROD: por ejemplo, el baseline scan de OWASP ZAP, que es gratuito y busca problemas comunes.
- [ ] Revisar que los errores no muestren información interna y que los logs no guarden datos sensibles (contraseñas, tokens).

### E9. Calidad
- [x] Arreglar los errores de lint que ya existían (eran 11). Hecho el 2026-11-09 (PR #1).
- [x] Tests unitarios del frontend con **Vitest**, solo de lógica pura (no pantallas). Hecho el 2026-11-09: fechas y horas en `src/lib/dates.ts` y perfil en `src/lib/profile.ts`, corren en el CI.
- [ ] 2 o 3 E2E con **Playwright** contra QA, cuando QA exista: entrar, reservar hasta llegar a Mercado Pago, reserva manual desde el panel. Más adelante, el cierre de caja. Pocos y críticos; cada test crea y limpia sus datos.
- [ ] Revisar que los tests del backend sigan cubriendo lo nuevo (hoy: service 82%, controllers 45%, repository 63%).
- [ ] Verificar mobile (375px y 768px).
- [ ] Probar el panel a mano (reserva manual, turno fijo, bloqueo, cancelar) y el flujo de reserva hasta el pago.
- [ ] Menores pendientes de la revisión de código (ver "Pendientes técnicos" en `RESUMEN-PROYECTO.md`).

### E10. Contenido y documentación
- [ ] Fotos faltantes (hero, "El complejo", padbol y beach vóley).
- [ ] Datos reales en `content.ts`: WhatsApp, link de Google Maps, dirección y cantidades.
- [ ] Reescribir `README.md` (hoy desactualizado).
- [ ] Borrar o rehacer `DESIGN-SYSTEM.md` (describe el diseño anterior) y borrar `frontend/README.md` (plantilla de Vite).

### E18. Monitoreo de PROD y dominio propio
- [ ] **Dominio propio** (p. ej. `h1canchas.com.ar`). Lo necesitan Resend (dominio verificado, para no caer en spam), Google Cloud (origen autorizado) y Mercado Pago (webhook y URLs de vuelta). Con dominio propio, cambiar de servidor no obliga a cambiar esas URLs.
- [ ] **Aviso si la app se cae:** monitor de disponibilidad gratuito (p. ej. UptimeRobot) contra `/health` de PROD, con aviso por mail.
- [ ] **Aviso de errores:** herramienta tipo Sentry en backend y frontend, para enterarse de un error con su detalle antes de que un cliente se queje.

### E19. Puesta en marcha
- [ ] **Carga de datos reales en PROD:** canchas, horarios, precios (D9), productos (E6) y cuentas del staff con contraseñas fuertes.
- [ ] **Guía corta para el recepcionista:** cómo reservar a mano, cobrar, cargar consumos y cerrar la caja.
- [ ] **Plan B si la app se cae** en plena jornada: cómo se anotan las reservas y los cobros (aunque sea en una planilla) para cargarlos después.

---

## Después de la inauguración

### E11. Avisos al cliente
Recordatorios de la reserva y mails de cancelación y de vencimiento, reusando el `EmailSender`.

### E12. Promociones para clientes frecuentes

### E13. Torneos y ranking

### E14. Partidos abiertos
Buscar jugadores con un ranking parecido. Quien reserva una cancha a cierta hora puede abrirla para que otros se sumen hasta completar la cantidad de jugadores.

### E15. Paginación
En `GET /bookings`, `GET /admin/bookings` y `GET /spaces`. Hace falta cuando crezca la cantidad de reservas.

### E16. Stock de productos
Control de stock del catálogo de E6: cantidades, compras al proveedor, ajustes por rotura o pérdida, aviso cuando se acaba y no alquilar lo que está en uso. Descartado para la inauguración (2026-11-09).

### E17. Más guardianes de calidad
- Umbral de cobertura en el CI, anclado a la medición real (no un número alto "porque sí").
- Análisis estático con CodeQL: busca vulnerabilidades leyendo el código. Gratis en repos públicos.

---

## Descartado por ahora
Para no volver a discutirlo sin un motivo nuevo:
- **Imágenes Docker y "build once, deploy many":** Vercel y Render construyen desde el repo, y `go.sum` y `package-lock.json` fijan las dependencias. Mucho trabajo para un riesgo chico a esta escala.
- **Blue-green, canary y feature flags:** pocos usuarios; una caída corta cuesta poco. Se reconsidera si el negocio crece.
- **Mutation testing:** solo como ejercicio puntual, no en el pipeline.

---

## Ideas sueltas

---

## Hecho
- **Mail y cuentas** (2026-11-09): mail de confirmación, login con Google, "Mi perfil", teléfono obligatorio para reservar. Falta solo su deploy (E5).
