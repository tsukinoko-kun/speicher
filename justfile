#!/usr/bin/env just --justfile

docs:
    go run ./cmd/docs > DOCS.md
