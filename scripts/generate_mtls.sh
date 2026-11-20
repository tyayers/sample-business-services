#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

OUT_DIR="${1:-${OUT_DIR:-${REPO_ROOT}/certs}}"
DOMAIN="${DOMAIN:-localhost}"
SERVICE_PORT="${SERVICE_PORT:-8080}"
CADDY_PORT="${CADDY_PORT:-8443}"
CLIENT_CN="${CLIENT_CN:-sample-business-services-client}"
CA_DAYS="${CA_DAYS:-3650}"
CERT_DAYS="${CERT_DAYS:-365}"

if ! command -v openssl >/dev/null 2>&1; then
  echo "Error: 'openssl' is required to generate mTLS certificates." >&2
  exit 1
fi

mkdir -p "${OUT_DIR}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

echo "Generating mTLS CA, server, and client certificates in ${OUT_DIR}..."

# 1. Certificate Authority (CA)
openssl genrsa -out "${OUT_DIR}/ca.key" 4096 2>/dev/null
openssl req -x509 -new -nodes \
  -key "${OUT_DIR}/ca.key" \
  -sha256 \
  -days "${CA_DAYS}" \
  -subj "/CN=sample-business-services-ca" \
  -addext "basicConstraints=critical,CA:TRUE" \
  -addext "keyUsage=critical,keyCertSign,cRLSign" \
  -out "${OUT_DIR}/ca.crt"

# 2. Server certificate (for Caddy when not using Let's Encrypt)
cat > "${TMP_DIR}/server.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:${DOMAIN},DNS:localhost,IP:127.0.0.1
EOF

openssl genrsa -out "${OUT_DIR}/server.key" 2048 2>/dev/null
openssl req -new \
  -key "${OUT_DIR}/server.key" \
  -subj "/CN=${DOMAIN}" \
  -out "${TMP_DIR}/server.csr"
openssl x509 -req \
  -in "${TMP_DIR}/server.csr" \
  -CA "${OUT_DIR}/ca.crt" \
  -CAkey "${OUT_DIR}/ca.key" \
  -CAcreateserial \
  -out "${OUT_DIR}/server.crt" \
  -days "${CERT_DAYS}" \
  -sha256 \
  -extfile "${TMP_DIR}/server.ext" 2>/dev/null

# 3. Client certificate (presented by callers to Caddy)
cat > "${TMP_DIR}/client.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature
extendedKeyUsage=clientAuth
EOF

openssl genrsa -out "${OUT_DIR}/client.key" 2048 2>/dev/null
openssl req -new \
  -key "${OUT_DIR}/client.key" \
  -subj "/CN=${CLIENT_CN}" \
  -out "${TMP_DIR}/client.csr"
openssl x509 -req \
  -in "${TMP_DIR}/client.csr" \
  -CA "${OUT_DIR}/ca.crt" \
  -CAkey "${OUT_DIR}/ca.key" \
  -CAcreateserial \
  -out "${OUT_DIR}/client.crt" \
  -days "${CERT_DAYS}" \
  -sha256 \
  -extfile "${TMP_DIR}/client.ext" 2>/dev/null

rm -f "${OUT_DIR}/ca.srl"
chmod 600 "${OUT_DIR}/ca.key" "${OUT_DIR}/server.key" "${OUT_DIR}/client.key"
chmod 644 "${OUT_DIR}/ca.crt" "${OUT_DIR}/server.crt" "${OUT_DIR}/client.crt"

cat <<EOF

Generated mTLS files in ${OUT_DIR}:
  - ca.crt / ca.key         : Certificate Authority (used by Caddy to verify client certs)
  - server.crt / server.key : Server TLS cert/key (SAN: ${DOMAIN}, localhost, 127.0.0.1)
  - client.crt / client.key : Client cert/key (CN=${CLIENT_CN})

================================================================================
Caddy mTLS Configuration Instructions (Debian Linux)
================================================================================

1. Install the CA cert (and server cert/key if not using public ACME/Let's Encrypt)
   where Caddy can read them:

   sudo mkdir -p /etc/caddy/certs
   sudo cp "${OUT_DIR}/ca.crt" /etc/caddy/certs/ca.crt
   sudo cp "${OUT_DIR}/server.crt" "${OUT_DIR}/server.key" /etc/caddy/certs/
   sudo chown -R caddy:caddy /etc/caddy/certs
   sudo chmod 640 /etc/caddy/certs/server.key

2. Configure /etc/caddy/Caddyfile:

   Option A — Public domain (Caddy manages Let's Encrypt server TLS, mTLS verifies clients):

     ${DOMAIN} {
         tls {
             client_auth {
                 mode require_and_verify
                 trusted_ca_cert_file /etc/caddy/certs/ca.crt
             }
         }
         reverse_proxy 127.0.0.1:${SERVICE_PORT}
     }

   Option B — Local / private endpoint using the generated server cert:

     https://${DOMAIN}:${CADDY_PORT} {
         tls /etc/caddy/certs/server.crt /etc/caddy/certs/server.key {
             client_auth {
                 mode require_and_verify
                 trusted_ca_cert_file /etc/caddy/certs/ca.crt
             }
         }
         reverse_proxy 127.0.0.1:${SERVICE_PORT}
     }

3. Validate and reload Caddy:

   sudo caddy validate --config /etc/caddy/Caddyfile
   sudo systemctl reload caddy

4. Test mTLS access with curl:

   # Should fail without a client certificate:
   curl -i --cacert "${OUT_DIR}/ca.crt" https://${DOMAIN}:${CADDY_PORT}/healthz

   # Should succeed with the client certificate and key:
   curl -s --cacert "${OUT_DIR}/ca.crt" \\
     --cert "${OUT_DIR}/client.crt" \\
     --key "${OUT_DIR}/client.key" \\
     https://${DOMAIN}:${CADDY_PORT}/api/v1/summary | jq

EOF
