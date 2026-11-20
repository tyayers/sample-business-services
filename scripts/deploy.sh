#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

GOOS="linux" GOARCH=amd64 go build

gcloud beta run deploy sample-business-services \
  --source . \
  --no-build \
  --base-image=osonly24 \
  --command=./sample-business-services \
  --no-iap \
  --no-allow-unauthenticated \
  --project "$GOOGLE_CLOUD_PROJECT" \
  --region "$GOOGLE_CLOUD_LOCATION"

CLOUD_RUN_URL="${CLOUD_RUN_URL:-$(gcloud run services describe sample-business-services \
  --project "$GOOGLE_CLOUD_PROJECT" \
  --region "$GOOGLE_CLOUD_LOCATION" \
  --format='value(status.url)')}"

SERVICE_ACCOUNT="${SERVICE_ACCOUNT:-ai-service}"
if [[ "$SERVICE_ACCOUNT" != *"@"* ]]; then
  SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT}@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com"
else
  SERVICE_ACCOUNT_EMAIL="$SERVICE_ACCOUNT"
fi
SERVICE_ACCOUNT_NAME="${SERVICE_ACCOUNT_EMAIL%%@*}"

if ! gcloud iam service-accounts describe "$SERVICE_ACCOUNT_EMAIL" --project "$GOOGLE_CLOUD_PROJECT" >/dev/null 2>&1; then
  echo "Creating service account ${SERVICE_ACCOUNT_EMAIL}..."
  gcloud iam service-accounts create "$SERVICE_ACCOUNT_NAME" \
    --display-name="Apigee Business Services Proxy SA" \
    --project "$GOOGLE_CLOUD_PROJECT"
fi

echo "Ensuring Cloud Run Invoker role for ${SERVICE_ACCOUNT_EMAIL}..."
gcloud run services add-iam-policy-binding sample-business-services \
  --project "$GOOGLE_CLOUD_PROJECT" \
  --region "$GOOGLE_CLOUD_LOCATION" \
  --member="serviceAccount:${SERVICE_ACCOUNT_EMAIL}" \
  --role="roles/run.invoker" >/dev/null

PROJECT_NUMBER="$(gcloud projects describe "$GOOGLE_CLOUD_PROJECT" --format="value(projectNumber)")"
echo "Ensuring Service Account Token Creator role for Apigee service agent..."
gcloud iam service-accounts add-iam-policy-binding "$SERVICE_ACCOUNT_EMAIL" \
  --project "$GOOGLE_CLOUD_PROJECT" \
  --member="serviceAccount:service-${PROJECT_NUMBER}@gcp-sa-apigee.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountTokenCreator" >/dev/null

APIGEE_ORG="${APIGEE_ORG:-$GOOGLE_CLOUD_PROJECT}"
APIGEE_ENV="${APIGEE_ENV:-eval}"

echo "Deploying Apigee proxies targeting ${CLOUD_RUN_URL} (org=${APIGEE_ORG}, env=${APIGEE_ENV}, sa=${SERVICE_ACCOUNT_EMAIL})..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

for proxy in apigee/*-v1.yaml; do
  rendered="${TMP_DIR}/$(basename "$proxy")"
  sed "s|{CLOUD_RUN_URL}|${CLOUD_RUN_URL}|g" "$proxy" > "$rendered"
  aft "$rendered" \
    --org "$APIGEE_ORG" \
    --env "$APIGEE_ENV" \
    --sa "$SERVICE_ACCOUNT_EMAIL" \
    --no-anim
done

echo "Deploying Apigee API products..."
for product in apigee/*-product.yaml; do
  aft "$product" \
    --org "$APIGEE_ORG" \
    --env "$APIGEE_ENV" \
    --no-anim
done
