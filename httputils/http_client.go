package httputils

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	netUrl "net/url"
	"strings"
	"time"
)

// HttpClient struct wrapper http client, provides some new feature.
// 1. Local DNS Cache.
// 2. Retry Exponential Backoff And Jitter Strategy. See: https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
// 3. Multi-Goroutine Download.
// 4. Allow Custom Max Redirects.
type HttpClient struct {
	MaxRetry         int
	RetryWaitTime    time.Duration
	MaxRetryWaitTime time.Duration

	AllowRedirect     bool
	MaxAllowRedirects int

	client *http.Client

	// Custom parse response HTTP Retry-After header.
	// See: https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
	// e.g. Retry-After: Fri, 31 Dec 1999 23:59:59 GMT
	// e.g. Retry-After: 120
	RetryAfterFunc RetryAfterFunc

	// Custom  judge response is need to retry.
	// you can implement your custom strategy, for example: response status code is not 200.
	RetryConditions []RetryConditionFunc
}

var DefaultHttpClient = &HttpClient{client: http.DefaultClient}

func NewHttpClient(config *HttpClientConfig) (*HttpClient, error) {
	// 1. configure some options in setting.
	config = options(config)

	// 2. init custom client
	httpClient := &HttpClient{
		AllowRedirect:     config.AllowRedirect,
		MaxAllowRedirects: config.MaxAllowRedirects,

		MaxRetry:         config.MaxRetry,
		RetryWaitTime:    time.Duration(config.RetryWaitTimeMs) * time.Millisecond,
		MaxRetryWaitTime: time.Duration(config.MaxRetryWaitTimeMs) * time.Millisecond,
	}

	// 3. init transport
	trans, err := createTransport(config)
	if nil != err {
		return nil, err
	}

	// 4. init cookies
	jar, _ := cookiejar.New(nil)

	// 5. new http client
	httpClient.client = &http.Client{
		Transport: trans,
		Timeout:   time.Duration(config.TimeoutMs) * time.Millisecond,
		Jar:       jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.Response == nil {
				return errors.New("expected non-nil Request.Response")
			}
			if !httpClient.AllowRedirect {
				return errors.New("not allow redirect")
			}
			if httpClient.MaxAllowRedirects > 0 && config.MaxAllowRedirects < len(via) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	return httpClient, nil
}

func createTransport(config *HttpClientConfig) (trans *http.Transport, err error) {
	var proxyUrl *netUrl.URL
	var proxyHeader http.Header
	if config.ProxyUrl != "" {
		if proxyUrl, err = netUrl.Parse(config.ProxyUrl); nil != err {
			return nil, err
		}
		if config.ProxyUname != "" && config.ProxyPasswd != "" {
			auth := config.ProxyUname + ":" + config.ProxyPasswd
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			proxyHeader.Set("Proxy-Authorization", basicAuth)
		}
	}

	trans = &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
			resolver := &DnsResolver{} // use local DNS cache.
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := resolver.LookupHost(ctx, host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				var dialer net.Dialer
				dialer.Timeout = time.Duration(config.TimeoutMs) * time.Millisecond
				dialer.KeepAlive = time.Duration(config.KeepAliveMs) * time.Millisecond
				dialer.DualStack = true
				conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
				if err == nil {
					break
				}
			}
			return
		},
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     time.Duration(config.IdleConnTimeoutMs) * time.Millisecond,
		TLSHandshakeTimeout: time.Duration(config.TLSHandshakeTimeoutMs) * time.Millisecond,
		Proxy:               http.ProxyURL(proxyUrl),
		ProxyConnectHeader:  proxyHeader,
	}
	return trans, nil
}

func (c *HttpClient) Cookies(u *netUrl.URL) []*http.Cookie {
	return c.client.Jar.Cookies(u)
}

func (c *HttpClient) Get(header map[string]string, url string, params map[string]string) (*Response, error) {
	body := netUrl.Values{}
	for key, value := range params {
		body.Set(key, value)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if nil != err {
		return nil, err
	}
	req.URL.RawQuery = body.Encode()

	for key, value := range header {
		req.Header.Set(key, value)
	}
	return c.Do(req)
}

func (c *HttpClient) PostForm(header map[string]string, url string, params map[string]string) (*Response, error) {
	values := netUrl.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(values.Encode()))
	if nil != err {
		return nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

func (c *HttpClient) Post(header map[string]string, url, contentType string, body []byte) (*Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if nil != err {
		return nil, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *HttpClient) Do(req *http.Request) (*Response, error) {
	var err error
	var rawRes *http.Response

	// 1. do request
	if 0 == c.MaxRetry {
		rawRes, err = c.client.Do(req)
	} else {
		execFunc := func() (*http.Response, error) {
			rawRes, err = c.client.Do(req)
			return rawRes, err
		}
		err = Backoff(execFunc,
			MaxRetries(c.MaxRetry),
			RetryWaitTime(c.RetryWaitTime),
			MaxRetryWaitTime(c.MaxRetryWaitTime),
			RetryAfterFun(c.RetryAfterFunc),
			RetryConditions(c.RetryConditions))
	}

	if nil != err {
		return nil, err
	}

	// 2. process cookies
	if len(rawRes.Cookies()) > 0 {
		c.client.Jar.SetCookies(req.URL, rawRes.Cookies())
	}

	// 3. read response body
	rawBody := rawRes.Body
	defer rawRes.Body.Close()
	if strings.EqualFold(rawRes.Header.Get("Content-Encoding"), "gzip") && rawRes.ContentLength > 0 {
		if rawBody, err = gzip.NewReader(rawBody); nil != err {
			return nil, err
		}
	}

	body, err := ioutil.ReadAll(rawBody)
	if nil != err {
		return nil, err
	}

	res := &Response{
		RawResponse: rawRes,
		StatusCode:  rawRes.StatusCode,
		Header:      rawRes.Header,
		Body:        body,
	}
	return res, err
}
