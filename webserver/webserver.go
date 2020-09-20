package webserver

type WebServer struct {
	WebConfig *WebServerConfig
}

type WebServerConfig struct {
	Addr             string `json:"addr"` // "e.g. 127.0.0.1:8090"
	SSL              bool
	Throttle         int
	ProcessTimeoutMs int
	ReadTimeoutMs    int `json:"read_timeout_ms"`
	WriteTimeoutMs   int `json:"write_timeout_ms"`
	MaxHeaderBytes   int `json:"max_header_bytes"`
}

//func main() {
//
//
//
//	router := fasthttprouter.New()
//	router.GET("/", HelloWorld)
//
//
//
//	log.Fatal(webserver.ListenAndServe(":8888", router.Handler))
//}
//
//func HelloWorld(ctx *webserver.RequestCtx)  {
//
//
//
//
//	ctx.Write([]byte(string("hello world")))
//}
