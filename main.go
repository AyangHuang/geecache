package main

import (
	"fmt"
	"geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
	"1":    "asdf",
	"2":    "sadf",
	"3":    "asdf",
	"4":    "asdf",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", "LRU", geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), 2<<10)
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHttpPool(addr)
	peers.Register(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

//这里是测试需求所以开了一个web（便于动态输入嘛，来模拟(实际中groups
//是被引入到程序中的）程序向group索求数据。
func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
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

/**
这里运行有bug，因为在同一个线程内，三次creatGroup，实际在全局变量groups存储的是peers3的，所以发给8002端口的
虽然被8002端口接收，但是处理getGroup取到的Get方法确实调用peers3，所以又会发给8002，造成死循环
eg: curl "http://localhost:9999/api?key=jack"  会死循环
	curl "http://localhost:9999/api?key=Tom"  正常
	curl "http://localhost:9999/api?key=Sam"  正常
	curl "http://localhost:9999/api?key=kkk"  正常

解决方案：启动三个进程分开运行。可以copy项目或者geetoto那样比较方便。
*/
func main() {

	apiAddr := "http://127.0.0.1:9999"

	addrMap := map[int]string{
		8001: "http://127.0.0.1:8001",
		8002: "http://127.0.0.1:8002",
		8003: "http://127.0.0.1:8003",
	}

	gee1 := createGroup()
	gee2 := createGroup()
	gee3 := createGroup()

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	//1, 2, 3
	// 相当于创建了实际3个工程中引入了3个分布式的缓存服务器
	go startCacheServer("http://127.0.0.1:8001", addrs, gee1)
	go startCacheServer("http://127.0.0.1:8002", addrs, gee2)
	go startCacheServer("http://127.0.0.1:8003", addrs, gee3)

	// 监听9999，且9999没加入分布式结点的Map中，这个类似于负载均衡，实际是模拟。具体看函数那里的注释
	go startAPIServer(apiAddr, gee1)
	c := make(chan bool)
	<-c

}
