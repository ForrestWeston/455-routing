package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
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
	table    map[string]routerInfo
	distv    map[string]linkInfo
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
		table:    make(map[string]routerInfo),
		distv:    make(map[string]linkInfo),
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
		fmt.Println("Recieved", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
}

func DisplayConfig(r *router) error {
	fmt.Printf("Name: %s Address: %s Port: %d\n", r.name, r.addr, r.port)
	fmt.Println("Router table:\n", r.table)
	fmt.Println("distance vec:\n", r.distv)
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

		li := linkInfo{
			cost:   cost,
			local:  local,
			remote: remote,
		}
		r.distv[name] = li
	}

	rtrpath := path.Join(r.testdir, "routers")
	fi, err = os.Open(rtrpath)
	if err != nil {
		return err
	}

	scan = bufio.NewScanner(fi)
	for scan.Scan() {
		fmt.Sscanf(scan.Text(), "%s %s %d", &name, &addr, &port)

		ri := routerInfo{
			addr: addr,
			port: port,
		}
		r.table[name] = ri
	}
	r.port = r.table[r.name].port

	return nil
}
