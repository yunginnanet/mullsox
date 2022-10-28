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

type Checker struct {
	m    map[string]MullvadServer
	size int
	url  string
	*sync.RWMutex
}

func NewRelays() *Checker {
	r := &Checker{
		m:       make(map[string]MullvadServer),
		RWMutex: &sync.RWMutex{},
		url:     EndpointRelays,
	}
	return r
}

func (r *Checker) Slice() []MullvadServer {
	r.RLock()
	defer r.RUnlock()
	var servers []MullvadServer
	for _, server := range r.m {
		servers = append(servers, server)
	}
	return servers
}

func (r *Checker) Has(hostname string) bool {
	r.RLock()
	_, ok := r.m[hostname]
	r.RUnlock()
	return ok
}

func (r *Checker) Add(server MullvadServer) {
	r.Lock()
	r.m[server.Hostname] = server
	r.Unlock()
}

func (r *Checker) Get(hostname string) MullvadServer {
	r.RLock()
	defer r.RUnlock()
	return r.m[hostname]
}

func (r *Checker) clear() {
	for k := range r.m {
		delete(r.m, k)
	}
}

func getContentSize(url string) int {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent(useragent)
	req.Header.SetMethod(http.MethodHead)
	req.SetRequestURI(url)
	if err := fasthttp.Do(req, res); err != nil {
		return -1
	}
	return res.Header.ContentLength()
}

func (r *Checker) Update() error {
	var serverSlice []MullvadServer
	if r.size > 0 {
		current := getContentSize(r.url)
		if current == r.size {
			return nil
		}
	}

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent(useragent)
	req.Header.SetContentType("application/json")
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(r.url)
	if err := fasthttp.Do(req, res); err != nil {
		return err
	}
	if err := json.Unmarshal(res.Body(), &serverSlice); err != nil {
		return err
	}
	r.Lock()
	r.clear()
	for _, server := range serverSlice {
		r.m[server.Hostname] = server
	}
	r.size = res.Header.ContentLength()
	r.Unlock()
	return nil
}
