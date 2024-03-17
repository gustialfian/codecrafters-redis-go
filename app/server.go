package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	config      map[string]string
	data        map[string]string
	opt         ServerOpt
	rdb         RDB
	replication replicationInfo
}

type ServerOpt struct {
	dbfilename string
	dir        string
	port       string
	replicaOf  string
}

type replicationInfo struct {
	masterReplid     string
	masterReplOffset int
	role             string
}

const (
	REPLICATION_ROLE_MASTER = "master"
	REPLICATION_ROLE_SLAVE  = "slave"
)

func startServer(opt ServerOpt) {
	srv := &Server{
		data:   make(map[string]string),
		config: make(map[string]string),
		opt:    opt,
	}

	srv.setupConfig()
	srv.loadRDB()
	srv.setReplicationInfo()
	if srv.replication.role == REPLICATION_ROLE_SLAVE {
		go srv.setupSlave()
	}

	addr := fmt.Sprintf("0.0.0.0:%s", srv.config["port"])
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("net.Listen:", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	log.Println("StartServer...")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("l.Accept:", err.Error())
			os.Exit(1)
		}

		go srv.HandleCon(conn)
	}
}

func (srv *Server) setupConfig() {
	srv.config["dir"] = srv.opt.dir
	srv.config["dbfilename"] = srv.opt.dbfilename
	srv.config["port"] = srv.opt.port
	srv.config["replicaOf"] = srv.opt.replicaOf

	log.Printf("setupConfig: %+v\n", srv.config)
}

func (srv *Server) loadRDB() {
	// log.Printf("loadRDB: %+v\n", srv.rdb.Databases)

	if srv.config["dir"] == "" || srv.config["dbfilename"] == "" {
		var db Database
		db.ID = 0
		db.Fields = map[string]Field{}

		srv.rdb.Databases = append(srv.rdb.Databases, db)
		return
	}

	path := filepath.Join(srv.config["dir"], srv.config["dbfilename"])
	_, err := os.Stat(path)
	if err != nil {
		return
	}
	srv.rdb = ParseRDB(path)
	for _, f := range srv.rdb.Databases[0].Fields {
		if f.ExpiredTime != 0 {
			expTime := time.UnixMilli(int64(f.ExpiredTime))
			if time.Now().After(expTime) {
				continue
			}

			duration := time.Until(expTime)
			go func() {
				time.AfterFunc(duration, func() {
					delete(srv.data, f.Key)
				})
			}()
		}

		srv.data[f.Key] = f.Value.(string)
	}
}

func (srv *Server) setReplicationInfo() {
	if srv.opt.replicaOf != "" {
		srv.replication.role = REPLICATION_ROLE_SLAVE
		return
	}
	srv.replication.role = REPLICATION_ROLE_MASTER
	srv.replication.masterReplOffset = 0
	srv.replication.masterReplid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
}

func (srv *Server) HandleCon(conn net.Conn) {
	for {
		m, err := ParseRESP(conn)
		if err != nil {
			break
		}

		log.Printf("incoming message: %+v\n", m)

		err = srv.RunMessage(conn, m)
		if err != nil {
			log.Fatalln(err)
			break
		}
	}
}

func (srv *Server) RunMessage(conn net.Conn, m Message) error {
	var resp string
	switch m.cmd {
	case "ping":
		resp = "+PONG\r\n"
	case "echo":
		resp = fmt.Sprintf("+%v\r\n", m.args[0])
	case "set":
		resp = srv.onSet(m.args)
	case "get":
		resp = srv.onGet(m.args)
	case "config":
		resp = srv.onConfig(m.args)
	case "keys":
		resp = srv.onKeys(m.args)
	case "info":
		resp = srv.onInfo(m.args)
	default:
		return fmt.Errorf("unknown command")
	}

	_, err := conn.Write([]byte(resp))
	return err
}

func (srv *Server) onSet(args []string) string {
	if len(args) == 2 {
		srv.data[args[0]] = args[1]
	}
	if len(args) == 4 {
		srv.data[args[0]] = args[1]

		ttl, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			<-time.After(time.Duration(ttl) * time.Millisecond)
			delete(srv.data, args[0])
		}()
	}
	return "+OK\r\n"
}

func (srv *Server) onGet(args []string) string {
	val := srv.data[args[0]]

	if len(val) == 0 {
		return "$-1\r\n"
	}

	return fmt.Sprintf("+%v\r\n", val)
}

func (srv *Server) onConfig(args []string) string {
	key := args[1]
	val := srv.config[args[1]]

	if len(val) == 0 {
		return "$-1\r\n"
	}
	return makeArrayBulkString([]string{key, val})
}

func (srv *Server) onKeys(args []string) string {
	switch args[0] {
	case "*":
		return makeArrayBulkString(srv.rdb.Databases[0].Keys)
	}
	return "*0"
}

func (srv *Server) onInfo(args []string) string {
	switch args[0] {
	case "replication":
		var sb strings.Builder
		sb.WriteString("# Replication\n")
		sb.WriteString(fmt.Sprintf("role:%s\n", srv.replication.role))
		sb.WriteString(fmt.Sprintf("master_replid:%s\n", srv.replication.masterReplid))
		sb.WriteString(fmt.Sprintf("master_repl_offset:%v", srv.replication.masterReplOffset))

		return fmt.Sprintf("$%d\r\n%s\r\n", len(sb.String()), sb.String())
	}
	return "*0"
}

func (srv *Server) setupSlave() {
	log.Println("setupSlave...")

	masterHost := strings.Join(strings.Split(srv.opt.replicaOf, " "), ":")
	master := MasterServer{host: masterHost}
	if err := master.Connect(); err != nil {
		log.Fatalln("setupSlave:", err)
	}

	res, err := master.Send("*1\r\n$4\r\nping\r\n")
	if err != nil {
		log.Fatalln("setupSlave:", err)
	}
	if res.raw != "+PONG" {
		fmt.Printf("%q\n", res.raw)
		log.Fatalln("setupSlave: unexpected ping res:", res.raw)
	}
	fmt.Printf("res ping:%+v\n", res)

	res, err = master.Send("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n")
	if err != nil {
		log.Fatalln("setupSlave:", err)
	}
	if res.raw != "+OK" {
		log.Fatalln("setupSlave: unexpected ping res:", res.raw)
	}
	fmt.Printf("res REPLCONF:%+v\n", res)

	res, err = master.Send("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n")
	if err != nil {
		log.Fatalln("setupSlave:", err)
	}
	if res.raw != "+OK" {
		log.Fatalln("setupSlave: unexpected ping res:", res.raw)
	}
	fmt.Printf("res REPLCONF:%+v\n", res)

	res, err = master.Send("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n")
	if err != nil {
		log.Fatalln("setupSlave:", err)
	}
	fmt.Printf("res PSYNC:%+v\n", res)
}

type MasterServer struct {
	host string
	conn net.Conn
}

func (ms *MasterServer) Connect() error {
	conn, err := net.Dial("tcp", ms.host)
	if err != nil {
		return fmt.Errorf("Connect: %w", err)
	}
	ms.conn = conn
	fmt.Println("Connect")
	return nil
}

func (ms *MasterServer) Send(b string) (Message, error) {
	_, err := ms.conn.Write([]byte(b))
	if err != nil {
		return Message{}, fmt.Errorf("Send: %w", err)
	}

	m, err := ParseRESP(ms.conn)
	if err != nil {
		return Message{}, fmt.Errorf("Send: %w", err)
	}
	fmt.Println("Send")
	return m, nil
}
