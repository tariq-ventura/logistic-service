# Logistic Service

Microservicio encargado de administrar **solicitudes logísticas de maquinaria pesada**, controlar su ciclo de vida y generar **recomendaciones de maquinaria disponible** consultando a `fleet-service`.

El servicio forma parte de una arquitectura de microservicios en la que:

- `logistic-service` administra las solicitudes de maquinaria para proyectos.
- `fleet-service` administra la flota y el estado operativo de cada equipo.
- PostgreSQL almacena las solicitudes y su historial de cambios de estado.
- El motor de recomendaciones cruza la ubicación y tipo requerido con los equipos disponibles de `fleet-service`.

## Tabla de contenido

- [Responsabilidades](#responsabilidades)
- [Arquitectura](#arquitectura)
- [Tecnologías](#tecnologías)
- [Requisitos](#requisitos)
- [Variables de entorno](#variables-de-entorno)
- [Levantar el proyecto localmente](#levantar-el-proyecto-localmente)
- [Modelo de dominio](#modelo-de-dominio)
- [Estados de una solicitud](#estados-de-una-solicitud)
- [Endpoints](#endpoints)
- [Casos de uso](#casos-de-uso)
- [Motor de recomendaciones](#motor-de-recomendaciones)
- [Integración con Fleet Service](#integración-con-fleet-service)
- [Observabilidad](#observabilidad)
- [Pruebas](#pruebas)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Notas conocidas de la implementación](#notas-conocidas-de-la-implementación)

---

## Responsabilidades

Actualmente `logistic-service` permite:

- crear solicitudes logísticas;
- consultar solicitudes;
- actualizar solicitudes mientras estén en estado `PENDING`;
- controlar el estado de una solicitud mediante transiciones válidas;
- almacenar el historial de cambios de estado;
- consultar maquinaria disponible en `fleet-service`;
- calcular un ranking de recomendaciones por solicitud;
- exponer health check HTTP;
- enviar trazas mediante OpenTelemetry;
- utilizar logging local o Google Cloud Logging/Error Reporting.

---

## Arquitectura

Flujo simplificado:

```text
                  +----------------------+
                  |      Cliente / UI    |
                  +----------+-----------+
                             |
                             | HTTP REST
                             v
                  +----------------------+
                  |   logistic-service   |
                  |       Go + Gin       |
                  +----+------------+----+
                       |            |
                       |            | HTTP REST
                       |            v
                       |     +------------------+
                       |     |  fleet-service   |
                       |     | maquinaria/flota |
                       |     +------------------+
                       |
                       | GORM
                       v
                 +-------------+
                 | PostgreSQL  |
                 +-------------+
```

Para generar recomendaciones el flujo es:

```text
GET /requests/{id}/recommendations
            |
            v
Buscar solicitud en PostgreSQL
            |
            v
¿Estado PENDING?
     |             |
    no            sí
     |             |
    409            v
              Consultar Fleet
                    |
                    v
          Equipos AVAILABLE del tipo
                    |
                    v
            Filtrar candidatos
                    |
                    v
      Calcular distancia + mantenimiento
              + combustible
                    |
                    v
              Calcular score
                    |
                    v
          Ordenar y responder 200
```

El repositorio incluye una explicación más detallada en [`recommendation-algorithm.md`](recommendation-algorithm.md) y su diagrama en [`recommendation-algorithm.png`](recommendation-algorithm.png).

---

## Tecnologías

### Lenguaje y framework

| Tecnología | Uso |
|---|---|
| Go 1.27 | Lenguaje principal |
| Gin | API HTTP REST |
| GORM | ORM y acceso a PostgreSQL |
| PostgreSQL 17 | Persistencia |
| UUID | Identificadores de entidades |

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
- una instancia accesible de `fleet-service` para el arranque y las recomendaciones;
- `curl` y `jq` si se ejecutará el script de pruebas.

Alternativamente se puede utilizar Docker y Docker Compose.

> `logistic-service` depende de `fleet-service`. La variable `FLEET_SERVICE_URL` es obligatoria durante el arranque aunque solamente se quieran probar los endpoints CRUD de solicitudes.

---

## Variables de entorno

| Variable | Requerida | Ejemplo | Descripción |
|---|---:|---|---|
| `DB_CONTEXT` | Sí | `postgresql` | Backend de base de datos. Actualmente solo PostgreSQL está implementado. |
| `DB_STRING` | Sí | `host=localhost user=mongo password=1234 dbname=backend_golang_gin port=5435 sslmode=disable` | DSN de PostgreSQL. |
| `FLEET_SERVICE_URL` | Sí | `http://localhost:3000` | URL base de `fleet-service`, sin `/api/v1`. |
| `TRACE_TYPE` | Sí | `STDOUT` | Exportador de trazas: `STDOUT`, `OTLP`, `GCP`, `NONE` o `DISABLED`. |
| `SERVICE_NAME` | Sí | `logistic-service` | Nombre utilizado por OpenTelemetry. |
| `PORT` | No | `3001` | Puerto HTTP. Por defecto `3001`. |
| `LOGGING_TYPE` | No | `local` | Usar `GCP` para Cloud Logging; cualquier otro valor usa logger local. |
| `SERVICE_VERSION` | No | `0.1.0` | Versión agregada al recurso OpenTelemetry. |
| `ENVIRONMENT` | No | `local` | Ambiente agregado a las trazas. |
| `GCP_PROJECT_ID` | Solo GCP | `my-project` | Proyecto utilizado por logging/tracing en Google Cloud. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Solo OTLP/GCP personalizado | `collector:4317` | Endpoint OTLP general. |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | Solo OTLP/GCP personalizado | — | Endpoint específico de trazas OTLP. |

Las variables `DB_USER`, `DB_PASS` y `DB_NAME` del `.env.example` son utilizadas principalmente para inicializar PostgreSQL desde Docker Compose.

### Ejemplo recomendado para desarrollo

```env
DB_USER=mongo
DB_PASS=1234
DB_NAME=backend_golang_gin
DB_CONTEXT=postgresql
DB_STRING=host=localhost user=mongo password=1234 dbname=backend_golang_gin port=5435 sslmode=disable

FLEET_SERVICE_URL=http://localhost:3000

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

Esta opción es útil cuando `fleet-service` ya está ejecutándose en `localhost:3000`.

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

export FLEET_SERVICE_URL='http://localhost:3000'

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

El repositorio contiene un `docker-compose.yml`, pero la versión actual requiere algunos ajustes para ejecutar el servicio correctamente de forma standalone:

1. PostgreSQL escucha internamente en `5432`, no en `5434`.
2. `FLEET_SERVICE_URL` es requerida por la aplicación y actualmente no se pasa al contenedor.
3. `fleet-service` debe ser alcanzable desde el contenedor de Logistics.

Un ejemplo mínimo de la parte relevante sería:

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
      FLEET_SERVICE_URL: ${FLEET_SERVICE_URL}
      TRACE_TYPE: ${TRACE_TYPE}
      SERVICE_NAME: logistic-service
      PORT: "3001"
```

Si ambos microservicios se levantan desde un Compose padre y el servicio se llama `fleet-service`, la URL recomendada es:

```env
FLEET_SERVICE_URL=http://fleet-service:3000
```

Luego:

```bash
docker compose up --build
```

---

## Modelo de dominio

Una solicitud logística se almacena en la tabla `logistics_requests`.

Ejemplo conceptual:

```json
{
  "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
  "equipmentType": "EXCAVATOR",
  "projectName": "Construcción tramo norte",
  "locationName": "San Miguel",
  "latitude": 13.4833,
  "longitude": -88.1833,
  "startDate": "2030-09-10T08:00:00Z",
  "endDate": "2030-09-15T18:00:00Z",
  "status": "PENDING",
  "createdAt": "2030-08-20T15:40:00Z",
  "updatedAt": "2030-08-20T15:40:00Z"
}
```

El historial se almacena en `request_status_history`.

```json
{
  "id": "5f37962e-d2ad-44ad-8388-f76978fda8ad",
  "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
  "fromStatus": "PENDING",
  "toStatus": "ASSIGNED",
  "reason": "Maquinaria confirmada para la solicitud",
  "changedAt": "2030-08-21T10:30:00Z"
}
```

GORM ejecuta `AutoMigrate` al iniciar para:

- `logistics_requests`;
- `request_status_history`.

---

## Estados de una solicitud

Estados válidos:

| Estado | Significado |
|---|---|
| `PENDING` | Solicitud creada y pendiente de asignación. |
| `ASSIGNED` | La solicitud ya tiene una asignación operacional. |
| `COMPLETED` | Trabajo finalizado. Estado terminal. |
| `CANCELLED` | Solicitud cancelada. Estado terminal. |

### Transiciones permitidas

```text
PENDING ───────> ASSIGNED ───────> COMPLETED
   │                │
   │                └────────────> CANCELLED
   │
   └─────────────────────────────> CANCELLED
```

Tabla:

| Estado actual | Estado siguiente | Permitido |
|---|---|---:|
| `PENDING` | `ASSIGNED` | Sí |
| `PENDING` | `CANCELLED` | Sí |
| `PENDING` | `COMPLETED` | No |
| `ASSIGNED` | `COMPLETED` | Sí |
| `ASSIGNED` | `CANCELLED` | Sí |
| `COMPLETED` | Cualquier otro | No |
| `CANCELLED` | Cualquier otro | No |

Cada transición válida crea un registro en `request_status_history`.

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
| `POST` | `/api/v1/requests` | Crear solicitud |
| `GET` | `/api/v1/requests` | Listar solicitudes |
| `GET` | `/api/v1/requests/:requestID` | Consultar solicitud |
| `PATCH` | `/api/v1/requests/:requestID` | Modificar una solicitud `PENDING` |
| `PATCH` | `/api/v1/requests/:requestID/status` | Cambiar estado |
| `GET` | `/api/v1/requests/:requestID/status-history` | Consultar historial |
| `GET` | `/api/v1/requests/:requestID/recommendations` | Generar recomendaciones |

---

## GET `/health`

Verifica que el proceso HTTP esté disponible.

### Request

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

## POST `/api/v1/requests`

Crea una nueva solicitud logística.

El estado inicial siempre es `PENDING` y `equipmentType` se almacena en mayúsculas.

### Request

```bash
curl -X POST http://localhost:3001/api/v1/requests \
  -H 'Content-Type: application/json' \
  -d '{
    "equipmentType": "EXCAVATOR",
    "projectName": "Proyecto San Miguel",
    "location": {
      "name": "San Miguel",
      "latitude": 13.4833,
      "longitude": -88.1833
    },
    "startDate": "2030-09-10T08:00:00Z",
    "endDate": "2030-09-15T18:00:00Z"
  }'
```

### Campos

| Campo | Tipo | Requerido | Validación |
|---|---|---:|---|
| `equipmentType` | string | Sí | No vacío |
| `projectName` | string | Sí | No vacío |
| `location.name` | string | Sí | No vacío |
| `location.latitude` | number | Sí* | `-90` a `90` |
| `location.longitude` | number | Sí* | `-180` a `180` |
| `startDate` | RFC3339 datetime | Sí | Fecha válida |
| `endDate` | RFC3339 datetime | Sí | Debe ser posterior a `startDate` |

`latitude` y `longitude` son valores numéricos y `0` es aceptado por las validaciones actuales.

### `201 Created`

```json
{
  "message": "Peticion registrada correctamente",
  "data": {
    "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "equipmentType": "EXCAVATOR",
    "projectName": "Proyecto San Miguel",
    "locationName": "San Miguel",
    "latitude": 13.4833,
    "longitude": -88.1833,
    "startDate": "2030-09-10T08:00:00Z",
    "endDate": "2030-09-15T18:00:00Z",
    "status": "PENDING"
  }
}
```

### `400 Bad Request`

Body inválido o campos requeridos faltantes:

```json
{
  "error": "invalid_request",
  "message": "Los datos enviados no son válidos",
  "detail": "..."
}
```

### `422 Unprocessable Entity`

Si `endDate <= startDate`:

```json
{
  "error": "endDate must be after startDate"
}
```

---

## GET `/api/v1/requests`

Lista solicitudes con paginación, ordenadas por `created_at DESC`.

### Query parameters

| Parámetro | Default | Descripción |
|---|---:|---|
| `page` | `1` | Página actual |
| `pageSize` | `20` | Registros solicitados por página |

El acceso a base de datos limita internamente `pageSize` a un máximo de `100`.

### Request

```bash
curl 'http://localhost:3001/api/v1/requests?page=1&pageSize=20'
```

### `200 OK`

```json
{
  "data": [
    {
      "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "equipmentType": "EXCAVATOR",
      "projectName": "Proyecto San Miguel",
      "locationName": "San Miguel",
      "latitude": 13.4833,
      "longitude": -88.1833,
      "startDate": "2030-09-10T08:00:00Z",
      "endDate": "2030-09-15T18:00:00Z",
      "status": "PENDING"
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

> Actualmente este endpoint no implementa filtros por estado, tipo de maquinaria, proyecto o fechas.

---

## GET `/api/v1/requests/:requestID`

Consulta una solicitud por UUID.

### Request

```bash
curl http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057
```

### `200 OK`

```json
{
  "data": {
    "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "equipmentType": "EXCAVATOR",
    "projectName": "Proyecto San Miguel",
    "locationName": "San Miguel",
    "latitude": 13.4833,
    "longitude": -88.1833,
    "startDate": "2030-09-10T08:00:00Z",
    "endDate": "2030-09-15T18:00:00Z",
    "status": "PENDING"
  }
}
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
  "error": "request_not_found",
  "message": "La peticion no existe"
}
```

---

## PATCH `/api/v1/requests/:requestID`

Actualiza parcialmente una solicitud.

Solo se pueden modificar solicitudes cuyo estado actual sea `PENDING`.

### Campos admitidos

```json
{
  "equipmentType": "BULLDOZER",
  "projectName": "Proyecto actualizado",
  "location": {
    "name": "Santa Ana",
    "latitude": 13.9942,
    "longitude": -89.5597
  },
  "startDate": "2030-10-01T08:00:00Z",
  "endDate": "2030-10-05T18:00:00Z"
}
```

Todos los campos son opcionales, pero debe enviarse al menos uno.

### Ejemplo

```bash
curl -X PATCH \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057 \
  -H 'Content-Type: application/json' \
  -d '{
    "projectName": "Proyecto actualizado"
  }'
```

### `200 OK`

```json
{
  "message": "Peticion actualizada correctamente",
  "data": {
    "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "projectName": "Proyecto actualizado",
    "status": "PENDING"
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

### `409 Conflict`

Si la solicitud ya no está `PENDING`:

```json
{
  "error": "request_cannot_be_updated",
  "message": "Solo se pueden modificar peticiones con estado PENDING"
}
```

---

## PATCH `/api/v1/requests/:requestID/status`

Cambia el estado de una solicitud y registra la transición en el historial.

La operación utiliza una transacción PostgreSQL y `SELECT ... FOR UPDATE` para reducir condiciones de carrera durante cambios concurrentes.

### Request

```bash
curl -X PATCH \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057/status \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "ASSIGNED",
    "reason": "Maquinaria confirmada para la solicitud"
  }'
```

### Campos

| Campo | Requerido | Validación |
|---|---:|---|
| `status` | Sí | `PENDING`, `ASSIGNED`, `COMPLETED` o `CANCELLED` |
| `reason` | Sí | Entre 3 y 250 caracteres |

### `200 OK`

```json
{
  "message": "Estado actualizado correctamente",
  "data": {
    "request": {
      "id": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "status": "ASSIGNED"
    },
    "transition": {
      "id": "5f37962e-d2ad-44ad-8388-f76978fda8ad",
      "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
      "fromStatus": "PENDING",
      "toStatus": "ASSIGNED",
      "reason": "Maquinaria confirmada para la solicitud"
    }
  }
}
```

### `400 Bad Request`

Estado no reconocido:

```json
{
  "error": "invalid_request",
  "message": "Los datos enviados no son válidos",
  "detail": "..."
}
```

### `422 Unprocessable Entity`

Transición no permitida, por ejemplo `PENDING -> COMPLETED`:

```json
{
  "error": "invalid_status_transition",
  "message": "invalid request status transition: PENDING -> COMPLETED"
}
```

También se rechaza intentar asignar nuevamente el mismo estado.

### `409 Conflict`

Puede ocurrir si otro proceso modificó el estado durante la operación:

```json
{
  "error": "concurrent_status_change",
  "message": "El estado de la petición cambió durante la operación. Consulte nuevamente."
}
```

---

## GET `/api/v1/requests/:requestID/status-history`

Devuelve el historial de transiciones de una solicitud, ordenado por `changed_at DESC`.

### Request

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
      "fromStatus": "PENDING",
      "toStatus": "ASSIGNED",
      "reason": "Maquinaria confirmada para la solicitud",
      "changedAt": "2030-08-21T10:30:00Z"
    }
  ]
}
```

Si la solicitud no existe responde `404 request_not_found`.

---

## GET `/api/v1/requests/:requestID/recommendations`

Genera un ranking de maquinaria para una solicitud `PENDING`.

Este endpoint consulta `fleet-service` en tiempo real.

### Request

```bash
curl \
  http://localhost:3001/api/v1/requests/d00271c7-6553-4461-97e2-ab6f5132b057/recommendations
```

### `200 OK`

```json
{
  "data": {
    "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "count": 1,
    "recommendations": [
      {
        "equipmentId": "a87277e1-0ab8-42cc-a475-5375de107feb",
        "code": "EXC-NEAR",
        "type": "EXCAVATOR",
        "brand": "Caterpillar",
        "model": "320",
        "serialNumber": "CAT320-0001",
        "year": 2025,
        "capacityTons": 23,
        "location": {
          "name": "San Miguel",
          "latitude": 13.4835,
          "longitude": -88.1828
        },
        "distanceKm": 0.06,
        "engineHours": 100,
        "nextMaintenanceHours": 600,
        "maintenanceHoursRemaining": 500,
        "fuelPercent": 95,
        "score": 99.23,
        "reasons": [
          "Maquinaria disponible",
          "Se encuentra a 0.06 km del proyecto",
          "Tiene 500.00 horas antes del próximo mantenimiento",
          "Nivel de combustible de 95.00%"
        ]
      }
    ]
  }
}
```

Una lista vacía sigue siendo una respuesta exitosa `200`:

```json
{
  "data": {
    "requestId": "d00271c7-6553-4461-97e2-ab6f5132b057",
    "count": 0,
    "recommendations": []
  }
}
```

### `409 Conflict`

```json
{
  "error": "request_not_pending",
  "message": "Solo se pueden generar recomendaciones para peticiones con estado PENDING"
}
```

### `502 Bad Gateway`

Si `fleet-service` no está disponible o responde de forma inesperada:

```json
{
  "error": "fleet_service_unavailable",
  "message": "No se pudo consultar la maquinaria disponible"
}
```

---

# Casos de uso

## Caso 1 — Registrar una necesidad de maquinaria

Un proyecto necesita una excavadora en San Miguel durante un rango de fechas.

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
Request = PENDING
```

La solicitud queda disponible para edición y recomendación.

---

## Caso 2 — Corregir datos antes de asignar maquinaria

Mientras la solicitud siga `PENDING` puede modificarse:

- tipo de equipo;
- nombre del proyecto;
- ubicación;
- fecha de inicio;
- fecha de finalización.

```http
PATCH /api/v1/requests/{requestID}
```

Una vez `ASSIGNED`, `COMPLETED` o `CANCELLED`, la modificación se rechaza.

---

## Caso 3 — Buscar la maquinaria más conveniente

```http
GET /api/v1/requests/{requestID}/recommendations
```

Logistics:

1. obtiene la solicitud;
2. comprueba que esté `PENDING`;
3. consulta equipos `AVAILABLE` del tipo requerido en Fleet;
4. descarta equipos con mantenimiento vencido;
5. calcula distancia Haversine;
6. calcula score;
7. devuelve los mejores candidatos primero.

---

## Caso 4 — Marcar una solicitud como asignada

```http
PATCH /api/v1/requests/{requestID}/status
```

```json
{
  "status": "ASSIGNED",
  "reason": "Se confirmó la excavadora EXC-001"
}
```

El cambio genera automáticamente un registro:

```text
PENDING -> ASSIGNED
```

---

## Caso 5 — Completar una solicitud

Solo una solicitud `ASSIGNED` puede pasar a `COMPLETED`.

```json
{
  "status": "COMPLETED",
  "reason": "Trabajo logístico completado"
}
```

Después de esto la solicitud queda en estado terminal.

---

## Caso 6 — Cancelar una solicitud

Puede cancelarse desde:

- `PENDING`;
- `ASSIGNED`.

Ejemplo:

```json
{
  "status": "CANCELLED",
  "reason": "Proyecto suspendido por el cliente"
}
```

---

# Motor de recomendaciones

El algoritmo es **determinista** y no utiliza IA generativa ni LLM para decidir.

## Reglas de elegibilidad

Un equipo solo participa cuando:

```text
status == AVAILABLE
AND type == request.equipmentType
AND nextMaintenanceHours - engineHours > 0
```

Aunque `fleet-service` ya recibe filtros por estado y tipo, Logistics valida nuevamente estos campos antes de calcular el ranking.

## Distancia

Se utiliza la fórmula de Haversine con un radio terrestre de `6371 km`.

```text
Proyecto (lat, lon)
       |
       | Haversine
       v
Maquinaria (lat, lon)
       |
       v
 distanceKm
```

La distancia es geográfica en línea recta; no representa distancia real por carretera.

## Score

El resultado se normaliza a un máximo de `100` puntos:

| Factor | Peso máximo |
|---|---:|
| Distancia | 60 |
| Margen antes del mantenimiento | 25 |
| Combustible | 15 |
| **Total** | **100** |

Fórmula:

```text
distanceFactor    = 1 - clamp(distanceKm / 200, 0, 1)
maintenanceFactor = clamp(maintenanceHoursRemaining / 500, 0, 1)
fuelFactor        = clamp(fuelPercent / 100, 0, 1)

score = distanceFactor * 60
      + maintenanceFactor * 25
      + fuelFactor * 15
```

Donde:

```text
maintenanceHoursRemaining = nextMaintenanceHours - engineHours
```

Los resultados se ordenan:

1. mayor `score` primero;
2. si el score empata, menor `distanceKm` primero.

Para la explicación completa consultar [`recommendation-algorithm.md`](recommendation-algorithm.md).

---

# Integración con Fleet Service

`logistic-service` utiliza un cliente HTTP con timeout de **5 segundos**.

Para obtener candidatos realiza llamadas como:

```http
GET {FLEET_SERVICE_URL}/api/v1/equipments?type=EXCAVATOR&status=AVAILABLE&page=1&pageSize=100
```

El cliente recorre todas las páginas indicadas por `pagination.totalPages`.

Formato esperado de Fleet:

```json
{
  "data": [
    {
      "id": "a87277e1-0ab8-42cc-a475-5375de107feb",
      "code": "EXC-001",
      "type": "EXCAVATOR",
      "status": "AVAILABLE",
      "location": {
        "name": "San Miguel",
        "latitude": 13.4835,
        "longitude": -88.1828
      },
      "engineHours": 100,
      "nextMaintenanceHours": 600,
      "fuelPercent": 95
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 100,
    "total": 1,
    "totalPages": 1
  }
}
```

También existe código de cliente para:

```http
GET /api/v1/equipments/{equipmentID}
PATCH /api/v1/equipments/{equipmentID}/status
```

Estos métodos están destinados al módulo de asignaciones que todavía está en desarrollo.

---

# Observabilidad

## Logging local

Por defecto `LOGGING_TYPE` usa el logger local basado en Logrus.

Los logs se escriben simultáneamente en:

```text
stdout
app.log
```

Ejemplo:

```env
LOGGING_TYPE=local
```

## Logging en GCP

```env
LOGGING_TYPE=GCP
GCP_PROJECT_ID=my-project
```

Utiliza:

- Google Cloud Logging;
- Google Error Reporting.

## Trazas

### STDOUT

```env
TRACE_TYPE=STDOUT
SERVICE_NAME=logistic-service
```

Las trazas se imprimen en consola.

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

El script prueba principalmente:

- creación de solicitudes;
- estado inicial `PENDING`;
- listado;
- consulta por UUID;
- actualización parcial;
- persistencia del cambio;
- rechazo de estados inválidos;
- transiciones válidas e inválidas;
- historial de estados;
- estados terminales;
- cancelación;
- UUID inválido;
- solicitud inexistente.

## Dependencias

```bash
curl
jq
```

## Ejecutar

```bash
chmod +x logistic-service-test.sh
./logistic-service-test.sh
```

Por defecto utiliza:

```text
http://localhost:3001/api/v1/requests
```

Se puede sobrescribir:

```bash
BASE_URL=http://localhost:3001/api/v1/requests \
./logistic-service-test.sh
```

> El script actual no cubre el endpoint de recomendaciones, ya que esa prueba necesita además datos controlados en `fleet-service`.

Para recomendaciones, el flujo end-to-end esperado consiste en crear equipos en Fleet, crear una solicitud en Logistics y consultar:

```bash
curl \
  http://localhost:3001/api/v1/requests/$REQUEST_ID/recommendations
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
│   ├── assigments/
│   │   ├── db/
│   │   ├── domain/
│   │   ├── dto/
│   │   └── handlers/
│   │
│   ├── clients/
│   │   └── fleet/
│   │
│   ├── database/
│   │   └── postgres/
│   │
│   ├── interfaces/
│   ├── logging/
│   │   ├── gcp/
│   │   └── local/
│   │
│   ├── recomendations/
│   ├── requests/
│   │   ├── db/
│   │   │   └── postgres/
│   │   ├── domain/
│   │   ├── dto/
│   │   └── handlers/
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

---