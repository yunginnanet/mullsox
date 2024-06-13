package mullvad

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	http "github.com/valyala/fasthttp"
)

var ErrNotMullvad = errors.New("your traffic is not being tunneled through mullvad")

type MyIPDetails struct {
	V4 *IPDetails `json:"ipv4,omitempty"`
	V6 *IPDetails `json:"ipv6,omitempty"`
}

func CheckIP4() (details *IPDetails, err error) {
	return checkIP(false)
}

func CheckIP6() (details *IPDetails, err error) {
	return checkIP(true)
}

func CheckIP(ctx context.Context) (*MyIPDetails, error) {
	type result struct {
		details *IPDetails
		ipv6    bool
	}

	var errGroup multierror.Group
	var resChan = make(chan result)

	check := func(resChan chan result, ipv6 bool) error {
		var err error
		var r = result{ipv6: ipv6}
		r.details, err = checkIP(r.ipv6)
		if err != nil {
			if r.ipv6 {
				err = fmt.Errorf("error checking ipv6: %w", err)
			} else {
				err = fmt.Errorf("error checking ipv4: %w", err)
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case resChan <- r:
			return nil
		}
	}

	errGroup.Go(func() error {
		return check(resChan, false)
	})
	errGroup.Go(func() error {
		return check(resChan, true)
	})

	var myip = new(MyIPDetails)

	var err error

	go func() {
		collect := func() {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-resChan:
				switch {
				case res.ipv6:
					myip.V6 = res.details
				case !res.ipv6:
					myip.V4 = res.details
				case !ok:
					return
				default:
					panic("malformed result")
				}
			}
		}
		for n := 0; n < 2; n++ {
			collect()
		}
	}()

	err = errGroup.Wait()
	err = err.(*multierror.Error).ErrorOrNil()
	close(resChan)

	return myip, err
}

// EnableIPV6 reenables ipv6 for `AmIMullvad` and `CheckIP`. As of writing (1718243316), mullvad brok the endpoints for ipv6.am.i.mullvad entirely. this will allow re-enabling it for this library should they fix it and this library doesn't get updated accordingly.
//
// To toggle: set `MULLSOX_ENABLE_V6` in your environment to any value
var EnableIPv6 = false

func init() {
	if os.Getenv("MULLSOX_ENABLE_V6") != "" {
		EnableIPv6 = true
	}
}

func checkIP(ipv6 bool) (details *IPDetails, err error) {
	var target string
	switch ipv6 {
	case true:
		if !EnableIPv6 {
			return &IPDetails{}, nil
		}
		target = EndpointCheck6
	default:
		target = EndpointCheck4
	}
	req := http.AcquireRequest()
	res := http.AcquireResponse()
	defer func() {
		http.ReleaseRequest(req)
		http.ReleaseResponse(res)
	}()
	req.SetRequestURI(target)
	req.Header.SetMethod(http.MethodGet)
	req.Header.SetUserAgent(useragent)
	client := http.Client{}
	client.DialDualStack = true

	err = client.DoTimeout(req, res, 15*time.Second)
	if err != nil {
		return
	}
	if res.StatusCode() != http.StatusOK {
		err = fmt.Errorf("got status code %d", res.StatusCode())
		return
	}

	err = json.Unmarshal(res.Body(), &details)
	return
}

// AmIMullvad checks if you are currently connecting through a Mullvad relay.
// Returns the mullvad server you are connected to if any, and any error that occured
//
//goland:noinspection GoNilness
func (c *Checker) AmIMullvad(ctx context.Context) ([]MullvadServer, error) {
	me, err := CheckIP(ctx)
	if me == nil || (me.V4 == nil && me.V6 == nil) {
		return []MullvadServer{}, ErrNotMullvad
	}
	if me.V4 != nil && !me.V4.MullvadExitIP {
		return []MullvadServer{}, err
	}
	//	if me.V6 != nil && !me.V6.MullvadExitIP {
	//		return []MullvadServer{}, err
	//	}

	err = c.update()
	if err != nil {
		return []MullvadServer{}, err
	}
	servs := make([]MullvadServer, 0, 2)

	isMullvad := false
	if me.V4 != nil && me.V4.MullvadExitIP {
		isMullvad = true
		if c.Has(me.V4.MullvadExitIPHostname) {
			servs = append(servs, c.Get(me.V4.MullvadExitIPHostname))
		}
	}
	if me.V6 != nil && me.V6.MullvadExitIP {
		isMullvad = true
		if c.Has(me.V6.MullvadExitIPHostname) {
			servs = append(servs, c.Get(me.V6.MullvadExitIPHostname))
		}
	}
	nils := 0
	for _, srv := range servs {
		if srv.Hostname == "" {
			nils++
		}
	}
	if nils == 2 || nils == len(servs) || len(servs) == 0 {
		switch isMullvad {
		case true:
			return servs,
				errors.New("could not find mullvad server in relay list, but you are connected to a mullvad exit ip")
		case false:
			return servs, ErrNotMullvad
		}
	}

	return servs, nil
}
