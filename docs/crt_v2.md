# crt.sh Documentation

## Table of Contents
- [**crt_v2.sh**](#crt_v2sh-crt_v2sh)
    <ul>
      <li><a href="#crt_v2sh-crt_v2sh-show_help">show_help()</a></li>
      <li><a href="#crt_v2sh-crt_v2sh-clean_results">clean_results()</a></li>
      <li><a href="#crt_v2sh-crt_v2sh-domain_search">domain_search()</a></li>
      <li><a href="#crt_v2sh-crt_v2sh-org_search">org_search()</a></li>
    </ul>


<h2 id="crt_v2sh-crt_v2sh">Module: crt_v2.sh</h2>

**Author:** az7rb, Arqsz

### Description

This script queries the crt.sh certificate transparency log to retrieve domain names associated with specific domains or organizations.

It processes the raw certificate data, cleaning and filtering it to provide unique and relevant domain results.



Key features include:

- Searching for domain names and subdomains.

- Searching for domains associated with specific organizations.

- Cleaning and de-duplicating search results.

- Outputting results to standard output or a specified file.



## Functions


<h4 id="crt_v2sh-crt_v2sh-show_help"><code>show_help</code></h4>


Displays the help message for the script, detailing its usage,
available options, and examples.





#### Example:

```bash
show_help
```




<h4 id="crt_v2sh-crt_v2sh-clean_results"><code>clean_results</code></h4>


Cleans and filters input data, typically piped to this function,
by removing unwanted characters and duplicates. It converts escaped
newlines to actual newlines, removes wildcard characters (*), filters out
email addresses and lines containing spaces, and ensures entries contain a dot.
Finally, it sorts the results and removes duplicates.





#### Example:

```bash
echo "example.com\\n*.test.com\\nuser@example.com" | clean_results
```



#### Returns:

<p>{string}
A multi-line string of cleaned, sorted, and unique domain names.
</p>



<h4 id="crt_v2sh-crt_v2sh-domain_search"><code>domain_search</code></h4>


Searches the crt.sh certificate transparency log for certificates
associated with a given domain name. It fetches common names and
name values from the crt.sh JSON API, then cleans and filters the results.



#### Arguments

| Name | Type | Description |
|------|-------------|-------------|
| `$1` | *string* | The domain name to search for (e.g., "example.com"). |
| `$2` | *boolean* | A flag indicating whether silent mode is enabled. If `true`, suppresses status messages printed to standard error. |




#### Example:

```bash
domain_search "hackerone.com" false
```



#### Returns:

<p>{string}
A multi-line string of unique and cleaned domain names found for the specified domain.
Returns an empty string if no results are found or the request fails.
</p>



<h4 id="crt_v2sh-crt_v2sh-org_search"><code>org_search</code></h4>


Searches the crt.sh certificate transparency log for certificates
associated with a given organization name. It fetches common names and
name values from the crt.sh JSON API, then cleans and filters the results.



#### Arguments

| Name | Type | Description |
|------|-------------|-------------|
| `$1` | *string* | The organization name to search for (e.g., "HackerOne, Inc."). This argument should be URL-encoded if it contains special characters. |
| `$2` | *boolean* | A flag indicating whether silent mode is enabled. If `true`, suppresses status messages printed to standard error. |




#### Example:

```bash
org_search "HackerOne%2C%20Inc." false
```



#### Returns:

<p>{string}
A multi-line string of unique and cleaned domain names found for the specified organization.
Returns an empty string if no results are found or the request fails.
</p>


