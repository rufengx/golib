package httputils

//func main() {
//  1. new http client
//	client, err := NewHttpClient(&HttpClientConfig{
//		TimeoutMs:          1000,
//		MaxRetry:           3,
//		RetryWaitTimeMs:    15,
//		MaxRetryWaitTimeMs: 50,
//	})
//
//  2. do request
//	res, err := client.Get(nil, "www.example.com", nil)
//	if nil != err {
//		panic(err)
//	}
//
//  3. parse response
//	fmt.Println(res.StatusCode)
//}
