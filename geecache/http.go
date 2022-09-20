package geecache

import (
	"fmt"
	"geecache/consistenhash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/geecache/"
	defaultReplicas = 50
)

func (h *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.selfHost, fmt.Sprintf(format, v...))
}

type HttpPool struct {
	selfHost string
	//相同的服务相同的basepath，因为一台服务器可能有很多服务，方便区分和归类
	// baseUrl:端口号/<basepath>/<groupname>/<key> required
	basePath string
	// 匿名字段会把方法集加到包含匿名字段的struct当中
	sync.Mutex
	peer *consistenhash.Map // 结点
	// 为什么要多余搞多个httpGetter，其实还是为了解耦（在前面解耦哈，不是在这里解耦）。
	// 未来可能不是通过http与远程服务器联系，但是只要实现了PeerGetter接口就可以了
	httpGetters map[string]*httpGetter
}

func NewHttpPool(selfUrl string) *HttpPool {
	return &HttpPool{
		selfHost: selfUrl,
		basePath: defaultBasePath,
		peer:     consistenhash.NewMap(defaultReplicas, nil),
	}
}

// Register 注册远程结点 ip:port
func (h *HttpPool) Register(url ...string) {
	h.Lock()
	defer h.Unlock()
	h.peer.Add(url...)
	h.httpGetters = make(map[string]*httpGetter, len(url))
	for _, peer := range url {
		h.httpGetters[peer] = &httpGetter{
			// ip:port/geeceche/...
			baseUrl: peer + h.basePath,
		}
	}
}

// PickPeer 即实现获取远程结点功能
func (h *HttpPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	h.Lock()
	defer h.Unlock()
	// 跟selfHost对比记得，不然有可能陷入无线循环
	if ip := h.peer.Get(key); ip != "" && ip != h.selfHost {
		h.Log("Pick peer %s", ip)
		return h.httpGetters[ip], true
	}
	return nil, false
}

// ServeHttp httpPool实现服务端功能，即别人来这里请求缓存
func (h *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	h.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.Split(r.URL.Path[len(h.basePath):], "/")
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	// 这里还是要调用到包里的函数，没完全解耦把
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

type httpGetter struct {
	// ip:端口号/geecache/
	baseUrl string
}

// Get 实现客户端功能，当本地缓存没有时，向其他结点发送http请求获取缓存
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseUrl,
		//QueryEscape 对字符串进行转义，以便可以安全地放置它
		// 在 URL 查询中。
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}
