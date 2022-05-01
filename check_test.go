package mullsox

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestCheckIP4(t *testing.T) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	v4, err := CheckIP4(ctx, http.DefaultClient)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	v4j, err4j := json.Marshal(v4)
	if err4j != nil {
		t.Fatalf("%s", err4j.Error())
	}
	t.Logf(string(v4j))
}

func TestCheckIP6(t *testing.T) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	v6, err := CheckIP6(ctx, http.DefaultClient)
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
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(15*time.Second))
	v4, v6, errs := CheckIP(ctx, http.DefaultClient)
	for _, err := range errs {
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
	}
	v4j, err4j := json.Marshal(v4)
	if err4j != nil {
		t.Fatalf("%s", err4j.Error())
	}
	v6j, err6j := json.Marshal(v6)
	if err6j != nil {
		t.Fatalf("%s", err6j.Error())
	}
	t.Logf(string(v4j))
	t.Logf(string(v6j))
}
