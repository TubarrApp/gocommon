package sharedconsts

// Environment variable strings.
const (
	// Allow arguments for FFmpeg (etc.) which are usually blocked for security.
	EnvAllowInsecureArguments = "ALLOW_INSECURE_ARGUMENTS"

	// Tubarr.
	EnvTubarrWebAddress      = "TUBARR_WEB_ADDRESS"       // --web-address.
	EnvTubarrWebAuth         = "TUBARR_WEB_AUTH"          // --web-auth.
	EnvTubarrWebHTTPS        = "TUBARR_WEB_HTTPS"         // --web-https.
	EnvTubarrWebTLSCert      = "TUBARR_WEB_TLS_CERT"      // --web-tls-cert.
	EnvTubarrWebTLSKey       = "TUBARR_WEB_TLS_KEY"       // --web-tls-key.
	EnvTubarrWebUsername     = "TUBARR_WEB_USERNAME"      // Sets the login at startup, e.g. for a forgotten password.
	EnvTubarrWebPassword     = "TUBARR_WEB_PASSWORD"      // Sets the login at startup, e.g. for a forgotten password.
	EnvTubarrLogToken        = "TUBARR_LOG_TOKEN"         // Passes the log token to Metarr.
	EnvTubarrLogURL          = "TUBARR_LOG_URL"           // Passes the URL Metarr sends its logs to.
	TubarrLogURLChannelParam = "channel"                  // TUBARR_LOG_URL's query parameter naming the channel a Metarr run is for.
	EnvTubarrWebAllowedHosts = "TUBARR_WEB_ALLOWED_HOSTS" // Comma-separated hostnames allowed alongside those saved in Settings.
)
