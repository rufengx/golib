package httputils

import (
	"context"
	"golang.org/x/sync/singleflight"
	"net"
	"sync"
	"time"
)

type DnsResolver struct {
	mutex sync.RWMutex
	once  sync.Once

	cache map[string]*cacheEntity
	TTL   time.Duration // default 5 min.
}

type cacheEntity struct {
	ips           []net.IP
	timestampNano int64
}

var lookupGroup singleflight.Group

func (r *DnsResolver) init() {
	if nil == r.cache {
		r.cache = make(map[string]*cacheEntity)
	}
}

func (r *DnsResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	r.once.Do(r.init)

	ips, err := r.queryCache(ctx, host)
	if nil != err {
		return nil, err
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); nil != ipv4 {
			addrs = append(addrs, ipv4.String())
		}
	}
	return addrs, nil
}

func (r *DnsResolver) GetAllEntities() []*cacheEntity {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	entities := []*cacheEntity{}
	for _, entity := range r.cache {
		entities = append(entities, entity)
	}
	return entities
}

func (r *DnsResolver) lookupFunc(host string) func() (interface{}, error) {
	return func() (interface{}, error) {
		return net.LookupIP(host)
	}
}

func (r *DnsResolver) queryCache(ctx context.Context, key string) (ips []net.IP, err error) {
	r.mutex.RLock()
	entry, found := r.cache[key]
	r.mutex.RUnlock()

	if found && time.Now().UnixNano() < entry.timestampNano+r.TTL.Nanoseconds() {
		return entry.ips, nil
	}

	c := lookupGroup.DoChan(key, r.lookupFunc(key))

	select {
	case <-ctx.Done():
		err = ctx.Err()
		if err == context.DeadlineExceeded {
			// When query DNS service timeout, we shouldn't waiting query complete.
			lookupGroup.Forget(key)
		}
	case res := <-c:
		if res.Shared {
			r.mutex.RLock()
			entry, found := r.cache[key]
			r.mutex.RUnlock()
			if found && time.Now().UnixNano() < entry.timestampNano+r.TTL.Nanoseconds() {
				return entry.ips, nil
			}
		}
		err := res.Err
		if nil == err {
			var ok bool
			ips, ok = res.Val.([]net.IP)
			if ok {
				// Update cache.
				r.mutex.Lock()
				r.cache[key] = &cacheEntity{
					ips:           ips,
					timestampNano: time.Now().UnixNano(),
				}
				r.mutex.Unlock()
			}
		}
	}
	return
}
