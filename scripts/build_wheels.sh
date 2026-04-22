#!/usr/bin/env bash
#
# Cross-compile the isetup Go binary for every supported GOOS/GOARCH and bake
# each result into a platform-tagged Python wheel. Produces six wheels (one per
# supported OS/arch pair) under ${WHEEL_OUT:-dist-wheels}/.
#
# Usage:
#   scripts/build_wheels.sh            # uses version from pyproject.toml
#   VERSION=1.2.3 scripts/build_wheels.sh
#
# Requires: go, python3, python3 -m build.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Version: prefer explicit env var, otherwise read from pyproject.toml.
VERSION="${VERSION:-$(grep -E '^version = ' pyproject.toml | head -1 | sed -E 's/version = "(.*)"/\1/')}"
if [[ -z "$VERSION" ]]; then
  echo "error: could not determine version" >&2
  exit 1
fi

LDFLAGS="-s -w -X github.com/host452b/isetup/cmd.Version=${VERSION}"
WHEEL_OUT="${WHEEL_OUT:-${REPO_ROOT}/dist-wheels}"
BIN_DIR="${REPO_ROOT}/python/isetup/bin"

echo "Building isetup v${VERSION} wheels → ${WHEEL_OUT}"

rm -rf "$WHEEL_OUT" "${REPO_ROOT}/build" "${REPO_ROOT}/dist" \
       "${REPO_ROOT}/python/isetup.egg-info" "${REPO_ROOT}/isetup.egg-info"
mkdir -p "$WHEEL_OUT"

# goos goarch python-platform-tag
TARGETS=(
  "linux   amd64 manylinux_2_17_x86_64"
  "linux   arm64 manylinux_2_17_aarch64"
  "darwin  amd64 macosx_11_0_x86_64"
  "darwin  arm64 macosx_11_0_arm64"
  "windows amd64 win_amd64"
  "windows arm64 win_arm64"
)

for line in "${TARGETS[@]}"; do
  # shellcheck disable=SC2086
  set -- $line
  goos="$1"; goarch="$2"; plat="$3"

  ext=""
  [[ "$goos" == "windows" ]] && ext=".exe"

  echo
  echo "==> $goos/$goarch ($plat)"

  # Cross-compile, replacing any previous binary so the wheel only ever
  # bundles the target we're about to package.
  rm -rf "$BIN_DIR"
  mkdir -p "$BIN_DIR"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -ldflags "$LDFLAGS" -o "${BIN_DIR}/isetup${ext}" .

  # setuptools's bdist_wheel accepts --plat-name directly; we bypass
  # `python -m build` here because it doesn't expose per-build plat tags.
  rm -rf "${REPO_ROOT}/build" "${REPO_ROOT}/dist"
  python3 setup.py bdist_wheel \
    --plat-name="$plat" \
    --dist-dir="$WHEEL_OUT" \
    >/dev/null
done

rm -rf "$BIN_DIR"
rm -rf "${REPO_ROOT}/build" "${REPO_ROOT}/dist" \
       "${REPO_ROOT}/python/isetup.egg-info" "${REPO_ROOT}/isetup.egg-info"

echo
echo "Wheels built:"
ls -1 "$WHEEL_OUT"
