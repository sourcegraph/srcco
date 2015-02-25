#!/bin/sh

set -e

# Create site dir if it does not exist
mkdir -p site

# Python server
cd site
python2 -m SimpleHTTPServer &
cd ..

# Kill python server on exit
trap "exit" INT TERM
trap "kill 0" EXIT

while true; do
  echo "Building site..."
  make install
  rm -rf .srclib-cache #SAMER: fix this
  srcco -v gen

  echo "Waiting for changes..."
  inotifywait -e modify -r .
done
