package mullsox

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/netip"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func (mvu *MullvadUser) GetIPv4Out(server WireguardServer) (string, error) {
	var dnsipas []netip.Addr //nolint:prealloc
	for _, dnsip := range []string{mvu.WGDNS} {
		dnsipas = append(dnsipas, netip.MustParseAddr(dnsip))
	}
	var localipas []netip.Addr //nolint:prealloc
	for _, localip := range []string{mvu.WGIPv4} {
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
	// fwmark := binary.BigEndian.Uint32([]byte("[redacted]"))
	dev := device.NewDevice(tun, conn.NewDefaultBind(), device.NewLogger(device.LogLevelVerbose, ""))
	err = dev.IpcSet(fmt.Sprintf(`private_key=%s
fwmark=%d
public_key=%s
endpoint=%s:%d
allowed_ip=0.0.0.0/0`,
		mvu.WGPrivateKey, fwmark, server.WGPublicKey, server.In4, server.Leapfrog))

	if err != nil {
		return "", err
	}
	err = dev.Up()
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: tnet.DialContext,
			/*func(ctx context.Context, network, addr string) (net.Conn, error) {
				addr = strings.ReplaceAll(
					addr, "ipv4.am.i.mullvad.net", "193.138.218.116")
				ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Second*20))
				println("\x1b[32mDialing: ", addr, "\x1b[0m")
				return tnet.DialContext(ctx, network, addr)
			},*/
			/*			TLSClientConfig: &tls.Config{
							ServerName:         "ipv4.am.i.mullvad.net",
							InsecureSkipVerify: false,
						},
			*/},
	}
	details, err := CheckIP4(context.Background(), client)
	if err != nil {
		return "", err
	}
	return details.IP, nil
}
