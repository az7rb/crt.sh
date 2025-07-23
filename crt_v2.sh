#!/usr/bin/env bash

# ==============================================================================
# @module crt_v2.sh
# @description
#   This script queries the crt.sh certificate transparency log to retrieve domain names associated with specific domains or organizations.
#   It processes the raw certificate data, cleaning and filtering it to provide unique and relevant domain results.
#
#   Key features include:
#     - Searching for domain names and subdomains.
#     - Searching for domains associated with specific organizations.
#     - Cleaning and de-duplicating search results.
#     - Outputting results to standard output or a specified file.
#
# @author
#   az7rb, Arqsz
# ==============================================================================

# Display banner
banner="
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      ..| search crt.sh v2.4 |..      |
+    site: crt.sh Certificate Search   +
|           Twitter: az7rb             |
|         Modified by: Arqsz           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
"

##!
# @description
#   Displays the help message for the script, detailing its usage,
#   available options, and examples.
#
# @example
#   show_help
#'
show_help() {
    echo "A script to query the crt.sh certificate transparency log."
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -d, --domain <domain>        Search for a specific domain name (e.g., hackerone.com)"
    echo "      --org <organization>     Search for a specific organization name (e.g., 'HackerOne, Inc.')"
    echo "  -o, --output <file>          File to save results. If not set, results are printed to stdout."
    echo "  -s, --silent                 Suppress banner and non-essential output."
    echo "  -h, --help                   Display this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --domain hackerone.com"
    echo "  $0 --org 'HackerOne, Inc.' --output ./results.txt"
    echo "  $0 -d example.com -s | grep '.com'"
    echo "  $0 -d example.com -s | httpx"
}

##!
# @description
#   Cleans and filters input data, typically piped to this function,
#   by removing unwanted characters and duplicates. It converts escaped
#   newlines to actual newlines, removes wildcard characters (*), filters out
#   email addresses and lines containing spaces, and ensures entries contain a dot.
#   Finally, it sorts the results and removes duplicates.
#
# @returns {string}
#   A multi-line string of cleaned, sorted, and unique domain names.
#
# @example
#   echo "example.com\\n*.test.com\\nuser@example.com" | clean_results
#'
clean_results() {
    sed 's/\\n/\n/g' |
        sed 's/\*.//g' |
        # Filter out entries that are likely email addresses
        grep -v '@' |
        # Filter out lines with spaces (e.g., certificate names)
        grep -v ' ' |
        # Filter out entries that do not contain a dot (i.e., not a full domain)
        grep '\.' |
        # Remove any leading/trailing whitespace
        sed 's/^[ \t]*//;s/[ \t]*$//' |
        sort -u
}

##!
# @description
#   Searches the crt.sh certificate transparency log for certificates
#   associated with a given domain name. It fetches common names and
#   name values from the crt.sh JSON API, then cleans and filters the results.
#
# @arg {string} $1 - The domain name to search for (e.g., "example.com").
# @arg {boolean} $2 - A flag indicating whether silent mode is enabled.
#   If `true`, suppresses status messages printed to standard error.
#
# @returns {string}
#   A multi-line string of unique and cleaned domain names found for the specified domain.
#   Returns an empty string if no results are found or the request fails.
#
# @example
#   domain_search "hackerone.com" false
#'
domain_search() {
    local domain_req="$1"
    local silent_mode="$2"

    if [[ "$silent_mode" = false ]]; then
        # Print status messages to stderr to not interfere with stdout
        echo "[*] Searching for domain: $domain_req" >&2
    fi

    local response
    response=$(curl -fs "https://crt.sh?q=%.$domain_req&output=json")

    # If request fails or returns no data, return nothing.
    if [[ $? -ne 0 || -z "$response" || "$response" == "[]" ]]; then
        return
    fi

    local results
    results=$(echo "$response" | jq -r '.[].common_name, .[].name_value' | clean_results)

    echo "$results"
}

##!
# @description
#   Searches the crt.sh certificate transparency log for certificates
#   associated with a given organization name. It fetches common names and
#   name values from the crt.sh JSON API, then cleans and filters the results.
#
# @arg {string} $1 - The organization name to search for (e.g., "HackerOne, Inc.").
#   This argument should be URL-encoded if it contains special characters.
# @arg {boolean} $2 - A flag indicating whether silent mode is enabled.
#   If `true`, suppresses status messages printed to standard error.
#
# @returns {string}
#   A multi-line string of unique and cleaned domain names found for the specified organization.
#   Returns an empty string if no results are found or the request fails.
#
# @example
#   org_search "HackerOne%2C%20Inc." false
#'
org_search() {
    local org_req="$1"
    local silent_mode="$2"

    if [[ "$silent_mode" = false ]]; then
        echo "[*] Searching for organization: $org_req" >&2
    fi

    local response
    response=$(curl -fs "https://crt.sh?q=$org_req&output=json")

    # If request fails or returns no data, return nothing.
    if [[ $? -ne 0 || -z "$response" || "$response" == "[]" ]]; then
        return
    fi

    local results
    results=$(echo "$response" | jq -r '.[].common_name, .[].name_value' | clean_results)

    echo "$results"
}

DOMAIN_SEARCH=""
ORG_SEARCH=""
OUTPUT_FILE=""
SILENT_MODE=false

if [[ "$#" -eq 0 ]]; then
    show_help
    exit 0
fi

while [[ "$#" -gt 0 ]]; do
    case "$1" in
    -h | --help)
        show_help
        exit 0
        ;;
    -s | --silent)
        SILENT_MODE=true
        ;;
    -d | --domain)
        if [[ -n "$2" && ! "$2" =~ ^- ]]; then
            DOMAIN_SEARCH="$2"
            shift
        else
            echo "Error: Missing argument for $1" >&2
            exit 1
        fi
        ;;
    --org)
        if [[ -n "$2" && ! "$2" =~ ^- ]]; then
            ORG_SEARCH="$2"
            shift
        else
            echo "Error: Missing argument for $1" >&2
            exit 1
        fi
        ;;
    -o | --output)
        if [[ -n "$2" && ! "$2" =~ ^- ]]; then
            OUTPUT_FILE="$2"
            shift
        else
            echo "Error: Missing argument for $1" >&2
            exit 1
        fi
        ;;
    *)
        echo "Error: Unknown parameter passed: $1" >&2
        show_help
        exit 1
        ;;
    esac
    shift
done

if [[ "$SILENT_MODE" = false ]]; then
    echo "${banner}" >&2
fi

if [[ -z "$DOMAIN_SEARCH" && -z "$ORG_SEARCH" ]]; then
    show_help
    exit 0
fi

if [[ -n "$DOMAIN_SEARCH" && -n "$ORG_SEARCH" ]]; then
    echo "Error: Please specify either a domain (-d) or an organization (--org), not both." >&2
    exit 1
elif [[ -z "$DOMAIN_SEARCH" && -z "$ORG_SEARCH" ]]; then
    echo "Error: You must specify a domain (-d) or an organization (--org) to search." >&2
    show_help
    exit 1
fi

RESULTS=""
if [[ -n "$DOMAIN_SEARCH" ]]; then
    RESULTS=$(domain_search "$DOMAIN_SEARCH" "$SILENT_MODE")
elif [[ -n "$ORG_SEARCH" ]]; then
    # URL-encode the organization string for the query
    ORG_SEARCH_ENCODED=$(echo -n "$ORG_SEARCH" | jq -sRr @uri)
    RESULTS=$(org_search "$ORG_SEARCH_ENCODED" "$SILENT_MODE")
fi

if [[ -z "$RESULTS" ]]; then
    if [[ "$SILENT_MODE" = false ]]; then
        echo "[-] No results found." >&2
    fi
    exit 1
fi

# Decide where to send the output based on whether -o was used
if [[ -n "$OUTPUT_FILE" ]]; then
    output_dir=$(dirname "$OUTPUT_FILE")
    mkdir -p "$output_dir"
    if [[ ! -d "$output_dir" ]]; then
        echo "Error: Could not create output directory for: $OUTPUT_FILE" >&2
        exit 1
    fi
    echo "$RESULTS" >"$OUTPUT_FILE"

    if [[ "$SILENT_MODE" = false ]]; then
        printf "[+] Total of %s unique domains found.\n" "$(echo "$RESULTS" | wc -l)" >&2
        printf "[+] Results saved to %s\n" "$OUTPUT_FILE" >&2
    fi
else
    echo "$RESULTS"
fi
