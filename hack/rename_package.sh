#!/bin/bash
set -ex
ORIGINAL="github.com\/jlewi\/pkg"
NEW="github.com\/jlewi\/p22h"
find ./backend/ -name "*.go"  -exec  sed -i ".bak" "s/${ORIGINAL}/${NEW}/g" {} ";"
sed -i ".bak" "s/${ORIGINAL}/${NEW}/g" backend/go.mod
find ./backend/ -name "*.bak" -exec rm {} ";"