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

func (_ CacheTests) FetchReturnsAndErrorOnInvalidLookup() {
	ips, err := New(0).Lookup("invalid.openmymind.io")
	Expect(ips).To.Equal(nil)
	Expect(err.Error()).To.Equal("lookup invalid.openmymind.io: no such host")
}

func (_ CacheTests) FetchReturnsAListOfIps() {
	ips, _ := New(0).Lookup("go-dnscache.openmymind.io")
	assertIps(ips, []string{"8.8.8.8", "8.8.4.4", "2404:6800:4005:8050::1014"})
}

func (_ CacheTests) Fetchv4ReturnsAListOfIps() {
	ips, _ := New(0).FetchV4("go-dnscache.openmymind.io")
	assertIps(ips, []string{"8.8.8.8", "8.8.4.4"})
}

func (_ CacheTests) CallingLookupAddsTheItemToTheCache() {
	r := New(0)
	r.Lookup("go-dnscache.openmymind.io")
	assertIps(r.cache["go-dnscache.openmymind.io"].ips, []string{"8.8.8.8", "8.8.4.4", "2404:6800:4005:8050::1014"})
}

func (_ CacheTests) FetchLoadsValueFromTheCache() {
	r := New(0)
	r.cache["invalid.openmymind.io"] = &value{
		ips:     []net.IP{net.ParseIP("1.1.2.3")},
		ipv4s:   []net.IP{net.ParseIP("1.1.2.3")},
		expires: time.Now(),
	}
	ips, _ := r.Fetch("invalid.openmymind.io")
	assertIps(ips, []string{"1.1.2.3"})
}

func (_ CacheTests) FetchOneLoadsAValue() {
	r := New(0)
	r.cache["something.openmymind.io"] = &value{
		ips:     []net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")},
		ipv4s:   []net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")},
		expires: time.Now(),
	}
	ip, _ := r.FetchOne("something.openmymind.io")
	if ip.String() != "100.100.102.103" && ip.String() != "1.1.2.3" {
		Fail("expected ip to be one of two ips")
	}
}

func (_ CacheTests) FetchOneStringLoadsAValue() {
	r := New(0)
	r.cache["something.openmymind.io"] = &value{
		ips:     []net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")},
		ipv4s:   []net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")},
		expires: time.Now(),
	}
	ip, _ := r.FetchOneString("something.openmymind.io")
	if ip != "100.100.102.103" && ip != "100.100.102.104" {
		Fail("expected ip to be one of two ips")
	}
}

func (_ CacheTests) FetchLoadsTheIpAndCachesIt() {
	r := New(0)
	ips, _ := r.Fetch("go-dnscache.openmymind.io")
	assertIps(ips, []string{"8.8.4.4", "8.8.8.8", "2404:6800:4005:8050::1014"})
	assertIps(r.cache["go-dnscache.openmymind.io"].ips, []string{"8.8.4.4", "8.8.8.8", "2404:6800:4005:8050::1014"})
}

func (_ CacheTests) ItReloadsTheIpsAtAGivenInterval() {
	r := New(time.Nanosecond)
	r.cache["go-dnscache.openmymind.io"] = &value{expires: time.Now().Add(-time.Minute)}
	r.Refresh()
	assertIps(r.cache["go-dnscache.openmymind.io"].ips, []string{"8.8.4.4", "8.8.8.8", "2404:6800:4005:8050::1014"})
}

func assertIps(actuals []net.IP, expected []string) {
	Expect(len(actuals)).To.Equal(len(expected))
	ips := make([]string, len(actuals))
	for i, ip := range actuals {
		ips[i] = ip.String()
	}
	sort.Strings(ips)
	sort.Strings(expected)

	for i, ip := range ips {
		Expect(ip).To.Equal(expected[i])
	}
}
