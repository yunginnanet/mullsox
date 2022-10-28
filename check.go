package mullsox

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-multierror"
)

type MyIPDetails struct {
	V4 *IPDetails
	V6 *IPDetails
}

func CheckIP4(ctx context.Context, h *http.Client) (details *IPDetails, err error) {
	return checkIP(ctx, h, false)
}

func CheckIP6(ctx context.Context, h *http.Client) (details *IPDetails, err error) {
	return checkIP(ctx, h, true)
}

func CheckIP(ctx context.Context, h *http.Client) (*MyIPDetails, error) {
	type result struct {
		details *IPDetails
		ipv6    bool
	}

	var errGroup multierror.Group
	var resChan = make(chan result)

	check := func(resChan chan result, ipv6 bool) error {
		var err error
		var r = result{ipv6: ipv6}
		r.details, err = checkIP(ctx, h, r.ipv6)
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
		for {
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
	}()

	err = errGroup.Wait()
	err = err.(*multierror.Error).ErrorOrNil()
	close(resChan)

	return myip, err
}

func checkIP(ctx context.Context, h *http.Client, ipv6 bool) (details *IPDetails, err error) {
	var (
		resp   *http.Response
		cytes  []byte
		target string
	)
	switch ipv6 {
	case true:
		target = EndpointCheck6
	default:
		target = EndpointCheck4
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", target, nil)
	resp, err = h.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("bad status code from %s : %s", target, resp.Status)
		return
	}
	cytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(cytes, &details)
	return
}

// AmIMullvad checks if you are currently connecting through a Mullvad relay.
// Returns the mullvad server you are connected to if any, and any error that occured
func AmIMullvad(ctx context.Context) (*MullvadServer, error) {
	// v4, v6, err := CheckIP(ctx, http.DefaultClient)
	// m
	// if err != nil {
	// }
	return nil, nil
}
