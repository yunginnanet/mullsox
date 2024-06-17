package mulltest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type OpState int

const (
	OpNull OpState = iota
	OpIsMullvad
	OpRelays
)

const (
	replyIsMullvadTrue  = `{"ip":"146.70.116.130","country":"The Nation of Fuck","city":"Vienna","longitude":-69.6969,"latitude":69.6969,"mullvad_exit_ip":true,"mullvad_exit_ip_hostname":"at-vie-ovpn-002","mullvad_server_type":"OpenVPN","blacklisted":{"blacklisted":false,"results":[]},"organization":"M247"}`
	replyIsMullvadFalse = `{"ip":"127.0.0.1","country":"","city":"","longitude":0,"latitude":0,"mullvad_exit_ip":false,"mullvad_exit_ip_hostname":"","mullvad_server_type":"","blacklisted":{"blacklisted":false,"results":[]},"organization":""}`
	testDataRelays      = `[{"hostname":"al-tia-wg-001","country_code":"al","country_name":"Albania","city_code":"tia","city_name":"Tirana","fqdn":"al-tia-wg-001.relays.mullvad.net","active":true,"owned":false,"provider":"iRegister","ipv4_addr_in":"31.171.153.66","ipv6_addr_in":"2a04:27c0:0:3::f001","network_port_speed":10,"stboot":true,"pubkey":"bPfJDdgBXlY4w3ACs68zOMMhLUbbzktCKnLOFHqbxl4=","multihop_port":3155,"socks_name":"al-tia-wg-socks5-001.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"al-tia-wg-002","country_code":"al","country_name":"Albania","city_code":"tia","city_name":"Tirana","fqdn":"al-tia-wg-002.relays.mullvad.net","active":true,"owned":false,"provider":"iRegister","ipv4_addr_in":"31.171.154.50","ipv6_addr_in":"2a04:27c0:0:4::f001","network_port_speed":10,"stboot":true,"pubkey":"/wPQafVa/60OIp8KqhC1xTTG+nQXZF17uo8XfdUnz2E=","multihop_port":3212,"socks_name":"al-tia-wg-socks5-002.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"at-vie-ovpn-002","country_code":"at","country_name":"Austria","city_code":"vie","city_name":"Vienna","fqdn":"at-vie-ovpn-002.relays.mullvad.net","active":false,"owned":false,"provider":"M247","ipv4_addr_in":"146.70.116.226","ipv6_addr_in":"2001:ac8:29:86::2f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"at-vie-wg-001","country_code":"at","country_name":"Austria","city_code":"vie","city_name":"Vienna","fqdn":"at-vie-wg-001.relays.mullvad.net","active":true,"owned":false,"provider":"M247","ipv4_addr_in":"146.70.116.98","ipv6_addr_in":"2001:ac8:29:84::a01f","network_port_speed":10,"stboot":true,"pubkey":"TNrdH73p6h2EfeXxUiLOCOWHcjmjoslLxZptZpIPQXU=","multihop_port":3543,"socks_name":"at-vie-wg-socks5-001.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"at-vie-wg-002","country_code":"at","country_name":"Austria","city_code":"vie","city_name":"Vienna","fqdn":"at-vie-wg-002.relays.mullvad.net","active":true,"owned":false,"provider":"M247","ipv4_addr_in":"146.70.116.130","ipv6_addr_in":"2001:ac8:29:85::a02f","network_port_speed":10,"stboot":true,"pubkey":"ehXBc726YX1N6Dm7fDAVMG5cIaYAFqCA4Lbpl4VWcWE=","multihop_port":3544,"socks_name":"at-vie-wg-socks5-002.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"at-vie-wg-003","country_code":"at","country_name":"Austria","city_code":"vie","city_name":"Vienna","fqdn":"at-vie-wg-003.relays.mullvad.net","active":true,"owned":false,"provider":"M247","ipv4_addr_in":"146.70.116.162","ipv6_addr_in":"2001:ac8:29:86::a03f","network_port_speed":10,"stboot":true,"pubkey":"ddllelPu2ndjSX4lHhd/kdCStaSJOQixs9z551qN6B8=","multihop_port":3545,"socks_name":"at-vie-wg-socks5-003.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-adl-ovpn-301","country_code":"au","country_name":"Australia","city_code":"adl","city_name":"Adelaide","fqdn":"au-adl-ovpn-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.214.20.146","ipv6_addr_in":"2404:f780:0:dee::c1f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-adl-ovpn-302","country_code":"au","country_name":"Australia","city_code":"adl","city_name":"Adelaide","fqdn":"au-adl-ovpn-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.214.20.162","ipv6_addr_in":"2404:f780:0:def::c2f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-adl-wg-301","country_code":"au","country_name":"Australia","city_code":"adl","city_name":"Adelaide","fqdn":"au-adl-wg-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.214.20.50","ipv6_addr_in":"2404:f780:0:deb::c1f","network_port_speed":10,"stboot":true,"pubkey":"rm2hpBiN91c7reV+cYKlw7QNkYtME/+js7IMyYBB2Aw=","multihop_port":3099,"socks_name":"au-adl-wg-socks5-301.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-adl-wg-302","country_code":"au","country_name":"Australia","city_code":"adl","city_name":"Adelaide","fqdn":"au-adl-wg-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.214.20.130","ipv6_addr_in":"2404:f780:0:dec::c2f","network_port_speed":10,"stboot":true,"pubkey":"e4jouH8n4e8oyi/Z7d6lJLd6975hlPZmnynJeoU+nWM=","multihop_port":3156,"socks_name":"au-adl-wg-socks5-302.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-bne-ovpn-301","country_code":"au","country_name":"Australia","city_code":"bne","city_name":"Brisbane","fqdn":"au-bne-ovpn-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.216.220.50","ipv6_addr_in":"2404:f780:4:dee::1f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-bne-ovpn-302","country_code":"au","country_name":"Australia","city_code":"bne","city_name":"Brisbane","fqdn":"au-bne-ovpn-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.216.220.66","ipv6_addr_in":"2404:f780:4:def::2f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-bne-wg-301","country_code":"au","country_name":"Australia","city_code":"bne","city_name":"Brisbane","fqdn":"au-bne-wg-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.216.220.18","ipv6_addr_in":"2404:f780:4:deb::f001","network_port_speed":10,"stboot":true,"pubkey":"1H/gj8SVNebAIEGlvMeUVC5Rnf274dfVKbyE+v5G8HA=","multihop_port":3220,"socks_name":"au-bne-wg-socks5-301.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-bne-wg-302","country_code":"au","country_name":"Australia","city_code":"bne","city_name":"Brisbane","fqdn":"au-bne-wg-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.216.220.34","ipv6_addr_in":"2404:f780:4:dec::a02f","network_port_speed":10,"stboot":true,"pubkey":"z+JG0QA4uNd/wRTpjCqn9rDpQsHKhf493omqQ5rqYAc=","multihop_port":3221,"socks_name":"au-bne-wg-socks5-302.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-mel-ovpn-301","country_code":"au","country_name":"Australia","city_code":"mel","city_name":"Melbourne","fqdn":"au-mel-ovpn-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.229.82","ipv6_addr_in":"2406:d501:f:def::1f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-mel-ovpn-302","country_code":"au","country_name":"Australia","city_code":"mel","city_name":"Melbourne","fqdn":"au-mel-ovpn-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.229.98","ipv6_addr_in":"2406:d501:f:dee::2f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-mel-wg-301","country_code":"au","country_name":"Australia","city_code":"mel","city_name":"Melbourne","fqdn":"au-mel-wg-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.229.50","ipv6_addr_in":"2406:d501:f:deb::a01f","network_port_speed":10,"stboot":true,"pubkey":"jUMZWFOgoFGhZjBAavE6jW8VgnnNpL4KUiYFYjc1fl8=","multihop_port":3307,"socks_name":"au-mel-wg-socks5-301.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-mel-wg-302","country_code":"au","country_name":"Australia","city_code":"mel","city_name":"Melbourne","fqdn":"au-mel-wg-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.229.66","ipv6_addr_in":"2406:d501:f:dec::a02f","network_port_speed":10,"stboot":true,"pubkey":"npTb63jWEaJToBfn0B1iVNbnLXEwwlus5SsolsvUhgU=","multihop_port":3308,"socks_name":"au-mel-wg-socks5-302.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]},{"hostname":"au-per-ovpn-301","country_code":"au","country_name":"Australia","city_code":"per","city_name":"Perth","fqdn":"au-per-ovpn-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.231.82","ipv6_addr_in":"2404:f780:8:def::1f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-per-ovpn-302","country_code":"au","country_name":"Australia","city_code":"per","city_name":"Perth","fqdn":"au-per-ovpn-302.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.231.98","ipv6_addr_in":"2404:f780:8:dee::2f","network_port_speed":10,"stboot":true,"type":"openvpn","status_messages":[]},{"hostname":"au-per-wg-301","country_code":"au","country_name":"Australia","city_code":"per","city_name":"Perth","fqdn":"au-per-wg-301.relays.mullvad.net","active":true,"owned":false,"provider":"hostuniversal","ipv4_addr_in":"103.108.231.50","ipv6_addr_in":"2404:f780:8:deb::a01f","network_port_speed":10,"stboot":true,"pubkey":"hQXsNk/9R2We0pzP1S9J3oNErEu2CyENlwTdmDUYFhg=","multihop_port":3309,"socks_name":"au-per-wg-socks5-301.relays.mullvad.net","socks_port":1080,"daita":false,"type":"wireguard","status_messages":[]}]`
)

var (
	existing = &atomic.Value{}
	noRace   = sync.Mutex{}
)

func TestModeEnabled() bool {
	return existing.Load() != nil
}

type Tester struct {
	Addr           string
	startOnce      sync.Once
	StateIsMullvad *atomic.Bool
	StateOpMode    *atomic.Value
}

func Init() *Tester {
	noRace.Lock()
	defer noRace.Unlock()
	if existing.Load() != nil {
		return existing.Load().(*Tester)
	}
	t := &Tester{
		StateIsMullvad: &atomic.Bool{},
		StateOpMode:    &atomic.Value{},
	}
	t.StateOpMode.Store(OpNull)
	t.StateIsMullvad.Store(true)
	t.StartServer()
	existing.Store(t)
	return t
}

func (t *Tester) OpState() OpState {
	return t.StateOpMode.Load().(OpState)
}

func (t *Tester) SetOpIsMullvad() {
	t.StateOpMode.Store(OpIsMullvad)
}

func (t *Tester) SetOpRelays() {
	t.StateOpMode.Store(OpRelays)
}

func (t *Tester) SetIsNotMullvad() {
	t.StateIsMullvad.Store(false)
}

func (t *Tester) SetIsMullvad() {
	t.StateIsMullvad.Store(true)
}

func (t *Tester) StartServer() {
	t.startOnce.Do(func() {
		_, _ = os.Stderr.WriteString("starting test server\n")
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch t.StateOpMode.Load().(OpState) {
			case OpIsMullvad:
				w.WriteHeader(http.StatusOK)
				switch t.StateIsMullvad.Load() {
				case true:
					_, _ = w.Write([]byte(replyIsMullvadTrue))
				default:
					_, _ = w.Write([]byte(replyIsMullvadFalse))
				}
			case OpRelays:
				var relays []map[string]interface{}
				_ = json.Unmarshal([]byte(testDataRelays), &relays)

				if len(relays) == 0 {
					panic("no relays found in static data")
				}

				if strings.Contains(r.RequestURI, "openvpn") {
					newRelays := make([]map[string]interface{}, 0, len(relays))
					for _, relay := range relays {
						if relay["type"] == "openvpn" {
							newRelays = append(newRelays, relay)
						}
					}
					relays = newRelays
				}

				dat, _ := json.Marshal(relays)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Content-Length", strconv.Itoa(len(dat)))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(dat)
			default:
				w.WriteHeader(http.StatusTeapot)
				_, _ = w.Write([]byte(`{"EEEEE": "` + strings.Repeat("e", 500) + `"}`))
				panic("invalid test mode")
			}
		}))
		if err := os.Setenv("MULLSOX_TEST_EP", testServer.URL); err != nil {
			panic(err)
		}
		t.Addr = testServer.URL
	})
	if t.Addr == "" {
		panic("failed to start test server")
	}

}
