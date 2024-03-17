package main

import (
	"fmt"
	"os"
)

func main() {
	flags := parseFlags()
	startServer(ServerOpt{
		dir:        flags.dir,
		dbfilename: flags.dbfilename,
		port:       flags.port,
		replicaOf:  flags.replicaof,
	})
}

type flags struct {
	port       string
	dir        string
	dbfilename string
	replicaof  string
}

func parseFlags() flags {
	result := flags{}
	args := os.Args
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--port":
			i += 1
			result.port = args[i]
		case "--dir":
			i += 1
			result.dir = args[i]
		case "--dbfilename":
			i += 1
			result.dbfilename = args[i]
		case "--replicaof":
			i += 2
			result.replicaof = fmt.Sprintf("%s %s", args[i-1], args[i])
		}
	}
	return result
}
