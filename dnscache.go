// Package dnscache caches DNS lookups
package dnscache

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

type value struct {
	ips     []net.IP
	ipv4s   []net.IP
	expires time.Time
}

type Resolver struct {
	sync.RWMutex
	stop       chan struct{}
	minTTL     time.Duration
	defaultTTL time.Duration
	cache      map[string]*value
	ttls       map[string]time.Duration
}

func New(defaultTTL time.Duration) *Resolver {
	resolver := &Resolver{
		minTTL:     defaultTTL,
		defaultTTL: defaultTTL,
		stop:       make(chan struct{}),
		cache:      make(map[string]*value),
		ttls:       make(map[string]time.Duration),
	}
	if defaultTTL > 0 {
		go resolver.autoRefresh()
	}
	return resolver
}

// Set a TTL for a specific address, overwriting the defaultTTL
func (r *Resolver) TTL(address string, ttl time.Duration) {
	r.ttls[address] = ttl
	if ttl < r.minTTL {
		r.minTTL = ttl
	}
}

// Get all of the addresses' ips
func (r *Resolver) Fetch(address string) ([]net.IP, error) {
	r.RLock()
	value, exists := r.cache[address]
	r.RUnlock()
	if exists {
		return value.ips, nil
	}

	return r.Lookup(address)
}

// Get one of the addresses' ips
func (r *Resolver) FetchOne(address string) (net.IP, error) {
	ips, err := r.Fetch(address)
	l := len(ips)
	if err != nil || l == 0 {
		return nil, err
	}
	if l == 1 {
		return ips[0], nil
	}
	return ips[rand.Intn(l)], nil
}

// Get one of the addresses' ips as a string
func (r *Resolver) FetchOneString(address string) (string, error) {
	ip, err := r.FetchOne(address)
	if err != nil || ip == nil {
		return "", err
	}
	return ip.String(), nil
}

// Get all of the addresses' ips
func (r *Resolver) FetchV4(address string) ([]net.IP, error) {
	r.RLock()
	value, exists := r.cache[address]
	r.RUnlock()
	if exists {
		return value.ipv4s, nil
	}
	r.Lookup(address)

	r.RLock()
	value, exists = r.cache[address]
	r.RUnlock()
	if exists {
		return value.ipv4s, nil
	}
	return nil, nil
}

// Get one of the addresses' ips
func (r *Resolver) FetchOneV4(address string) (net.IP, error) {
	ips, err := r.FetchV4(address)
	l := len(ips)
	if err != nil || l == 0 {
		return nil, err
	}
	if l == 1 {
		return ips[0], nil
	}
	return ips[rand.Intn(l)], nil
}

// Get one of the addresses' ips as a string
func (r *Resolver) FetchOneV4String(address string) (string, error) {
	ip, err := r.FetchOneV4(address)
	if err != nil || ip == nil {
		return "", err
	}
	return ip.String(), nil
}

// Refresh expired items (called automatically by default)
func (r *Resolver) Refresh() {
	now := time.Now()
	r.RLock()
	addresses := make([]string, 0, len(r.cache))
	for key, value := range r.cache {
		if value.expires.Before(now) {
			addresses = append(addresses, key)
		}
	}
	r.RUnlock()

	for _, address := range addresses {
		r.Lookup(address)
		time.Sleep(time.Nanosecond * 10000000) //10ms
	}
}

// Lookup an address' ip, circumventing the cache
func (r *Resolver) Lookup(address string) ([]net.IP, error) {
	ips, err := net.LookupIP(address)
	if err != nil {
		return nil, err
	}

	v4s := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ip.To4() != nil {
			v4s = append(v4s, ip)
		}
	}

	ttl, ok := r.ttls[address]
	if ok == false {
		ttl = r.defaultTTL
	}
	expires := time.Now().Add(ttl)
	r.Lock()
	r.cache[address] = &value{
		ips:     ips,
		ipv4s:   v4s,
		expires: expires,
	}
	r.Unlock()
	return ips, nil
}

// Stops the background refresher. Once stopped, it cannot be started again
func (r *Resolver) Stop() {
	r.stop <- struct{}{}
}

func (r *Resolver) autoRefresh() {
	for {
		select {
		case <-r.stop:
			return
		case <-time.After(r.minTTL):
			r.Refresh()
		}
	}
}
