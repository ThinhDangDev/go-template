#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)"
PROTO_DIR="$ROOT_DIR/proto"
THIRD_PARTY_DIR="$ROOT_DIR/third_party/googleapis"
OUT_DIR="$ROOT_DIR/protogen"
DOCS_DIR="$ROOT_DIR/internal/docs"

PROTOC_BIN="${PROTOC_BIN:-$(command -v protoc)}"
if [[ -z "$PROTOC_BIN" ]]; then
  echo "protoc is required" >&2
  exit 1
fi

PROTOC_INCLUDE="${PROTOC_INCLUDE:-$(dirname "$(dirname "$PROTOC_BIN")")/include}"

mkdir -p "$OUT_DIR" "$DOCS_DIR"
rm -f "$OUT_DIR"/*.pb.go "$OUT_DIR"/*.gw.go "$DOCS_DIR"/*.swagger.json

protoc \
  -I "$PROTO_DIR" \
  -I "$THIRD_PARTY_DIR" \
  -I "$PROTOC_INCLUDE" \
  --go_out "$OUT_DIR" \
  --go_opt paths=source_relative \
  --go-grpc_out "$OUT_DIR" \
  --go-grpc_opt paths=source_relative \
  --grpc-gateway_out "$OUT_DIR" \
  --grpc-gateway_opt paths=source_relative \
  --openapiv2_out "$DOCS_DIR" \
  --openapiv2_opt logtostderr=true,allow_merge=true,merge_file_name=api \
  "$PROTO_DIR/api.proto"

echo "generated protogen/*.go and internal/docs/api.swagger.json"
