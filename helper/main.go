package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack"
	"github.com/yuriy-gagarin/netstring"
)

var domain string
var ip uint

type Message struct {
	Domain string `msgpack:"domain" json:"domain"`
	Ip     uint32 `msgpack:"ip" json:"ip"`
}

func main() {
	var msgs []Message

	f, err := os.Open("input.json")
	if err != nil {
		log.Panic(err)
	}

	jdec := json.NewDecoder(f)
	err = jdec.Decode(&msgs)
	if err != nil {
		log.Panic(err)
	}

	var mbuf bytes.Buffer
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	menc := msgpack.NewEncoder(&mbuf)
	for _, v := range msgs {
		err := menc.Encode(v)
		if err != nil {
			log.Println(err)
			continue
		}

		chunk, err := ioutil.ReadAll(&mbuf)
		if err != nil {
			log.Println(err)
			continue
		}

		nc := netstring.Encode(chunk)

		conn.Write(nc)
		fmt.Printf("sent %v\n", string(nc))

		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}

	conn.Close()

	ch := make(chan string, 0)

	go func() {
		for {
			r := bufio.NewReader(os.Stdin)
			text, _ := r.ReadString('\n')
			ch <- text
		}
	}()

	conn, err = net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	for t := range ch {
		ts := strings.Split(t, " ")
		if len(ts) < 2 {
			fmt.Print("Usage: <string> <uint32>\n")
			continue
		}

		tx := strings.Trim(ts[1], "\n ")
		ip, err := strconv.ParseUint(tx, 10, 32)
		if err != nil {
			fmt.Print("Usage: <string> <uint32>\n")
			continue
		}

		var buf bytes.Buffer
		msg := Message{ts[0], uint32(ip)}
		menc := msgpack.NewEncoder(&buf)
		err = menc.Encode(msg)
		if err != nil {
			log.Println(err)
			continue
		}

		chunk, err := ioutil.ReadAll(&buf)
		if err != nil {
			log.Println(err)
			continue
		}

		nc := netstring.Encode(chunk)
		conn.Write(nc)
		fmt.Printf("sent %v\n", string(nc))
	}
}
