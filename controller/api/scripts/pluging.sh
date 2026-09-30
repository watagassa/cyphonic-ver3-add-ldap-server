#!/bin/bash -e

tools="$(
  cd $(dirname $0)
  pwd
)/../../.tool-versions"

echo "🏃 Installing asdf tools in progress ..."
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
awk '{ print $1; }' ${tools} | while read tool; do
  asdf plugin add ${tool}
done
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
asdf install
