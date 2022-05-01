package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const AimHost = "am.i.mullvad.net"
const Ipv4Endpoint = `https://ipv4.` + AimHost
const Ipv6Endpoint = `https://ipv6.` + AimHost

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
		cytes  []byte
		target string
	)
	switch ipv6 {
	case true:
		target = Ipv6Endpoint + "/json"
	default:
		target = Ipv4Endpoint + "/json"
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
	err = json.Unmarshal(cytes, &details)
	return
}

type MyIPDetails struct {
	Ip                    string  `json:"ip"`
	Country               string  `json:"country"`
	City                  string  `json:"city"`
	Longitude             float64 `json:"longitude"`
	Latitude              float64 `json:"latitude"`
	MullvadExitIp         bool    `json:"mullvad_exit_ip"`
	MullvadExitIpHostname string  `json:"mullvad_exit_ip_hostname"`
	MullvadServerType     string  `json:"mullvad_server_type"`
	Blacklisted           struct {
		Blacklisted bool `json:"blacklisted"`
		Results     []struct {
			Name        string `json:"name"`
			Link        string `json:"link"`
			Blacklisted bool   `json:"blacklisted"`
		} `json:"results"`
	} `json:"blacklisted"`
	Organization string `json:"organization"`
}

type MullvadServer struct {
	Hostname         string `json:"hostname"`
	CountryCode      string `json:"country_code"`
	CountryName      string `json:"country_name"`
	CityCode         string `json:"city_code"`
	CityName         string `json:"city_name"`
	Active           bool   `json:"active"`
	Owned            bool   `json:"owned"`
	Provider         string `json:"provider"`
	Ipv4AddrIn       string `json:"ipv4_addr_in"`
	Ipv6AddrIn       string `json:"ipv6_addr_in"`
	NetworkPortSpeed int    `json:"network_port_speed"`
	Pubkey           string `json:"pubkey"`
	MultihopPort     int    `json:"multihop_port"`
	SocksName        string `json:"socks_name"`
}
