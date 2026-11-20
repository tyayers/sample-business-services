#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

: "${GOOGLE_CLOUD_PROJECT:?Error: GOOGLE_CLOUD_PROJECT must be set}"
: "${MTLS_BACKEND_URL:?Error: MTLS_BACKEND_URL must be set (e.g. https://api.example.com:8443)}"

APIGEE_ORG="${APIGEE_ORG:-$GOOGLE_CLOUD_PROJECT}"
APIGEE_ENV="${APIGEE_ENV:-eval}"
MTLS_CERTS_DIR="${MTLS_CERTS_DIR:-${REPO_ROOT}/certs}"
MTLS_KEYSTORE="${MTLS_KEYSTORE:-mtls-keystore}"
MTLS_KEYSTORE_REF="${MTLS_KEYSTORE_REF:-mtls-keystore-ref}"
MTLS_KEYSTORE_ALIAS="${MTLS_KEYSTORE_ALIAS:-mtls-client}"

for cmd in gcloud curl aft; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: '$cmd' is required." >&2
    exit 1
  fi
done

for cert_file in client.crt client.key; do
  if [[ ! -f "${MTLS_CERTS_DIR}/${cert_file}" ]]; then
    echo "Error: Missing ${MTLS_CERTS_DIR}/${cert_file}. Run ./scripts/generate_mtls.sh first." >&2
    exit 1
  fi
done

ACCESS_TOKEN="$(gcloud auth application-default print-access-token)"
APIGEE_ENV_API="https://apigee.googleapis.com/v1/organizations/${APIGEE_ORG}/environments/${APIGEE_ENV}"

apigee_api() {
  local method="$1"
  local url="$2"
  shift 2
  curl -sS -f -X "$method" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    "$@" \
    "$url"
}

apigee_status() {
  local url="$1"
  curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    "$url"
}

ensure_keystore() {
  local name="$1"
  local status
  status="$(apigee_status "${APIGEE_ENV_API}/keystores/${name}")"
  if [[ "$status" != "200" ]]; then
    echo "Creating Apigee keystore '${name}' in ${APIGEE_ORG}/${APIGEE_ENV}..."
    apigee_api POST "${APIGEE_ENV_API}/keystores" \
      -H "Content-Type: application/json" \
      -d "{\"name\":\"${name}\"}" >/dev/null
  else
    echo "Apigee keystore '${name}' already exists."
  fi
}

ensure_reference() {
  local ref_name="$1"
  local store_name="$2"
  local status
  local payload="{\"name\":\"${ref_name}\",\"refers\":\"${store_name}\",\"resourceType\":\"KeyStore\"}"
  status="$(apigee_status "${APIGEE_ENV_API}/references/${ref_name}")"
  if [[ "$status" == "200" ]]; then
    echo "Updating Apigee reference '${ref_name}' -> '${store_name}'..."
    apigee_api PUT "${APIGEE_ENV_API}/references/${ref_name}" \
      -H "Content-Type: application/json" \
      -d "$payload" >/dev/null
  else
    echo "Creating Apigee reference '${ref_name}' -> '${store_name}'..."
    apigee_api POST "${APIGEE_ENV_API}/references" \
      -H "Content-Type: application/json" \
      -d "$payload" >/dev/null
  fi
}

# 1. KeyStore (client certificate + key for Apigee to authenticate to Caddy)
ensure_keystore "${MTLS_KEYSTORE}"
if [[ "$(apigee_status "${APIGEE_ENV_API}/keystores/${MTLS_KEYSTORE}/aliases/${MTLS_KEYSTORE_ALIAS}")" == "200" ]]; then
  echo "Alias '${MTLS_KEYSTORE_ALIAS}' already exists in keystore '${MTLS_KEYSTORE}'; skipping upload."
else
  echo "Uploading client cert/key to keystore '${MTLS_KEYSTORE}' (alias=${MTLS_KEYSTORE_ALIAS})..."
  apigee_api POST "${APIGEE_ENV_API}/keystores/${MTLS_KEYSTORE}/aliases?alias=${MTLS_KEYSTORE_ALIAS}&format=keycertfile" \
    -F "certFile=@${MTLS_CERTS_DIR}/client.crt" \
    -F "keyFile=@${MTLS_CERTS_DIR}/client.key" >/dev/null
fi
ensure_reference "${MTLS_KEYSTORE_REF}" "${MTLS_KEYSTORE}"

# 2. Render and deploy mtls-service-proxy
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

RENDERED="${TMP_DIR}/mtls-service-proxy.yaml"
sed \
  -e "s|{MTLS_BACKEND_URL}|${MTLS_BACKEND_URL}|g" \
  -e "s|{KEYSTORE_REF}|${MTLS_KEYSTORE_REF}|g" \
  -e "s|{KEYSTORE_ALIAS}|${MTLS_KEYSTORE_ALIAS}|g" \
  apigee/mtls-service-proxy.yaml > "${RENDERED}"

echo "Deploying mtls-service-proxy targeting ${MTLS_BACKEND_URL} (org=${APIGEE_ORG}, env=${APIGEE_ENV})..."
aft "${RENDERED}" \
  --org "${APIGEE_ORG}" \
  --env "${APIGEE_ENV}" \
  --no-anim
