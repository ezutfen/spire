#!/usr/bin/env bash

set -euo pipefail

REPO="ezutfen/spire"
BASE_BRANCH="master"

usage() {
	echo "Usage: $0 --title <title> [--body-file <path>] [--body <text>] [--draft] [additional gh pr create args]"
	echo
	echo "Creates a pull request against ${REPO}:${BASE_BRANCH}."
	echo "Any extra arguments are forwarded to \`gh pr create\`."
}

require_command() {
	local command_name="$1"

	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "Missing required command: $command_name" >&2
		exit 1
	fi
}

require_command gh

TITLE=""
BODY_FILE=""
BODY_TEXT=""
FORWARD_ARGS=()

while [[ $# -gt 0 ]]; do
	case "$1" in
		--title)
			TITLE="${2:-}"
			shift 2
			;;
		--body-file)
			BODY_FILE="${2:-}"
			shift 2
			;;
		--body)
			BODY_TEXT="${2:-}"
			shift 2
			;;
		--draft)
			FORWARD_ARGS+=("$1")
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			FORWARD_ARGS+=("$1")
			shift
			;;
	esac
done

if [[ -z "$TITLE" ]]; then
	echo "--title is required" >&2
	usage >&2
	exit 1
fi

if [[ -n "$BODY_FILE" && -n "$BODY_TEXT" ]]; then
	echo "Use only one of --body-file or --body" >&2
	exit 1
fi

CURRENT_BRANCH=$(git branch --show-current)

if [[ -z "$CURRENT_BRANCH" ]]; then
	echo "Unable to determine the current branch" >&2
	exit 1
fi

COMMAND=(
	gh pr create
	--repo "$REPO"
	--base "$BASE_BRANCH"
	--head "$CURRENT_BRANCH"
	--title "$TITLE"
)

if [[ -n "$BODY_FILE" ]]; then
	COMMAND+=(--body-file "$BODY_FILE")
elif [[ -n "$BODY_TEXT" ]]; then
	COMMAND+=(--body "$BODY_TEXT")
fi

COMMAND+=("${FORWARD_ARGS[@]}")

printf 'Creating PR against %s from branch %s\n' "$REPO" "$CURRENT_BRANCH"
"${COMMAND[@]}"
