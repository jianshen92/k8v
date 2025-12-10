#!/usr/bin/env bash
set -euo pipefail

# Install k8v by downloading a release binary from GitHub.
# Usage:
#   ./scripts/install.sh                 # install latest release
#   VERSION=v0.1.0 ./scripts/install.sh  # install specific version
#   ./scripts/install.sh v0.1.0          # version as arg

OWNER="user"
REPO="k8v"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

fail() {
  echo "error: $*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

need_cmd curl
need_cmd uname

version="${VERSION:-${1:-latest}}"

detect_platform() {
  case "$(uname -s)" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) os="windows" ;;
    *) fail "unsupported OS $(uname -s)" ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) fail "unsupported architecture $(uname -m)" ;;
  esac

  if [[ "$os" == "windows" ]]; then
    ext=".exe"
  else
    ext=""
  fi
}

fetch_latest_tag() {
  curl -fsSL "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" \
    | sed -n 's/ *"tag_name": *"\\(.*\\)",/\\1/p'
}

detect_platform

if [[ "$version" == "latest" ]]; then
  version="$(fetch_latest_tag)"
  [[ -n "$version" ]] || fail "could not determine latest release tag"
fi

asset="k8v-${os}-${arch}${ext}"
url="https://github.com/${OWNER}/${REPO}/releases/download/${version}/${asset}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
outfile="${tmpdir}/${asset}"

echo "Downloading ${url}"
curl -fL "$url" -o "$outfile"

if [[ "$ext" != ".exe" ]]; then
  chmod +x "$outfile"
fi

mkdir -p "$INSTALL_DIR"
dest="${INSTALL_DIR}/k8v${ext}"

if install -m 755 "$outfile" "$dest" 2>/dev/null; then
  :
else
  echo "install to ${dest} requires elevated permissions; retrying with sudo"
  sudo install -m 755 "$outfile" "$dest"
fi

echo "k8v installed to ${dest}"
echo "Run with: ${dest}"
