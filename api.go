package mullsox

import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	baseDomain   = "mullvad.net"
	baseEndpoint = "am.i." + baseDomain
	ipv4Endpoint = `https://ipv4.` + baseEndpoint
	ipv6Endpoint = `https://ipv6.` + baseEndpoint
	servEndpoint = `https://api.` + baseDomain + `www/relays/all/`
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
	Hostname             string        `json:"hostname"`
	CountryCode          string        `json:"country_code"`
	CountryName          string        `json:"country_name"`
	CityCode             string        `json:"city_code"`
	CityName             string        `json:"city_name"`
	Active               bool          `json:"active"`
	Owned                bool          `json:"owned"`
	Provider             string        `json:"provider"`
	Ipv4AddrIn           string        `json:"ipv4_addr_in"`
	Ipv6AddrIn           *string       `json:"ipv6_addr_in"`
	NetworkPortSpeed     int           `json:"network_port_speed"`
	Type                 string        `json:"type"`
	StatusMessages       []interface{} `json:"status_messages"`
	Pubkey               string        `json:"pubkey,omitempty"`
	MultihopPort         int           `json:"multihop_port,omitempty"`
	SocksName            string        `json:"socks_name,omitempty"`
	SshFingerprintSha256 string        `json:"ssh_fingerprint_sha256,omitempty"`
	SshFingerprintMd5    string        `json:"ssh_fingerprint_md5,omitempty"`
}
