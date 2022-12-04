package mullvad

import (
	"sync"

	jsoniter "github.com/json-iterator/go"
	http "github.com/valyala/fasthttp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (mvs MullvadServer) String() string {
	return mvs.Hostname
}

type Checker struct {
	m          map[string]MullvadServer
	cachedSize int
	url        string
	*sync.RWMutex
}

func NewChecker() *Checker {
	r := &Checker{
		m:       make(map[string]MullvadServer),
		RWMutex: &sync.RWMutex{},
		url:     EndpointRelays,
	}
	return r
}

func (c *Checker) Slice() []MullvadServer {
	c.RLock()
	defer c.RUnlock()
	var servers []MullvadServer
	for _, server := range c.m {
		servers = append(servers, server)
	}
	return servers
}

func (c *Checker) Has(hostname string) bool {
	c.RLock()
	_, ok := c.m[hostname]
	c.RUnlock()
	return ok
}

func (c *Checker) Add(server MullvadServer) {
	c.Lock()
	c.m[server.Hostname] = server
	c.Unlock()
}

func (c *Checker) Get(hostname string) MullvadServer {
	c.RLock()
	defer c.RUnlock()
	return c.m[hostname]
}

func (c *Checker) clear() {
	for k := range c.m {
		delete(c.m, k)
	}
}

func getContentSize(url string) int {
	req := http.AcquireRequest()
	res := http.AcquireResponse()
	defer func() {
		http.ReleaseRequest(req)
		http.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent(useragent)
	req.Header.SetMethod(http.MethodHead)
	req.SetRequestURI(url)
	if err := http.Do(req, res); err != nil {
		return -1
	}
	return res.Header.ContentLength()
}

func (c *Checker) update() error {
	var serverSlice []MullvadServer
	if c.cachedSize > 0 {
		latestSize := getContentSize(c.url)
		if latestSize == c.cachedSize {
			return nil
		}
	}

	req := http.AcquireRequest()
	res := http.AcquireResponse()
	defer func() {
		http.ReleaseRequest(req)
		http.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent(useragent)
	req.Header.SetContentType("application/json")
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(c.url)
	if err := http.Do(req, res); err != nil {
		return err
	}
	if err := json.Unmarshal(res.Body(), &serverSlice); err != nil {
		return err
	}
	c.Lock()
	c.clear()
	for _, server := range serverSlice {
		c.m[server.Hostname] = server
	}
	c.cachedSize = res.Header.ContentLength()
	c.Unlock()
	return nil
}

func (c *Checker) GetRelays() ([]MullvadServer, error) {
	if err := c.update(); err != nil {
		return nil, err
	}
	return c.Slice(), nil
}
