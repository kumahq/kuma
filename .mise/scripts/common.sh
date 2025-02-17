#!/usr/bin/env bash

common::export_build_info() {
  # Declare and populate an associative array
  declare -gA BUILD_INFO

  # Extract build information from the version script
  read -r BUILD_INFO["version"] \
    BUILD_INFO["git_tag"] \
    BUILD_INFO["git_commit"] \
    BUILD_INFO["build_date"] \
    BUILD_INFO["release"] <<< "$(tools/releases/version.sh)"

  # Export the associative array for use in the environment
  export BUILD_INFO
}

common::get_resources() {
  local resources_dir="$1"
  local resources
  local -a resources_array

  # Find directories and exclude specific names
  # -maxdepth 1: Search only one level deep
  # -mindepth 1: Skip the base directory itself
  # -type d: Include only directories
  # ! -name: Exclude directories named "core", "mesh", and "system"
  # -exec basename {} \;: Extract only the base names of the directories
  resources=$(find "$resources_dir" \
    -maxdepth 1 \
    -mindepth 1 \
    -type d \
    ! -name "core" \
    ! -name "mesh" \
    ! -name "system" \
    -exec basename {} \;)

  # Sort and store in an array
  mapfile -t resources_array < <(echo "$resources" | sort)

  echo "${resources_array[@]}"
}
