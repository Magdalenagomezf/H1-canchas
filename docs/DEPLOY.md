# Deploy: QA y producción

Cómo se publica H1 Canchas, cómo se arma cada entorno y cómo se vuelve atrás.

## Cómo funciona

```
merge a main → CI verde → deploy a QA → smoke → integración (API) → E2E contra QA
            → aprobación manual → deploy a PROD → smoke
```

- **Solo el pipeline publica.** Render y Vercel tienen el deploy automático apagado. Si algo no pasó por el pipeline, no está publicado.
- **Siempre el commit verificado.** Cada corrida queda atada al commit del merge que la disparó. Render recibe ese commit en el deploy hook (parámetro `ref=<sha>`) y Vercel construye el frontend desde un checkout de ese mismo commit. Nunca se publica "lo último de `main`".
- **Si QA falla, PROD ni se ofrece.** Si el smoke, la integración o los E2E de QA fallan, el job de PROD no arranca y no pide aprobación.
- **No se vuelve atrás por accidente.** Si una corrida vieja se aprueba tarde y PROD ya corre un commit más nuevo, el deploy se rechaza. Es un control de mejor esfuerzo: si el backend de PROD no responde o informa un commit desconocido, el deploy sigue (es la forma de arreglar un PROD caído).

Archivos: `.github/workflows/deploy.yml` (las etapas), `.github/actions/deploy/action.yml` (deploy de un entorno + smoke) y `scripts/smoke.sh`.

| Entorno | Base | Backend | Frontend | Mercado Pago |
|---|---|---|---|---|
| QA | Supabase (el proyecto que existía) | Render (el servicio que existía) | Vercel (el proyecto que existía) | credenciales de prueba |
| PROD | Supabase (proyecto nuevo) | Render (servicio nuevo) | Vercel (proyecto nuevo) | credenciales de prueba hasta decidir D3, después las reales |

## Configuración inicial

Hacela **antes de mergear el PR que agrega el pipeline**. Ese merge dispara el primer deploy y, a partir de ahí, Vercel deja de publicar solo.

### 1. QA (los servicios que ya existen)

1. **Render (backend QA):** Settings → Deploy Hook → copiá la URL. Es un secreto: quien la tenga puede disparar deploys.
   En Settings → Build & Deploy, poné **Auto-Deploy: Off**.
2. **Vercel (frontend QA):** Settings → General → copiá el **Project ID**. En Settings de tu cuenta o equipo, copiá el **Team ID** (o el ID de la cuenta personal): es el `VERCEL_ORG_ID`.
3. **Token de Vercel:** Account Settings → Tokens → creá uno (por ejemplo `github-deploy`). Sirve para los dos entornos.
4. **Admin para los E2E:** desde el panel de QA, creá un usuario `admin` solo para los tests (por ejemplo "E2E Admin"), con una contraseña larga y aleatoria. Los tests lo usan para crear y borrar sus propios datos. Si un E2E falla a mitad de camino, puede dejar datos sueltos en QA.

### 2. PROD (servicios nuevos)

1. **Supabase:** creá un proyecto nuevo y aplicá las migraciones de `backend/migrations/` en orden (001 → 011). El plan gratuito permite 2 proyectos activos, así que con QA + PROD queda completo.
2. **Render:** creá un Web Service nuevo desde el mismo repo.
   - Root directory: `backend`
   - Build: `go build -tags netgo -ldflags '-s -w' -o app ./cmd`
   - Start: `./app`
   - Variables: las de `backend/.env.example` con valores de PROD. `DATABASE_URL` es la de la base nueva; `ALLOWED_ORIGIN` y `FRONTEND_URL` son la URL del frontend de PROD; `MP_WEBHOOK_URL` es `<backend PROD>/payments/webhook`. `JWT_SECRET` tiene que ser **distinto** del de QA. `MP_ACCESS_TOKEN` y `MP_WEBHOOK_SECRET` van con credenciales de prueba hasta decidir D3.
   - Apenas se cree, en Settings → Build & Deploy poné **Auto-Deploy: Off**. Si no, cualquier merge a `main` llega a PROD sin aprobación hasta el primer deploy del pipeline.
   - Copiá su Deploy Hook.
3. **Vercel:** creá un proyecto nuevo desde el mismo repo, con root directory `frontend`. Variables (entorno Production): `VITE_API_URL` = URL del backend de PROD y `VITE_GOOGLE_CLIENT_ID`. Copiá su Project ID.
4. **Google Cloud:** en el cliente OAuth, agregá la URL del frontend de PROD a los orígenes autorizados de JavaScript.
5. **Mercado Pago:** creá una **aplicación nueva** en Mercado Pago Developers para PROD. El webhook se configura en el panel, uno por aplicación (el backend no manda `notification_url` en cada pago). Si QA y PROD compartieran la aplicación, los avisos de pago llegarían a un solo backend. En la aplicación de PROD, configurá Webhooks → URL `<backend PROD>/payments/webhook` (evento Pagos) y usá sus credenciales de prueba y su clave secreta en `MP_ACCESS_TOKEN` y `MP_WEBHOOK_SECRET`.
6. **Primer admin:** registrate desde el frontend de PROD y, en el SQL editor de Supabase, ejecutá `UPDATE users SET role = 'admin' WHERE phone = '+549...';` con tu teléfono.

### 3. GitHub

En Settings → Environments, creá dos environments:

**`qa`** (sin revisores)

| Tipo | Nombre | Valor |
|---|---|---|
| Secret | `RENDER_DEPLOY_HOOK_URL` | deploy hook del backend QA |
| Secret | `VERCEL_TOKEN` | token de Vercel |
| Secret | `VERCEL_ORG_ID` | Team ID / ID de la cuenta |
| Secret | `VERCEL_PROJECT_ID` | Project ID del frontend QA |
| Secret | `E2E_ADMIN_PHONE` | teléfono del admin de los E2E y la integración |
| Secret | `E2E_ADMIN_PASSWORD` | contraseña del admin de los E2E y la integración |
| Variable | `API_URL` | URL del backend QA, sin `/` final |
| Variable | `FRONTEND_URL` | URL del frontend QA |

**`production`**

- Required reviewers: vos. El job de PROD queda en pausa hasta que lo apruebes.
- Deployment branches: solo `main`.

| Tipo | Nombre | Valor |
|---|---|---|
| Secret | `RENDER_DEPLOY_HOOK_URL` | deploy hook del backend PROD |
| Secret | `VERCEL_TOKEN` | token de Vercel |
| Secret | `VERCEL_ORG_ID` | Team ID / ID de la cuenta |
| Secret | `VERCEL_PROJECT_ID` | Project ID del frontend PROD |
| Variable | `API_URL` | URL del backend PROD, sin `/` final |
| Variable | `FRONTEND_URL` | URL del frontend PROD |

Los secrets de PROD viven solo en `production`: ningún otro job puede leerlos.

## Pasar un cambio a PROD

1. Mergeá el PR a `main`.
2. Esperá a que el workflow **Deploy** termine el job de QA (deploy, smoke, integración y E2E). Si querés, probalo a mano en QA.
3. En Actions → la corrida → **Review deployments** → marcá `production` → **Approve and deploy**.
4. El job de PROD despliega y corre el smoke. Verde = publicado.

Si no querés publicar esa versión, rechazala. La próxima corrida trae todo lo anterior. Si hay una corrida vieja esperando aprobación y querés publicar una más nueva, **rechazá primero la vieja**: mientras espera, la nueva queda en cola detrás.

### Migraciones

El pipeline **no** aplica migraciones: siguen siendo manuales. Si un PR agrega una migración:

1. Aplicala en la base de QA **antes de mergear**.
2. Aplicala en la base de PROD **antes de aprobar**.

Escribí migraciones compatibles con la versión anterior del código (agregar columnas o tablas, no renombrar ni borrar en el mismo paso), así el código viejo sigue andando mientras se despliega el nuevo.

## Volver atrás (rollback)

- **Lo normal:** revertí el PR (`git revert`, en un PR nuevo). Pasa por el pipeline como cualquier cambio.
- **Emergencia (PROD roto ya):** en Vercel, Deployments → el deploy anterior → **Instant Rollback**. En Render, Events → el deploy anterior → **Rollback**. Después igual revertí el PR, porque si no la próxima corrida vuelve a publicar el código roto.
- Si el problema es una migración, el rollback del código no la deshace: corregila con una migración nueva.

## Keep-alive (planes gratuitos)

Supabase pausa una base gratuita tras una semana sin uso, y queda caída hasta reanudarla a mano desde el dashboard. `.github/workflows/keep-alive.yml` corre `scripts/smoke.sh` contra QA y PROD cada 3 días: `/health` hace una consulta real a la base y eso cuenta como uso. También se puede correr a mano (Actions → Keep alive → Run workflow).

- Si falla, ese entorno está caído: GitHub te manda un mail. Revisá primero si Supabase pausó la base.
- Cuando la base de un entorno pase a un plan pago, sacalo de la matriz del workflow; cuando no quede ninguno gratis, borrá el workflow.
- GitHub desactiva los workflows programados si el repo pasa 60 días sin actividad. Se reactiva desde Actions.

## Qué no hacer

- **No uses "Manual Deploy"** en Render ni redeploys desde el panel de Vercel en el día a día: se saltean los tests y QA. Son salidas de emergencia.
- No vuelvas a prender el auto-deploy de los proveedores.
- No pongas secrets de PROD en el environment `qa` ni como secrets del repositorio.
