#!/usr/bin/env bash
# 仅构建并同步 web/manager 静态资源（不编译 Go、不 migrate、不重启服务）
set -euo pipefail

cd "$(dirname "$0")"

DO_DEPLOY=1
DO_BUILD=1

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

section() {
  echo
  echo "── $1 ──"
}

info() {
  printf "  %-10s %s\n" "$1" "$2"
}

status() {
  printf "  %s  %s\n" "$1" "$2"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-deploy)
      DO_DEPLOY=0
      ;;
    --no-build)
      DO_BUILD=0
      ;;
    -h|--help)
      echo "Usage: $0 [--no-build] [--no-deploy]"
      echo "  默认：yarn build web/manager + rsync dist 到远端 ${DEPLOY_PATH}/web/manager/dist/"
      exit 0
      ;;
    *)
      echo "❌ usage: $0 [--no-build] [--no-deploy]" >&2
      exit 1
      ;;
  esac
  shift
done

MANAGER_DIR="web/manager"
DIST_DIR="${MANAGER_DIR}/dist"

section "Manager 前端"

if [[ "${DO_BUILD}" == "1" ]]; then
  status "🎨" "Building ${MANAGER_DIR} …"
  if [[ ! -d "${MANAGER_DIR}/node_modules" ]]; then
    (cd "${MANAGER_DIR}" && yarn install)
  fi
  (cd "${MANAGER_DIR}" && yarn build)
  status "✅" "Build OK"
else
  status "⏭️" "Skipped build (--no-build)"
fi

if [[ ! -d "${DIST_DIR}" ]]; then
  echo "❌ missing ${DIST_DIR} — run build first" >&2
  exit 1
fi

if [[ "${DO_DEPLOY}" -eq 0 ]]; then
  section "Deploy"
  status "⏭️" "Skipped (--no-deploy)"
  section "Done"
  status "🎉" "Manager build complete"
  echo
  exit 0
fi

if [[ ! -f "${DEPLOY_ENV}" ]]; then
  section "Deploy"
  status "⚠️" "Skipped (no ${DEPLOY_ENV})"
  row_hint="cp deploy.env.example deploy.env"
  info "Hint" "${row_hint}"
  section "Done"
  status "🎉" "Manager build complete"
  echo
  exit 0
fi

DEST="${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/web/manager/dist/"
section "Deploy"
status "📤" "Uploading manager dist …"
info "Target" "${DEST}"
ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "mkdir -p ${DEPLOY_PATH}/web/manager/dist"
rsync -az --delete "${DIST_DIR}/" "${DEST}"
status "✅" "Upload OK"

section "Done"
status "🎉" "Manager deploy complete"
echo
