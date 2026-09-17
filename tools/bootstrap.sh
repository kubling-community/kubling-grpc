#!/usr/bin/env bash

set -euo pipefail

BUF_VERSION="1.59.0"

TOOLS_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${TOOLS_DIR}/bin"

mkdir -p "${BIN_DIR}"

if [ ! -x "${BIN_DIR}/buf" ] || [ "$("${BIN_DIR}/buf" --version)" != "$BUF_VERSION" ]; then
    echo "Downloading buf ${BUF_VERSION}..."
    case "$(uname -s)" in
      Linux) platform=Linux ;;
      Darwin) platform=Darwin ;;
      *) echo "Unsupported platform; run the build on Linux or macOS." >&2; exit 1 ;;
    esac
    case "$(uname -m)" in
      x86_64|amd64) architecture=x86_64 ;;
      aarch64|arm64) architecture=arm64 ;;
      *) echo "Unsupported architecture." >&2; exit 1 ;;
    esac
    if [[ "$platform" == Linux && "$architecture" == arm64 ]]; then
      architecture=aarch64
    fi
    download="$(mktemp "${BIN_DIR}/buf.download.XXXXXX")"
    trap 'rm -f "$download"' EXIT
    curl -fsSL --retry 3 \
      "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-${platform}-${architecture}" \
      -o "$download"
    chmod +x "$download"
    test "$("$download" --version)" = "$BUF_VERSION"
    mv "$download" "${BIN_DIR}/buf"
fi

echo "buf ${BUF_VERSION} ready"
