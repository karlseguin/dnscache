package dnscache

import (
	"net"
	"sort"
	"testing"
	"time"
	. "github.com/karlseguin/expect"
)

type CacheTests struct {}

func Test_Cache(t *testing.T) {
	Expectify(new(CacheTests), t)
}

func (c *CacheTests) FetchReturnsAndErrorOnInvalidLookup() {
	ips, err := New(0).Lookup("invalid.viki.io")
	Expect(ips).To.Equal(nil)
	Expect(err.Error()).To.Equal("lookup invalid.viki.io: no such host")
}

func (c *CacheTests) FetchReturnsAListOfIps() {
	ips, _ := New(0).Lookup("dnscache.go.test.viki.io")
	assertIps(ips, []string{"1.123.58.13", "31.85.32.110"})
}

func (c *CacheTests) CallingLookupAddsTheItemToTheCache() {
	r := New(0)
	r.Lookup("dnscache.go.test.viki.io")
	assertIps(r.cache["dnscache.go.test.viki.io"], []string{"1.123.58.13", "31.85.32.110"})
}

func (c *CacheTests) FetchLoadsValueFromTheCache() {
	r := New(0)
	r.cache["invalid.viki.io"] = []net.IP{net.ParseIP("1.1.2.3")}
	ips, _ := r.Fetch("invalid.viki.io")
	assertIps(ips, []string{"1.1.2.3"})
}

func (c *CacheTests) FetchOneLoadsTheFirstValue() {
	r := New(0)
	r.cache["something.viki.io"] = []net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")}
	ip, _ := r.FetchOne("something.viki.io")
	assertIps([]net.IP{ip}, []string{"1.1.2.3"})
}

func (c *CacheTests) FetchOneStringLoadsTheFirstValue() {
	r := New(0)
	r.cache["something.viki.io"] = []net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")}
	ip, _ := r.FetchOneString("something.viki.io")
	Expect(ip).To.Equal("100.100.102.103")
}

func (c *CacheTests) FetchLoadsTheIpAndCachesIt() {
	r := New(0)
	ips, _ := r.Fetch("dnscache.go.test.viki.io")
	assertIps(ips, []string{"1.123.58.13", "31.85.32.110"})
	assertIps(r.cache["dnscache.go.test.viki.io"], []string{"1.123.58.13", "31.85.32.110"})
}

func (c *CacheTests) ItReloadsTheIpsAtAGivenInterval() {
	r := New(time.Nanosecond * 10000000)
	r.cache["dnscache.go.test.viki.io"] = nil
	time.Sleep(time.Nanosecond * 20000000)
	assertIps(r.cache["dnscache.go.test.viki.io"], []string{"1.123.58.13", "31.85.32.110"})
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
