package httputils

type HttpClientConfig struct {
	TimeoutMs int64 // default timeout 30 seconds, contain connection timeout and request timeout.

	MaxRetry           int // default not retry.
	RetryWaitTimeMs    int64
	MaxRetryWaitTimeMs int64

	AllowRedirect     bool
	MaxAllowRedirects int // default allow 10 redirects.

	KeepAliveMs           int64 // default keep-alive time is 30 seconds.
	MaxIdleConns          int   // default max idle connections is 100.
	MaxIdleConnsPerHost   int   // default per-host max idle connections is 0, no limit.
	IdleConnTimeoutMs     int   // default idle connections timeout is 90 seconds.
	TLSHandshakeTimeoutMs int   // default TLS hand shake timeout is 10 seconds.

	ProxyUrl    string // support http, https, socks proxy.
	ProxyUname  string
	ProxyPasswd string
}

var DefaultHttpClientConfig = &HttpClientConfig{
	TimeoutMs:             30000,
	MaxRetry:              0,
	RetryWaitTimeMs:       0,
	MaxRetryWaitTimeMs:    0,
	AllowRedirect:         false,
	KeepAliveMs:           30000,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   0,
	IdleConnTimeoutMs:     90000,
	TLSHandshakeTimeoutMs: 10000,
}

func options(config *HttpClientConfig) *HttpClientConfig {
	if nil == config {
		return DefaultHttpClientConfig
	}

	if config.TimeoutMs == 0 {
		config.TimeoutMs = 30000
	}

	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 100
	}

	if config.IdleConnTimeoutMs == 0 {
		config.IdleConnTimeoutMs = 90000
	}

	if config.TLSHandshakeTimeoutMs == 0 {
		config.TLSHandshakeTimeoutMs = 10000
	}

	if config.AllowRedirect && config.MaxAllowRedirects == 0 {
		config.MaxAllowRedirects = 10
	}
	return config
}
