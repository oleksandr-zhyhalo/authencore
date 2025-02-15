use crate::error::Result;
use crate::error::{ConfigError, Error};
use crate::settings::Profile;
use reqwest::{Certificate, Client, Identity};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::Path;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CredentialsResponse {
    pub credentials: AwsCredentials,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
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
    let load_pem = |path: &Path| -> Result<Vec<u8>> {
        fs::read(path).map_err(|e| {
            Error::Config(ConfigError::FileNotFound {
                file: path.to_path_buf(),
                description: format!("Failed to read file: {e}"),
            })
        })
    };

    let ca = load_pem(&profile.ca_path)?;
    let cert = load_pem(&profile.cert_path)?;
    let key = load_pem(&profile.key_path)?;

    let identity = Identity::from_pem(
        &[cert.as_slice(), key.as_slice()].concat()
    )?;
    let ca_cert = Certificate::from_pem(&ca)?;

    Client::builder()
        .add_root_certificate(ca_cert)
        .identity(identity)
        .build()
        .map_err(Into::into)
}

pub async fn fetch_credentials(profile: &Profile, client: &Client) -> Result<CredentialsResponse> {
    let url = format!(
        "https://{}/role-aliases/{}/credentials",
        profile.aws_iot_endpoint, profile.role_alias
    );

    let response = client.get(&url).send().await?.error_for_status()?;

    response.json().await.map_err(Into::into)
}

pub fn format_credentials(creds: &CredentialsResponse) -> Result<()> {
    let output = serde_json::json!({
        "Version": 1,
        "AccessKeyId": creds.credentials.access_key_id,
        "SecretAccessKey": creds.credentials.secret_access_key,
        "SessionToken": creds.credentials.session_token,
        "Expiration": creds.credentials.expiration
    });

    println!("{}", serde_json::to_string(&output)?);
    Ok(())
}
