package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/einsier/go-cache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {

				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheHTTPServer(port int, addrMap map[int]string, g *cache.Group) {
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	addr := addrMap[port]

	peers := cache.NewHTTPPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startCacheGRPCServer(port int, addrMap map[int]string, g *cache.Group) {
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	addr := addrMap[port]

	peers := cache.NewGRPCPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("cache is running at", addr)
	peers.Run()
}

func startAPIServer(apiAddr string, g *cache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

var (
	port = flag.Int("port", 8001, "cache server port")
	api  = flag.Bool("api", false, "Start a api server?")
)

func main() {
	flag.Parse()

	apiAddr := "http://localhost:9999"

	g := createGroup()
	if *api {
		go startAPIServer(apiAddr, g)
	}

	// use http
	startCacheHTTPServer(*port, map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}, g)

	// // use grpc
	// startCacheGRPCServer(*port, map[int]string{
	// 	8001: ":8001",
	// 	8002: ":8002",
	// 	8003: ":8003",
	// }, g)
}
