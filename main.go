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
	DisplayConfig(r)

	ListenForMsg(r)

	//DistanceVector()
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(0)
	}
}

func HandleRouterUpdateMsg(r *router, msg string) error {
	// U d1 cost1 d2 cost2 ... dn costn
	// 0 1  2     3  4
	fmt.Println("Update Router\n")
	words := strings.Fields(msg)
	for i := 0; i < len(words); {
		c, err := strconv.Atoi(words[i+1])
		CheckError(err)
		if words[i] == r.name {
			r.distv[words[i]] = 0
			i += 2
			continue
		}
		r.distv[words[i]] = c
		i += 2
	}
	DisplayConfig(r)
	return nil
}

func HandleLinkUpdateMsg(r *router, msg string) error {
	fmt.Println("Update Link\n")
	return nil
}

func HandlePrintMsg(r *router, msg string) error {
	fmt.Println("Print Msg\n")
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
		fmt.Println("Recieved", string(buf[0]), string(buf[1:n]), " from ", addr)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		switch buf[0] {
		case 'U': //Router Update Message
			HandleRouterUpdateMsg(r, string(buf[1:n]))
		case 'L': //Link Update Message
			HandleLinkUpdateMsg(r, string(buf[1:n]))
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
	cfgpath := path.Join(r.testdir, r.name+".cfg")
	fi, err := os.Open(cfgpath)
	if err != nil {
		return err
	}
	defer fi.Close()

	var name, addr string
	var cost, local, remote, port int

	scan := bufio.NewScanner(fi)
	for scan.Scan() {
		fmt.Sscanf(scan.Text(),
			"%s %d %d %d", &name, &cost, &local, &remote)

		r.distv[name] = cost
	}

	rtrpath := path.Join(r.testdir, "routers")
	fi, err = os.Open(rtrpath)
	if err != nil {
		return err
	}

	scan = bufio.NewScanner(fi)
	for scan.Scan() {
		fmt.Sscanf(scan.Text(), "%s %s %d", &name, &addr, &port)

		ri := &routerInfo{
			addr: addr,
			port: port,
		}
		r.table[name] = ri
	}
	r.port = r.table[r.name].port

	return nil
}
