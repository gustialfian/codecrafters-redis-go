package main

import "flag"

func main() {
	dir := flag.String("dir", "", "The directory where RDB files are stored")
	dbfilename := flag.String("dbfilename", "", "The name of the RDB file")
	flag.Parse()

	startServer(serverOpt{dir: *dir, dbfilename: *dbfilename})
}
