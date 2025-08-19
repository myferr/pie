# pie

[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![Install with Go](https://img.shields.io/badge/Install-go%20install-brightgreen)](https://golang.org/doc/install)
[![Docker Required](https://img.shields.io/badge/Docker-Required-blue?logo=docker&logoColor=white)](https://www.docker.com/)
[![Python Compatible](https://img.shields.io/badge/Python-3.10%2B-yellow?logo=python&logoColor=white)](https://www.python.org/)
[![License MIT](https://img.shields.io/badge/License-MIT-purple)](LICENSE)

## What is pie?

**pie** is a fun and lightweight CLI tool that makes Python project setup **easy, reproducible, and isolated**.  
It runs your Python scripts in **Docker containers**, manages dependencies automatically, and ensures your projects don’t break due to conflicting environments.

## Features

- Runs Python scripts in **isolated Docker environments**
- Supports **multiple Python versions** using Docker images
- Easy CLI with `init`, YAML-based *optional* configuration, and Docker-based containerization
- Perfect for Python projects of all sizes  

## Installation

Requires [Go 1.18+](https://golang.org/dl/) and [Docker](https://www.docker.com/).

```bash
# Install Pie globally
go install github.com/myferr/pie@latest
```

## Usage

```bash
pie -h
pie [file] [flags]
pie [command]
```

### Commands

* `init` – Initialize a new Pie project
* `help` – Show help about any command
* `completion` – Generate autocompletion script (Cobra built-in)

### Flags

* `-d, --dependencies string` – Comma-separated list of dependencies
* `--dispose` – Remove the container after it runs
* `-f, --file string` – Specify a custom `pie.yml` (default `"pie.yml"`)
* `-h, --help` – Show help for Pie
* `-p, --python-version string` – Specify the Python version
* `-v, --verbose` – Show full Docker build output

---

## Examples

```bash
# Run a Python script
pie test/main.py

# Specify importants without pie.yml
pie test/main.py -p 3.12 -d {colorama}
# Initialize a new Pie project
pie init
```

## License

[MIT License](LICENSE)
