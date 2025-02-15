mod fetch;
mod configure;
mod cache;
mod completions;
mod fetch;

pub use fetch::execute_fetch;
pub use configure::execute_configure;
pub use cache::execute_cache;
pub use completions::execute_completions;

use crate::cli::{Commands, ConfigActions, CacheActions, Shell};
use crate::error::Result;

pub async fn handle_command(command: Commands) -> Result<()> {
    match command {
        Commands::Fetch { force } => execute_fetch(force).await,
        Commands::Configure { action } => execute_configure(action).await,
        Commands::Cache { action } => execute_cache(action).await,
        Commands::Completions { shell } => execute_completions(shell),
    }
}