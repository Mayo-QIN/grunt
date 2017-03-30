#!/bin/sh

VERSION=$(git describe --tags --always --abbrev=0)
VERSIONFULL=$(git describe --tags --always)
DATE=$(date -u '+%Y-%m-%d-%H%M UTC')
HASH=$(git rev-parse --verify HEAD)
HASH_SHORT=$(git rev-parse --verify --short HEAD)
cat <<EOF > version.go
package main

var VersionInfo = struct {
	Version     string
	FullVersion string
	Hash        string
	ShortHash   string
	Date        string
}{
  "$VERSION",
  "$VERSIONFULL",
  "$HASH",
  "$HASH_SHORT",
  "$DATE",
}

EOF
