package mullsox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.tcp.direct/kayos/mullsox/mulltest"
	"git.tcp.direct/kayos/mullsox/mullvad"
)

/*
const MullvadInternalDNS4 = "10.64.0.1:53"
const MullvadInternalDNS6 = "[fc00:bbbb:bbbb:bb01::2b:e7d3]:53"
*/

type RelayFetcher interface {
	GetRelays() ([]mullvad.Server, error)
}

func GetSOCKS(fetcher RelayFetcher) ([]netip.AddrPort, error) {
	relays, rerr := fetcher.GetRelays()
	switch {
	case rerr != nil:
		return nil, rerr
	case len(relays) == 0:
		return nil, fmt.Errorf("no relays found")
	default:
	}

	var (
		done     = make(chan struct{})
		errs     = make(chan error, len(relays))
		multiErr error
	)

	var tmpMap = make(map[netip.AddrPort]struct{})
	var tmpMapMu = &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	var resolved = make(chan netip.AddrPort, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wg.Add(len(relays))
	for _, serv := range relays {
		go func(host string, port int) {
			defer wg.Done()
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return
			}
			if len(ips) == 0 {
				return
			}
			for _, ipa := range ips {
				port := uint16(port)
				if port == 0 {
					port = 1080
				}
				ip, err := netip.ParseAddr(ipa.String())
				if err != nil {
					continue
				}
				ap := netip.AddrPortFrom(ip, port)
				if ap.IsValid() && ap.Port() > 0 {
					resolved <- ap
					return
				}
				switch {
				case !ap.IsValid():
					errs <- fmt.Errorf("invalid address/port combo: %s", ap.String())
					continue
				case ap.Port() == 0:
					errs <- fmt.Errorf("invalid port: %d", ap.Port())
					continue
				}
			}
		}(serv.SocksName, serv.SocksPort)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	var sox = make([]netip.AddrPort, 0, len(relays))
	for {
		select {
		case ap := <-resolved:
			tmpMapMu.RLock()
			_, ok := tmpMap[ap]
			tmpMapMu.RUnlock()
			if ok {
				continue
			}
			tmpMapMu.Lock()
			tmpMap[ap] = struct{}{}
			sox = append(sox, ap)
			tmpMapMu.Unlock()
		case err := <-errs:
			multiErr = errors.Join(multiErr, err)
		case <-done:
			return sox, multiErr
		}
	}
}

func checker(candidate netip.AddrPort, verified chan netip.AddrPort, errs chan error, working *atomic.Int64) {
	for working.Load() > 10 {
		time.Sleep(10 * time.Millisecond)
	}
	working.Add(1)
	defer func() {
		time.Sleep(10 * time.Millisecond)
		working.Add(-1)
	}()
	if !candidate.IsValid() {
		errs <- fmt.Errorf("invalid address/port combo: %s", candidate.String())
		return
	}
	if mulltest.TestModeEnabled() {
		addruri := mulltest.Init().Addr
		if addruri == "" {
			panic("no test server address")
		}
		serv := strings.TrimSuffix(strings.Split(addruri, "http://")[1], "/")
		candidate = netip.MustParseAddrPort(serv)
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
		working  = &atomic.Int64{}
	)
	working.Store(0)

	for _, prx := range sox {
		go checker(prx, verified, errs, working)
	}
	go func() {
		for !working.CompareAndSwap(0, 0) {
			time.Sleep(100 * time.Millisecond)
		}
		close(errs)
		close(verified)
	}()
	return verified, errs
}
