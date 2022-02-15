#! /usr/bin/env bash

lastGitTag=$(git describe --abbrev=0 --tags)

IFS=. read -r major minor patch <<< "$lastGitTag"
echo "$major.$((++minor)).0-rc1"