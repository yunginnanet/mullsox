package mullsox

import (
	"testing"

	"git.tcp.direct/kayos/mullsox/mulltest"
	"git.tcp.direct/kayos/mullsox/mullvad"
)

func TestChecker_GetSOCKS(t *testing.T) {
	mt := mulltest.Init()
	mt.SetOpRelays()

	t.Logf("test server: %s", mt.Addr)

	c := mullvad.NewChecker()
	t.Run("GetSOCKS", func(t *testing.T) {
		gotSox, err := GetSOCKS(c)
		if err != nil {
			t.Error(err)
		}
		if len(gotSox) == 0 {
			t.Error("expected non-zero length")
		}
		t.Logf("got %d socks", len(gotSox))
		t.Logf("%v", gotSox)
	})

	t.Run("GetAndVerifySOCKS", func(t *testing.T) {
		gotSox, errs := GetAndVerifySOCKS(c)
		count := 0
		for sox := range gotSox {
			select {
			case err := <-errs:
				if err != nil {
					t.Error(err)
				}
			default:
				t.Logf("got verified: %s", sox.String())
				count++
			}
		}
		t.Logf("got %d active mullvad SOCKS5 servers", count)
	})
}
