#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $(basename "$0") <version>" >&2
  echo "example: $(basename "$0") 0.2.1" >&2
  exit 1
fi

VERSION="$1"
TAG="v${VERSION}"

if ! command -v git >/dev/null 2>&1; then
  echo "missing command: git" >&2
  exit 1
fi
if ! command -v go >/dev/null 2>&1; then
  echo "missing command: go" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "working tree is not clean. commit or stash changes first." >&2
  exit 1
fi

git fetch --tags origin

if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  echo "tag already exists locally: ${TAG}" >&2
  exit 1
fi
if git ls-remote --tags origin "refs/tags/${TAG}" | grep -q .; then
  echo "tag already exists on origin: ${TAG}" >&2
  exit 1
fi

echo "==> run tests"
go test ./...

echo "==> create tag ${TAG}"
git tag -a "${TAG}" -m "release: ${TAG}"

echo "==> push tag ${TAG}"
git push origin "${TAG}"

echo "Done: ${TAG}"
echo "GitHub Actions will auto-create/update the release notes."
