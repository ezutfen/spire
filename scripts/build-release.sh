#!/usr/bin/env bash

set -euo pipefail

usage() {
	echo "Usage: $0 [--output <path>] [--skip-frontend] [--release]"
	echo
	echo "Builds frontend/dist and embeds it into a fresh Go binary."
	echo "Default output: ./spire"
	echo
	echo "Options:"
	echo "  --output <path>    Write the local binary to this path"
	echo "  --skip-frontend    Reuse the existing frontend/dist contents"
	echo "  --release          Run release packaging/publish targets after the local build"
}

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
OUTPUT_PATH="$ROOT_DIR/spire"
BUILD_FRONTEND=1
RELEASE_BUILD=0

while [[ $# -gt 0 ]]; do
	case "$1" in
		--output)
			if [[ $# -lt 2 ]]; then
				echo "--output requires a path" >&2
				exit 1
			fi
			OUTPUT_PATH="$2"
			shift 2
			;;
		--skip-frontend)
			BUILD_FRONTEND=0
			shift
			;;
		--release)
			RELEASE_BUILD=1
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			echo "Unknown argument: $1" >&2
			usage >&2
			exit 1
			;;
	esac
done

ensure_release_version_is_new() {
	local latest_tag local_version

	latest_tag=$(curl -fsSL "https://api.github.com/repos/EQEmu/spire/tags" | jq -r '.[0].name' | sed 's/^v//')
	local_version=$(jq -r '.version' "$ROOT_DIR/package.json")

	if [[ "$latest_tag" == "$local_version" ]]; then
		echo "Version tag matches the latest release ($local_version); refusing release build."
		exit 1
	fi

	echo "Local version $local_version differs from latest release $latest_tag; continuing release build."
}

require_command() {
	local command_name="$1"

	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "Missing required command: $command_name" >&2
		exit 1
	fi
}

install_frontend_dependencies() {
	if [[ -f "$ROOT_DIR/frontend/package-lock.json" ]]; then
		npm ci --legacy-peer-deps
	else
		npm install --legacy-peer-deps
	fi
}

if [[ $BUILD_FRONTEND -eq 1 ]]; then
	echo "Installing frontend dependencies with legacy peer resolution..."
	cd "$ROOT_DIR/frontend"
	install_frontend_dependencies
	echo "Building frontend/dist..."
	npm run build
fi

echo "Building local Spire binary at $OUTPUT_PATH..."
cd "$ROOT_DIR"
go build -o "$OUTPUT_PATH"

if [[ $RELEASE_BUILD -eq 1 ]]; then
	require_command curl
	require_command jq
	require_command gh-release
	require_command zip
	ensure_release_version_is_new
	echo "Running release packaging targets..."
	cd "$ROOT_DIR"
	make build-binary
	make build-installer-binary
	make release-binary
fi
