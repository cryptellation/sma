package configs

const (
	// DefaultDBDSN is the default database DSN.
	DefaultDBDSN = "host=localhost " +
		"user=cryptellation_sma " +
		"password=cryptellation_sma " +
		"dbname=cryptellation_sma " +
		"sslmode=disable"

	// DefaultBinanceAPIKey is the default Binance API key.
	DefaultBinanceAPIKey = ""

	// DefaultBinanceSecretKey is the default Binance secret key.
	DefaultBinanceSecretKey = ""

	// DefaultTemporalAddress is the default Temporal address.
	DefaultTemporalAddress = "localhost:7233"

	// DefaultHealthAddress is the default health address.
	DefaultHealthAddress = ":9000"
)
