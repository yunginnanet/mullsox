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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var ips []netip.Addr
	if hostname == "" {
		return ips
	}
	for n := 0; n < 5; n++ {
		var err error
		var res []netip.Addr
		go func() {
			res, err = net.DefaultResolver.LookupNetIP(ctx, "ip", hostname)
			if err == nil && res != nil && len(res) > 0 {
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
	var tmpMap = make(map[netip.AddrPort]struct{})
	var tmpMapMu = &sync.RWMutex{}
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
				tmpMapMu.RLock()
				_, ok := tmpMap[ap]
				if ap.IsValid() && ap.Port() > 0 && !ok {
					sox = append(sox, ap)
					tmpMapMu.RUnlock()
					tmpMapMu.Lock()
					tmpMap[ap] = struct{}{}
					tmpMapMu.Unlock()
					continue
				}
				tmpMapMu.RUnlock()
				if !ap.IsValid() {
					err = fmt.Errorf("invalid address/port combo: %s", ap.String())
				}
			}
		}(&serv)
	}
	wg.Wait()
	return
}

func (c *Checker) GetAndVerifySOCKS() (chan netip.AddrPort, chan error) {
	sox, err := c.GetSOCKS()
	var errs = make(chan error, len(sox)+1)
	var verified = make(chan netip.AddrPort, len(sox))
	if err != nil || len(sox) == 0 {
		errs <- err
		close(errs)
		return nil, errs
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(sox))
	for _, prx := range sox {
		time.Sleep(250 * time.Millisecond)
		go func(prx netip.AddrPort) {
			defer wg.Done()
			var conn net.Conn
			conn, err = net.DialTimeout("tcp", prx.String(), 10*time.Second)
			if err != nil {
				errs <- err
			}
			if conn != nil {
				_ = conn.Close()
			}
			if err == nil {
				verified <- prx
			}
		}(prx)
	}
	go func() {
		wg.Wait()
		close(errs)
		close(verified)
	}()
	return verified, errs
}
