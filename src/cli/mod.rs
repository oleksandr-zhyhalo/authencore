use clap::{Parser, Subcommand, ValueEnum};
#[derive(Parser, Debug)]
#[command(
    author,
    version,
    about,
    long_about = "AWS IoT Credential Provider with mTLS authentication"
)]

pub struct Cli {
    pub command: Commands,
    pub profile: Option<String>,
    pub verbose: u8,
}

#[derive(Debug, Subcommand)]
pub enum Commands {
    Fetch {
        #[arg(short, long)]
        force: bool,
    },
    Configure {
        action: ConfigActions,
    },
    Cache {
        #[command(subcommand)]
        action: CacheActions,
    },
    Completions {
        #[arg(value_enum)]
        shell: Shell,
    },
}

#[derive(Subcommand, Debug)]
pub(crate) enum CacheActions {
    Clear,
}
#[derive(Subcommand, Debug)]
pub enum ConfigActions {
    Validate,
    SetEnv,
    List,
}
#[derive(Debug, Clone, ValueEnum)]
pub enum Shell {
    Bash,
    Zsh,
    Fish,
    PowerShell,
}
