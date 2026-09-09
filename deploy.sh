#!/usr/bin/env bash
# Upload pre-built binaries + manager dist (no compile). Build first with ./build.sh
set -euo pipefail

cd "$(dirname "$0")"

DEPLOY_ENV="${DEPLOY_ENV:-deploy.env}"
if [[ -f "${DEPLOY_ENV}" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "${DEPLOY_ENV}"
  set +a
fi

DEPLOY_HOST="${DEPLOY_HOST:-kylebing.cn}"
DEPLOY_USER="${DEPLOY_USER:-root}"
DEPLOY_PATH="${DEPLOY_PATH:-/var/www/html/portal-go}"
DEPLOY_RESTART="${DEPLOY_RESTART:-1}"
DEPLOY_SERVICES="${DEPLOY_SERVICES:-portal-go portal-ws}"
DEPLOY_FRONTEND="${DEPLOY_FRONTEND:-1}"

OUT_DIR="bin/linux"
PORTAL_BIN="${OUT_DIR}/portal"
WS_BIN="${OUT_DIR}/ws"
CRON_BIN="${OUT_DIR}/cron"

section() {
  echo
  echo "── $1 ──"
}

status() {
  printf "  %s  %s\n" "$1" "$2"
}

info() {
  printf "  %-10s %s\n" "$1" "$2"
}

row() {
  printf "      %-10s %s\n" "$1" "$2"
}

ok() {
  printf "      %s  %s\n" "$1" "$2"
}

run_nested() {
  local tmp ec
  tmp="$(mktemp)"
  set +e
  "$@" >"${tmp}" 2>&1
  ec=$?
  set -e
  if [[ -s "${tmp}" ]]; then
    sed 's/^/      /' "${tmp}"
  fi
  rm -f "${tmp}"
  return "${ec}"
}

VERSION_FILE="internal/version/VERSION"
APP_VERSION="unknown"
if [[ -f "${VERSION_FILE}" ]]; then
  APP_VERSION="$(tr -d '[:space:]' < "${VERSION_FILE}")"
fi

echo
echo "Portal Go"
section "Deploy"
info "Version" "${APP_VERSION}"

if [[ ! -f "${DEPLOY_ENV}" ]]; then
  status "❌" "Missing ${DEPLOY_ENV}"
  row "Hint" "cp deploy.env.example deploy.env"
  echo
  exit 1
fi

for f in "${PORTAL_BIN}" "${WS_BIN}" "${CRON_BIN}"; do
  if [[ ! -f "${f}" ]]; then
    status "❌" "Missing ${f}"
    row "Hint" "run ./build.sh first"
    echo
    exit 1
  fi
done

if [[ "${DEPLOY_FRONTEND}" == "1" ]]; then
  if [[ ! -d web/manager/dist ]] || [[ ! -f web/manager/dist/index.html ]]; then
    status "❌" "Missing web/manager/dist"
    row "Hint" "run ./build.sh  (or DEPLOY_FRONTEND=0)"
    echo
    exit 1
  fi
fi

DEST="${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/"

status "📤" "Uploading binaries …"
row "Target" "${DEST}"
run_nested ssh "${DEPLOY_USER}@${DEPLOY_HOST}" \
  "mkdir -p ${DEPLOY_PATH}/bin ${DEPLOY_PATH}/web/manager ${DEPLOY_PATH}/config ${DEPLOY_PATH}/migrations ${DEPLOY_PATH}/upload ${DEPLOY_PATH}/temp"
rsync -az "${PORTAL_BIN}" "${WS_BIN}" "${CRON_BIN}" "${DEST}bin/"
ok "✅" "Binaries OK"

status "📤" "Uploading migrations/init.sql …"
rsync -az migrations/init.sql "${DEST}migrations/"
ok "✅" "init.sql OK"

if [[ "${DEPLOY_FRONTEND}" == "1" ]]; then
  status "📤" "Uploading manager frontend …"
  rsync -az --delete web/manager/dist/ "${DEST}web/manager/dist/"
  ok "✅" "Frontend upload OK"
fi

status "🗄️" "Running migrations …"
run_nested ssh "${DEPLOY_USER}@${DEPLOY_HOST}" \
  "chmod +x ${DEPLOY_PATH}/bin/portal ${DEPLOY_PATH}/bin/ws ${DEPLOY_PATH}/bin/cron && cd ${DEPLOY_PATH} && ./bin/portal migrate"
ok "✅" "Migrate OK"

if [[ "${DEPLOY_RESTART}" == "1" ]]; then
  status "🔄" "Restarting services …"
  for svc in ${DEPLOY_SERVICES}; do
    row "Service" "${svc}"
    run_nested ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "systemctl restart ${svc}"
  done
  ok "✅" "Services restarted"
fi

section "Done"
info "Version" "${APP_VERSION}"
status "🎉" "Deploy complete"
echo
