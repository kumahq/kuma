#!/usr/bin/env bash

# Script for parsing and formatting test reports with optional color output and JSON format.

# Default configuration
input="/dev/stdin"
format="pretty"
use_color=true

# ANSI escape codes for color output
bold=""
reset=""
gray=""

# Associative array to store spec counts per suite
declare -A suite_counts
total_specs=0

# Parse command-line arguments
parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --input-file) input=$2; shift 2 ;;
      --format) format=$2; shift 2 ;;
      --no-color) use_color=false; shift ;;
      *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
  done

  if [[ -n "$format" && "$format" != "json" && "$format" != "pretty" ]]; then
    echo "Invalid format: $format. Supported formats: json, pretty" >&2
    exit 1
  fi
}

# Setup colors if enabled
setup_colors() {
  if $use_color; then
    bold=$'\e[1m'
    reset=$'\e[0m'
    gray=$'\e[90m'
  fi
}

# Extract JSON data into a temporary file
extract_json_data() {
  tmpfile=$(mktemp)
  jq -r '
    map({
      key: (.SuiteDescription + "§" + (.SuitePath | sub("^.*github.com/[^/]+?/[^/]+?/"; ""))),
      value: [
        .SpecReports[]
        | select(.State != "passed")
        | {
            Description: (.ContainerHierarchyTexts + [.LeafNodeText] | join(" ")),
            Paths: ([.ContainerHierarchyLocations[], .LeafNodeLocation]
              | map((.FileName | sub("^.*github.com/[^/]+?/[^/]+?/"; "")) + ":" + (.LineNumber | tostring)))
          }
      ] | select(length > 0)
    })
    | from_entries
    | to_entries[]
    | "\(.key)⸬\(.value | tojson)"
  ' "$input" > "$tmpfile"
}

# Pretty print output for each suite
pretty_print_suite() {
  local suite_desc=$1
  local suite_path=$2
  local specs_json=$3
  local bar; bar=$(printf '%.0s=' {1..100})

  printf '%b%s%b\n' "$bold$gray" "$bar" "$reset"
  printf '%b%*s%b\n' "$bold" $((( ${#suite_desc} + 100 ) / 2)) "$suite_desc" "$reset"
  printf '%b%*s%b\n' "$gray" $((( ${#suite_path} + 102 ) / 2)) "($suite_path)" "$reset"
  printf '%b%s%b\n\n' "$bold$gray" "$bar" "$reset"

  jq -r '.[] | "\(.Description)⸬\(.Paths|join(","))"' <<< "$specs_json" | while IFS='⸬' read -r title locations; do
    printf '%b%s%b\n' "$bold" "$title" "$reset"

    IFS=',' read -r -a loc_array <<< "$locations"
    for idx in "${!loc_array[@]}"; do
      case $idx in
        0) prefix='  ';;
        1) prefix='  └─ ';;
        *) prefix=$(printf ' %*s└─ ' $((idx * 2)) '') ;;
      esac
      printf '%b%s%s%b\n' "$gray" "$prefix" "${loc_array[$idx]}" "$reset"
    done
    echo
  done
}

# Pretty print summary
generate_summary() {
  local summary_bar; summary_bar=$(printf '%.0s=' {1..100})

  printf '%b%s%b\n' "$bold$gray" "$summary_bar" "$reset"
  printf '%b%*s%b\n' "$bold" 54 "Summary" "$reset"
  printf '%b%s%b\n' "$bold$gray" "$summary_bar" "$reset"

  printf '\n%bTotal disabled specs: %s%b\n\n' "$bold" "$total_specs" "$reset"

  for suite_key in "${!suite_counts[@]}"; do
    IFS='§' read -r suite_desc suite_path <<< "$suite_key"
    printf '  %b%s%b %s%b: %d\n' "$bold" "$suite_desc" "$reset$gray" "($suite_path)" "$reset" "${suite_counts[$suite_key]}"
  done

  printf '\n%b' "$bold$gray"; printf '%.0s=' {1..100}; printf '%b' "$reset\n"
}

# Main function
main() {
  parse_args "$@"
  setup_colors

  if [[ "$format" == "json" ]]; then
    jq -r '
      map({
        Suite: {
          Description: .SuiteDescription,
          Path: (.SuitePath | sub("^.*github.com/[^/]+?/[^/]+?/"; ""))
        },
        Specs: [
          .SpecReports[] | select(.State != "passed") | {
            Description: (.ContainerHierarchyTexts + [.LeafNodeText] | join(" ")),
            Paths: ([.ContainerHierarchyLocations[], .LeafNodeLocation]
              | map((.FileName | sub("^.*github.com/[^/]+?/[^/]+?/"; "")) + ":" + (.LineNumber | tostring)))
          }
        ] | select(length > 0)
      })
    ' "$input"
    exit 0
  fi

  extract_json_data

  while IFS='⸬' read -r suite_key specs_json; do
    IFS='§' read -r suite_desc suite_path <<< "$suite_key"

    local count; count=$(jq length <<< "$specs_json")
    suite_counts["$suite_key"]=$count
    (( total_specs += count ))

    pretty_print_suite "$suite_desc" "$suite_path" "$specs_json"
  done < "$tmpfile"

  rm -f "$tmpfile"
  generate_summary
}

main "$@"
