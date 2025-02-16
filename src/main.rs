use clap::Parser;
use crate::commands::handle_command;

mod cli;
mod commands;
mod settings;
mod core;
pub mod error;

#[tokio::main]
async fn main() {
    let cli = cli::Cli::parse();
    settings::logging::setup_logging(cli.verbose).expect("TODO: panic message");
    handle_command(cli.command).await.expect("Error: ");

}