package badcase

// This section code not print x, because it's will running forever.
// Why? when main goroutine sleep 1 second, go scheduler will start execute other sub-goroutine,
// because sub-goroutine numbers equals os CUP numbers, as the same time, every sub-goroutine is not exit.
// so every thread exec a goroutine, don't print x.
//
//func main() {
//	var x int
//	threads := runtime.GOMAXPROCS(0)
//	for i := 0; i < threads; i++ {
//		go func() {
//			for {
//				x++
//			}
//		}()
//	}
//	time.Sleep(time.Second)
//	fmt.Println("x =", x)
//}
