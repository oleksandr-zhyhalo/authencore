use std::fs;
use std::path::PathBuf;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use crate::core::aws::AwsCredentials;
#[derive(Debug, Serialize, Deserialize)]
struct CachedCredentials {
    credentials: AwsCredentials,
    expiration: DateTime<Utc>,
}
pub fn get_cached_credentials() -> Result<Option<AwsCredentials>> {
    let cache_path = get_cache_path()?;

    if !cache_path.exists() {
        return Ok(None);
    }

    let content = fs::read_to_string(&cache_path)
        .map_err(|e| Error::Cache(format!("Failed to read cache file: {}", e)))?;

    let cached: CachedCredentials = serde_json::from_str(&content)
        .map_err(|e| Error::Cache(format!("Invalid cache format: {}", e)))?;

    Ok(Some(cached.credentials))
}
pub fn credentials_expired(creds: &AwsCredentials) -> bool {
    match DateTime::parse_from_rfc3339(&creds.expiration) {
        Ok(expiration) => Utc::now() > expiration.with_timezone(&Utc),
        Err(_) => true,
    }
}
fn get_cache_path() -> Result<PathBuf> {
    let base_dir = if cfg!(target_os = "linux") {
        PathBuf::from("/var/cache/authencore")
    } else {
        dirs::cache_dir()
            .ok_or_else(|| Error::Cache("Could not determine cache directory".to_string()))?
            .join("authencore")
    };

    Ok(base_dir.join("credentials.json"))
}