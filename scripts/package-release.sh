#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 <auth|componentry|security|server|site> <vX.Y.Z>" >&2
  exit 1
}

[[ $# -eq 2 ]] || usage

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

package="$1"
version="$2"
owner="${PACKAGE_REPO_OWNER:-go-sum}"
token="${PACKAGE_SYNC_TOKEN:-}"
gh_token="${GH_TOKEN:-${token}}"

case "${package}" in
  auth|componentry|security|server|site)
    ;;
  *)
    usage
    ;;
esac

if [[ ! "${version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-].+)?$ ]]; then
  echo "version must be a semver-like tag such as v0.1.0" >&2
  exit 1
fi

if [[ -z "${token}" ]]; then
  echo "PACKAGE_SYNC_TOKEN is required" >&2
  exit 1
fi

prefix="pkg/${package}"
repo="${owner}/${package}"
remote_url="https://x-access-token:${token}@github.com/${repo}.git"
split_sha="$(git subtree split --prefix="${prefix}/")"

echo "Releasing ${prefix} to ${repo} as ${version} (${split_sha})"
git push "${remote_url}" \
  "${split_sha}:refs/heads/main" \
  "${split_sha}:refs/tags/${version}"

if command -v gh >/dev/null 2>&1; then
  if [[ -n "${gh_token}" ]]; then
    export GH_TOKEN="${gh_token}"
    gh release create "${version}" \
      --repo "${repo}" \
      --title "${version}" \
      --notes "Released from go-sum/forge subtree ${prefix}"
  else
    echo "GH_TOKEN not set; skipping GitHub Release creation"
  fi
else
  echo "gh CLI not available; skipping GitHub Release creation"
fi
