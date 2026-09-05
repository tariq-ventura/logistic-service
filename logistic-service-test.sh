#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:3001/api/v1/requests}"
BODY_FILE="$(mktemp)"

trap 'rm -f "$BODY_FILE"' EXIT

HTTP_STATUS=""

call_api() {
  local method="$1"
  local path="$2"
  local payload="${3:-}"

  if [[ -n "$payload" ]]; then
    HTTP_STATUS="$(curl -sS \
      -o "$BODY_FILE" \
      -w '%{http_code}' \
      -X "$method" \
      -H 'Content-Type: application/json' \
      --data "$payload" \
      "${BASE_URL}${path}")"
  else
    HTTP_STATUS="$(curl -sS \
      -o "$BODY_FILE" \
      -w '%{http_code}' \
      -X "$method" \
      "${BASE_URL}${path}")"
  fi
}

expect_status() {
  local expected="$1"
  local name="$2"

  if [[ "$HTTP_STATUS" != "$expected" ]]; then
    echo "FAIL: $name"
    echo "Esperado: HTTP $expected"
    echo "Recibido: HTTP $HTTP_STATUS"
    jq . "$BODY_FILE" 2>/dev/null || cat "$BODY_FILE"
    exit 1
  fi

  echo "PASS: $name (HTTP $HTTP_STATUS)"
}

assert_json() {
  local expression="$1"
  local name="$2"

  if ! jq -e "$expression" "$BODY_FILE" >/dev/null; then
    echo "FAIL: $name"
    jq . "$BODY_FILE"
    exit 1
  fi

  echo "PASS: $name"
}

for dependency in curl jq; do
  if ! command -v "$dependency" >/dev/null 2>&1; then
    echo "ERROR: se requiere $dependency"
    exit 1
  fi
done

echo "Probando: $BASE_URL"

UNIQUE_SUFFIX="$(date +%s)"

#
# 1. Crear request
#

CREATE_PAYLOAD="$(jq -n \
  --arg projectName "Proyecto Test ${UNIQUE_SUFFIX}" \
  '{
    equipmentType: "EXCAVATOR",
    projectName: $projectName,
    location: {
      name: "San Miguel",
      latitude: 13.4833,
      longitude: -88.1833
    },
    startDate: "2030-09-10T08:00:00Z",
    endDate: "2030-09-15T18:00:00Z"
  }')"

call_api POST "" "$CREATE_PAYLOAD"
expect_status 201 "crear request"

REQUEST_ID="$(jq -r '.data.id // .id // empty' "$BODY_FILE")"

if [[ -z "$REQUEST_ID" ]]; then
  echo "FAIL: la respuesta no contiene el ID"
  jq . "$BODY_FILE"
  exit 1
fi

if [[ "$REQUEST_ID" == "00000000-0000-0000-0000-000000000000" ]]; then
  echo "FAIL: la API generó uuid.Nil"
  exit 1
fi

echo "Request creado: $REQUEST_ID"

assert_json \
  '(.data.status // .status) == "PENDING"' \
  "el estado inicial es PENDING"

#
# 2. Listar requests
#

call_api GET ""
expect_status 200 "listar requests"

if ! jq -e --arg id "$REQUEST_ID" \
  '[.. | objects | .id? | select(. == $id)] | length > 0' \
  "$BODY_FILE" >/dev/null; then

  echo "FAIL: el request creado no aparece en el listado"
  jq . "$BODY_FILE"
  exit 1
fi

echo "PASS: el request aparece en el listado"

#
# 3. Consultar request por ID
#

call_api GET "/$REQUEST_ID"
expect_status 200 "consultar request por ID"

if ! jq -e --arg id "$REQUEST_ID" \
  '(.data.id // .id) == $id' \
  "$BODY_FILE" >/dev/null; then

  echo "FAIL: se devolvió un request diferente"
  jq . "$BODY_FILE"
  exit 1
fi

echo "PASS: se devolvió el request correcto"

#
# 4. Actualizar request
#

UPDATED_PROJECT="Proyecto Test ${UNIQUE_SUFFIX} actualizado"

UPDATE_PAYLOAD="$(jq -n \
  --arg projectName "$UPDATED_PROJECT" \
  '{projectName: $projectName}')"

call_api PATCH "/$REQUEST_ID" "$UPDATE_PAYLOAD"
expect_status 200 "actualizar request"

if ! jq -e --arg projectName "$UPDATED_PROJECT" \
  '(.data.projectName // .projectName) == $projectName' \
  "$BODY_FILE" >/dev/null; then

  echo "FAIL: la respuesta no contiene el nombre actualizado"
  jq . "$BODY_FILE"
  exit 1
fi

echo "PASS: el PATCH devuelve el registro actualizado"

#
# 5. Comprobar persistencia
#

call_api GET "/$REQUEST_ID"
expect_status 200 "consultar request actualizado"

if ! jq -e --arg projectName "$UPDATED_PROJECT" \
  '(.data.projectName // .projectName) == $projectName' \
  "$BODY_FILE" >/dev/null; then

  echo "FAIL: la actualización no fue persistida"
  jq . "$BODY_FILE"
  exit 1
fi

echo "PASS: la actualización fue persistida"

#
# 6. Estado desconocido
#

call_api PATCH "/$REQUEST_ID/status" \
  '{
    "status": "UNKNOWN",
    "reason": "Estado inválido para prueba"
  }'

expect_status 400 "rechazar estado desconocido"

#
# 7. Transición inválida PENDING -> COMPLETED
#

call_api PATCH "/$REQUEST_ID/status" \
  '{
    "status": "COMPLETED",
    "reason": "Transición inválida para prueba"
  }'

expect_status 422 "rechazar PENDING a COMPLETED"

#
# 8. Transición válida PENDING -> ASSIGNED
#

ASSIGNED_REASON="Maquinaria confirmada para la solicitud"

ASSIGNED_PAYLOAD="$(jq -n \
  --arg reason "$ASSIGNED_REASON" \
  '{
    status: "ASSIGNED",
    reason: $reason
  }')"

call_api PATCH "/$REQUEST_ID/status" "$ASSIGNED_PAYLOAD"
expect_status 200 "cambiar PENDING a ASSIGNED"

assert_json \
  '[.. | objects | .status? |
    select(. == "ASSIGNED")] | length > 0' \
  "la respuesta contiene ASSIGNED"

#
# 9. Consultar historial
#

call_api GET "/$REQUEST_ID/status-history"
expect_status 200 "consultar historial"

if ! jq -e --arg reason "$ASSIGNED_REASON" \
  '[.. | objects |
    select(
      .fromStatus? == "PENDING" and
      .toStatus? == "ASSIGNED" and
      .reason? == $reason
    )
  ] | length > 0' "$BODY_FILE" >/dev/null; then

  echo "FAIL: no aparece PENDING -> ASSIGNED"
  jq . "$BODY_FILE"
  exit 1
fi

echo "PASS: historial PENDING -> ASSIGNED"

#
# 10. Repetir el mismo estado
#

call_api PATCH "/$REQUEST_ID/status" \
  '{
    "status": "ASSIGNED",
    "reason": "Intento de estado repetido"
  }'

expect_status 422 "rechazar ASSIGNED a ASSIGNED"

#
# 11. No permitir editar un request ASSIGNED
#

call_api PATCH "/$REQUEST_ID" \
  '{
    "projectName": "Cambio no permitido"
  }'

expect_status 409 "rechazar edición de request ASSIGNED"

#
# 12. Transición válida ASSIGNED -> COMPLETED
#

COMPLETED_REASON="Trabajo logístico completado"

COMPLETED_PAYLOAD="$(jq -n \
  --arg reason "$COMPLETED_REASON" \
  '{
    status: "COMPLETED",
    reason: $reason
  }')"

call_api PATCH "/$REQUEST_ID/status" "$COMPLETED_PAYLOAD"
expect_status 200 "cambiar ASSIGNED a COMPLETED"

assert_json \
  '[.. | objects | .status? |
    select(. == "COMPLETED")] | length > 0' \
  "la respuesta contiene COMPLETED"

#
# 13. Comprobar historial completo
#

call_api GET "/$REQUEST_ID/status-history"
expect_status 200 "consultar historial completo"

assert_json \
  '[.. | objects |
    select(has("fromStatus") and has("toStatus"))
  ] | length >= 2' \
  "el historial contiene dos transiciones"

#
# 14. COMPLETED es terminal
#

call_api PATCH "/$REQUEST_ID/status" \
  '{
    "status": "CANCELLED",
    "reason": "Intento desde estado terminal"
  }'

expect_status 422 "rechazar transición desde COMPLETED"

#
# 15. Probar PENDING -> CANCELLED con otro request
#

CANCEL_CREATE_PAYLOAD="$(jq -n \
  --arg projectName "Proyecto cancelable ${UNIQUE_SUFFIX}" \
  '{
    equipmentType: "EXCAVATOR",
    projectName: $projectName,
    location: {
      name: "Santa Ana",
      latitude: 13.9942,
      longitude: -89.5597
    },
    startDate: "2030-10-01T08:00:00Z",
    endDate: "2030-10-05T18:00:00Z"
  }')"

call_api POST "" "$CANCEL_CREATE_PAYLOAD"
expect_status 201 "crear request cancelable"

CANCEL_REQUEST_ID="$(
  jq -r '.data.id // .id // empty' "$BODY_FILE"
)"

CANCEL_REASON="Proyecto suspendido por el cliente"

CANCEL_PAYLOAD="$(jq -n \
  --arg reason "$CANCEL_REASON" \
  '{
    status: "CANCELLED",
    reason: $reason
  }')"

call_api PATCH \
  "/$CANCEL_REQUEST_ID/status" \
  "$CANCEL_PAYLOAD"

expect_status 200 "cambiar PENDING a CANCELLED"

call_api GET "/$CANCEL_REQUEST_ID/status-history"
expect_status 200 "consultar historial de cancelación"

if ! jq -e --arg reason "$CANCEL_REASON" \
  '[.. | objects |
    select(
      .fromStatus? == "PENDING" and
      .toStatus? == "CANCELLED" and
      .reason? == $reason
    )
  ] | length > 0' "$BODY_FILE" >/dev/null; then

  echo "FAIL: no aparece PENDING -> CANCELLED"
  jq . "$BODY_FILE"
  exit 1
fi

echo "PASS: historial PENDING -> CANCELLED"

#
# 16. UUID inválido
#

call_api GET "/not-a-uuid"
expect_status 400 "rechazar UUID inválido"

#
# 17. UUID válido inexistente
#

NOT_FOUND_ID="11111111-1111-4111-8111-111111111111"

call_api GET "/$NOT_FOUND_ID"
expect_status 404 "request inexistente"

call_api GET "/$NOT_FOUND_ID/status-history"
expect_status 404 "historial de request inexistente"

echo
echo "Todas las pruebas finalizaron correctamente."