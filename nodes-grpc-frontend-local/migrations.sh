#!/bin/bash
set -xe

if ! command -v migrate 2>&1 >/dev/null; then
  exit
fi

migrate -path ./src/migrations -database "postgres://root:root@localhost:5432/root?sslmode=disable" down
migrate -path ./src/migrations -database "postgres://root:root@localhost:5432/root?sslmode=disable" up
