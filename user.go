package mullsox

type MullvadUser struct {
	ID             string                   `json:"id"`
	Account        int                      `json:"account"`
	WGPrivateKey   string                   `json:"private_key"`
	WireguardPorts map[int]*WireguardServer `json:"wireguard_ports"`
}

func NewMullvadUser(account int, id, privateKey string) (*MullvadUser, error) {
	k, err := encodeBase64ToHex(privateKey)
	if err != nil {
		return nil, err
	}
	mvu := &MullvadUser{
		ID:           id,
		Account:      account,
		WGPrivateKey: k,
	}
	return mvu, nil
}
