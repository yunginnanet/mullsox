package mullsox

import "testing"

func TestGetMullvadServers(t *testing.T) {
	servers, err := GetMullvadServers()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	t.Logf("got %d servers", len(servers.Slice()))
}
