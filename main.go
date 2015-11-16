package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
)

const INF = 64

type routerInfo struct {
	addr string
	port int
}

type linkInfo struct {
	cost   int
	local  int
	remote int
}

type router struct {
	name     string
	testdir  string
	poisoned bool
	table    map[string]*routerInfo
	distv    map[string]int
	addr     string
	port     int
}

func main() {
	dashp := flag.Bool("p", false, "do p things")
	flag.Parse()

	testdir := flag.Arg(0)
	routername := flag.Arg(1)

	flag.Args()

	r := &router{
		name:     routername,
		testdir:  testdir,
		poisoned: *dashp,
		table:    make(map[string]*routerInfo),
		distv:    make(map[string]int),
		addr:     "localhost",
		port:     0,
	}

	ReadConfigFiles(r)

	ListenForMsg(r)

	//DistanceVector()
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(0)
	}
}

func HandleRouterUpdateMsg(r *router, msg string, addr string) error {
	var sendUpdate bool
	words := strings.Fields(msg)
	fmt.Println("ADDR:", addr)
	for i := 0; i < len(words); {
		cost, err := strconv.Atoi(words[i+1])
		node := words[i]
		CheckError(err)

		if node == r.name {
			i += 2
			continue
		}

		r.distv[node] = cost
		i += 2
	}
	if sendUpdate {
		SendRouterUpdateMsg(r)
	}
	DisplayConfig(r)
	return nil
}

func SendRouterUpdateMsg(r *router) error {

}

func HandleLinkUpdateMsg(r *router, msg string, addr string) error {
	fmt.Println("Update Link\n")
	words := strings.Fields(msg)
	c, err := strconv.Atoi(words[1])
	CheckError(err)
	r.distv[words[0]] = c
	DisplayConfig(r)
	return nil
}

func HandlePrintMsg(r *router, msg string) error {
	words := strings.Fields(msg)
	if len(words) > 0 {
		if r.distv[words[0]] != 0 || words[0] == r.name {
			fmt.Printf("Router: %s\nRouting Entry:\n", r.name)
			fmt.Printf("\tDest: %s Cost: %d\n",
				words[0], r.distv[words[0]])
		} else {
			fmt.Printf("Router: %s has no entry: %s\n",
				r.name, words[0])
		}
	} else {
		fmt.Printf("Router: %s\nRouting Table:\n", r.name)
		for key, val := range r.distv {
			fmt.Printf("\tDest: %s Cost: %d\n", key, val)
		}
	}
	fmt.Printf("------------------\n")
	return nil
}

func ListenForMsg(r *router) error {
	ServerAddr, err :=
		net.ResolveUDPAddr("udp", ":"+fmt.Sprint(r.port))

	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		fmt.Print(addr.IP.String())

		switch buf[0] {
		case 'U': //Router Update Message
			HandleRouterUpdateMsg(r, string(buf[1:n]), addr.String())
		case 'L': //Link Update Message
			HandleLinkUpdateMsg(r, string(buf[1:n]), addr.String())
		case 'P': //Print Message
			HandlePrintMsg(r, string(buf[1:n]))
		default:
			panic("Unrecognized Message")
		}

	}
}

func DisplayConfig(r *router) error {
	fmt.Printf("(Host) Name: %s Cost: %d Address: %s Port: %d\n",
		r.name, r.distv[r.name], r.addr, r.port)

	for key, val := range r.distv {
		if key == r.name {
			continue
		}
		fmt.Printf("(Neighbor) Name: %s Cost: %d\n", key, val)
	}

	return nil
}

func ReadConfigFiles(r *router) error {
	var name, addr string
	var cost, local, remote, port int

	rtrpath := path.Join(r.testdir, "routers")
	fi, err := os.Open(rtrpath)
	if err != nil {
		return err
	}
	defer fi.Close()

	scan := bufio.NewScanner(fi)
	for scan.Scan() {
		fmt.Sscanf(scan.Text(), "%s %s %d", &name, &addr, &port)

		ri := &routerInfo{
			addr: addr,
			port: port,
		}
		r.table[name] = ri
		r.distv[name] = 64
	}
	r.port = r.table[r.name].port

	cfgpath := path.Join(r.testdir, r.name+".cfg")
	fi, err = os.Open(cfgpath)
	if err != nil {
		return err
	}

	scan = bufio.NewScanner(fi)
	for scan.Scan() {
		fmt.Sscanf(scan.Text(),
			"%s %d %d %d", &name, &cost, &local, &remote)

		r.distv[name] = cost
	}
	r.distv[r.name] = 0

	return nil
}
