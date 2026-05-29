#!/usr/bin/env bash
# Regenerate the Go stubs for the fetch-record wire format from
# proto/fetch_record.proto.
#
# The generated file (internal/fetchrecord/fetch_record.pb.go) is
# committed, so the normal build (and the Docker image) needs no proto
# toolchain — only run this when proto/fetch_record.proto changes.
#
# Requires:
#   - protoc            (brew install protobuf)
#   - protoc-gen-go     (go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

for tool in protoc protoc-gen-go; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "error: $tool not found on PATH; see the header of this script for install hints" >&2
    exit 1
  fi
done

mkdir -p internal/fetchrecord

protoc \
  --proto_path=proto \
  --go_out=internal/fetchrecord \
  --go_opt=paths=source_relative \
  proto/fetch_record.proto

echo "generated internal/fetchrecord/fetch_record.pb.go"
