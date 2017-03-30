#!/bin/sh

VERSION=$(git describe --tags --always --dirty)
DATE=$(date -u '+%Y-%m-%d-%H%M UTC')
HASH=$(git rev-parse --verify HEAD)
HASH_SHORT=$(git rev-parse --verify --short HEAD)
cat <<EOF > version.go

package main

var version = "$VERSION"
var version_hash = "$HASH"
var version_hash_short = "$HASH_SHORT"
var version_date = "$DATE"

EOF
