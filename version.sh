#!/bin/sh

VERSION=$(git describe --tags --always --abbrev=0)
VERSIONFULL=$(git describe --tags --always)
DATE=$(date -u '+%Y-%m-%d-%H%M UTC')
HASH=$(git rev-parse --verify HEAD)
HASH_SHORT=$(git rev-parse --verify --short HEAD)
cat <<EOF > version.go
package main


var Version string = "$VERSION"
var FullVersion string = "$VERSIONFULL"
var BuildTimestamp string = "$DATE"
var Hash string = "$HASH"

EOF
