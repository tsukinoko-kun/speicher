#!/usr/bin/env just --justfile

docs:
    go run ./cmd/docs -o ../speicher.wiki/
    (cd ../speicher.wiki/ && git add . && git commit -m update && git push)
