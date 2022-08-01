package mullsox

type WireguardServer struct {
	Parent      *MullvadServer
	WGPublicKey string
	Leapfrog    int
	In4         string
	In6         string
	// Out addresses should be immutable
	Out4 *string
	Out6 *string
}

func (wgs WireguardServer) String() string {
	return wgs.Parent.Hostname
}
