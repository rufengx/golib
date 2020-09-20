package requestid

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

var (
	localIP4 uint32
)

type RequestID struct {
	RequestTimeNano uint64
	LocalIPv4       net.IP
}

func init() {
	addrs, err := net.InterfaceAddrs()
	if nil != err {
		log.Fatalf("[init] gen request id fail, cause: %v", err)
	}

	var ip net.IP
	for _, addr := range addrs {
		if ip, _, err = net.ParseCIDR(addr.String()); err != nil {
			continue
		}
		if ip != nil && (ip.To4() != nil) && ip.IsGlobalUnicast() {
			break
		}
	}

	if nil == ip {
		log.Fatal("gen request id, get local ip fail.")
	}

	ip4 := ip.To4()
	localIP4 = (uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3]))
	log.Printf("[init] request id, ip part %+v %s %x", ip, ip4.String(), localIP4)
}

// GenRequestID 生成请求ID，前 64 位为时间戳，单位：纳秒，后 32 位为服务本地IP
func GenRequestID() string {
	return fmt.Sprintf("%016x%08x", time.Now().UnixNano(), localIP4)
}

func ParseRequestID(requestID string) (*RequestID, error) {
	data, err := hex.DecodeString(requestID)
	if nil != err {
		return nil, err
	}

	tsNano, err := strconv.ParseUint(fmt.Sprintf("%016x", data[:8]), 16, 64)
	if nil != err {
		return nil, err
	}

	ip, err := strconv.ParseUint(fmt.Sprintf("%08x", data[8:12]), 16, 32)
	var bytes [4]byte
	bytes[0] = byte(ip & 0xFF)
	bytes[1] = byte((ip >> 8) & 0xFF)
	bytes[2] = byte((ip >> 16) & 0xFF)
	bytes[3] = byte((ip >> 24) & 0xFF)

	return &RequestID{
		RequestTimeNano: tsNano,
		LocalIPv4:       net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0]),
	}, nil
}
