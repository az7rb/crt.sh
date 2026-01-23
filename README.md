### Updated Documentation for crt.sh v2.1 ([@TheArqsz](https://github.com/TheArqsz))

The logic of the script was changed a little bit so one can use it more freely in the automated environment and it aligns more with popular security tools.

Updated usage:

```bash
A script to query the crt.sh certificate transparency log.

Usage: ./crt_v2.1.sh [OPTIONS]

Options:
  -d, --domain <domain>        Search for a specific domain name (e.g., hackerone.com)
      --org <organization>     Search for a specific organization name (e.g., 'HackerOne, Inc.')
  -o, --output <file>          File to save results. If not set, results are printed to stdout.
  -s, --silent                 Suppress banner and non-essential output.
  -h, --help                   Display this help message

Examples:
  ./crt_v2.1.sh --domain hackerone.com
  ./crt_v2.1.sh --org 'HackerOne, Inc.' --output ./results.txt
  ./crt_v2.1.sh -d example.com -s | grep '.com'
  ./crt_v2.1.sh -d example.com -s | httpx
```

All the original credits go to [az7rb](https://github.com/az7rb)

### Updated Documentation for crt.sh v2.0

The **v2.0** version of the script, now named `crt_v2.sh`, introduces significant improvements in performance, reliability, and documentation. Below are the updated instructions for setting up and using both the original script `crt.sh` and the new `crt_v2.sh`:

---

## crt.sh and crt_v2.sh

These bash scripts are designed to simplify saving and parsing output from the [crt.sh](https://crt.sh) website, allowing for easy integration with tools like `httpx` for further analysis.

### Usage Overview

**crt.sh**: The original version of the script with basic functionality.

**crt_v2.sh**: The enhanced version (v2.0) with improved performance, better error handling, and comprehensive documentation.

### Installation and Setup

#### Step 1: Clone the Repository and Set Permissions

To install the scripts, clone the repository from GitHub and set the appropriate execution permissions.

```bash
git clone https://github.com/az7rb/crt.sh.git && cd crt.sh/
chmod +x crt.sh crt_v2.sh
```

#### Step 2: Display Help and Options

To view the usage options and get started with either script:

For the original script:

```bash
./crt.sh -h
```

For the updated script:

```bash
./crt_v2.sh -h
```

### Example Usage

**Original Script**:
```bash
./crt.sh -d hackerone.com | httpx
```

**Updated Script (v2.0)**:
```bash
./crt_v2.sh -d hackerone.com | httpx
```

Both commands will enumerate subdomains for `hackerone.com` and output them in a format ready to be piped into other tools like `httpx`.

### Notes:

- **crt_v2.sh** is the recommended version due to its enhanced features and reliability.
- **Output**: Both scripts will save the enumerated subdomains to a specified output file, making the data ready for further processing or use with other tools.

### Screenshot

For a quick visual guide, refer to the screenshot below:

![Help Screenshot](https://raw.githubusercontent.com/az7rb/crt.sh/main/Screenshot/Screenshot_Help.png)

---
### Additional Resources

Don't forget to visit [BugBountyzip](https://github.com/BugBountyzip) for more tools and resources.

Happy hunting! 🎯
