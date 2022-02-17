#! /usr/bin/env bash

# make-release-tag.sh: This script assumes that you are on a branch and
# have otherwise prepared the release. It creates an annotated tag with
# a message containing the shortlog of commits from the previous version.
#
# Usage: make-release-tag.sh OLDVERS NEWVERS
#
# Since the log from OLDVERS to NEWVERS is used to generate the commit
# message for the annotated tag, OLDVERS should be the last "fully released"
# version, i.e. it's not that helpful to see the log from a last RC in
# the final release tag.

PROGNAME=$(basename "$0")
readonly PROGNAME
readonly OLDVERS="$1"
readonly NEWVERS="$2"

if [ -z "$OLDVERS" ] || [ -z "$NEWVERS" ]; then
    printf "Usage: %s OLDVERS NEWVERS\n" "$PROGNAME"
    exit 1
fi

shopt -s extglob # Enable extended pattern matching in case switches.

set -o errexit
set -o nounset
set -o pipefail

if [ -z "$(git tag --list "$OLDVERS")" ]; then
    printf "%s: tag '%s' does not exist\n" "$PROGNAME" "$OLDVERS"
    exit 1
fi

if [ -n "$(git tag --list "$NEWVERS")" ]; then
    printf "%s: tag '%s' already exists\n" "$PROGNAME" "$NEWVERS"
    exit 1
fi

case "$NEWVERS" in
v+([0-9.])?([-]*))
    ;;
*)
    printf "%s: tag '%s' must be of the form vX.Y.X\n" "$PROGNAME" "$NEWVERS"
    exit 1
    ;;
esac

git tag -F - "$NEWVERS" <<EOF
Tag $NEWVERS release.

$(git shortlog "$OLDVERS..HEAD")
EOF

printf "Created tag '%s'.\n" "$NEWVERS"

# People set up their remotes in different ways, so we need to check
# which one to dry run against. Choose a remote name that pushes to the
# kumahq org repository (i.e. not the user's Github fork).
REMOTE=$(git remote -v | awk '$2~/kumahq\/kuma/ && $3=="(push)" {print $1}' | head -n 1)
readonly REMOTE
if [ -z "$REMOTE" ]; then
    printf "%s: unable to determine remote for %s\n" "$PROGNAME" "kumahq/kuma"
    exit 1
fi

printf "Testing whether tag '%s' can be pushed.\n" "$NEWVERS"
git push --dry-run "$REMOTE" "$NEWVERS"

printf "Run 'git push %s %s' to push the tag if you are happy.\n" "$REMOTE" "$NEWVERS"
