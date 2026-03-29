#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage:
  scripts/workspace.sh list
  scripts/workspace.sh exec <command> [args...]
  scripts/workspace.sh cover [go test args...]

commands:
  list   Print modules from go.work, one per line
  exec   Run a command in each workspace module
  cover  Run go test coverage in each workspace module and print summaries

environment:
  COVERAGE_FILE  Base coverage filename for the cover command (default: coverage.out)
EOF
  exit 1
}

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

workspace_modules() {
  awk '
    $1 == "use" && $2 == "(" { in_use = 1; next }
    in_use && $1 == ")" { in_use = 0; next }
    in_use { print $1; next }
    $1 == "use" && $2 != "(" { print $2 }
  ' go.work
}

run_each() {
  local module

  [[ $# -ge 1 ]] || usage

  while IFS= read -r module; do
    echo "==> ${module}"
    (
      cd "${repo_root}/${module}"
      "$@"
    )
  done < <(workspace_modules)
}

run_cover() {
  local coverage_file coverage_base module safe_name out_file

  coverage_file="${COVERAGE_FILE:-coverage.out}"
  coverage_base="${coverage_file%.out}"

  while IFS= read -r module; do
    safe_name="${module//\//_}"
    out_file="${repo_root}/${coverage_base}.${safe_name}.out"
    echo "==> ${module}"
    (
      cd "${repo_root}/${module}"
      go test -coverpkg=./... -coverprofile="${out_file}" "$@" ./...
    )
    go tool cover -func="${out_file}"
  done < <(workspace_modules)
}

[[ $# -ge 1 ]] || usage

case "$1" in
  list)
    shift
    [[ $# -eq 0 ]] || usage
    workspace_modules
    ;;
  exec)
    shift
    run_each "$@"
    ;;
  cover)
    shift
    run_cover "$@"
    ;;
  *)
    usage
    ;;
esac
