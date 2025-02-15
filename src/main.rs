use clap::Parser;

mod cli;
mod utils;
mod commands;
mod settings;
mod core;
pub mod error;

#[tokio::main]
async fn main() {
    let cli = cli::Cli::parse();
    if let Err(e) = utils::logging::setup_logging(&log_config) {
         eprintln!("Failed to initialize logging: {e}");
        std::process::exit(1);
    }
    if let Err(e) = cli::execute(cli).await {
        tracing::error!(error = ?e, "Application error");
        std::process::exit(1);
    }
}