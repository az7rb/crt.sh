#!/bin/bash

# Display banner
echo "
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      ..| search crt.sh v 2.0 |..    |
+   site : crt.sh Certificate Search  +
|            Twitter: az7rb           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
"

# Function: Help
# Purpose: Display the help message with usage instructions.
Help() {
    echo "Options:"
    echo ""
    echo "-h     Help"
    echo "-d     Search Domain Name       | Example: $0 -d hackerone.com"
    echo "-o     Search Organization Name | Example: $0 -o hackerone+inc"
    echo ""
}

# Function: CleanResults
# Purpose: Clean and filter the results by removing unwanted characters and duplicates.
# - Converts escaped newlines to actual newlines.
# - Removes wildcard characters (*).
# - Filters out email addresses.
# - Sorts the results and removes duplicates.
CleanResults() {
    sed 's/\\n/\n/g' | \
    sed 's/\*.//g' | \
    sed -r 's/([A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,4})//g' | \
    sort | uniq
}

# Function: Domain
# Purpose: Search for certificates associated with a specific domain name.
# - Sends a request to crt.sh with the specified domain.
# - Processes the JSON response to extract common names and domain values.
# - Cleans and filters the results using CleanResults function.
# - Saves the results to an output file and displays them.
Domain() {
    # Check if the domain name is provided
    if [ -z "$req" ]; then
        echo "Error: Domain name is required."
        exit 1
    fi
    
    # Perform the search request to crt.sh
    response=$(curl -s "https://crt.sh?q=%.$req&output=json")
    
    # Check if the response is empty
    if [ -z "$response" ]; then
        echo "No results found for domain $req"
        exit 1
    fi
    
    # Process the response, clean it, and store the results
    results=$(echo "$response" | jq -r ".[].common_name,.[].name_value" | CleanResults)
    
    # Check if there are any valid results after cleaning
    if [ -z "$results" ]; then
        echo "No valid results found."
        exit 1
    fi
    
    # Define the output file name based on the domain
    output_file="output/domain.$req.txt"
    
    # Save the results to the output file
    echo "$results" > "$output_file"
    
    # Display the results and summary
    echo ""
    echo "$results"
    echo ""
    echo -e "\e[32m[+]\e[0m Total Save will be \e[31m$(echo "$results" | wc -l)\e[0m Domain only"
    echo -e "\e[32m[+]\e[0m Output saved in $output_file"
}

# Function: Organization
# Purpose: Search for certificates associated with a specific organization name.
# - Sends a request to crt.sh with the specified organization name.
# - Processes the JSON response to extract common names.
# - Cleans and filters the results using CleanResults function.
# - Saves the results to an output file and displays them.
Organization() {
    # Check if the organization name is provided
    if [ -z "$req" ]; then
        echo "Error: Organization name is required."
        exit 1
    fi
    
    # Perform the search request to crt.sh
    response=$(curl -s "https://crt.sh?q=$req&output=json")
    
    # Check if the response is empty
    if [ -z "$response" ]; then
        echo "No results found for organization $req"
        exit 1
    fi
    
    # Process the response, clean it, and store the results
    results=$(echo "$response" | jq -r ".[].common_name" | CleanResults)
    
    # Check if there are any valid results after cleaning
    if [ -z "$results" ]; then
        echo "No valid results found."
        exit 1
    fi
    
    # Define the output file name based on the organization name
    output_file="output/org.$req.txt"
    
    # Save the results to the output file
    echo "$results" > "$output_file"
    
    # Display the results and summary
    echo ""
    echo "$results"
    echo ""
    echo -e "\e[32m[+]\e[0m Total Save will be \e[31m$(echo "$results" | wc -l)\e[0m Domain only"
    echo -e "\e[32m[+]\e[0m Output saved in $output_file"
}

# Main Script Logic

# If no arguments are provided, display the help message
if [ -z "$1" ]; then
    Help
    exit
fi

# Parse command-line options using getopts
while getopts "h:d:o:" option; do
    case $option in
        h) # Display help
            Help
            ;;
        d) # Search for domain name
            req=$OPTARG
            Domain
            ;;
        o) # Search for organization name
            req=$OPTARG
            Organization
            ;;
        *) # Invalid option, display help
            Help
            ;;
    esac
done
