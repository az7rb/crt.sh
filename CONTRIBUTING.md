# Contributing to crt.sh

Thanks for your interest in contributing.

## Getting Started

```bash
git clone https://github.com/az7rb/crt.sh
cd crt.sh
go build -o crt.sh .
```

## How to Contribute

### Reporting Bugs
Open an issue using the **Bug Report** template. Include:
- Command you ran
- Expected vs actual output
- OS and Go version

### Suggesting Features
Open an issue using the **Feature Request** template.

### Pull Requests
1. Fork the repo
2. Create a branch: `git checkout -b feature/your-feature`
3. Make your changes and test them
4. Submit a PR with a clear description of what changed and why

## Adding a New Source

To propose a new CT log source, open an issue first. Include:
- API endpoint
- Response format
- Rate limits
- Evidence it works (curl output)

Sources must be free and require no API key to be considered.

## Code Style

Follow standard Go conventions. Run `go vet ./...` before submitting.
