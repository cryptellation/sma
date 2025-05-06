package configs

import "github.com/spf13/viper"

// EnvSQLDSN is the environment variable name for the database DSN in the config.
const EnvSQLDSN = "SQL_DSN"

// EnvBinanceAPIKey is the environment variable name for the Binance API key in the config.
const EnvBinanceAPIKey = "BINANCE_API_KEY"

// EnvBinanceSecretKey is the environment variable name for the Binance secret key in the config.
const EnvBinanceSecretKey = "BINANCE_SECRET_KEY"

// EnvTemporalAddress is the environment variable name for the Temporal address in the config.
const EnvTemporalAddress = "TEMPORAL_ADDRESS"

// EnvHealthAddress is the environment variable name for the health address in the config.
const EnvHealthAddress = "HEALTH_ADDRESS"

func init() {
	// Tell viper to read environment variables
	viper.AutomaticEnv()

	// Set default values for the config
	viper.SetDefault(EnvSQLDSN, DefaultDBDSN)
	viper.SetDefault(EnvBinanceAPIKey, DefaultBinanceAPIKey)
	viper.SetDefault(EnvBinanceSecretKey, DefaultBinanceSecretKey)
	viper.SetDefault(EnvTemporalAddress, DefaultTemporalAddress)
	viper.SetDefault(EnvHealthAddress, DefaultHealthAddress)
}
