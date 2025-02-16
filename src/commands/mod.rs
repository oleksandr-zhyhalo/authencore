pub mod fetch;

use crate::cli::{Commands, Shell};
use crate::error::Result;

pub async fn handle_command(command: Commands) -> Result<()> {
    match command {
        Commands::Fetch { force } => fetch::execute_fetch(force).await,
        // Commands::Configure { action } => configure::execute(action).await,
        // Commands::Cache { action } => cache::execute(action).await,
        // Commands::Completions { shell } => completions::execute(shell),
        _ => {
            Ok(())
        }
    }
}