package mullsox

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
	// "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (mvs MullvadServer) String() string {
	return mvs.Hostname
}

type Relays []MullvadServer

func (r *Relays) Slice() []MullvadServer {
	return *r
}

func GetMullvadServers() (*Relays, error) {
	var servers = new(Relays)
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent("mulls0x/v0.0.1")
	req.Header.SetContentType("application/json")
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(EndpointRelays)
	if err := fasthttp.Do(req, res); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(res.Body(), servers); err != nil {
		return nil, err
	}
	return servers, nil
}
