#!/usr/bin/env bash
# Regenerate the API method reference under docs/reference.
#
# Source of truth is github.com/gotd/td at the pinned ref below. Override with
# TD_DIR=/path/to/td to generate from a local checkout (its tg/ subdir is used).
set -euo pipefail

# Pinned td version. Keep in sync with the CI drift-check workflow.
TD_REF="${TD_REF:-v0.154.0}"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/../.." && pwd)"

if [[ -n "${TD_DIR:-}" ]]; then
  td_dir="${TD_DIR}"
else
  cache="${script_dir}/.cache/td-${TD_REF}"
  if [[ ! -d "${cache}/tg" ]]; then
    echo "cloning gotd/td@${TD_REF} into ${cache}"
    rm -rf "${cache}"
    git clone --quiet --depth 1 --branch "${TD_REF}" https://github.com/gotd/td "${cache}"
  fi
  td_dir="${cache}"
fi

echo "generating reference from ${td_dir}/tg"
cd "${script_dir}"
go run . --tg "${td_dir}/tg" --out "${repo_root}/docs/reference"
