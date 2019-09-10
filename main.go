package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/go-redis/redis"
)

const prefix = "msg:"
const ttl = 10 * time.Second

var port int
var host string

var redisaddr string

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
	scn.Split(bufio.ScanLines)

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

var LUA = `
local result = {};
for i, key in ipairs(KEYS) do
	local value = redis.call('GET', key);
	local jvalue = cjson.encode(cmsgpack.unpack(value));
	table.insert(result, jvalue);
end;
return result;`

var SHA string

func ListValues(r *redis.Client) error {
	var cursor uint64
	var err error

	for {
		var keys []string
		keys, cursor, err = r.Scan(cursor, prefix+"*", 50).Result()
		if err != nil {
			return fmt.Errorf("failed scan: %v", err)
		}

		if cursor == 0 {
			if len(keys) == 0 {
				break
			}

			vals, err := r.EvalSha(SHA, keys).Result()
			if err != nil {
				return fmt.Errorf("failed eval: %v", err)
			}

			s := fmt.Sprintln("Current values:")
			for _, v := range vals.([]interface{}) {
				s += fmt.Sprintln(v.(string))
			}

			log.Print(s)

			break
		}
	}

	return nil
}

func main() {
	r := redis.NewClient(&redis.Options{Addr: redisaddr, Password: "", DB: 0})
	_, err := r.Ping().Result()
	if err != nil {
		log.Panic(err)
	}

	sha, err := r.ScriptLoad(LUA).Result()
	if err != nil {
		log.Panic(err)
	}

	SHA = sha
	log.Printf("DEBUG: sha for deserialize script is %s\n", SHA)

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
