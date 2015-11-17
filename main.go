package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const INF = 64

type routerInfo struct {
	addr string
	port int
}

type router struct {
	name     string
	testdir  string
	poisoned bool
	table    map[string]*routerInfo
	distv    map[string]int
	neigh    map[string]*routerInfo
	pton     map[int]string
	addr     string
	port     int
	serv     *net.UDPConn
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
		neigh:    make(map[string]*routerInfo),
		pton:     make(map[int]string),
		addr:     "localhost",
		port:     0,
		serv:     nil,
	}

	ReadConfigFiles(r)

	go UpdateTimer(r)
	ListenForMsg(r)

	//DistanceVector()
}

func UpdateTimer(r *router) error {
	for range time.Tick(time.Second * 30) {
		SendRouterUpdateMsg(r)
	}
	return nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(0)
	}
}

func HandleRouterUpdateMsg(r *router, msg string, sender string) error {
	var sendUpdate bool
	var node string
	words := strings.Fields(msg)
	for i := 0; i < len(words); {
		cost, err := strconv.Atoi(words[i+1])
		node = words[i]
		CheckError(err)

		if node == r.name {
			i += 2
			continue
		}

		if r.distv[node] > (r.distv[sender] + cost) {
			r.distv[node] = cost + r.distv[sender]
			sendUpdate = true
		}
		i += 2
	}
	if sendUpdate {
		SendRouterUpdateMsg(r)
		fmt.Printf("%s - dest: %s cost: %d nexthop %s\n", r.name, node, r.distv[node], sender)
	}
	//DisplayConfig(r)
	return nil
}

func SendRouterUpdateMsg(r *router) error {
	var buffer bytes.Buffer
	buffer.WriteString("U")

	for key, val := range r.distv {
		fmt.Fprintf(&buffer, " %s %d", key, val)
	}
	for _, val := range r.neigh {
		ServerAddr, err :=
			net.ResolveUDPAddr("udp", ":"+fmt.Sprint(val.port))
		CheckError(err)

		r.serv.WriteToUDP(buffer.Bytes(), ServerAddr)
	}
	return nil
}

func HandleLinkUpdateMsg(r *router, msg string, sender string) error {
	words := strings.Fields(msg)
	c, err := strconv.Atoi(words[1])
	CheckError(err)
	r.distv[words[0]] = c
	SendRouterUpdateMsg(r)
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
	r.serv = ServerConn

	defer ServerConn.Close()

	buf := make([]byte, 1024)
	for {

		n, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		switch buf[0] {
		case 'U': //Router Update Message
			HandleRouterUpdateMsg(r, string(buf[1:n]), r.pton[addr.Port])
		case 'L': //Link Update Message
			HandleLinkUpdateMsg(r, string(buf[1:n]), r.pton[addr.Port])
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
		r.neigh[name] = r.table[name]
		r.pton[r.table[name].port] = name
	}
	r.distv[r.name] = 0

	return nil
}
