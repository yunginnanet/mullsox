package mullsox

import (
	"testing"
)

func TestChecker_GetSOCKS(t *testing.T) {
	c := NewChecker()
	t.Run("GetSOCKS", func(t *testing.T) {
		if err := c.Update(); err != nil {
			t.Fatalf("%s", err.Error())
		}
		gotSox, err := c.GetSOCKS()
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
		if err := c.Update(); err != nil {
			t.Fatalf("%s", err.Error())
		}
		gotSox, errs := c.GetAndVerifySOCKS()
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
