package httputils

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHttpClient(t *testing.T) {
	client, err := NewHttpClient(&HttpClientConfig{})
	if nil != err {
		t.Error(err)
	}

	res, err := client.Get(nil, "http://www.google.com", nil)
	if nil != err {
		t.Error(err)
	}
	t.Log(res.StatusCode)
	t.Log(res.Header)
}

func TestBackoff(t *testing.T) {
	client, err := NewHttpClient(&HttpClientConfig{
		TimeoutMs:          1000,
		MaxRetry:           3,
		RetryWaitTimeMs:    15,
		MaxRetryWaitTimeMs: 50,
	})
	if nil != err {
		t.Error(err)
	}

	res, err := client.Get(nil, "http://www.google.com", nil)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(res.StatusCode)
	t.Log(res.Header)
}

func TestRetryAfterFun(t *testing.T) {
	client, err := NewHttpClient(&HttpClientConfig{
		TimeoutMs:          10,
		MaxRetry:           3,
		RetryWaitTimeMs:    15,
		MaxRetryWaitTimeMs: 50000,
	})
	if nil != err {
		t.Error(err)
	}

	client.RetryAfterFunc = func(response *http.Response) time.Duration {
		// this wait time greater than config max retry wait time, use config max retry wait time.
		waitTime := time.Duration(3) * time.Second
		return waitTime
	}
	res, err := client.Get(nil, "http://www.google.com", nil)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(res.StatusCode)
	t.Log(res.Header)
}

func TestRetryConditions(t *testing.T) {
	client, err := NewHttpClient(&HttpClientConfig{
		TimeoutMs:          1000,
		MaxRetry:           3,
		RetryWaitTimeMs:    15,
		MaxRetryWaitTimeMs: 50,
	})
	if nil != err {
		t.Error(err)
	}

	retryConditionFunc := func(response *http.Response, err error) bool {
		return true
	}
	client.RetryConditions = []RetryConditionFunc{retryConditionFunc}

	res, err := client.Get(nil, "http://www.google.com", nil)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log(res.StatusCode)
	t.Log(res.Header)
}
