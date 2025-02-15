use crate::error::{ConfigError, Error, Result};
use serde::Deserialize;
use std::path::{Path, PathBuf};

#[derive(Debug, Deserialize)]
pub struct AppConfig {
    pub active_profile: String,
    pub profiles: Vec<Profile>
}

impl AppConfig {
    pub fn active_profile(&self) -> Result<&Profile> {
        self.profiles
            .iter()
            .find(|p| p.name == self.active_profile)
            .ok_or_else(|| Error::Config(ConfigError::MissingEnvironment(
                format!("Active profile '{}' not found in configuration", self.active_profile)
            )))    }
}

#[derive(Debug, Deserialize, Clone)]
pub struct Profile {
    pub name: String,
    pub aws_iot_endpoint: String,
    pub role_alias: String,
    pub ca_path: PathBuf,
    pub cert_path: PathBuf,
    pub key_path: PathBuf,
}

impl AppConfig {
    pub fn load() -> Result<Self> {
        let config_path = Self::find_config_file()?;

        let settings = config::Config::builder()
            .add_source(config::File::from(config_path))
            .build()
            .map_err(|e| Error::Config(ConfigError::LoadError(e.to_string())))?;

        settings
            .try_deserialize()
            .map_err(|e| Error::Config(ConfigError::LoadError(e.to_string())))
    }
    pub fn validate(&self) -> Result<()> {
        if self.active_profile != "" {
            Ok(println!("TODO Validation"));
        } Err("Profile does not exist")?
    }
    pub fn list_profiles(&self) {
        for profile in &self.profiles {
            println!("Profile: {}", profile.name);
        }
    }
    fn find_config_file() -> Result<PathBuf> {
        let paths = ["/etc/authencore/authencore.toml", "./authencore.toml"];

        for path in &paths {
            let path = PathBuf::from(path);
            if path.exists() {
                return Ok(path);
            }
        }

        Err(Error::Config(ConfigError::LoadError(
            "No configuration file found in default locations".to_string(),
        )))
    }

    pub fn validate_paths(&self) -> Result<()> {
        let profile = self.active_profile()?;

        let validate = |path: &Path, desc: &str| {
            if !path.exists() {
                Err(Error::Config(ConfigError::FileNotFound {
                    file: path.to_path_buf(),
                    description: desc.to_string(),
                }))
            } else {
                Ok(())
            }
        };

        validate(&profile.cert_path, "Client certificate")?;
        validate(&profile.key_path, "Private key")?;
        validate(&profile.ca_path, "CA certificate")?;

        Ok(())
    }
}

pub mod logging {
    use crate::error::{Error, Result};
    use tracing_subscriber::{fmt, EnvFilter};

    pub fn setup_logging(verbosity: u8) -> Result<()> {
        let filter = match verbosity {
            0 => "info",
            1 => "debug",
            _ => "trace",
        };

        tracing_subscriber::fmt()
            .with_env_filter(EnvFilter::new(filter))
            .with_writer(std::io::stderr)
            .try_init()
            .map_err(|e| Error::Logging(e.to_string()))?;

        Ok(())
    }
}