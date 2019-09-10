package main

import (
	"flag"
	"math"
	"net"

	"github.com/vmihailenco/msgpack"
)

var domain string
var ip uint

func init() {

	flag.StringVar(&domain, "d", "", "domain")
	flag.UintVar(&ip, "ip", 0, "ip")
	flag.Parse()

	if ip > math.MaxInt32 {
		ip = math.MaxInt32
	}
}

type Message struct {
	Domain string `msgpack:"domain"`
	Ip     uint32 `msgpack:"ip"`
}

func main() {
	msg := Message{domain, uint32(ip)}

	b, err := msgpack.Marshal(&msg)
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.Write(b)
}
