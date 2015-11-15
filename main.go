package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
)

type router struct {
	name     string
	testdir  string
	poisoned bool
	table    map[string]int
	distv    map[string]int
	sock     int
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
		table:    make(map[string]int),
		distv:    make(map[string]int),
		sock:     0,
	}

	ReadConfigFiles(r)

	//CreateConnectAllSockets()

	//DistanceVector()
}

func ReadConfigFiles(r *router) error {
	rpath := path.Join(r.testdir, r.name+".cfg")
	fi, err := os.Open(rpath)
	if err != nil {
		return err
	}
	defer fi.Close()

	var a string
	var b, c, d int

	scan := bufio.NewScanner(fi)
	//foreach line in the file
	for scan.Scan() {
		fmt.Sscanf(scan.Text(), "%s %d %d %d", &a, &b, &c, &d)
		fmt.Println(a, b, c, d)
	}

	//cfg, err := ioutil.ReadFile("/" + r.testdir + "/" + r.name + ".cfg")
	//rtr, err := ioutil.ReadFile("/" + r.testdir + "/routers")
	return nil
}
