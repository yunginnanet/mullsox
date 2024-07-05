# mullsox

[![GoDoc](https://godoc.org/git.tcp.direct/kayos/prox5?status.svg)](https://pkg.go.dev/git.tcp.direct/kayos/mullsox) [![Go Report Card](https://goreportcard.com/badge/github.com/yunginnanet/prox5)](https://goreportcard.com/report/github.com/yunginnanet/mullsox) [![IRC](https://img.shields.io/badge/ircd.chat-%23tcpdirect-blue.svg)](ircs://ircd.chat:6697/#tcpdirect)

mullsox is an overengineered toolkit to work with the [mullvad](https://mullvad.net/) API. It's designed for use when already connected to a mullvad VPN endpoint. 

More specifically, it was built to help the user make use of all of the [SOCKS](https://mullvad.net/en/help/socks5-proxy/) proxies that are available via the internal `10.0.0.0/8` subnet while connected to mullvad servers. this allows you to utilize somewhere around 75-100 different outgoing IP addresses via SOCKS proxies all while only connected to one mullvad VPN server.

###  usage

```golang

package main

import (
	"time"

	"git.tcp.direct/kayos/mullsox"
	"git.tcp.direct/kayos/mullsox/mullvad"
)

func main() {
	fetcher := mullvad.NewChecker()
	socksCh, errCh := mullsox.GetAndVerifySOCKS(fetcher) 
	
	// or, with options:
	
	/* 
	socksCh, errCh := mullsox.GetAndVerifySOCKS(fetcher,
	    mullsox.CheckerWorkerLimit(25), 
	    mullsox.CheckerTimeout(3*time.Second),
	)
	*/
	
	go func() {
		for err := range errCh {
			println(err.Error())
		}
	}()
	for sock := range socksCh {
		println(sock.String())
	}
	fmt.Printf("loaded %d p")
}
```


##### 5 5 5 5 5

mullsox works great with [prox5](https://git.tcp.direct/kayos/prox5).
