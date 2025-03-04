#!/usr/bin/env bash

# Function to replace $HOME with ~ in a given string
shorten_home() {
  echo "${1//$HOME/\~}"
}

log() {
  # Use the provided prefix or default to "___task_" if no prefix is given
  local prefix="${1:-___task_}"

  # Declare an associative array to store the table rows
  declare -A table_rows
  local var_name

  while read -r var_name; do
    # Skip variables that don't match the expected pattern (with optional parameters prefix)
    if ! [[ "$var_name" =~ $prefix([ar]{1,2}_)?([a-zA-Z0-9_]+) ]]; then
      continue
    fi

    # Extract the variable name without the prefix and convert it to lowercase
    local name="${BASH_REMATCH[-1],,}"

    # Dynamically reference the variable using `declare -n`
    declare -n full_var="$var_name"

    # Check if the variable is not an array (using grep to check for 'a' flag in `declare`)
    if ! declare -p "$var_name" 2>/dev/null | grep -qE 'declare -.*a.* '; then
      table_rows["$name"]="$(shorten_home "$full_var")"
      continue
    fi

    # For array-like variables, iterate through each element
    table_rows["${name}[]"]=$(
      for i in "${!full_var[@]}"; do
        echo "$(gum style --foreground 240 "$i:") $(shorten_home "${full_var[i]}")"
      done
    )
  done < <(declare -p | cut -d ' ' -f 3 | grep "^$prefix" | cut -d '=' -f 1)

  {
    echo "Variable,Value"
    for key in $(printf "%s\n" "${!table_rows[@]}" | sort); do
      echo "$key,\"${table_rows[$key]}\""
    done
  } | gum table --print
}

variables::verify() {
  local error_messages=()
  local var_name
  local prefix="${1:-___task_}"  # Use the argument as prefix or default to ___task_

  for var_name in $(declare -p | cut -d ' ' -f 3 | grep "^$prefix"); do
    var_name="${var_name%%=*}" # Remove everything after '=' including '='

    # Parameters explanation:
    #   r: Marks a variable as required. The function checks if it is set
    #   a: Indicates the variable is specified via arguments and not flag
    # Variables must use the prefix and follow the format: <prefix><parameters>_<name>
    if ! [[ "$var_name" =~ ${prefix}([ar]{1,2})_([a-zA-Z0-9_]+) ]]; then
      continue
    fi

    local parameters="${BASH_REMATCH[1]}"
    local name="${BASH_REMATCH[2]}"

    # Skip the loop if variable is not required
    if [[ "$parameters" != *r* ]]; then
      continue
    fi

    if declare -p "$var_name" 2>/dev/null | grep -qE 'declare \-.*a.*\ '; then
      # Check if the array is non-empty, prevent unbound variable error with ':-'
      [[ "${!var_name[*]:-}" ]] && continue
    elif [[ -n ${!var_name} ]]; then
      # Check if the string variable is non-empty
      continue
    fi

    local flag="--${name//_/-}" # Replace underscores with dashes
    local env="${name^^}"       # Convert to uppercase

    if [[ "$parameters" != *a* ]]; then
      error_messages+=("Error: $name not specified; provide via $flag flag or $env env var")
    else
      error_messages+=("Error: $name missing; provide as argument(s) or $env env var")
    fi
  done

  # If there are errors, log all of them and exit
  if [[ ${#error_messages[@]} -gt 0 ]]; then
    printf "%s\n" "${error_messages[@]}" >&2
    exit 1
  fi

  case "${VERBOSE_LOG:-}" in
    true|1|on|enabled)
      log "$prefix" ;;
  esac
}
