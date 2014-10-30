package dnscache

import (
	. "github.com/karlseguin/expect"
	"net"
	"sort"
	"testing"
	"time"
)

type CacheTests struct{}

func Test_Cache(t *testing.T) {
	Expectify(new(CacheTests), t)
}

func (_ *CacheTests) FetchReturnsAndErrorOnInvalidLookup() {
	ips, err := New(0).Lookup("invalid.viki.io")
	Expect(ips).To.Equal(nil)
	Expect(err.Error()).To.Equal("lookup invalid.viki.io: no such host")
}

func (_ *CacheTests) FetchReturnsAListOfIps() {
	ips, _ := New(0).Lookup("dnscache.go.test.viki.io")
	assertIps(ips, []string{"1.123.58.13", "31.85.32.110"})
}

func (_ *CacheTests) CallingLookupAddsTheItemToTheCache() {
	r := New(0)
	r.Lookup("dnscache.go.test.viki.io")
	assertIps(r.cache["dnscache.go.test.viki.io"].ips, []string{"1.123.58.13", "31.85.32.110"})
}

func (_ *CacheTests) FetchLoadsValueFromTheCache() {
	r := New(0)
	r.cache["invalid.viki.io"] = &value{[]net.IP{net.ParseIP("1.1.2.3")}, time.Now()}
	ips, _ := r.Fetch("invalid.viki.io")
	assertIps(ips, []string{"1.1.2.3"})
}

func (_ *CacheTests) FetchOneLoadsTheFirstValue() {
	r := New(0)
	r.cache["something.viki.io"] = &value{[]net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")}, time.Now()}
	ip, _ := r.FetchOne("something.viki.io")
	assertIps([]net.IP{ip}, []string{"1.1.2.3"})
}

func (_ *CacheTests) FetchOneStringLoadsTheFirstValue() {
	r := New(0)
	r.cache["something.viki.io"] = &value{[]net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")}, time.Now()}
	ip, _ := r.FetchOneString("something.viki.io")
	Expect(ip).To.Equal("100.100.102.103")
}

func (_ *CacheTests) FetchLoadsTheIpAndCachesIt() {
	r := New(0)
	ips, _ := r.Fetch("dnscache.go.test.viki.io")
	assertIps(ips, []string{"1.123.58.13", "31.85.32.110"})
	assertIps(r.cache["dnscache.go.test.viki.io"].ips, []string{"1.123.58.13", "31.85.32.110"})
}

func (_ *CacheTests) ItReloadsTheIpsAtAGivenInterval() {
	r := New(time.Nanosecond)
	r.cache["dnscache.go.test.viki.io"] = &value{expires: time.Now().Add(-time.Minute)}
	r.Refresh()
	assertIps(r.cache["dnscache.go.test.viki.io"].ips, []string{"1.123.58.13", "31.85.32.110"})
}

func assertIps(actuals []net.IP, expected []string) {
	Expect(len(actuals)).To.Equal(len(expected))
	sort.Strings(expected)
	for _, ip := range actuals {
		if sort.SearchStrings(expected, ip.String()) == -1 {
			Fail("Got an unexpected ip: %v:", actuals[0])
		}
	}
}
