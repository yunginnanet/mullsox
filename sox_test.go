package mullsox

import (
	"testing"
)

func TestChecker_GetSOCKS(t *testing.T) {
	c := NewChecker()
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
}
