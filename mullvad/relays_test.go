package mullvad

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"git.tcp.direct/kayos/mullsox/mulltest"
)

func TestGetMullvadServers(t *testing.T) {
	mt := mulltest.Init()
	mt.SetOpRelays()
	mt.SetIsMullvad()
	servers := NewChecker()

	update := func(srv *Checker) {
		err := srv.update()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		t.Logf("got %d servers for uri %s", len(srv.Slice()), srv.url)
	}

	t.Run("GetMullvadServers", func(t *testing.T) {
		update(servers)
		t.Log(spew.Sdump(servers.Slice()))
	})
	var last int
	var lastSlice []Server
	t.Run("GetMullvadServersCached", func(t *testing.T) {
		update(servers)
		update(servers)
		update(servers)
		update(servers)
		update(servers)
		update(servers)
		update(servers)
		last = servers.cachedSize
		lastSlice = servers.Slice()
	})
	t.Run("GetMullvadServersChanged", func(t *testing.T) {
		servers.url = servers.url + "/openvpn/"
		t.Logf("changing url to %s", servers.url)
		update(servers)
		if last == servers.cachedSize {
			t.Fatalf("expected %d to not equal %d", last, servers.cachedSize)
		}
		if len(servers.Slice()) == len(lastSlice) {
			t.Fatalf("expected %d to not equal %d", len(lastSlice), len(servers.Slice()))
		}
	})
}
