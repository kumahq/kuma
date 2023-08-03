#!/bin/bash

REPO=$1
PR_NUMBER=$2
BASE_REF=$3
HEAD_REF=$4
PREDEFINED_GLOBS=$5
LABEL_TO_ADD=$6

# Enable recursive globs, needs bash 4!
shopt -s globstar

# Convert the comma-separated globs to an array
IFS=',' read -ra PREDEFINED_GLOBS_ARR <<< "$PREDEFINED_GLOBS"

# Get the changed files in the PR using git diff
CHANGED_FILES=$(git diff --name-only "$BASE_REF" "$HEAD_REF")
echo "Changed files:"
echo "$CHANGED_FILES"

# Collect matching files in an array
MATCHING_FILES=()
for glob in "${PREDEFINED_GLOBS_ARR[@]}"; do
  for file in $CHANGED_FILES; do
    # shellcheck disable=SC2053
    if [[ "$file" == $glob ]]; then
      MATCHING_FILES+=("$file")
    fi
  done
done

# Check if there are any matching files before updating the issue
if [ ${#MATCHING_FILES[@]} -gt 0 ]; then
  echo "Matching files:"
  printf '%s\n' "${MATCHING_FILES[@]}"

  echo "Adding '$LABEL_TO_ADD' label to the pull request..."
  gh issue edit "$PR_NUMBER" --add-label "$LABEL_TO_ADD" -R "$REPO"
else
  echo "No matching files found. Not adding any label."
fi
