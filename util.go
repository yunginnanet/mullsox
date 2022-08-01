package mullsox

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

func encodeBase64ToHex(key string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", errors.New("invalid base64 string: " + key)
	}
	if len(decoded) != 32 {
		return "", errors.New("key should be 32 bytes: " + key)
	}
	return hex.EncodeToString(decoded), nil
}

const padding = "+-+-+-+-+-+-+-+-+"

var arrowLeft = padding[:len(padding)-2] + "> "
var arrowRight = " <" + padding[:len(padding)-2]

func prettyPrint(srvs []MullvadServer) { //goland:noinspection ALL
	for _, srv := range srvs {
		border := padding + strings.Repeat("-", len(srv.String())) + padding
		println(border + "\n" + arrowLeft + srv.String() + arrowRight + "\n" + border)
		pp, _ := json.MarshalIndent(srv, "", "\t")
		println(strings.ReplaceAll(string(pp), `"`, ""))
		println("\n+" + strings.Repeat("-+", (len(border)/2)-1) + "\n")
	}
}
