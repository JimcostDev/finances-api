# finances-api

API REST en **Go** para la app **MyFinances**: usuarios, autenticación por sesión (JWT en cookie HttpOnly o cabecera `Authorization`), categorías, reportes mensuales, balance general y reportes anuales (ingresos, gastos, diezmos/ofrendas según configuración del usuario).

Arquitectura en capas (**Clean Architecture**): `handlers` → `services` → `repositories` → MongoDB.

## Stack

| Componente | Detalle |
|--------------|---------|
| Lenguaje | Go 1.26 |
| HTTP | [Fiber v2](https://gofiber.io/) |
| Base de datos | MongoDB (driver oficial), base `finances` |
| Auth | JWT (HMAC), contraseñas con **bcrypt** |
| CORS | Credenciales habilitadas para el front con cookies cross-origin |

## Requisitos

- Go 1.26+ (ver `go.mod`)
- Instancia de **MongoDB** accesible (URI en variable de entorno)

## Variables de entorno

| Variable | Obligatoria | Descripción |
|----------|-------------|-------------|
| `MONGO_URI` | Sí | Cadena de conexión MongoDB (p. ej. `mongodb://...`) |
| `JWT_SECRET_KEY` | Sí | Secreto para firmar y verificar JWT |
| `CORS_ORIGINS` | No | Orígenes permitidos separados por **coma** (por defecto incluye `localhost:4321` y el dominio del front). Tras proxy (Koyeb, etc.) el servidor usa `X-Forwarded-Proto` para cookies `Secure`. |
| `COOKIE_SECURE` | No | Si vale `true`, la cookie de sesión se marca `Secure` (HTTPS recomendado en producción) |

## Ejecución local

```bash
cd finances-api
export MONGO_URI="mongodb://localhost:27017"
export JWT_SECRET_KEY="tu-secreto-largo-y-aleatorio"
go run .
```

El servidor escucha en el puerto **3000** (`:3000`).

- Health: `GET /` → `{"message":"Hola Mundo"}`

## Docker

```bash
docker build -t finances-api .
docker run --rm -p 3000:3000 -e MONGO_URI=... -e JWT_SECRET_KEY=... finances-api
```

## Rutas API (resumen)

Prefijos bajo el mismo host (ej. `https://tu-api.com`).

### Autenticación — `api/auth`

| Método | Ruta | Protegida |
|--------|------|-----------|
| POST | `/api/auth/register` | No |
| POST | `/api/auth/login` | No |
| POST | `/api/auth/logout` | No |
| GET | `/api/auth/me` | Sí (JWT) |

El token se envía en la cookie **`finances_access_token`** (HttpOnly) o como `Authorization: Bearer <token>`.

### Reportes — `api/reports` (todas protegidas)

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/reports/general-balance` | Balance histórico |
| GET | `/api/reports/annual` | Reporte anual |
| GET | `/api/reports/by-month` | Filtro por mes/año |
| GET | `/api/reports` | Listado |
| POST | `/api/reports` | Crear reporte |
| GET/PUT/DELETE | `/api/reports/:id` | CRUD por ID |
| POST/DELETE | `/api/reports/:id/income`, `.../income/:income_id` | Ingresos |
| POST/DELETE | `/api/reports/:id/expense`, `.../expense/:expense_id` | Gastos |

### Usuarios — `api/users` (protegidas)

| Método | Ruta |
|--------|------|
| GET | `/api/users/profile` |
| PUT | `/api/users/profile` |
| DELETE | `/api/users/profile` |

### Categorías — `api/categories` (protegidas)

| Método | Ruta |
|--------|------|
| GET | `/api/categories` |

## Estructura del repositorio

| Carpeta | Rol |
|---------|-----|
| `config/` | Conexión MongoDB (`ConnectDB`) |
| `handlers/` | HTTP: entrada/salida JSON, valida inputs |
| `services/` | Lógica de negocio y cálculos |
| `repositories/` | Acceso a MongoDB |
| `models/` | Structs BSON/JSON |
| `routes/` | Registro de rutas e inyección de dependencias |
| `middleware/` | JWT, cookie de sesión (`AuthCookieName`) |
| `main.go` | Fiber, CORS, DB, rutas |

Flujo: `Request` → `Handler` → `Service` → `Repository` → MongoDB.

## Frontend

El cliente web está en **`../finances-ui`** (Astro + React); debe apuntar la URL base del API en su configuración y usar el mismo origen CORS que declares en `CORS_ORIGINS`.
