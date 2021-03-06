# Tarp

Tarp generates HTML coverage reports from files generated by `go tool cover`. It is a replacement for the `go tool cover -html=` command.
Tarp generates a tree view which resembles the actual file structure of the project, allowing users to intuitively find the files they are looking for, instead of having to navigate the long dropdown of the traditional report.

## Installation

**Install via go**
`go install github.com/dylandreimerink/tarp/cmd/tarp@latest`

## Usage

```
Usage:
  tarp {coverage file 1} [coverage file N...] [flags]

Flags:
  -h, --help            help for tarp
  -o, --output string   The generated coverage report (default "./coverage.html")
```