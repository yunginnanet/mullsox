package mullsox

import (
	"testing"
)

func TestGetMullvadServers(t *testing.T) {
	servers := NewChecker()

	update := func() {
		err := servers.Update()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		t.Logf("got %d servers", len(servers.Slice()))
	}

	t.Run("GetMullvadServers", func(t *testing.T) {
		update()
		// t.Logf(spew.Sdump(servers.Slice()))
	})
	var last int
	var lastSlice []MullvadServer
	t.Run("GetMullvadServersCached", func(t *testing.T) {
		update()
		update()
		update()
		update()
		update()
		update()
		update()
		last = servers.cachedSize
		lastSlice = servers.Slice()
	})
	t.Run("GetMullvadServersChanged", func(t *testing.T) {
		servers.url = "https://api.mullvad.net/www/relays/openvpn/"
		update()
		if last == servers.cachedSize {
			t.Fatalf("expected %d to not equal %d", last, servers.cachedSize)
		}
		if len(servers.Slice()) == len(lastSlice) {
			t.Fatalf("expected %d to not equal %d", len(lastSlice), len(servers.Slice()))
		}
	})
}
