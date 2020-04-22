#!/usr/bin/env bash
set -e
cd _tools/linter
set -x
go run ../ls-imports/main.go -u -f magefile.go | xargs -tI % go install -v %
