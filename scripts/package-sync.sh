#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 <auth|componentry|security|server|site>" >&2
  exit 1
}

[[ $# -eq 1 ]] || usage

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

package="$1"
owner="${PACKAGE_REPO_OWNER:-go-sum}"
token="${PACKAGE_SYNC_TOKEN:-}"

case "${package}" in
  auth|componentry|security|server|site)
    ;;
  *)
    usage
    ;;
esac

if [[ -z "${token}" ]]; then
  echo "PACKAGE_SYNC_TOKEN is required" >&2
  exit 1
fi

prefix="pkg/${package}"
repo="${owner}/${package}"
remote_url="https://x-access-token:${token}@github.com/${repo}.git"

if [[ ! -d "${prefix}" ]]; then
  echo "package prefix not found: ${prefix}" >&2
  exit 1
fi

split_sha="$(git subtree split --prefix="${prefix}/")"

echo "Syncing ${prefix} to ${repo}@main (${split_sha})"
git push "${remote_url}" "${split_sha}:refs/heads/main"
