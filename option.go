package blinker

import (
	"log"
	"net"
	"time"
)

type option func(*blinkerNode)

func WithTimeout(timeout time.Duration) option {
	//
	return func(node *blinkerNode) {
		//
		node.timeout = timeout
	}
}

func WithLogger(fn func() (int, *log.Logger)) option {
	//
	return func(node *blinkerNode) {
		//
		if nil != fn {
			//
			node.loglevel, node.logger = fn()
		}
	}
}

func WithServerAddress(network, address string) option {
	//
	return func(node *blinkerNode) {
		//
		if "" != network && "" != address {
			//
			switch network {
			case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
				//
				node.network = network
				node.address = address
			}
		}
	}
}

func WithConn(conn net.Conn) option {
	//
	return func(node *blinkerNode) {
		//
		if nil != conn {
			//
			node.conn = conn
		}
	}
}

func WithResolveFailCallback(fn func(string, string)) option {
	//
	return func(node *blinkerNode) {
		//
		if nil != fn {
			//
			node.cb_resolve_fail = fn
		}
	}
}

func WithPowerSetCallback(fn func(bool)) option {
	//
	return func(node *blinkerNode) {
		//
		if nil != fn {
			//
			node.cb_power_set = fn
		}
	}
}

func WithUpdateCallback(fn func() bool) option {
	//
	return func(node *blinkerNode) {
		//
		if nil != fn {
			//
			node.cb_update = fn
		}
	}
}
