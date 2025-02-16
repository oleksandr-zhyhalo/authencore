use crate::core::aws;
use crate::core::cache;
use crate::error::Result;
use crate::settings::AppConfig;

pub async fn execute_fetch(force: bool) -> Result<()> {
    let config = AppConfig::load()?;
    config.validate_paths()?;
    let profile = config.active_profile()?;

    let client = aws::create_mtls_client(profile).await?;

    let credentials = if force {
        let creds = aws::fetch_credentials(profile, &client).await?;
        cache::store(&creds)?;
        creds
    } else {
        match cache::get_cached()? {
            Some(cached) => cached,
            None => {
                let creds = aws::fetch_credentials(profile, &client).await?;
                cache::store(&creds)?;
                creds
            }
        }
    };

    aws::format_credentials(&credentials)
}