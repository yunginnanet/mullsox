package mullsox

type MullvadUser struct {
	ID             string                   `json:"id"`
	Account        int                      `json:"account"`
	WGPrivateKey   string                   `json:"private_key"`
	WGIPv4         string                   `json:"ipv4"`
	WGDNS          string                   `json:"dns"`
	WireguardPorts map[int]*WireguardServer `json:"wireguard_ports"`
}

func NewMullvadUser(id, privateKey, mvip, dns string) (*MullvadUser, error) {
	k, err := encodeBase64ToHex(privateKey)
	if err != nil {
		return nil, err
	}
	mvu := &MullvadUser{
		ID:           id,
		WGIPv4:       mvip,
		WGDNS:        dns,
		WGPrivateKey: k,
		// Account:      account,
	}
	return mvu, nil
}
