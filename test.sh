#!/bin/sh

set -e

# Create site dir if it does not exist
mkdir -p docs

# Python server
cd docs
python2 -m SimpleHTTPServer &
cd ..

# Kill python server on exit
trap "exit" INT TERM
trap "kill 0" EXIT

while true; do
  echo "Building docs..."
  make install
  rm -rf .srclib-cache #SAMER: fix this
  srcco -v .

  echo "Waiting for changes..."
  inotifywait -e modify -r .
done
