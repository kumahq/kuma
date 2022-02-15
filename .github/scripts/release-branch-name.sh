#! /usr/bin/env bash

lastGitTag=$(git describe --abbrev=0 --tags)

IFS=. read -r major minor patch <<< "$lastGitTag"
echo "release-$major.$((++minor))"