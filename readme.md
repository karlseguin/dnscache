### A DNS cache for Go
CGO is used to lookup domain names. Given enough concurrent requests and the slightest hiccup in name resolution, it's quite easy to end up with blocked/leaking goroutines.

The issue is documented at <https://code.google.com/p/go/issues/detail?id=5625>

The Go team's singleflight solution (which isn't in stable yet) is rather elegant. However, it only eliminates concurrent lookups (thundering herd problems). Many systems can live with slightly stale resolve names, which means we can cacne DNS lookups and refresh them in the background.

### Installation
Install using the "go get" command:

    go get github.com/karlseguin/dnscache

### Usage
The cache is thread safe. Create a new instance by specifying how long each entry should be cached (in seconds). Items will be refreshed in the background.

```go
//refresh items every 5 minutes
resolver := dnscache.New(time.Minute * 5)

//get an array of net.IP
ips, _ := resolver.Fetch("openmymind.io")

//get the first net.IP
ip, _ := resolver.FetchOne("openmymind.io")

//get the first net.IP as string
ip, _ := resolver.FetchOneString("openmymind.io")
```

If you are using an `http.Transport`, you can use this cache by speficifying a
`Dial` function:

```go
transport := &http.Transport {
  MaxIdleConnsPerHost: 64,
  Dial: func(network string, address string) (net.Conn, error) {
    separator := strings.LastIndex(address, ":")
    ip, _ := dnscache.FetchString(address[:separator])
    return net.Dial("tcp", ip + address[separator:])
  },
}
```

## TTLs

By default, items are cached for the TTL specified when creating the cache object. This can be overwritten on a per-address basis via the `TTL` method:

```go
resolver.TTL("algorithms.openmymind.net", time.Second * 30)
```

Note that unlike the other methods, `TTL` is not thread-safe.
