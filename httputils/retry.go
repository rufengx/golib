package httputils

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxRetries  = 3
	defaultWaitTime    = time.Duration(100) * time.Millisecond
	defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
)

type (
	ConfigureFunc func(options *Options)

	// Judge response it's need to retry.
	RetryConditionFunc func(response *http.Response, err error) bool

	// Custom parse response HTTP Retry-After header.
	// If Retry-After time greater than config max retry wait time, use config max retry wait time.
	// See: https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
	RetryAfterFunc func(response *http.Response) time.Duration
)

type Options struct {
	maxRetries       int
	retryWaitTime    time.Duration
	maxRetryWaitTime time.Duration
	retryConditions  []RetryConditionFunc
	retryAfterFunc   RetryAfterFunc
}

func MaxRetries(value int) ConfigureFunc {
	return func(o *Options) {
		o.maxRetries = value
	}
}

func RetryWaitTime(value time.Duration) ConfigureFunc {
	return func(o *Options) {
		o.retryWaitTime = value
	}
}

func MaxRetryWaitTime(value time.Duration) ConfigureFunc {
	return func(o *Options) {
		o.maxRetryWaitTime = value
	}
}

func RetryAfterFun(retryAfterFunc RetryAfterFunc) ConfigureFunc {
	return func(o *Options) {
		o.retryAfterFunc = retryAfterFunc
	}
}

func RetryConditions(conditions []RetryConditionFunc) ConfigureFunc {
	return func(o *Options) {
		o.retryConditions = conditions
	}
}

func defaultRetryAfterFunc(response *http.Response) time.Duration {
	if nil == response {
		return 0
	}
	// Retry-After: Fri, 31 Dec 1999 23:59:59 GMT
	// Retry-After: 120
	rf := response.Header.Get("Retry-After")
	if len(rf) == 0 {
		return 0
	}
	waitTimeSec, err := strconv.Atoi(rf)
	if nil == err && waitTimeSec != 0 {
		return time.Duration(waitTimeSec) * time.Second
	} else if nil != err && len(strings.TrimSpace(rf)) > 0 {
		waitTime, err := time.Parse(time.RFC1123, rf)
		if nil == err {
			return waitTime.Sub(time.Now())
		}
		return 0
	}
	return 0
}

func Backoff(execFunc func() (*http.Response, error), configureFuncs ...ConfigureFunc) error {
	// Default options.
	options := &Options{
		maxRetries:       defaultMaxRetries,
		retryWaitTime:    defaultWaitTime,
		maxRetryWaitTime: defaultMaxWaitTime,
	}

	// Configure some options in the settings.
	for _, configureFunc := range configureFuncs {
		configureFunc(options)
	}

	var err error
	var res *http.Response
	for retryCount := 0; retryCount < options.maxRetries; retryCount++ {
		// 1. Exec func
		res, err = execFunc()

		// 2. Judge it's need retry
		needRetry := err != nil
		for _, condition := range options.retryConditions {
			needRetry = condition(res, err)
			if needRetry {
				break
			}
		}

		if !needRetry {
			return err
		}

		// 3. Custom Parse Response Retry-After header.
		// See: https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#14.37
		retryAfterTime := time.Duration(0)
		retryAfterFunc := options.retryAfterFunc
		if nil == retryAfterFunc {
			retryAfterFunc = defaultRetryAfterFunc
		}
		retryAfterTime = retryAfterFunc(res)
		waitTime := timeDuration(options.retryWaitTime, options.maxRetryWaitTime, retryAfterTime, retryCount)

		// just only wait.
		<-time.After(waitTime)
	}
	return err
}

// About timeout, we need consider "Exponential Backoff And Jitter"
// See: https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
func timeDuration(minWaitTime, maxWaitTime, retryAfterTime time.Duration, retryCount int) time.Duration {
	const maxInt = 1<<31 - 1 // max int for arch 386
	// 1. calculate wait time.
	min := float64(minWaitTime)
	max := float64(maxWaitTime)

	temp := math.Min(max, min*math.Exp2(float64(retryCount)))
	ri := int(temp / 2)
	if ri <= 0 {
		ri = maxInt // max int for arch 386
	}
	result := time.Duration(math.Abs(float64(ri + rand.Intn(ri))))

	if result < minWaitTime {
		result = minWaitTime
	}

	if 0 == retryAfterTime {
		return result
	}

	// 2. use response Retry-After header retry time.
	if retryAfterTime < 0 || maxWaitTime < retryAfterTime {
		retryAfterTime = maxWaitTime
	}
	if retryAfterTime < minWaitTime {
		retryAfterTime = minWaitTime
	}
	return retryAfterTime
}
