package mullsox

import (
	"net/http"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (mvs MullvadServer) String() string {
	return mvs.Hostname
}

type Relays struct {
	m    map[string]MullvadServer
	size int
	*sync.RWMutex
}

func NewRelays() *Relays {
	r := &Relays{
		m:       make(map[string]MullvadServer),
		RWMutex: &sync.RWMutex{},
	}
	return r
}

func (r *Relays) Slice() []MullvadServer {
	r.RLock()
	defer r.RUnlock()
	var servers []MullvadServer
	for _, server := range r.m {
		servers = append(servers, server)
	}
	return servers
}

func (r *Relays) Has(hostname string) bool {
	r.RLock()
	_, ok := r.m[hostname]
	r.RUnlock()
	return ok
}

func (r *Relays) Add(server MullvadServer) {
	r.Lock()
	r.m[server.Hostname] = server
	r.Unlock()
}

func (r *Relays) Get(hostname string) MullvadServer {
	r.RLock()
	defer r.RUnlock()
	return r.m[hostname]
}

func GetMullvadServers() (*Relays, error) {
	var servers = NewRelays()
	var serverSlice []MullvadServer
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
	if err := json.Unmarshal(res.Body(), &serverSlice); err != nil {
		return nil, err
	}
	for _, server := range serverSlice {
		servers.Add(server)
	}
	return servers, nil
}
