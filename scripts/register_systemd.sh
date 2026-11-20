#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

SERVICE_NAME="${SERVICE_NAME:-sample-business-services}"
INSTALL_DIR="${INSTALL_DIR:-/opt/${SERVICE_NAME}}"
SERVICE_USER="${SERVICE_USER:-${SERVICE_NAME}}"
PORT="${PORT:-8080}"
SEED_DIR="${SEED_DIR:-}"
UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

if ! command -v go >/dev/null 2>&1; then
  echo "Error: 'go' is required to build ${SERVICE_NAME}." >&2
  exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
  echo "Error: 'systemctl' is required to register the systemd service." >&2
  exit 1
fi

if [[ "${EUID}" -eq 0 ]]; then
  SUDO=""
else
  SUDO="sudo"
fi

TMP_BIN="$(mktemp)"
trap 'rm -f "${TMP_BIN}"' EXIT

echo "Building ${SERVICE_NAME} from ${REPO_ROOT}..."
(
  cd "${REPO_ROOT}"
  go build -o "${TMP_BIN}" .
)

if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
  echo "Creating system user '${SERVICE_USER}'..."
  ${SUDO} useradd --system --no-create-home --shell /usr/sbin/nologin "${SERVICE_USER}"
fi

echo "Installing binary to ${INSTALL_DIR}/${SERVICE_NAME}..."
${SUDO} install -d -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0755 "${INSTALL_DIR}"
${SUDO} install -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0755 "${TMP_BIN}" "${INSTALL_DIR}/${SERVICE_NAME}"

LOCAL_ENV_FILE="${REPO_ROOT}/.env"
INSTALLED_ENV_FILE="${INSTALL_DIR}/.env"
if [[ -f "${LOCAL_ENV_FILE}" ]]; then
  echo "Installing environment file from ${LOCAL_ENV_FILE} to ${INSTALLED_ENV_FILE}..."
  TMP_ENV="$(mktemp)"
  # Strip CRLF and bash 'export ' prefixes so systemd EnvironmentFile parses cleanly.
  sed -e 's/\r$//' -e 's/^[[:space:]]*export[[:space:]]\+//' "${LOCAL_ENV_FILE}" > "${TMP_ENV}"
  ${SUDO} install -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0600 "${TMP_ENV}" "${INSTALLED_ENV_FILE}"
  rm -f "${TMP_ENV}"
else
  echo "No local .env found at ${LOCAL_ENV_FILE}; skipping EnvironmentFile copy."
fi

SEED_ENV_LINE=""
if [[ -n "${SEED_DIR}" ]]; then
  SEED_ENV_LINE="Environment=\"SEED_DIR=${SEED_DIR}\""
fi

echo "Writing systemd unit ${UNIT_FILE}..."
${SUDO} tee "${UNIT_FILE}" >/dev/null <<EOF
[Unit]
Description=Sample Business Service (REST + OpenAPI + MCP)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/${SERVICE_NAME}
Environment="PORT=${PORT}"
EnvironmentFile=-${INSTALLED_ENV_FILE}
${SEED_ENV_LINE}
Restart=on-failure
RestartSec=5
KillSignal=SIGTERM
TimeoutStopSec=10
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
EOF

echo "Reloading systemd and enabling ${SERVICE_NAME}..."
${SUDO} systemctl daemon-reload
${SUDO} systemctl enable "${SERVICE_NAME}"
${SUDO} systemctl restart "${SERVICE_NAME}"

echo "Service '${SERVICE_NAME}' registered and started."
${SUDO} systemctl --no-pager status "${SERVICE_NAME}" || true
