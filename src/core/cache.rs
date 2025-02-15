use crate::error::{Error, Result};
use serde::{Deserialize, Serialize};
use std::{fs, path::PathBuf};
use chrono::{DateTime, Utc};
use crate::core::aws::CredentialsResponse;

#[derive(Debug, Serialize, Deserialize)]
struct CachedCredentials {
    credentials: super::aws::CredentialsResponse,
    expiration: DateTime<Utc>,
}

pub fn get_cached() -> Result<Option<super::aws::CredentialsResponse>> {
    let path = cache_path()?;
    if !path.exists() {
        return Ok(None);
    }

    let data = fs::read_to_string(&path)?;
    let cached: CachedCredentials = serde_json::from_str(&data)?;

    if Utc::now() > cached.expiration {
        Ok(None)
    } else {
        Ok(Some(cached.credentials))
    }
}

pub fn store(creds: &CredentialsResponse) -> Result<()> {
    let path = cache_path()?;
    let dir = path.parent().ok_or(Error::Cache("Invalid path".into()))?;

    let expiration = DateTime::parse_from_rfc3339(&creds.credentials.expiration)
        .map(|dt| dt.with_timezone(&Utc))
        .map_err(|_| Error::CredentialsFormat)?;

    let cached = CachedCredentials {
        credentials: **creds.clone(),
        expiration,
    };

    fs::create_dir_all(dir)?;
    let data = serde_json::to_string(&cached)?;
    fs::write(path, data)?;

    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mut perms = fs::metadata(&path)?.permissions();
        perms.set_mode(0o600);
        fs::set_permissions(&path, perms)?;
    }

    Ok(())
}

fn cache_path() -> Result<PathBuf> {
    let base = dirs::cache_dir()
        .ok_or(Error::Cache("Failed to find cache directory".into()))?
        .join("authencore");

    Ok(base.join("credentials.json"))
}