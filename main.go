package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/vmihailenco/msgpack"

	"github.com/yuriy-gagarin/netstring"

	"github.com/google/uuid"

	"github.com/go-redis/redis"
)

const prefix = "msg:"
const ttl = 10 * time.Second

var port int
var host string

var redisaddr string

var scriptIndexSearch string

type Message struct {
	Domain string `msgpack:"domain" json:"domain"`
	Ip     uint32 `msgpack:"ip" json:"ip"`
}

func init() {
	flag.IntVar(&port, "tcpport", 6000, "port")
	flag.StringVar(&host, "tcpaddr", "0.0.0.0", "host")
	flag.Parse()

	redisaddr = os.Getenv("REDIS")
	if redisaddr == "" {
		redisaddr = "localhost:6379"
	}
}

func HandleConn(r *redis.Client, conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection: %s\n", conn.RemoteAddr().String())

	scn := bufio.NewScanner(conn)
	scn.Split(netstring.SplitNetstring)

	for scn.Scan() {
		id, err := uuid.NewRandom()
		if err != nil {
			continue
		}

		r.SetNX(prefix+id.String(), scn.Bytes(), ttl)
	}

	if err := scn.Err(); err != nil {
		log.Println(err)
	}
}

func Ticker(r *redis.Client) {
	ticker := time.Tick(time.Second)
	for _ = range ticker {
		err := ListValues(r)
		if err != nil {
			log.Println(err)
		}
	}
}

func ListValues(r *redis.Client) error {
	vals, err := r.EvalSha(scriptIndexSearch, []string{}, prefix+"*").Result()
	if err != nil {
		return fmt.Errorf("failed eval: %v", err)
	}

	if len(vals.([]interface{})) == 0 {
		log.Print("Current values:\nnone\n\n")
		return nil
	}

	s := fmt.Sprintln("Current values:")
	for _, v := range vals.([]interface{}) {
		var msg Message
		err = msgpack.Unmarshal([]byte(v.(string)), &msg)
		if err != nil {
			log.Println(err)
			continue
		}
		s += fmt.Sprintf("Domain: %s\t Ip: %d\n", msg.Domain, msg.Ip)
	}

	log.Println(s)

	return nil
}

func loadScript(r *redis.Client, path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return r.ScriptLoad(string(b)).Result()
}

func main() {
	r := redis.NewClient(&redis.Options{Addr: redisaddr, Password: "", DB: 0})
	_, err := r.Ping().Result()
	if err != nil {
		log.Panic(err)
	}

	scriptIndexSearch, err = loadScript(r, "indexSearch.lua")
	if err != nil {
		log.Panic(err)
	}

	listener, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()

	go Ticker(r)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go HandleConn(r, conn)
	}
}
