package mullsox

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic/decoder"
)

func CheckIP4(ctx context.Context, h *http.Client) (details *MyIPDetails, err error) {
	return checkIP(ctx, h, false)
}

func CheckIP6(ctx context.Context, h *http.Client) (details *MyIPDetails, err error) {
	return checkIP(ctx, h, true)
}

func CheckIP(ctx context.Context, h *http.Client) (v4details *MyIPDetails, v6details *MyIPDetails, errs []error) {
	type result struct {
		details *MyIPDetails
		ipv6    bool
		err     error
	}

	var (
		resChan  = make(chan result)
		finished = 0
	)

	check := func(resChan chan result, ipv6 bool) {
		var r = result{ipv6: ipv6}
		r.details, r.err = checkIP(ctx, h, r.ipv6)
		select {
		case <-ctx.Done():
			r.err = ctx.Err()
			resChan <- r
		case resChan <- r:
			//
		}
	}

	go check(resChan, false)
	go check(resChan, true)

	for {
		if finished == 2 {
			break
		}
		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return
		case res := <-resChan:
			switch {
			case res.err != nil:
				prefix := "(v4)"
				if res.ipv6 {
					prefix = "(v6)"
				}
				errs = append(errs, fmt.Errorf("%s %s", prefix, res.err.Error()))
			case res.ipv6:
				v6details = res.details
			case !res.ipv6:
				v4details = res.details
			default:
				panic("malformed result")
			}
			finished++
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
	return
}

func checkIP(ctx context.Context, h *http.Client, ipv6 bool) (details *MyIPDetails, err error) {
	var (
		resp   *http.Response
		target string
	)
	switch ipv6 {
	case true:
		target = EndpointCheck6 + "/json"
	default:
		target = EndpointCheck4 + "/json"
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
	fizz := decoder.NewStreamDecoder(resp.Body)
	err = fizz.Decode(&details)
	return
}
