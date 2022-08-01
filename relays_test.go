package mullsox

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetRelays(t *testing.T) {
	srvs, err := GetMullvadServers()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	// prettyPrint(srvs)
	t.Run("GetWireguardServers", func(t *testing.T) {
		wgs, err := srvs.getWireguards()
		if err != nil {
			t.Fatalf("%s", err.Error())
		}
		for _, wg := range wgs {
			pp, _ := json.MarshalIndent(wg, "", "\t")
			println(strings.ReplaceAll(string(pp), `"`, ""))
		}
	})
}
