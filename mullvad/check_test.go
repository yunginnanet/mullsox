package mullvad

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"git.tcp.direct/kayos/mullsox/mulltest"
)

var tester = mulltest.Init()

func TestCheckIP4(t *testing.T) {
	tester.SetOpIsMullvad()

	t.Run("is mullvad", func(t *testing.T) {
		tester.SetIsMullvad()
		v4, err := CheckIP4()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		v4j, err4j := json.Marshal(v4)
		if err4j != nil {
			t.Fatalf("%s", err4j.Error())
		}
		t.Logf(string(v4j))
	})

	t.Run("is not mullvad", func(t *testing.T) {
		tester.SetIsNotMullvad()
		v4, err := CheckIP4()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		v4j, err4j := json.Marshal(v4)
		if err4j != nil {
			t.Fatalf("%s", err4j.Error())
		}
		t.Logf(string(v4j))
	})
}

func TestCheckIP6(t *testing.T) {
	t.Skip("skipping ip6 check as mullvad seems to have broken it")
	tester.SetOpIsMullvad()
	v6, err := CheckIP6()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	v6j, err6j := json.Marshal(v6)
	if err6j != nil {
		t.Fatalf("%s", err6j.Error())
	}
	t.Logf(string(v6j))
}

func TestCheckIPConcurrent(t *testing.T) {
	t.Skip("skipping as ipv6 is broken on mullvad's end for the check")
	tester.SetOpIsMullvad()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	me, err := CheckIP(ctx)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	v4j, err4j := json.Marshal(me.V4)
	if err4j != nil {
		t.Fatalf("%s", err4j.Error())
	}
	v6j, err6j := json.Marshal(me.V6)
	if err6j != nil {
		t.Fatalf("%s", err6j.Error())
	}
	unmarshaled := &MyIPDetails{}
	unv4 := &IPDetails{}
	unv6 := &IPDetails{}

	if err = json.Unmarshal(v4j, unv4); err != nil {
		t.Fatalf("%s", err.Error())
	}
	if err = json.Unmarshal(v6j, unv6); err != nil {
		t.Fatalf("%s", err.Error())
	}
	unmarshaled.V4 = unv4
	unmarshaled.V6 = unv6

	t.Logf(spew.Sdump(unmarshaled.V4))
	t.Logf(spew.Sdump(unmarshaled.V6))
	cancel()
}

func TestAmIMullvad(t *testing.T) {
	tester.SetOpIsMullvad()

	t.Run("is mullvad", func(t *testing.T) {
		tester.SetIsMullvad()
		servers := NewChecker()
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
		am, err := servers.AmIMullvad(ctx)
		if err != nil {
			t.Errorf("%s", err.Error())
		}
		if err != nil {
			t.Errorf("failed is mullvad check: %s", err.Error())
		}
		if len(am) == 0 {
			t.Errorf("expected non-zero length")
		}
		if len(am) > 0 && am[0].Hostname == "" {
			t.Errorf("expected hostname to be set")
		}
		cancel()
	})

	t.Run("is not mullvad", func(t *testing.T) {
		tester.SetIsNotMullvad()
		servers := NewChecker()
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
		am, err := servers.AmIMullvad(ctx)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if len(am) != 0 {
			t.Errorf("expected zero length")
		}
		cancel()
	})
}
