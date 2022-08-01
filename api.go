package mullsox

import (
	"net/http"

	json "github.com/pquerna/ffjson/ffjson"
	"github.com/valyala/fasthttp"
)

const (
	baseDomain     = "mullvad.net"
	baseEndpoint   = "am.i." + baseDomain
	EndpointCheck4 = `https://ipv4.` + baseEndpoint
	EndpointCheck6 = `https://ipv6.` + baseEndpoint
	EndpointRelays = `https://api.` + baseDomain + `/www/relays/all/`
)

type MyIPDetails struct {
	IP                    string  `json:"ip"`
	Country               string  `json:"country"`
	City                  string  `json:"city"`
	Longitude             float64 `json:"longitude"`
	Latitude              float64 `json:"latitude"`
	MullvadExitIP         bool    `json:"mullvad_exit_ip"`
	MullvadExitIPHostname string  `json:"mullvad_exit_ip_hostname"`
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
	Hostname             string `json:"hostname"`
	CountryCode          string `json:"country_code"`
	CountryName          string `json:"country_name"`
	CityCode             string `json:"city_code"`
	CityName             string `json:"city_name"`
	Active               bool   `json:"active"`
	Owned                bool   `json:"owned"`
	Provider             string `json:"provider"`
	Ipv4AddrIn           string `json:"ipv4_addr_in"`
	Ipv6AddrIn           string `json:"ipv6_addr_in"`
	NetworkPortSpeed     int    `json:"network_port_speed"`
	Type                 string `json:"type"`
	Pubkey               string `json:"pubkey,omitempty"`
	MultihopPort         int    `json:"multihop_port,omitempty"`
	SocksName            string `json:"socks_name,omitempty"`
	SSHFingerprintSHA256 string `json:"ssh_fingerprint_sha256,omitempty"`
	SSHFingerprintMD5    string `json:"ssh_fingerprint_md5,omitempty"`
}

func (mvs MullvadServer) String() string {
	return mvs.Hostname
}

type relays []MullvadServer

func GetMullvadServers() (*relays, error) {
	var servers = new(relays)
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent("mulls0x/v0.0.1")
	req.Header.SetContentType("application/json")
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(EndpointRelays)
	if err := fasthttp.Do(req, res); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(res.Body(), servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func GetWireguardServers() ([]WireguardServer, error) {
	srvs, err := GetMullvadServers()
	if err != nil {
		return nil, err
	}
	return srvs.getWireguards()
}

func (servers *relays) getWireguards() ([]WireguardServer, error) {
	var wgs []WireguardServer
	for _, srv := range *servers {
		if srv.Type != "wireguard" {
			continue
		}
		pub, err := encodeBase64ToHex(srv.Pubkey)
		if err != nil {
			pub = srv.Pubkey
		}
		wgs = append(wgs, WireguardServer{
			Parent:      &srv,
			WGPublicKey: pub,
			In4:         srv.Ipv4AddrIn,
			In6:         srv.Ipv6AddrIn,
		})
	}
	return wgs, nil
}
