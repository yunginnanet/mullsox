package mullsox

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestCheckIP4(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	v4, err := CheckIP4(ctx, http.DefaultClient)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	v4j, err4j := json.Marshal(v4)
	if err4j != nil {
		t.Fatalf("%s", err4j.Error())
	}
	t.Logf(string(v4j))
	cancel()
}

func TestCheckIP6(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	v6, err := CheckIP6(ctx, http.DefaultClient)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	v6j, err6j := json.Marshal(v6)
	if err6j != nil {
		t.Fatalf("%s", err6j.Error())
	}
	t.Logf(string(v6j))
	cancel()
}

func TestCheckIPConcurrent(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	me, err := CheckIP(ctx, http.DefaultClient)
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
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	am, err := AmIMullvad(ctx, http.DefaultClient)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	indented, err := json.MarshalIndent(am, "", "  ")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	t.Logf(string(indented))
	cancel()
}
