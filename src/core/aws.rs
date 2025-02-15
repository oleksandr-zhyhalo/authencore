use std::fs;
use std::fs::OpenOptions;
use std::path::Path;
use std::time::Duration;
use super::super::error::Result;
use super::super::settings::Profile;
use reqwest::{Certificate, Client, Identity};
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
pub struct AwsCredentials {
    #[serde(rename = "accessKeyId")]
    pub access_key_id: String,
    #[serde(rename = "secretAccessKey")]
    pub secret_access_key: String,
    #[serde(rename = "sessionToken")]
    pub session_token: String,
    pub expiration: String,
}

pub async fn create_mtls_client(profile: &Profile) -> Result<Client> {
    let ca_cert = load_pem(&profile.ca_path).map_err(|e| Error::LoadCaCert {
        path: profile.ca_path.clone(),
        source: e,
    })?;

    let client_cert = load_pem(&profile.cert_path).map_err(|e| Error::LoadClientCert {
        path: profile.cert_path.clone(),
        source: e,
    })?;

    let client_key = load_pem(&profile.key_path).map_err(|e| Error::LoadPrivateKey {
        path: profile.key_path.clone(),
        source: e,
    })?;

    let identity = Identity::from_pem(&[client_cert, client_key].concat())
        .map_err(|e| Error::HttpClient(e))?;

    let ca_cert = Certificate::from_pem(&ca_cert).map_err(|e| Error::HttpClient(e))?;

    Client::builder()
        .use_rustls_tls()
        .add_root_certificate(ca_cert)
        .identity(identity)
        .timeout(Duration::from_secs(15))
        .build()
        .map_err(Error::HttpClient)
}

pub async fn fetch_credentials(
    profile: &Profile,
    client: &Client,
) -> Result<AwsCredentials> {
    let url = format!(
        "https://{}/role-aliases/{}/credentials",
        profile.aws_iot_endpoint, profile.role_alias
    );

    let max_attempts = 3;
    let mut attempts = 0;
    let mut delay = Duration::from_secs(1);

    loop {
        attempts += 1;
        match client.get(&url).send().await {
            Ok(response) => {
                if response.status().is_success() {
                    return response
                        .json::<AwsCredentials>()
                        .await
                        .map_err(Error::HttpClient);
                }

                return Err(Error::CredentialsRequest {
                    url: url.clone(),
                    status: response.status(),
                });
            }
            Err(e) => {
                tracing::error!(error = ?e, url = %url, "Failed to send credentials request");

                if attempts >= max_attempts {
                    return Err(Error::HttpClient(e));
                }

                tracing::info!("Retrying AWS credentials request in {:?}", delay);
                tokio::time::sleep(delay).await;
                delay *= 2;
            }
        }
    }
}

pub async  fn format_credentials(aws_credentials: &AwsCredentials) -> Result<()> {
    let output = serde_json::json!({
        "Version": 1,
        "AccessKeyId": aws_credentials.access_key_id,
        "SecretAccessKey": aws_credentials.secret_access_key,
        "SessionToken": aws_credentials.session_token,
        "Expiration": aws_credentials.expiration
    });

    // Print JSON to stdout
    println!("{}", serde_json::to_string(&output).map_err(Error::JsonParse)?);

    // Log success message to stderr
    tracing::info!("Successfully formatted credentials for output");
    Ok(())
}
pub async fn get_cached_credentials() -> Result<Option<AwsCredentials>> {
    let cache_path = Path::new("/var/cache/creds.json").to_path_buf();
    if !cache_path.exists() {
        return Ok(None);
    }

    let file = OpenOptions::new()
        .read(true)
        .open(&cache_path)
        .map_err(|e| {
            Error::Cache(format!(
                "Failed to open cache file {}: {}",
                cache_path.display(),
                e
            ))
        })?;

    file.lock_shared().map_err(|e| {
        Error::Cache(format!(
            "Failed to acquire shared lock on cache file: {}",
            e
        ))
    })?;

    let mut data = String::new();
    {
        use std::io::BufReader;
        let mut reader = BufReader::new(&file);
        reader
            .read_to_string(&mut data)
            .map_err(|e| Error::Cache(format!("Failed to read cache file: {}", e)))?;
    }

    file.unlock()
        .map_err(|e| Error::Cache(format!("Failed to release lock on cache file: {}", e)))?;

    let creds = serde_json::from_str(&data).map_err(Error::JsonParse)?;

    Ok(Some(creds))
}

fn load_pem(path: &Path) -> std::io::Result<Vec<u8>> {
    fs::read(path)
}
