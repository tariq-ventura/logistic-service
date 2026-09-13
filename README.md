# Logistic Service

Microservicio que administra el inventario de **maquinaria Prisma** y las **solicitudes de maquinaria** de los proyectos, junto con su asignación, liberación e historial de estados.

El servicio forma parte de la plataforma Entropy, en la que:

- `logistic-service` administra la maquinaria Prisma y las solicitudes de los proyectos.
- `fleet-service` administra los vehículos, tareas, mantenimientos y geocercas de Startrack.
- `mcp-server` construye la proyección unificada que correlaciona ambos inventarios.
- PostgreSQL almacena la maquinaria, las solicitudes y su historial de cambios de estado.

Está desarrollado en **Go**, expone una API HTTP con **Gin**, utiliza **PostgreSQL** mediante **GORM** y cuenta con instrumentación de observabilidad basada en **OpenTelemetry**.

> **Cambio importante:** el motor de recomendaciones y toda la integración HTTP hacia `fleet-service` fueron retirados. Este servicio ya no consulta a Fleet ni requiere `FLEET_SERVICE_URL`. La asignación de maquinaria ahora se resuelve internamente contra la tabla `prisma_machinery`.

## Tabla de contenido

- [Responsabilidades](#responsabilidades)
- [Arquitectura](#arquitectura)
- [Tecnologías](#tecnologías)
- [Requisitos](#requisitos)
- [Variables de entorno](#variables-de-entorno)
- [Levantar el proyecto localmente](#levantar-el-proyecto-localmente)
- [Modelo de dominio](#modelo-de-dominio)
- [Estados](#estados)
- [Endpoints](#endpoints)
- [API de maquinaria](#api-de-maquinaria)
- [API de solicitudes](#api-de-solicitudes)
- [Asignación de maquinaria](#asignación-de-maquinaria)
- [Casos de uso](#casos-de-uso)
- [Carga de datos sintéticos](#carga-de-datos-sintéticos)
- [Manejo de errores](#manejo-de-errores)
- [Observabilidad](#observabilidad)
- [CORS](#cors)
- [Pruebas](#pruebas)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Notas conocidas de la implementación](#notas-conocidas-de-la-implementación)

---

## Responsabilidades

Actualmente `logistic-service` permite:

- registrar maquinaria Prisma con su número de activo, clase y estado;
- consultar maquinaria con paginación, filtros y búsqueda libre;
- actualizar los datos maestros de una maquinaria;
- cambiar el estado operativo de una maquinaria;
- eliminar maquinaria mediante borrado lógico;
- crear solicitudes de maquinaria para proyectos;
- consultar y listar solicitudes con filtros y paginación;
- realizar búsquedas avanzadas por múltiples estados mediante `POST /requests/search`;
- actualizar solicitudes mientras estén en estado `Pendiente`;
- eliminar solicitudes pendientes sin maquinaria asignada;
- asignar una maquinaria a una solicitud de forma transaccional;
- liberar la maquinaria asignada y devolver la solicitud a `Pendiente`;
- almacenar el historial de cambios de estado de cada solicitud;
- exponer health check HTTP;
- utilizar logging local o Google Cloud Logging/Error Reporting.

---

## Arquitectura

```text
                  +----------------------+
                  |      Cliente / UI    |
                  |     n8n / MCP        |
                  +----------+-----------+
                             |
                             | HTTP REST
                             v
                  +----------------------+
                  |   logistic-service   |
                  |       Go + Gin       |
                  +----+------------+----+
                       |            |
                       | GORM       | GORM
                       v            v
              +----------------+  +--------------------------+
              | prisma_machinery|  | prisma_machinery_requests|
              +----------------+  +--------------------------+
                                        |
                                        v
                          +-------------------------------+
                          | prisma_request_status_history  |
                          +-------------------------------+
```

La asignación de maquinaria ocurre en una sola transacción de PostgreSQL que toca las tres tablas:

```text
PATCH /requests/{id}/assignment
            |
            v
SELECT ... FOR UPDATE sobre la solicitud
            |
            v
SELECT ... FOR UPDATE sobre la maquinaria
            |
            v
¿Solicitud Pendiente?  ──no──> 409 assignment_conflict
            |sí
            v
¿Maquinaria Disponible? ──no──> 409 assignment_conflict
            |sí
            v
¿type == equipmentClass? ──no──> 409 assignment_conflict
            |sí
            v
Solicitud: machinery = assetNumber, status = Aprobada
Maquinaria: status = Ocupada
Historial: Pendiente -> Aprobada
            |
            v
          200 OK
```

Cada módulo de dominio sigue la misma estructura de cuatro capas:

```text
domain/    Modelo GORM, enums y nombre de tabla
dto/       Contratos de entrada y sus validaciones
db/        Interfaz de persistencia + implementación PostgreSQL
handlers/  Adaptadores HTTP de Gin
```

---

## Tecnologías

### Lenguaje y framework

| Tecnología | Uso |
|---|---|
| Go 1.27 | Lenguaje principal |
| Gin 1.12 | API HTTP REST |
| GORM 1.31 | ORM y acceso a PostgreSQL |
| PostgreSQL 17 | Persistencia |
| Google UUID | Identificadores de entidades |

### Observabilidad

| Tecnología | Uso |
|---|---|
| OpenTelemetry | Instrumentación y trazas |
| OTLP gRPC | Exportación estándar de trazas |
| Google Cloud Telemetry | Exportación de trazas a GCP |
| Google Cloud Logging | Logging en GCP |
| Google Error Reporting | Reporte de errores en GCP |
| Logrus | Logging local |

### Desarrollo y ejecución

| Tecnología | Uso |
|---|---|
| Docker | Construcción del servicio |
| Docker Compose | Entorno local |
| curl | Pruebas HTTP |
| jq | Validación de respuestas en scripts |

---

## Requisitos

Para ejecución local directa:

- Go `1.27` compatible con el `go.mod`;
- PostgreSQL;
- `curl` y `jq` si se ejecutarán scripts de prueba o de carga de datos.

Alternativamente se puede utilizar Docker y Docker Compose.

> Este servicio ya **no** depende de `fleet-service` en tiempo de arranque ni de ejecución. Puede levantarse de forma completamente independiente.

---

## Variables de entorno

| Variable | Requerida | Ejemplo | Descripción |
|---|---:|---|---|
| `DB_CONTEXT` | Sí | `postgresql` | Backend de base de datos. Actualmente solo PostgreSQL está implementado. |
| `DB_STRING` | Sí | `host=localhost user=mongo password=1234 dbname=backend_golang_gin port=5435 sslmode=disable` | DSN de PostgreSQL. |
| `TRACE_TYPE` | Sí | `STDOUT` | Exportador de trazas: `STDOUT`, `OTLP`, `GCP`, `NONE` o `DISABLED`. |
| `SERVICE_NAME` | Sí | `logistic-service` | Nombre utilizado por OpenTelemetry. |
| `PORT` | No | `3001` | Puerto HTTP. Por defecto `3001`. |
| `LOGGING_TYPE` | No | `local` | Usar `GCP` para Cloud Logging; cualquier otro valor usa logger local. |
| `SERVICE_VERSION` | No | `0.1.0` | Versión agregada al recurso OpenTelemetry. |
| `ENVIRONMENT` | No | `local` | Ambiente agregado a las trazas. |
| `GCP_PROJECT_ID` | Solo GCP | `my-project` | Proyecto utilizado por logging/tracing en Google Cloud. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Solo OTLP | `collector:4317` | Endpoint OTLP general. |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Solo OTLP | — | Endpoint específico de trazas OTLP. |

Las variables `DB_USER`, `DB_PASS` y `DB_NAME` del `.env.example` se utilizan principalmente para inicializar PostgreSQL desde Docker Compose.

> `FLEET_SERVICE_URL` ya no se lee en ninguna parte del código. Sigue declarada en los manifiestos de `cdeploy/`, donde puede retirarse.

### Ejemplo recomendado para desarrollo

```env
DB_USER=mongo
DB_PASS=1234
DB_NAME=backend_golang_gin
DB_CONTEXT=postgresql
DB_STRING=host=localhost user=mongo password=1234 dbname=backend_golang_gin port=5435 sslmode=disable

TRACE_TYPE=STDOUT
SERVICE_NAME=logistic-service
SERVICE_VERSION=0.1.0
ENVIRONMENT=local
LOGGING_TYPE=local
PORT=3001
```

---

## Levantar el proyecto localmente

### Opción 1: Go + PostgreSQL en Docker

#### 1. Crear PostgreSQL

```bash
docker run --name logistic-postgres \
  -e POSTGRES_DB=backend_golang_gin \
  -e POSTGRES_USER=mongo \
  -e POSTGRES_PASSWORD=1234 \
  -p 5435:5432 \
  -d postgres:17
```

#### 2. Configurar variables

```bash
export DB_CONTEXT=postgresql
export DB_STRING='host=localhost user=mongo password=1234 dbname=backend_golang_gin port=5435 sslmode=disable'

export TRACE_TYPE=STDOUT
export SERVICE_NAME=logistic-service
export SERVICE_VERSION=0.1.0
export ENVIRONMENT=local
export LOGGING_TYPE=local
export PORT=3001
```

#### 3. Descargar dependencias

```bash
go mod download
go mod tidy
```

Si se presenta un problema HTTP/2 contra `sum.golang.org` o el proxy de módulos:

```bash
GODEBUG=http2client=0 go mod tidy
```

#### 4. Ejecutar

```bash
go run ./cmd/services/main.go
```

La API quedará disponible en:

```text
http://localhost:3001
```

Al iniciar, GORM ejecuta `AutoMigrate` para las tres tablas del servicio.

#### 5. Health check

```bash
curl http://localhost:3001/health
```

Respuesta esperada:

```json
{
  "Status": "Up and Running"
}
```

---

### Opción 2: Docker Compose

El repositorio contiene un `docker-compose.yml`, pero la versión actual publica PostgreSQL como `5435:5434` cuando el contenedor escucha internamente en `5432`. Un ejemplo mínimo corregido de la parte relevante:

```yaml
services:
  postgres:
    image: postgres:17
    ports:
      - "5435:5432"
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASS}

  logistic-service:
    build: .
    ports:
      - "3001:3001"
    environment:
      DB_STRING: "host=postgres user=${DB_USER} password=${DB_PASS} dbname=${DB_NAME} port=5432 sslmode=disable"
      DB_CONTEXT: postgresql
      TRACE_TYPE: ${TRACE_TYPE}
      SERVICE_NAME: logistic-service
      PORT: "3001"
```

Luego:

```bash
docker compose up --build
```

---

## Modelo de dominio

### Maquinaria

Tabla `prisma_machinery`.

```json
{
  "id": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
  "company": "ECON",
  "assetNumber": "SYN-DEMO-01",
  "name": "Excavadora hidráulica 320",
  "equipmentClass": "Excavadora",
  "status": "Disponible",
  "createdAt": "2026-09-13T20:00:00Z",
  "updatedAt": "2026-09-13T20:00:00Z"
}
```

| Campo | Tipo | Notas |
|---|---|---|
| `id` | UUID | Generado por el servicio |
| `company` | string | Empresa propietaria |
| `assetNumber` | string | Número de activo. **Único** |
| `name` | string | Nombre descriptivo |
| `equipmentClass` | string | Clase de maquinaria. Debe coincidir con el `type` de la solicitud para poder asignarla |
| `status` | enum | Ver [Estados](#estados) |

### Solicitud

Tabla `prisma_machinery_requests`.

```json
{
  "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
  "project": "Ampliación carretera Los Chorros",
  "type": "Excavadora",
  "requester": "Gerencia Técnica",
  "location": "",
  "latitude": 0,
  "longitude": 0,
  "startDate": "2030-01-10T08:00:00Z",
  "endDate": "2030-01-25T17:00:00Z",
  "status": "Pendiente",
  "createdAt": "2026-09-13T20:00:00Z",
  "updatedAt": "2026-09-13T20:00:00Z"
}
```

| Campo | Tipo | Notas |
|---|---|---|
| `id` | UUID | Generado por el servicio |
| `project` | string | Nombre del proyecto |
| `type` | string | Clase de maquinaria requerida |
| `requester` | string | Área o persona solicitante |
| `location` | string | Ubicación textual. Ver [Notas conocidas](#notas-conocidas-de-la-implementación) |
| `latitude` | float | Grados decimales. Ver [Notas conocidas](#notas-conocidas-de-la-implementación) |
| `longitude` | float | Grados decimales. Ver [Notas conocidas](#notas-conocidas-de-la-implementación) |
| `startDate` | fecha | RFC 3339, almacenada en UTC |
| `endDate` | fecha | Debe ser posterior a `startDate` |
| `status` | enum | `Pendiente` o `Aprobada` |
| `machinery` | string, nullable | `assetNumber` de la maquinaria asignada. Se omite cuando es nulo |

> `machinery` guarda el **número de activo** en texto, no el UUID de la maquinaria. No existe llave foránea entre ambas tablas.

### Historial de estados

Tabla `prisma_request_status_history`.

```json
{
  "id": "5f37962e-d2ad-44ad-8388-f76978fda8ad",
  "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
  "fromStatus": "Pendiente",
  "toStatus": "Aprobada",
  "reason": "Maquinaria confirmada para la solicitud",
  "changedAt": "2026-09-13T21:00:00Z"
}
```

---

## Estados

### Estados de la maquinaria

| Valor | Significado |
|---|---|
| `Disponible` | Libre para ser asignada. Valor por defecto al crear |
| `Ocupada` | Asignada a una solicitud aprobada |
| `Mant. preventivo` | En mantenimiento programado |
| `Mant. correctivo` | En reparación |
| `Obsoletas` | Fuera de servicio |

El servicio normaliza los valores recibidos, sin distinguir mayúsculas:

```text
"disponible" | "available"                      -> Disponible
"ocupada" | "ocupado" | "occupied"              -> Ocupada
"mant. preventivo" | "mantenimiento preventivo" -> Mant. preventivo
"mant. correctivo" | "mantenimiento correctivo" -> Mant. correctivo
"obsoletas" | "obsoleta" | "obsolete"           -> Obsoletas
```

Cualquier otro valor responde `400 invalid_status`.

### Estados de la solicitud

Solo existen dos:

| Valor | Significado |
|---|---|
| `Pendiente` | Solicitud creada, sin maquinaria asignada. Valor por defecto al crear |
| `Aprobada` | Solicitud con maquinaria asignada |

Normalización aceptada:

```text
"pendiente" | "pending"                 -> Pendiente
"aprobada" | "aprobado" | "approved"    -> Aprobada
```

### Transiciones permitidas

```text
Pendiente <────────> Aprobada
```

Ambas direcciones son válidas. Las reglas adicionales son:

| Desde | Hacia | Condición |
|---|---|---|
| `Pendiente` | `Aprobada` | La solicitud ya debe tener `machinery` asignada |
| `Aprobada` | `Pendiente` | Libera la maquinaria: la deja en `Disponible` y pone `machinery` en `null` |
| Cualquiera | El mismo estado | Rechazado |

> Los estados `ASSIGNED`, `COMPLETED` y `CANCELLED` de la versión anterior ya no existen. La cancelación de una asignación se modela liberando la maquinaria.

Cada transición válida crea un registro en `prisma_request_status_history`.

---

# Endpoints

Base URL:

```text
http://localhost:3001
```

Base API:

```text
http://localhost:3001/api/v1
```

## Resumen

| Método | Endpoint | Descripción |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/equipments` | Registrar maquinaria |
| `GET` | `/api/v1/equipments` | Listar y filtrar maquinaria |
| `GET` | `/api/v1/equipments/:id` | Consultar maquinaria |
| `PATCH` | `/api/v1/equipments/:id` | Actualizar datos maestros |
| `PATCH` | `/api/v1/equipments/:id/status` | Cambiar estado de la maquinaria |
| `DELETE` | `/api/v1/equipments/:id` | Eliminar maquinaria |
| `POST` | `/api/v1/requests` | Crear solicitud |
| `GET` | `/api/v1/requests` | Listar y filtrar solicitudes |
| `POST` | `/api/v1/requests/search` | Búsqueda avanzada |
| `GET` | `/api/v1/requests/:requestID` | Consultar solicitud |
| `PATCH` | `/api/v1/requests/:requestID` | Modificar una solicitud `Pendiente` |
| `DELETE` | `/api/v1/requests/:requestID` | Eliminar una solicitud `Pendiente` sin maquinaria |
| `PATCH` | `/api/v1/requests/:requestID/status` | Cambiar estado |
| `GET` | `/api/v1/requests/:requestID/status-history` | Consultar historial |
| `PATCH` | `/api/v1/requests/:requestID/assignment` | Asignar maquinaria |
| `DELETE` | `/api/v1/requests/:requestID/assignment` | Liberar maquinaria |

---

## GET `/health`

```bash
curl http://localhost:3001/health
```

### `200 OK`

```json
{
  "Status": "Up and Running"
}
```

---

# API de maquinaria

## POST `/api/v1/equipments`

Registra una maquinaria en el inventario Prisma.

### Request

```bash
curl -X POST http://localhost:3001/api/v1/equipments \
  -H 'Content-Type: application/json' \
  -d '{
    "company": "ECON",
    "assetNumber": "SYN-DEMO-01",
    "name": "Excavadora hidráulica 320",
    "equipmentClass": "Excavadora",
    "status": "Disponible"
  }'
```

### Campos

| Campo | Tipo | Requerido | Validación |
|---|---|---:|---|
| `company` | string | Sí | Máximo 150 caracteres |
| `assetNumber` | string | Sí | Máximo 200 caracteres. Único |
| `name` | string | Sí | Máximo 200 caracteres |
| `equipmentClass` | string | Sí | Máximo 100 caracteres |
| `status` | string | No | Uno de los estados válidos. Si se omite, `Disponible` |

### `201 Created`

```json
{
  "message": "Maquinaria registrada correctamente",
  "data": {
    "id": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
    "company": "ECON",
    "assetNumber": "SYN-DEMO-01",
    "name": "Excavadora hidráulica 320",
    "equipmentClass": "Excavadora",
    "status": "Disponible",
    "createdAt": "2026-09-13T20:00:00Z",
    "updatedAt": "2026-09-13T20:00:00Z"
  }
}
```

### `400 Bad Request`

```json
{
  "error": "invalid_status",
  "message": "El estado de maquinaria no es válido"
}
```

### `409 Conflict`

```json
{
  "error": "equipment_already_exists",
  "message": "Ya existe una maquinaria con ese número de activo"
}
```

---

## GET `/api/v1/equipments`

### Query parameters

| Parámetro | Default | Descripción |
|---|---:|---|
| `page` | `1` | Página actual |
| `pageSize` | `20` | Registros por página; máximo efectivo `100` |
| `equipmentClass` | — | Coincidencia exacta, sin distinguir mayúsculas |
| `status` | — | Coincidencia exacta, sin distinguir mayúsculas |
| `search` | — | Busca en `company`, `assetNumber`, `name` y `equipmentClass` |

```bash
curl 'http://localhost:3001/api/v1/equipments?equipmentClass=Excavadora&status=Disponible&search=SYN-DEMO'
```

### `200 OK`

```json
{
  "data": [
    {
      "id": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
      "company": "ECON",
      "assetNumber": "SYN-DEMO-01",
      "name": "Excavadora hidráulica 320",
      "equipmentClass": "Excavadora",
      "status": "Disponible"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

Los resultados se ordenan por `createdAt` descendente.

---

## GET `/api/v1/equipments/:id`

```bash
curl http://localhost:3001/api/v1/equipments/7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d
```

### `400 Bad Request`

```json
{
  "error": "invalid_id",
  "message": "El identificador no es un UUID válido"
}
```

### `404 Not Found`

```json
{
  "error": "equipment_not_found",
  "message": "La maquinaria no existe"
}
```

---

## PATCH `/api/v1/equipments/:id`

Actualiza los datos maestros. **No** modifica el estado; para eso existe `/status`.

Campos admitidos:

```text
company
assetNumber
name
equipmentClass
```

```bash
curl -X PATCH http://localhost:3001/api/v1/equipments/7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d \
  -H 'Content-Type: application/json' \
  -d '{"name": "Excavadora hidráulica 320 GX"}'
```

### `200 OK`

```json
{
  "message": "Maquinaria actualizada correctamente",
  "data": {
    "id": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
    "name": "Excavadora hidráulica 320 GX"
  }
}
```

### `400 Bad Request` — actualización vacía

```json
{
  "error": "empty_update",
  "message": "Debe enviar al menos un campo"
}
```

---

## PATCH `/api/v1/equipments/:id/status`

```bash
curl -X PATCH http://localhost:3001/api/v1/equipments/7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d/status \
  -H 'Content-Type: application/json' \
  -d '{"status": "Mant. preventivo"}'
```

El único campo es `status` y es obligatorio.

### `200 OK`

La respuesta devuelve la maquinaria completa, sin envoltorio adicional:

```json
{
  "message": "Estado actualizado correctamente",
  "data": {
    "id": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
    "assetNumber": "SYN-DEMO-01",
    "status": "Mant. preventivo"
  }
}
```

Este endpoint no valida transiciones: cualquier estado válido puede pasar a cualquier otro, y tampoco comprueba si la maquinaria está asignada a una solicitud aprobada.

---

## DELETE `/api/v1/equipments/:id`

```bash
curl -i -X DELETE http://localhost:3001/api/v1/equipments/7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d
```

Responde **204 No Content**. El borrado es lógico. Si no existe, `404 equipment_not_found`.

---

# API de solicitudes

## POST `/api/v1/requests`

Crea una solicitud de maquinaria. El estado inicial siempre es `Pendiente`.

### Request

```bash
curl -X POST http://localhost:3001/api/v1/requests \
  -H 'Content-Type: application/json' \
  -d '{
    "project": "Ampliación carretera Los Chorros",
    "type": "Excavadora",
    "requester": "Gerencia Técnica",
    "startDate": "2030-01-10T08:00:00Z",
    "endDate": "2030-01-25T17:00:00Z"
  }'
```

### Campos

| Campo | Tipo | Requerido | Validación |
|---|---|---:|---|
| `project` | string | Sí | Máximo 250 caracteres |
| `type` | string | Sí | Máximo 100 caracteres. Debe coincidir con el `equipmentClass` de la maquinaria a asignar |
| `requester` | string | Sí | Máximo 200 caracteres |
| `startDate` | RFC3339 datetime | Sí | Fecha válida |
| `endDate` | RFC3339 datetime | Sí | Debe ser posterior a `startDate` |
| `status` | string | No | Si se envía, debe normalizar a `Pendiente` |

> El modelo tiene `location`, `latitude` y `longitude`, pero **no forman parte del contrato de creación**. Si se envían, se descartan. Ver [Notas conocidas](#notas-conocidas-de-la-implementación).

### `201 Created`

```json
{
  "message": "Solicitud registrada correctamente",
  "data": {
    "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "project": "Ampliación carretera Los Chorros",
    "type": "Excavadora",
    "requester": "Gerencia Técnica",
    "location": "",
    "latitude": 0,
    "longitude": 0,
    "startDate": "2030-01-10T08:00:00Z",
    "endDate": "2030-01-25T17:00:00Z",
    "status": "Pendiente"
  }
}
```

### `400 Bad Request`

```json
{
  "error": "invalid_request",
  "message": "Los datos enviados no son válidos",
  "detail": "..."
}
```

Si se envía un `status` distinto de `Pendiente`:

```json
{
  "error": "invalid_status",
  "message": "Una solicitud nueva debe iniciar Pendiente"
}
```

### `422 Unprocessable Entity`

Si `endDate <= startDate`:

```json
{
  "error": "invalid_period",
  "message": "endDate debe ser posterior a startDate"
}
```

---

## GET `/api/v1/requests`

### Query parameters

| Parámetro | Default | Descripción |
|---|---:|---|
| `page` | `1` | Página actual |
| `pageSize` | `20` | Registros por página; máximo efectivo `100` |
| `status` | — | `Pendiente` o `Aprobada`. Un valor no reconocido devuelve una lista vacía, no un error |
| `type` | — | Coincidencia exacta, sin distinguir mayúsculas |
| `search` | — | Busca en `project`, `type`, `requester` y `machinery` |

```bash
curl 'http://localhost:3001/api/v1/requests?status=Pendiente&type=Excavadora&page=1&pageSize=20'
```

### `200 OK`

```json
{
  "data": [
    {
      "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "project": "Ampliación carretera Los Chorros",
      "type": "Excavadora",
      "requester": "Gerencia Técnica",
      "startDate": "2030-01-10T08:00:00Z",
      "endDate": "2030-01-25T17:00:00Z",
      "status": "Pendiente"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

Ordenado por `createdAt` descendente.

---

## POST `/api/v1/requests/search`

Búsqueda avanzada. A diferencia de `GET /requests`, permite filtrar por **varios estados a la vez** y por solicitante parcial.

### Request

```bash
curl -X POST http://localhost:3001/api/v1/requests/search \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "Los Chorros",
    "statuses": ["Pendiente", "Aprobada"],
    "type": "Excavadora",
    "requester": "Gerencia",
    "page": 1,
    "pageSize": 20
  }'
```

### Campos

| Campo | Tipo | Requerido | Descripción |
|---|---|---:|---|
| `query` | string | No | Busca en `project`, `type`, `requester` y `machinery`. Máximo 500 caracteres |
| `statuses` | array de string | No | Los valores no reconocidos se ignoran. Si ninguno es válido, el resultado es vacío |
| `type` | string | No | Coincidencia exacta, sin distinguir mayúsculas |
| `requester` | string | No | Coincidencia parcial |
| `page` | int | No | Default `1` |
| `pageSize` | int | No | Default `20`; máximo `100` |

Un body vacío `{}` es válido y devuelve la primera página sin filtros.

### `200 OK`

La forma de la respuesta es distinta a la del listado: los resultados van dentro de `data.requests`.

```json
{
  "data": {
    "count": 1,
    "requests": [
      {
        "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
        "project": "Ampliación carretera Los Chorros",
        "type": "Excavadora",
        "requester": "Gerencia Técnica",
        "startDate": "2030-01-10T08:00:00Z",
        "endDate": "2030-01-25T17:00:00Z",
        "status": "Pendiente",
        "createdAt": "2026-09-13T20:00:00Z",
        "updatedAt": "2026-09-13T20:00:00Z"
      }
    ]
  },
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

### `400 Bad Request`

```json
{
  "error": "invalid_query",
  "message": "La consulta no puede superar 500 caracteres"
}
```

---

## GET `/api/v1/requests/:requestID`

```bash
curl http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057
```

### `404 Not Found`

```json
{
  "error": "request_not_found",
  "message": "La solicitud no existe"
}
```

---

## PATCH `/api/v1/requests/:requestID`

Actualiza parcialmente una solicitud. Solo se permite cuando el estado actual es `Pendiente`.

Campos admitidos:

```text
project
type
requester
startDate
endDate
```

```bash
curl -X PATCH \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057 \
  -H 'Content-Type: application/json' \
  -d '{"project": "Ampliación carretera Los Chorros - fase 2"}'
```

### `200 OK`

```json
{
  "message": "Solicitud actualizada correctamente",
  "data": {
    "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "project": "Ampliación carretera Los Chorros - fase 2",
    "status": "Pendiente"
  }
}
```

### `400 Bad Request` — actualización vacía

```json
{
  "error": "empty_update",
  "message": "Debe enviar al menos un campo"
}
```

### `422 Unprocessable Entity`

Si el rango resultante deja `endDate <= startDate`, considerando los valores ya almacenados:

```json
{
  "error": "invalid_period",
  "message": "endDate debe ser posterior a startDate"
}
```

### `409 Conflict`

```json
{
  "error": "request_cannot_be_updated",
  "message": "Solo se pueden modificar solicitudes pendientes"
}
```

---

## DELETE `/api/v1/requests/:requestID`

Elimina una solicitud mediante borrado lógico. Solo se permite cuando está `Pendiente` **y** no tiene maquinaria asignada.

```bash
curl -i -X DELETE \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057
```

### `204 No Content`

Sin body.

### `409 Conflict`

```json
{
  "error": "request_cannot_be_deleted",
  "message": "Solo se pueden eliminar solicitudes pendientes sin maquinaria asignada"
}
```

---

## PATCH `/api/v1/requests/:requestID/status`

Cambia el estado de una solicitud y registra la transición. Usa una transacción con `SELECT ... FOR UPDATE`.

### Request

```bash
curl -X PATCH \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057/status \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "Pendiente",
    "reason": "Se libera la maquinaria por cambio de cronograma"
  }'
```

### Campos

| Campo | Requerido | Validación |
|---|---:|---|
| `status` | Sí | `Pendiente` o `Aprobada`, con las variantes normalizadas |
| `reason` | Sí | Entre 3 y 250 caracteres |

### `200 OK`

```json
{
  "message": "Estado actualizado correctamente",
  "data": {
    "request": {
      "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "status": "Pendiente"
    },
    "transition": {
      "id": "5f37962e-d2ad-44ad-8388-f76978fda8ad",
      "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "fromStatus": "Aprobada",
      "toStatus": "Pendiente",
      "reason": "Se libera la maquinaria por cambio de cronograma",
      "changedAt": "2026-09-13T22:00:00Z"
    }
  }
}
```

Al pasar de `Aprobada` a `Pendiente`, la maquinaria asociada vuelve automáticamente a `Disponible` y `machinery` queda en `null`.

### `400 Bad Request`

```json
{
  "error": "invalid_status",
  "message": "El estado de solicitud no es válido"
}
```

### `409 Conflict`

El código HTTP para una transición rechazada es **409**, no 422. El mensaje varía según la causa:

```json
{
  "error": "invalid_status_transition",
  "message": "invalid request status transition"
}
```

```json
{
  "error": "invalid_status_transition",
  "message": "machinery is required before approval"
}
```

> Para aprobar una solicitud normalmente no se usa este endpoint, sino `PATCH /requests/:requestID/assignment`, que asigna la maquinaria y aprueba en una sola operación.

---

## GET `/api/v1/requests/:requestID/status-history`

```bash
curl http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057/status-history
```

### `200 OK`

```json
{
  "data": [
    {
      "id": "5f37962e-d2ad-44ad-8388-f76978fda8ad",
      "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "fromStatus": "Pendiente",
      "toStatus": "Aprobada",
      "reason": "Maquinaria confirmada para la solicitud",
      "changedAt": "2026-09-13T21:00:00Z"
    }
  ]
}
```

Ordenado por `changedAt` descendente. Si la solicitud no existe responde `404 request_not_found`.

---

# Asignación de maquinaria

## PATCH `/api/v1/requests/:requestID/assignment`

Asigna una maquinaria a una solicitud y la aprueba, en una sola transacción.

### Request

```bash
curl -X PATCH \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057/assignment \
  -H 'Content-Type: application/json' \
  -d '{
    "equipmentId": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
    "reason": "Maquinaria confirmada para la solicitud"
  }'
```

### Campos

| Campo | Requerido | Validación |
|---|---:|---|
| `equipmentId` | Sí | UUID de la maquinaria |
| `reason` | No | Máximo 250 caracteres. Se guarda en el historial |

### Condiciones verificadas

Las tres se evalúan dentro de la transacción, con ambas filas bloqueadas:

1. la solicitud debe estar `Pendiente`;
2. la maquinaria debe estar `Disponible`;
3. `request.type` debe ser igual a `equipment.equipmentClass`, sin distinguir mayúsculas.

### Efectos

```text
Solicitud    machinery = assetNumber de la maquinaria
             status    = Aprobada
Maquinaria   status    = Ocupada
Historial    Pendiente -> Aprobada
```

### `200 OK`

```json
{
  "message": "Maquinaria asignada y solicitud aprobada correctamente",
  "data": {
    "request": {
      "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "project": "Ampliación carretera Los Chorros",
      "type": "Excavadora",
      "status": "Aprobada",
      "machinery": "SYN-DEMO-01"
    },
    "equipment": {
      "id": "7a1b2c3d-4e5f-4a6b-8c9d-0e1f2a3b4c5d",
      "assetNumber": "SYN-DEMO-01",
      "equipmentClass": "Excavadora",
      "status": "Ocupada"
    },
    "transition": {
      "id": "5f37962e-d2ad-44ad-8388-f76978fda8ad",
      "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "fromStatus": "Pendiente",
      "toStatus": "Aprobada",
      "reason": "Maquinaria confirmada para la solicitud",
      "changedAt": "2026-09-13T21:00:00Z"
    }
  }
}
```

### `400 Bad Request`

```json
{
  "error": "invalid_equipment_id",
  "message": "equipmentId debe ser un UUID válido"
}
```

### `404 Not Found`

```json
{
  "error": "request_or_equipment_not_found",
  "message": "La solicitud o la maquinaria no existe"
}
```

### `409 Conflict`

Un solo código de error cubre las tres condiciones; el `message` indica cuál falló:

```json
{
  "error": "assignment_conflict",
  "message": "request is not pending"
}
```

```json
{
  "error": "assignment_conflict",
  "message": "equipment is not available"
}
```

```json
{
  "error": "assignment_conflict",
  "message": "equipment class does not match request type"
}
```

---

## DELETE `/api/v1/requests/:requestID/assignment`

Libera la maquinaria asignada y devuelve la solicitud a `Pendiente`.

### Request

```bash
curl -X DELETE \
  'http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057/assignment?reason=cancelacion-de-proyecto'
```

El parámetro de query `reason` es opcional; si se omite se registra `Liberación solicitada`.

### Efectos

```text
Maquinaria   status    = Disponible
Solicitud    machinery = null
             status    = Pendiente
Historial    Aprobada -> Pendiente
```

### `200 OK`

A diferencia de la asignación, aquí `data` es la solicitud directamente:

```json
{
  "message": "Maquinaria liberada correctamente",
  "data": {
    "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "project": "Ampliación carretera Los Chorros",
    "status": "Pendiente"
  }
}
```

### `409 Conflict`

```json
{
  "error": "release_conflict",
  "message": "request has no assigned machinery"
}
```

---

# Casos de uso

## Caso 1 — Cargar el inventario de maquinaria

```text
POST /api/v1/equipments
        │
        ▼
Se valida company, assetNumber, name y equipmentClass
        │
        ▼
Estado inicial Disponible
        │
        ▼
201 Created
```

---

## Caso 2 — Registrar una necesidad de maquinaria

```text
Cliente
  |
  | POST /requests
  v
logistic-service
  |
  | INSERT
  v
PostgreSQL
  |
  v
Solicitud = Pendiente
```

La solicitud queda disponible para edición y asignación.

---

## Caso 3 — Corregir datos antes de asignar

Mientras la solicitud siga `Pendiente` pueden modificarse el proyecto, el tipo, el solicitante y las fechas:

```http
PATCH /api/v1/requests/{requestID}
```

Una vez `Aprobada`, la modificación se rechaza con `409 request_cannot_be_updated`. Para volver a editarla hay que liberar primero la maquinaria.

---

## Caso 4 — Asignar maquinaria y aprobar

```text
GET /api/v1/equipments?equipmentClass=Excavadora&status=Disponible
        │
        ▼
Elegir el assetNumber adecuado
        │
        ▼
PATCH /api/v1/requests/{requestID}/assignment
        │
        ▼
Solicitud Aprobada + maquinaria Ocupada + historial
```

Ejemplo:

```bash
curl -X PATCH \
  http://localhost:3001/api/v1/requests/$REQUEST_ID/assignment \
  -H 'Content-Type: application/json' \
  -d '{"equipmentId":"'"$EQUIPMENT_ID"'","reason":"Se confirmó la excavadora"}'
```

---

## Caso 5 — Cerrar el trabajo y devolver la maquinaria al inventario

Cuando el trabajo termina, la maquinaria se libera:

```http
DELETE /api/v1/requests/{requestID}/assignment?reason=trabajo-finalizado
```

Como alternativa, si la maquinaria quedó fuera de servicio, puede marcarse directamente:

```http
PATCH /api/v1/equipments/{id}/status
```

```json
{ "status": "Mant. correctivo" }
```

---

## Caso 6 — Cancelar una asignación

El modelo actual no tiene un estado `Cancelada`. Una cancelación se representa liberando la maquinaria, lo que devuelve la solicitud a `Pendiente` y deja el motivo en el historial:

```bash
curl -X DELETE \
  'http://localhost:3001/api/v1/requests/'"$REQUEST_ID"'/assignment?reason=proyecto-suspendido'
```

Si además la solicitud ya no aplica, puede eliminarse:

```bash
curl -i -X DELETE http://localhost:3001/api/v1/requests/$REQUEST_ID
```

---

## Caso 7 — Auditar el ciclo de una solicitud

```http
GET /api/v1/requests/{requestID}/status-history
```

Cada asignación y cada liberación deja su propia fila, con el motivo enviado en `reason`.

---

# Carga de datos sintéticos

El script `seed-entropy-synthetic-data.sh` (mantenido fuera de este repositorio, junto a los demás servicios de Entropy) carga un conjunto correlacionado de datos de prueba usando únicamente HTTP.

Contra este servicio ejecuta:

| Llamada | Cantidad aproximada |
|---|---|
| `POST /api/v1/equipments` | 25 maquinarias, incluida una sin pareja en Fleet |
| `POST /api/v1/requests` | 10 solicitudes de proyectos |
| `PATCH /api/v1/requests/{id}/assignment` | 4 asignaciones |
| `PATCH /api/v1/equipments/{id}/status` | Cambios de estado del caso completado |
| `DELETE /api/v1/requests/{id}/assignment` | 1 liberación, para el caso cancelado |

Además llama a `GET /health` antes de empezar y coordina cada maquinaria con su vehículo equivalente en `fleet-service` y con la proyección unificada del `mcp-server`.

Configuración por variables de entorno, sin valores dentro del repositorio:

```bash
FLEET_URL=http://localhost:3000 \
LOGISTIC_URL=http://localhost:3001 \
MCP_URL=http://localhost:3002 \
SEED_RUN=SYNTH-DEMO \
./seed-entropy-synthetic-data.sh
```

`SEED_RUN` identifica la corrida y se incrusta en `assetNumber` y en el nombre del proyecto, de modo que todo el conjunto puede aislarse después con `search`.

Al finalizar, el conjunto esperado en este servicio es de 10 solicitudes: 7 `Pendiente` y 3 `Aprobada`.

> El script envía una cabecera `Authorization: Bearer`. Este servicio **no** valida credenciales: la cabecera se ignora. No incrustes tokens reales en scripts versionados.

---

# Manejo de errores

Formato general:

```json
{
  "error": "error_code",
  "message": "Descripción legible del error"
}
```

En los errores de binding también se incluye el detalle del validador:

```json
{
  "detail": "detalle técnico de validación"
}
```

Errores relevantes:

| HTTP | Código | Descripción |
|---:|---|---|
| `400` | `invalid_request` | Payload inválido |
| `400` | `invalid_id` | UUID inválido en la ruta |
| `400` | `invalid_equipment_id` | `equipmentId` no es un UUID |
| `400` | `invalid_status` | Estado fuera del catálogo |
| `400` | `invalid_query` | `query` de búsqueda mayor a 500 caracteres |
| `400` | `empty_update` | PATCH sin campos |
| `404` | `request_not_found` | Solicitud inexistente |
| `404` | `equipment_not_found` | Maquinaria inexistente |
| `404` | `request_or_equipment_not_found` | Falta alguna de las dos al asignar |
| `409` | `equipment_already_exists` | `assetNumber` duplicado |
| `409` | `request_cannot_be_updated` | La solicitud no está `Pendiente` |
| `409` | `request_cannot_be_deleted` | La solicitud no está `Pendiente` o tiene maquinaria asignada |
| `409` | `invalid_status_transition` | Transición no permitida o falta maquinaria para aprobar |
| `409` | `assignment_conflict` | Estado o clase incompatibles al asignar |
| `409` | `release_conflict` | La solicitud no tiene maquinaria asignada |
| `422` | `invalid_period` | `endDate` no es posterior a `startDate` |
| `500` | `database_error` | Error interno de persistencia |

---

# Observabilidad

## Logging local

Por defecto `LOGGING_TYPE` usa el logger local basado en Logrus.

Los logs se escriben simultáneamente en:

```text
stdout
app.log
```

```env
LOGGING_TYPE=local
```

## Logging en GCP

```env
LOGGING_TYPE=GCP
GCP_PROJECT_ID=my-project
```

Utiliza Google Cloud Logging y Google Error Reporting.

## Trazas

### STDOUT

```env
TRACE_TYPE=STDOUT
SERVICE_NAME=logistic-service
```

### OTLP

```env
TRACE_TYPE=OTLP
SERVICE_NAME=logistic-service
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
```

### Google Cloud

```env
TRACE_TYPE=GCP
SERVICE_NAME=logistic-service
GCP_PROJECT_ID=my-project
```

El exportador utiliza Application Default Credentials y, si no se configura otro endpoint, envía las trazas a:

```text
telemetry.googleapis.com:443
```

### Deshabilitar trazas

```env
TRACE_TYPE=NONE
```

O:

```env
TRACE_TYPE=DISABLED
```

> El proveedor de OpenTelemetry se inicializa correctamente, pero ningún handler ni capa de base de datos abre spans, y el router no instala middleware de instrumentación. En la práctica no se emite ninguna traza. Ver [Notas conocidas](#notas-conocidas-de-la-implementación).

---

# CORS

La configuración actual permite como origen:

```text
http://localhost:8080
```

Métodos permitidos:

```text
GET
POST
PUT
DELETE
OPTIONS
PATCH
```

Headers permitidos:

```text
Origin
Content-Type
Accept
Authorization
Access-Control-Allow-Origin
```

---

# Pruebas

El repositorio contiene:

```text
logistic-service-test.sh
```

> **Este script está desactualizado.** Fue escrito para el modelo anterior: envía `equipmentType`, `projectName` y `location`, y espera los estados `PENDING`, `ASSIGNED`, `COMPLETED` y `CANCELLED`. Falla en el primer paso, porque `POST /requests` ahora exige `project`, `type` y `requester` y responde `400`. Necesita reescribirse contra el modelo de dos estados y el flujo de asignación.

## Dependencias

```bash
curl
jq
```

## Flujo de humo manual

Mientras el script se actualiza, el ciclo completo puede verificarse a mano:

```bash
BASE=http://localhost:3001/api/v1

# 1. Crear maquinaria
curl -X POST $BASE/equipments \
  -H 'Content-Type: application/json' \
  -d '{"company":"ECON","assetNumber":"SMOKE-01","name":"Excavadora de prueba","equipmentClass":"Excavadora"}'

# 2. Crear solicitud del mismo tipo
curl -X POST $BASE/requests \
  -H 'Content-Type: application/json' \
  -d '{"project":"Proyecto de humo","type":"Excavadora","requester":"QA","startDate":"2030-01-10T08:00:00Z","endDate":"2030-01-25T17:00:00Z"}'

# 3. Asignar
curl -X PATCH $BASE/requests/<REQUEST_ID>/assignment \
  -H 'Content-Type: application/json' \
  -d '{"equipmentId":"<EQUIPMENT_ID>","reason":"Prueba de humo"}'

# 4. Verificar que la maquinaria quedó Ocupada
curl $BASE/equipments/<EQUIPMENT_ID>

# 5. Liberar
curl -X DELETE "$BASE/requests/<REQUEST_ID>/assignment?reason=fin-de-prueba"

# 6. Revisar el historial
curl $BASE/requests/<REQUEST_ID>/status-history
```

---

# Estructura del proyecto

```text
.
├── cmd/
│   └── services/
│       └── main.go
│
├── internal/
│   ├── equipments/          # maquinaria Prisma
│   │   ├── db/
│   │   │   └── postgres/
│   │   ├── domain/
│   │   ├── dto/
│   │   └── handlers/
│   │
│   ├── requests/            # solicitudes y asignaciones
│   │   ├── db/
│   │   │   └── postgres/
│   │   ├── domain/
│   │   ├── dto/
│   │   └── handlers/
│   │
│   ├── database/
│   │   └── postgres/
│   │
│   ├── interfaces/
│   ├── logging/
│   │   ├── gcp/
│   │   └── local/
│   │
│   ├── router/
│   ├── trace/
│   │   ├── exporter/
│   │   │   ├── gcp/
│   │   │   ├── otlp/
│   │   │   └── stdout/
│   │   └── otel/
│   │
│   └── validations/
│
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── logistic-service-test.sh
├── recommendation-algorithm.md
└── recommendation-algorithm.png
```

Los paquetes `clients/fleet`, `recomendations/` y `assigments/` fueron eliminados. La lógica de asignación vive ahora en `internal/requests/db/postgres/requests.go`.

### Tablas creadas

| Módulo | Tabla |
|---|---|
| Maquinaria | `prisma_machinery` |
| Solicitudes | `prisma_machinery_requests` |
| Historial de estados | `prisma_request_status_history` |

Las tres usan borrado lógico (`deleted_at`).

---

# Notas conocidas de la implementación

Puntos detectados durante la revisión del código que conviene tener presentes o corregir.

## La ubicación de la solicitud nunca se guarda

El modelo `Request` tiene `location`, `latitude` y `longitude`, pero el DTO `CreateRequest` **no incluye esos campos** y `UpdateRequest` tampoco. Si el cliente los envía, Gin los descarta en silencio y la solicitud se persiste con `location: ""`, `latitude: 0` y `longitude: 0`. El script de datos sintéticos los envía en cada `POST /api/v1/requests` y ninguno se almacena. Es el mismo hueco que existe en `fleet-service` con las coordenadas del vehículo.

## Documentación del motor de recomendaciones obsoleta

`recommendation-algorithm.md` y `recommendation-algorithm.png` describen el algoritmo de scoring, la fórmula de Haversine y la integración HTTP con Fleet. Ese código ya no existe en el repositorio. Ambos archivos deberían eliminarse o marcarse como histórico.

## `FLEET_SERVICE_URL` sigue declarada en despliegue

La variable ya no se lee en el código, pero continúa presente en `cdeploy/manifests/dev.yaml` y `cdeploy/manifests/prod.yaml`. No rompe nada, pero induce a error sobre las dependencias reales del servicio.

## Las trazas no se emiten

`trace.NewTrace` construye el `TracerProvider` y lo registra globalmente, pero ningún handler ni capa de persistencia llama a `StartSpan`, y el router no instala middleware de OpenTelemetry para Gin. El resultado es que, con cualquier `TRACE_TYPE`, no se produce ninguna traza. `fleet-service` sí abre spans explícitos en sus módulos `equipments` y `fleets`.

## `PATCH /equipments/:id/status` no considera las asignaciones

El endpoint cambia el estado de una maquinaria sin verificar si está asignada a una solicitud aprobada. Es posible dejar una maquinaria en `Obsoletas` mientras `prisma_machinery_requests.machinery` sigue apuntando a ella, o devolverla a `Disponible` y permitir que se asigne dos veces.

## La relación entre solicitud y maquinaria es por texto

`Request.Machinery` guarda el `assetNumber` como cadena, sin llave foránea. Si una maquinaria cambia de `assetNumber` mediante `PATCH /equipments/:id`, las solicitudes que la referencian quedan apuntando a un valor que ya no existe, y las liberaciones posteriores no encontrarán la fila a devolver a `Disponible`.

## `pgvector` declarado sin uso

`go.mod` incluye `github.com/pgvector/pgvector-go` como dependencia directa, pero ningún archivo del proyecto lo importa. Probablemente quedó de una exploración de búsqueda semántica; puede retirarse con `go mod tidy`.

## `logistic-service-test.sh` desactualizado

El script prueba el contrato anterior por completo. Ver [Pruebas](#pruebas).

## El servicio no valida autenticación

No hay middleware de autenticación ni de API key. Las cabeceras `Authorization` que envían los clientes de integración se ignoran. El control de acceso depende por completo de la capa de red o del gateway que exponga el servicio.
