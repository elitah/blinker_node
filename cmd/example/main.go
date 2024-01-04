package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elitah/blinker_node"
)

func main() {
	//
	var vFlag bool
	//
	var network, address, key string
	//
	flag.StringVar(&network, "n", "", "your server protocol name")
	flag.StringVar(&address, "a", "", "your server address")
	flag.StringVar(&key, "k", "", "your blinker node key")
	//
	flag.Parse()
	//
	if "" == network || "" == address || "" == key {
		//
		flag.Usage()
		//
		return
	}
	//
	node := blinker.NewBlinkerNode(
		blinker.WithLogger(func() (int, *log.Logger) {
			//
			return blinker.LogLevelInfo, log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
		}),
		blinker.WithResolveFailCallback(func(network, address string) {
			//
			fmt.Printf("resolve fail: %s, %s\n", network, address)
		}),
		blinker.WithServerAddress(network, address),
		blinker.WithPowerSetCallback(func(v bool) {
			//
			vFlag = v
		}),
		blinker.WithUpdateCallback(func() bool {
			//
			return vFlag
		}),
	)
	//
	if err := node.Loop(key); nil != err {
		//
		fmt.Println(err)
	}
	//
	fmt.Println(node.IsRunning(), node.IsConnected())
	//
	go func() {
		//
		time.Sleep(5 * time.Second)
		//
		node.Close()
	}()
	//
	node.WaitDone()
	//
	fmt.Println(node.IsRunning(), node.IsConnected())
	//
	time.Sleep(1 * time.Second)
	//
	node.Reset()
	//
	if err := node.Loop(key); nil != err {
		//
		fmt.Println(err)
	}
	//
	node.WaitDone(10000)
	//
	node.Close()
}
