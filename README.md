# Authencore

<div align="center">

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/oleksandr-zhyhalo/authencore)](https://github.com/oleksandr-zhyhalo/authencore/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/oleksandr-zhyhalo/authencore)](https://golang.org/doc/devel/release)

Secure credential management utility for AWS IoT authentication.

</div>

## Overview

Authencore is a CLI utility that handles AWS credential retrieval and rotation for IoT devices using mutual TLS authentication. It implements the AWS credential process specification for seamless integration with AWS CLI tools.

## Features

- Mutual TLS authentication with AWS IoT Core
- Credential caching with configurable validity period
- Multi-environment configuration support
- AWS CLI credential process integration
- Cross-platform compatibility (Linux/macOS)

## Installation

### Automated Installation

```bash
curl -sSL https://raw.githubusercontent.com/oleksandr-zhyhalo/authencore/main/install.sh | sudo bash
```
### Manual Installation
1. Download the appropriate binary from Releases page
2. Extract and install:
```bash
tar xzf authencore-<os>-<arch>.tar.gz
sudo install -m 755 authencore /usr/local/bin/
sudo mkdir -p /etc/authencore /var/log/authencore
sudo chown ${USER}:$(id -gn) /etc/authencore /var/log/authencore
sudo chmod 750 /etc/authencore /var/log/authencore
```
## Configuration
### AWS IoT Requirements
1. Provision IoT Thing certificates 
2. Create IAM Role Alias 
3. Configure certificate policies in AWS IoT Core

Reference: AWS IoT Credential Provider Documentation

### Configuration File
Create `/etc/authencore/authencore.yml`:

```yaml
current_env: dev

dev:
    iot_endpoint: "dev-ats.iot.us-west-2.amazonaws.com"
    role_alias: "dev-role-alias"
    cert_path: "/etc/authencore/certs/device.crt"
    key_path: "/etc/authencore/certs/device.key"
    ca_path: "/etc/authencore/certs/root-ca.pem"
    cache_buffer_minutes: 5
    max_retries: 3
    retry_delay_ms: 1000

prod:
    iot_endpoint: "prod-ats.iot.us-west-2.amazonaws.com"
    role_alias: "prod-role-alias"
    cert_path: "/etc/authencore/certs/prod-device.crt"
    key_path: "/etc/authencore/certs/prod-device.key"
    ca_path: "/etc/authencore/certs/prod-root-ca.pem"
    cache_buffer_minutes: 5
    max_retries: 3
    retry_delay_ms: 1000
```
### AWS CLI Integration
Add to your AWS CLI configuration (`~/.aws/config`):

```ini
[profile iot-device]
credential_process = /usr/local/bin/authencore
region = us-west-2
```
### File Structure
```
/etc/authencore/
├── authencore.yml         # Main configuration
└── authencore.yml.sample  # Example configuration

/var/log/authencore/
└── error.log             # Application error log

~/.cache/authencore/
└── cache.json            # Cached credentials
```
### Verification
Test credential retrieval:

```bash
AWS_PROFILE=iot-device aws sts get-caller-identity
```
Inspect logs:

```bash
sudo tail -f /var/log/authencore/error.log
```
### License
Distributed under the MIT License. See LICENSE for details.

<div align="center"> <sub>Report issues via <a href="https://github.com/oleksandr-zhyhalo/authencore/issues">GitHub Issues</a></sub> </div>
