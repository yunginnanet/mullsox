package mullvad

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"sync"

	http "github.com/valyala/fasthttp"

	"git.tcp.direct/kayos/mullsox/mulltest"
)

func (mvs Server) String() string {
	return mvs.Hostname
}

type Checker struct {
	m          map[string]Server
	cachedSize int
	url        string
	*sync.RWMutex
}

func NewChecker() *Checker {
	r := &Checker{
		m:       make(map[string]Server),
		RWMutex: &sync.RWMutex{},
		url:     EndpointRelays,
	}

	if mulltest.TestModeEnabled() {
		mt := mulltest.Init()
		mt.SetOpRelays()
		_, _ = os.Stderr.WriteString("running in test mode, using addr: " + mt.Addr + "\n")
		r.url = mulltest.Init().Addr
		if r.url == "" {
			panic("no test server address")
		}
	}
	return r
}

func (c *Checker) Slice() []Server {
	c.RLock()
	defer c.RUnlock()
	var servers []Server
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

func (c *Checker) Add(server Server) {
	c.Lock()
	key := server.Hostname
	key = strings.ToLower(key)
	if key == "" {
		panic("empty hostname")
	}
	c.m[server.Hostname] = server
	c.Unlock()
}

func (c *Checker) Get(hostname string) Server {
	hostname = strings.ToLower(hostname)
	hostname = strings.TrimSpace(hostname)
	found := c.Has(hostname)
	if !found {
		hostname = strings.Split(hostname, ".")[0]
		found = c.Has(hostname)
	}
	if !found {
		return Server{}
	}
	c.RLock()
	srv, _ := c.m[hostname]
	c.RUnlock()
	return srv
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
	if mulltest.TestModeEnabled() {
		current := mulltest.Init().OpState()
		defer mulltest.Init().StateOpMode.Store(current)
		mulltest.Init().SetOpRelays()
		if !strings.Contains(c.url, mulltest.Init().Addr) {
			u, _ := url.Parse(c.url)
			c.url = mulltest.Init().Addr + u.Path
		}
	}

	var serverSlice []Server
	if c.cachedSize > 0 {
		latestSize := getContentSize(c.url)
		if latestSize == c.cachedSize {
			return nil
		}
	}

	req := http.AcquireRequest()
	res := http.AcquireResponse()
	req.SetRequestURI(c.url)
	defer func() {
		http.ReleaseRequest(req)
		http.ReleaseResponse(res)
	}()
	req.Header.SetUserAgent(useragent)
	req.Header.SetContentType("application/json")
	req.Header.SetMethod(http.MethodGet)
	if mulltest.TestModeEnabled() {
		mulltest.Init().SetOpRelays()
		c.url = mulltest.Init().Addr
	}

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

func (c *Checker) GetRelays() ([]Server, error) {
	if err := c.update(); err != nil {
		return nil, err
	}
	return c.Slice(), nil
}
