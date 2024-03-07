package main

import "flag"

func main() {
	dir := flag.String("dir", "", "The directory where RDB files are stored")
	dbfilename := flag.String("dbfilename", "", "The name of the RDB file")
	port := flag.String("port", "6379", "The PORT of the Server")
	flag.Parse()

	startServer(serverOpt{
		dir:        *dir,
		dbfilename: *dbfilename,
		port:       *port,
	})
}
