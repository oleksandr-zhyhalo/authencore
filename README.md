# authencore 🛡️

<div align="center">

![GitHub release (latest by date)](https://img.shields.io/github/v/release/oleksandr-zhyhalo/authencore)
![Rust Version](https://img.shields.io/badge/rust-1.70%2B-blue.svg)

A secure credentials manager for AWS IoT devices with local caching and multi-environments setup.

[Installation](#Installation) •
[Features](#features) •
[Usage](#usage) •
[Configuration](#configuration) •
[Contributing](#contributing)

</div>

## ✨ Features

- 🔐 **Secure mTLS Authentication**: Uses device certificates for AWS IoT authentication
- 🔄 **Automatic Credential Management**: Handles AWS credential rotation
- 📦 **AWS CLI Integration**: Works seamlessly with AWS CLI credential_process
- 💻 **Cross-Platform**: Statically linked binary works on any Linux system
- 💾 **Credential Caching**: Local caching mechanism for AWS credentials to reduce unnecessary network calls

## 🚀 Installation

### Using Install Script (Recommended)

```bash
curl -o- https://raw.githubusercontent.com/oleksandr-zhyhalo/authencore/main/install.sh | sudo bash
# or with wget
wget -qO- https://raw.githubusercontent.com/oleksandr-zhyhalo/authencore/main/install.sh | sudo bash
```

### Manual Installation

1. Download the latest release for your platform from [releases page](https://github.com/oleksandr-zhyhalo/authencore/releases)
2. Install manually:
```bash
# Extract the archive
tar xzf authencore-linux-*.tar.gz

# Create directories
sudo mkdir -p /etc/authencore /var/log/authencore

# Install binary
sudo install -m 755 authencore/authencore /usr/local/bin/

# Set up settings and logs (replace 'your-username' with your actual username)
sudo chown your-username:your-username /etc/authencore /var/log/authencore
sudo chmod 700 /etc/authencore /var/log/authencore

# Install settings if needed
sudo install -m 600 -o your-username authencore/authencore.toml.sample /etc/authencore/authencore.conf
```

## 📚 Configuration

### AWS IoT Setup

1. Create an IoT thing and download certificates
2. Create a role alias in AWS IoT
3. Attach appropriate policies to your certificates

For more details read:
[Authorizing direct calls to AWS services using AWS IoT Core credential provider
   ](https://docs.aws.amazon.com/iot/latest/developerguide/authorizing-direct-aws.html)

### Configuration File

Create or edit `/etc/authencore/authencore.toml`:
```toml
cache_dir = "/var/cache/authencore"
log_dir = "/var/log/authencore"

[environment]
current = "dev"

[environment.dev]
aws_iot_endpoint = "dev-ats.iot.us-west-2.amazonaws.com"
role_alias = "dev-role-alias"
cert_path = "/etc/authencore/dev/cert.pem"
key_path = "/etc/authencore/dev/key.pem"
ca_path = "/etc/authencore/dev/root-ca.pem"

[environment.prod]
aws_iot_endpoint = "prod-ats.iot.us-west-2.amazonaws.com"
role_alias = "prod-role-alias"
cert_path = "/etc/authencore/prod/cert.pem"
key_path = "/etc/authencore/prod/key.pem"
ca_path = "/etc/authencore/prod/root-ca.pem"
```

### AWS CLI Integration

Add to your AWS CLI config (`~/.aws/config`):
```ini
[profile your-profile]
credential_process = /usr/local/bin/authencore
```

## 📂 Directory Structure

```
/etc/authencore/
├── authencore.toml         # Main configuration
└── authencore.toml.sample  # Sample configuration

/var/log/authencore/
├── authencore.log         # Current application log
└── authencore.log.*      # Rotated log files

/var/cache/authencore/
├── creds_cache.json     # Cached credentials
└── cb_state.json       # Circuit breaker state

/usr/local/bin/
└── authencore            # Binary executable
```

## 🔨 Usage

### Testing Configuration

Verify your setup:
```bash
# Test credential retrieval
aws sts get-caller-identity --profile your-profile

# Check logs
tail -f /var/log/authencore/authencore.log
```

### Common Operations

```bash
# Direct credential retrieval
authencore

# With debug output
AWS_PROFILE=your-profile aws sts get-caller-identity --debug
```

## 📄 License

See the [LICENSE](LICENSE) file for details.

---

<div align="center">
Made with ❤️ for secure AWS IoT authentication

[Report Bug](https://github.com/oleksandr-zhyhalo/authencore/issues) • [Request Feature](https://github.com/oleksandr-zhyhalo/authencore/issues)
</div>
