package mullsox

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"git.tcp.direct/kayos/mullsox/mullvad"
)

const MullvadInternalDNS4 = "10.64.0.1:53"
const MullvadInternalDNS6 = "[fc00:bbbb:bbbb:bb01::2b:e7d3]:53"

type RelayFetcher interface {
	GetRelays() ([]mullvad.MullvadServer, error)
}

func GetSOCKS(fetcher RelayFetcher) (sox []netip.AddrPort, err error) {
	relays, err := fetcher.GetRelays()
	switch {
	case err != nil:
		return nil, err
	case len(relays) == 0:
		return nil, fmt.Errorf("no relays found")
	default:
	}
	var tmpMap = make(map[netip.AddrPort]struct{})
	var tmpMapMu = &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	for _, serv := range relays {
		wg.Add(1)
		go func(endpoint *mullvad.MullvadServer) {
			defer wg.Done()
			var ips []net.IP
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			ips, err = net.DefaultResolver.LookupIP(ctx, "ip", endpoint.SocksName)
			if err != nil {
				return
			}
			if len(ips) == 0 {
				return
			}
			for _, ipa := range ips {
				var ip netip.Addr
				port := uint16(endpoint.SocksPort)
				if port == 0 {
					port = 1080
				}
				ip, err = netip.ParseAddr(ipa.String())
				if err != nil {
					continue
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

func checker(candidate netip.AddrPort, verified chan netip.AddrPort, errs chan error, working *int64) {
	atomic.AddInt64(working, 1)
	defer func() {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(working, -1)
	}()
	if !candidate.IsValid() {
		errs <- fmt.Errorf("invalid address/port combo: %s", candidate.String())
		return
	}
	var conn net.Conn
	conn, err := net.DialTimeout("tcp", candidate.String(), 15*time.Second)
	if err != nil {
		errs <- err
	}
	if conn != nil {
		_ = conn.Close()
	}
	if err == nil {
		verified <- candidate
	}
}

func GetAndVerifySOCKS(fetcher RelayFetcher) (chan netip.AddrPort, chan error) {
	sox, err := GetSOCKS(fetcher)
	var errs = make(chan error, len(sox)+1)
	switch {
	case len(sox) == 0:
		err = fmt.Errorf("no relays found")
		fallthrough
	case err != nil:
		go func() {
			errs <- err
		}()
		return nil, errs
	default:
	}

	var (
		verified = make(chan netip.AddrPort, len(sox))
		working  = new(int64)
	)
	atomic.StoreInt64(working, 0)

	for _, prx := range sox {
		for atomic.LoadInt64(working) > 10 {
			time.Sleep(50 * time.Millisecond)
		}
		checker(prx, verified, errs, working)
	}
	go func() {
		for atomic.LoadInt64(working) > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		close(errs)
		close(verified)
	}()
	return verified, errs
}
