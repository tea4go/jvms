#!/usr/bin/env bash
set -euo pipefail

APP_NAME="jvms"
APP_VERSION="${1:-3.0.11}"
BUILD_TIME="$(date '+%Y-%m-%d(%H:%M:%S)')"
DIST_DIR="dist"

targets=(
  "windows amd64 zip"
  "linux amd64 tar.gz"
  "linux arm64 tar.gz"
  "darwin arm64 tar.gz"
  "darwin amd64 tar.gz"
)

echo "Build time: ${BUILD_TIME}"
echo "Version: ${APP_VERSION}"

mkdir -p "${DIST_DIR}"

for target in "${targets[@]}"; do
  read -r goos goarch archive_type <<<"${target}"

  output_name="${APP_NAME}"
  if [[ "${goos}" == "windows" ]]; then
    output_name="${APP_NAME}.exe"
  fi

  output_path="${DIST_DIR}/${output_name}"
  archive_base="${APP_NAME}${APP_VERSION}.${goos}-${goarch}"

  echo "============================================================="
  echo "Building ${goos}/${goarch}"

  rm -f "${output_path}"
  GO111MODULE=on CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
    go build -o "${output_path}" \
      -ldflags "-X main.AppVersion=${APP_VERSION} -X main.BuildTime=${BUILD_TIME}" .

  if [[ "${archive_type}" == "zip" ]]; then
    rm -f "${DIST_DIR}/${archive_base}.zip"
    (cd "${DIST_DIR}" && zip -q "${archive_base}.zip" "${output_name}")
  else
    rm -f "${DIST_DIR}/${archive_base}.tar.gz"
    tar -czf "${DIST_DIR}/${archive_base}.tar.gz" -C "${DIST_DIR}" "${output_name}"
  fi

  rm -f "${output_path}"
done

echo "============================================================="
echo "Build artifacts:"
ls -lh "${DIST_DIR}"/"${APP_NAME}${APP_VERSION}".*
