package spf

import (
	"context"
	"net"
	"time"
)

func NewDNSWithResolver(nameserver string) {
	resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(5000),
			}
			return d.DialContext(ctx, "udp", net.JoinHostPort(nameserver, "53"))
		},
	}
	lookupTXT = lookupTXTWithResolver
	lookupMX = lookupMXWithResolver
	lookupIP = lookupIPWithResolver
	lookupAddr = lookupAddrWithResolver
}

func lookupTXTWithResolver(domain string) (txts []string, err error) {
	ctx := context.Background()
	return resolver.LookupTXT(ctx, domain)
}

func lookupMXWithResolver(domain string) (mxs []*net.MX, err error) {
	ctx := context.Background()
	return resolver.LookupMX(ctx, domain)
}

func lookupIPWithResolver(host string) (ips []net.IP, err error) {
	ctx := context.Background()
	return resolver.LookupIP(ctx, "ip", host)
}

func lookupAddrWithResolver(host string) (addrs []string, err error) {
	ctx := context.Background()
	return resolver.LookupAddr(ctx, host)
}
