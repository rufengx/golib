package httputils

import (
	"context"
	"testing"
	"time"
)

func ExampleResolver_LookupHost() {

}

func TestResolver_LookupHost(t *testing.T) {
	reslover := &DnsResolver{
		TTL: time.Duration(5) * time.Minute,
	}
	hosts, err := reslover.LookupHost(context.Background(), "www.google.com")
	if nil != err {
		t.Error(err)
	}
	for _, host := range hosts {
		t.Log(host)
	}
}

func TestDnsResolver_GetAllEntities(t *testing.T) {
	resolver := &DnsResolver{
		TTL: time.Duration(5) * time.Minute,
	}
	_, err := resolver.LookupHost(context.Background(), "www.google.com")
	if nil != err {
		t.Error(err)
	}

	_, err = resolver.LookupHost(context.Background(), "www.baidu.com")
	if nil != err {
		t.Error(err)
	}

	for _, entity := range resolver.GetAllEntities() {
		t.Log(entity.ips, entity.timestampNano)
	}
}
