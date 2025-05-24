#!/usr/bin/env bash
set -euo pipefail
cd /var/www/toolcenter/api
/usr/local/go/bin/go mod tidy
/usr/local/go/bin/go build -o toolcenter main.go
exec ./toolcenter
# This script is used to start the API server for the ToolCenter application. MAIS A LA MAIN !