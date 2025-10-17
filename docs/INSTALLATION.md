# CQLAI Installation Guide

CQLAI can be installed through various package managers, Docker, or built from source. Choose the method that best suits your environment.

## Table of Contents
- [Package Manager Installation](#package-manager-installation)
  - [APT (Debian/Ubuntu)](#apt-debianubuntu)
  - [YUM/DNF (RHEL/CentOS/Fedora)](#yumdnf-rhelcentosfedora)
- [Docker Installation](#docker-installation)
- [Binary Installation](#binary-installation)
- [Building from Source](#building-from-source)
- [Verifying Installation](#verifying-installation)
- [Configuration](#configuration)

## Package Manager Installation

### APT (Debian/Ubuntu)

Add the CQLAI repository and install:

```bash
# Add the repository key
curl -fsSL https://packages.axonops.com/apt/KEY.gpg | sudo gpg --dearmor -o /usr/share/keyrings/axonops-archive-keyring.gpg

# Add the repository
echo "deb [signed-by=/usr/share/keyrings/axonops-archive-keyring.gpg] https://packages.axonops.com/apt stable main" | sudo tee /etc/apt/sources.list.d/cqlai.list

# Update package list
sudo apt update

# Install CQLAI
sudo apt install cqlai
```

#### Specific Version Installation
```bash
# List available versions
apt list -a cqlai

# Install specific version
sudo apt install cqlai=0.0.5
```

#### Upgrade
```bash
sudo apt update && sudo apt upgrade cqlai
```

### YUM/DNF (RHEL/CentOS/Fedora)

Add the CQLAI repository and install:

```bash
# Add the repository
sudo tee /etc/yum.repos.d/cqlai.repo <<EOF
[cqlai]
name=CQLAI Repository
baseurl=https://packages.axonops.com/rpm/stable/\$basearch
enabled=1
gpgcheck=1
gpgkey=https://packages.axonops.com/rpm/KEY.gpg
EOF

# Install CQLAI
sudo yum install cqlai
# or for newer systems
sudo dnf install cqlai
```

#### Specific Version Installation
```bash
# List available versions
yum list available cqlai --showduplicates
# or
dnf list available cqlai --showduplicates

# Install specific version
sudo yum install cqlai-0.0.5
# or
sudo dnf install cqlai-0.0.5
```

#### Upgrade
```bash
sudo yum update cqlai
# or
sudo dnf upgrade cqlai
```

## Docker Installation

CQLAI is available as a Docker image for containerized deployments.

### Pull the Latest Image
```bash
docker pull registry.axonops.com/axonops-public/axonops-docker/cqlai:latest
```

### Pull Specific Version
```bash
docker pull registry.axonops.com/axonops-public/axonops-docker/cqlai:0.0.5
```

### Running CQLAI in Docker

#### Interactive Mode
```bash
# Connect to Cassandra on host network
docker run -it --rm --network host \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest \
  -h localhost -p 9042 -u cassandra

# With custom configuration
docker run -it --rm \
  -v $(pwd)/cqlai.json:/app/cqlai.json:ro \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest
```

#### Batch Mode
```bash
# Execute a single query
docker run --rm --network host \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest \
  -h localhost -u cassandra \
  -e "SELECT * FROM system.local;"

# Execute queries from file
docker run --rm \
  -v $(pwd)/queries.cql:/queries.cql:ro \
  --network host \
  registry.axonops.com/axonops-public/axonops-docker/cqlai:latest \
  -h localhost -u cassandra \
  -f /queries.cql
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  cqlai:
    image: registry.axonops.com/axonops-public/axonops-docker/cqlai:latest
    container_name: cqlai
    network_mode: host
    volumes:
      - ./cqlai.json:/app/cqlai.json:ro
      - ./queries:/queries:ro
    environment:
      - CQLAI_HOST=localhost
      - CQLAI_PORT=9042
      - CQLAI_USERNAME=cassandra
    stdin_open: true
    tty: true
```

## Binary Installation

Download pre-built binaries from GitHub releases:

### Linux (amd64)
```bash
# Download latest release
curl -L https://github.com/axonops/cqlai/releases/latest/download/cqlai-linux-amd64 -o cqlai

# Make executable
chmod +x cqlai

# Move to PATH (optional)
sudo mv cqlai /usr/local/bin/
```

### macOS (Intel)
```bash
# Download latest release
curl -L https://github.com/axonops/cqlai/releases/latest/download/cqlai-darwin-amd64 -o cqlai

# Make executable
chmod +x cqlai

# Move to PATH (optional)
sudo mv cqlai /usr/local/bin/
```

### macOS (Apple Silicon)
```bash
# Download latest release
curl -L https://github.com/axonops/cqlai/releases/latest/download/cqlai-darwin-arm64 -o cqlai

# Make executable
chmod +x cqlai

# Move to PATH (optional)
sudo mv cqlai /usr/local/bin/
```

### Windows
Download the executable from:
```
https://github.com/axonops/cqlai/releases/latest/download/cqlai-windows-amd64.exe
```

## Building from Source

### Prerequisites
- Go 1.21 or later
- Git

### Build Steps
```bash
# Clone the repository
git clone https://github.com/axonops/cqlai.git
cd cqlai

# Download dependencies
go mod download

# Build the binary
go build -o cqlai cmd/cqlai/main.go

# Install to PATH (optional)
sudo cp cqlai /usr/local/bin/
```

### Build with Specific Version
```bash
# Set version during build
go build -ldflags "-X main.Version=0.0.5" -o cqlai cmd/cqlai/main.go
```

## Verifying Installation

After installation, verify CQLAI is working:

```bash
# Check version
cqlai --version

# Test connection to Cassandra
cqlai -h localhost -p 9042 -u cassandra -e "SELECT release_version FROM system.local;"
```

## Configuration

### Quick Start
```bash
# Create default configuration
cqlai --generate-config > cqlai.json

# Edit configuration
nano cqlai.json
```

### Configuration Locations
CQLAI looks for configuration in the following order:
1. `./cqlai.json` (current directory)
2. `~/.cqlai.json` (user home directory)
3. `~/.config/cqlai/config.json` (XDG config directory)

### Environment Variables
Configuration can also be set via environment variables:
```bash
export CQLAI_HOST=localhost
export CQLAI_PORT=9042
export CQLAI_USERNAME=cassandra
export CQLAI_PASSWORD=cassandra
export CQLAI_KEYSPACE=system
```

## Troubleshooting

### Common Issues

#### Permission Denied
If you get permission errors when installing:
```bash
# Ensure you have sudo privileges
sudo apt install cqlai
```

#### Repository Not Found
If package repositories are not found:
```bash
# Update repository cache
sudo apt update
# or
sudo yum makecache
```

#### Docker Network Issues
If Docker cannot connect to Cassandra:
```bash
# Use host network mode
docker run --network host ...

# Or specify Cassandra container name
docker run --link cassandra:cassandra ...
```

## Support

For issues and support:
- GitHub Issues: https://github.com/axonops/cqlai/issues
- Documentation: https://github.com/axonops/cqlai/tree/main/docs
- Release Notes: https://github.com/axonops/cqlai/releases

## License

CQLAI is distributed under the Apache License 2.0. See LICENSE file for details.