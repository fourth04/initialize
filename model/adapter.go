// Package model provides ...
package model

type Adapter struct {
	name       string
	flags      []string
	mtu        int
	inet       string
	netmask    string
	broadcast  string
	inet6      string
	prefixlen  int
	ether      string
	txqueuelen int
}
