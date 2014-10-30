// Package dnscache caches DNS lookups
package dnscache

import (
	"net"
	"sync"
	"time"
	"math/rand"
)

type value struct {
	ips     []net.IP
	expires time.Time
}

type Resolver struct {
	sync.RWMutex
	minTTL     time.Duration
	defaultTTL time.Duration
	cache      map[string]*value
	ttls       map[string]time.Duration
}

func New(defaultTTL time.Duration) *Resolver {
	resolver := &Resolver{
		minTTL:     defaultTTL,
		defaultTTL: defaultTTL,
		cache:      make(map[string]*value),
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

	ttl, ok := r.ttls[address]
	if ok == false {
		ttl = r.defaultTTL
	}
	now := time.Now()
	r.Lock()
	defer r.Unlock()
	r.cache[address] = &value{ips, now.Add(ttl)}
	return ips, nil
}

func (r *Resolver) autoRefresh() {
	for {
		time.Sleep(r.minTTL)
		r.Refresh()
	}
}
