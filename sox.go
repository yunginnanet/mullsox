package mullsox

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"
)

func persistentResolver(hostname string) []netip.Addr {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var ips []netip.Addr
	for n := 0; n < 5; n++ {
		var err error
		var res []netip.Addr
		go func() {
			res, err = net.DefaultResolver.LookupNetIP(ctx, "ip", hostname)
			if err == nil && len(res) > 0 {
				ips = res
				cancel()
			}
		}()
		time.Sleep(1 * time.Second)
	}
	<-ctx.Done()
	return ips
}

func (c *Checker) GetSOCKS() (sox []netip.AddrPort, err error) {
	if err = c.Update(); err != nil {
		return
	}
	wg := &sync.WaitGroup{}
	for _, serv := range c.m {
		wg.Add(1)
		go func(endpoint *MullvadServer) {
			defer wg.Done()
			ips := persistentResolver(endpoint.SocksName)
			for _, ip := range ips {
				port := uint16(endpoint.SocksPort)
				if port == 0 {
					port = 1080
				}
				ap := netip.AddrPortFrom(ip, port)
				if ap.IsValid() && ap.Port() > 0 {
					sox = append(sox, ap)
					continue
				}
				err = fmt.Errorf("invalid address/port combo: %s", ap.String())
			}
		}(&serv)
	}
	wg.Wait()
	return
}
