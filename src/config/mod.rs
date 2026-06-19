use config::{Config, ConfigError};

pub fn create() -> Result<Config, ConfigError> {
    let file_path =
        std::env::var("BOT_CONFIG_PATH").unwrap_or_else(|_| "./config.yaml".to_string());

    Config::builder()
        .add_source(config::File::with_name(file_path.as_str()))
        .build()
}
