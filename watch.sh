#!/bin/sh
# Simple "watch" script for changes.
find . -name \*.go | entr goapp test
