package main

import (
	"context"
	"net/http"
	"os"

	"git.tcp.direct/kayos/mullsox"
)

func main() {
	current, err := mullsox.CheckIP4(context.Background(), http.DefaultClient)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	println("control group: " + current.IP)
	mvu, err := mullsox.NewMullvadUser(
		"default",
		os.Args[1],
		os.Args[2],
		os.Args[3],
	)
	srvs, err := mullsox.GetWireguardServers()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	for _, srv := range srvs {
		ip, err := mvu.GetIPv4Out(srv)
		if err != nil {
			println(err.Error())
			continue
		}
		println(srv.String() + ": " + ip)
	}
}
