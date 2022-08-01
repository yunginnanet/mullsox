package mullsox

import (
	"io"
	"log"
	"net/http"
	"net/netip"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func (mvu *MullvadUser) getIP(localips []string, dnsips []string, server WireguardServer) (string, error) {
	var dnsipas []netip.Addr
	for _, dnsip := range dnsips {
		dnsipas = append(dnsipas, netip.MustParseAddr(dnsip))
	}
	var localipas []netip.Addr
	for _, localip := range localips {
		localipas = append(localipas, netip.MustParseAddr(localip))
	}
	tun, tnet, err := netstack.CreateNetTUN(
		// VPN Adapter IPs
		localipas,
		// DNS
		dnsipas,
		// MTU
		1420,
	)
	if err != nil {
		log.Panic(err)
	}
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelVerbose, ""))
	err = dev.IpcSet(
		`private_key=` + mvu.WGPrivateKey +
			`public_key=` + server.WGPublicKey +
			`endpoint=` + server.In4 +
			`allowed_ip=0.0.0.0/0`,
	)
	if err != nil {
		return "", err
	}
	err = dev.Up()
	if err != nil {
		return "", err
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: tnet.DialContext,
		},
	}
	resp, err := client.Get("https://tcp.ac/ip")
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
