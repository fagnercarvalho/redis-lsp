package main

import (
	"context"
	"flag"
	"github.com/fagnercarvalho/redis-lsp/server"
	"github.com/sourcegraph/jsonrpc2"
	"io"
	"log"
	"os"
)

func main() {
	var address, username, password, logFile string
	var database int
	var debugLogEnabled, dbCacheEnabled bool
	flag.StringVar(&address, "address", "localhost:6379", "Redis instance address for caching data for autocompletion.")
	flag.StringVar(&username, "username", "", "Redis instance username for caching data for autocompletion.")
	flag.StringVar(&password, "password", "", "Redis instance password for caching data for autocompletion.")
	flag.IntVar(&database, "database", 0, "Redis database for caching data for autocompletion.")
	flag.StringVar(&logFile, "logFile", "c:/server.log", "Path for log file.")
	flag.BoolVar(&debugLogEnabled, "debugLogEnabled", false, "Enables debug logging.")
	flag.BoolVar(&dbCacheEnabled, "dbCacheEnabled", false, "Enables keys and users autocompletion.")
	flag.Parse()

	if debugLogEnabled {
		f, err := os.OpenFile("c:/server.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	} else {
		log.SetOutput(io.Discard)
	}

	log.Println("starting server")
	server, err := server.New(address, username, password, database, dbCacheEnabled)
	if err != nil {
		panic(err)
	}
	handler := jsonrpc2.HandlerWithError(server.Handle)
	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(StdIo{}, jsonrpc2.VSCodeObjectCodec{}),
		handler).DisconnectNotify()
	log.Println("stopping server")
}

type StdIo struct{}

func (StdIo) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (StdIo) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (StdIo) Close() error {
	err := os.Stdin.Close()
	if err != nil {
		return err
	}

	return os.Stdout.Close()
}
