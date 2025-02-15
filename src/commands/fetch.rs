use crate::{core::aws};
use crate::error::Result;
use crate::settings::AppConfig;

pub async fn execute_fetch(force: bool) -> Result<()> {
    let config = AppConfig::load()?;
    let profile = config.active_profile()?;

    let client = aws::create_mtls_client(profile).await?;

    if force {
        let credentials = aws::fetch_credentials(&profile, &client).await?;
        aws::format_credentials(&credentials)
    } else {
        match aws::get_cached_credentials()? {
            Some(creds) if !aws::credentials_expired(&creds) => {
                aws::format_credentials(&creds)
            }
            _ => {
                let credentials = aws::fetch_credentials(&profile, &client).await?;
                aws::cache_credentials(&credentials)?;
                aws::format_credentials(&credentials)
            }
        }
    }
}