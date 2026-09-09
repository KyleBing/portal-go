#!/usr/bin/env bash
# Build only (no upload). Deploy with ./deploy.sh after this.
set -euo pipefail

cd "$(dirname "$0")"

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-amd64}"
DEPLOY_FRONTEND="${DEPLOY_FRONTEND:-1}"

DEPLOY_ENV="${DEPLOY_ENV:-deploy.env}"
if [[ -f "${DEPLOY_ENV}" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "${DEPLOY_ENV}"
  set +a
fi

OUT_DIR="bin/linux"
PORTAL_BIN="${OUT_DIR}/portal"
WS_BIN="${OUT_DIR}/ws"
CRON_BIN="${OUT_DIR}/cron"

file_size_human() {
  ls -lh "$1" | awk '{print $5}'
}

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

MIN_GO_MAJOR=1
MIN_GO_MINOR=22

pick_go() {
  local candidates=()
  if [[ -n "${GO:-}" ]]; then
    candidates+=("${GO}")
  fi
  candidates+=(
    "$(command -v go 2>/dev/null || true)"
    "/opt/homebrew/bin/go"
    "/usr/local/opt/go/bin/go"
  )
  local candidate version major minor
  for candidate in "${candidates[@]}"; do
    [[ -z "${candidate}" || ! -x "${candidate}" ]] && continue
    version="$("${candidate}" env GOVERSION 2>/dev/null || true)"
    [[ -z "${version}" ]] && version="$("${candidate}" version 2>/dev/null | awk '{print $3}')"
    version="${version#go}"
    major="${version%%.*}"
    minor="${version#*.}"
    minor="${minor%%.*}"
    if [[ "${major}" -gt "${MIN_GO_MAJOR}" ]] \
      || { [[ "${major}" -eq "${MIN_GO_MAJOR}" ]] && [[ "${minor}" -ge "${MIN_GO_MINOR}" ]]; }; then
      GO="${candidate}"
      return 0
    fi
  done
  return 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    arm64|aarch64)
      GOARCH=arm64
      ;;
    amd64)
      GOARCH=amd64
      ;;
    --no-frontend)
      DEPLOY_FRONTEND=0
      ;;
    *)
      echo "❌ usage: $0 [amd64|arm64] [--no-frontend]" >&2
      exit 1
      ;;
  esac
  shift
done

VERSION_FILE="internal/version/VERSION"
if [[ ! -f "${VERSION_FILE}" ]]; then
  echo "❌ missing ${VERSION_FILE}" >&2
  exit 1
fi
APP_VERSION="$(tr -d '[:space:]' < "${VERSION_FILE}")"
if [[ -z "${APP_VERSION}" ]]; then
  echo "❌ ${VERSION_FILE} is empty" >&2
  exit 1
fi

echo
echo "Portal Go"
section "Build"
info "Version" "${APP_VERSION}"

if ! pick_go; then
  status "❌" "Need Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+"
  row "Current" "$(go version 2>/dev/null || echo 'go not found')"
  echo
  exit 1
fi
info "Go" "$("${GO}" version)"

if [[ "${DEPLOY_FRONTEND}" == "1" ]]; then
  status "🎨" "Building manager frontend …"
  if [[ ! -d web/manager/node_modules ]]; then
    run_nested bash -lc "cd web/manager && yarn install"
  fi
  run_nested bash -lc "cd web/manager && yarn build"
  ok "✅" "Frontend OK"
fi

mkdir -p "${OUT_DIR}"
status "🔨" "Compiling ${GOOS}/${GOARCH} …"
CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" "${GO}" build -o "${PORTAL_BIN}" ./cmd/portal
CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" "${GO}" build -o "${WS_BIN}" ./cmd/ws
CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" "${GO}" build -o "${CRON_BIN}" ./cmd/cron
ok "✅" "Compile OK"
row "portal" "$(file_size_human "${PORTAL_BIN}")"
row "ws" "$(file_size_human "${WS_BIN}")"
row "cron" "$(file_size_human "${CRON_BIN}")"

section "Done"
info "Version" "${APP_VERSION}"
status "🎉" "Build complete — run ./deploy.sh to upload"
echo
