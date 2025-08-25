package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/yinheli/udppunch"
)

var (
	l        = log.New(os.Stdout, "", log.LstdFlags)
	port     = flag.Int("port", 19993, "udp punch port")
	webPort  = flag.Int("web-port", 8080, "web interface port")
	webPass  = flag.String("web-pass", "admin", "web interface password")
	version  = flag.Bool("version", false, "show version")
	peers    *lru.Cache
)

type PeerInfo struct {
	Key       string    `json:"key"`
	Address   string    `json:"address"`
	Timestamp time.Time `json:"timestamp"`
}

func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != *webPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="UDP Punch Monitor"`)
			w.WriteHeader(401)
			w.Write([]byte("需要认证才能访问"))
			return
		}
		handler(w, r)
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>UDP Punch 客户端监控</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .status { color: #28a745; font-weight: bold; }
        .info { margin-bottom: 20px; padding: 10px; background: #e9ecef; border-radius: 5px; }
    </style>
    <script>
        function refreshPage() {
            location.reload();
        }
        setInterval(refreshPage, 10000); // 自动刷新，每10秒
    </script>
</head>
<body>
    <h1>UDP Punch 服务器监控</h1>
    <div class="info">
        <p><strong>服务器状态:</strong> <span class="status">运行中</span></p>
        <p><strong>UDP端口:</strong> {{.UDPPort}}</p>
        <p><strong>Web端口:</strong> {{.WebPort}}</p>
        <p><strong>活跃客户端数量:</strong> {{.ClientCount}}</p>
        <p><strong>最后更新:</strong> {{.UpdateTime}}</p>
    </div>
    
    <h2>已连接的客户端</h2>
    {{if .Peers}}
    <table>
        <tr>
            <th>客户端公钥</th>
            <th>IP地址</th>
            <th>端口</th>
            <th>最后活跃时间</th>
        </tr>
        {{range .Peers}}
        <tr>
            <td>{{.Key}}</td>
            <td>{{.Address}}</td>
            <td>{{.Port}}</td>
            <td>{{.Timestamp}}</td>
        </tr>
        {{end}}
    </table>
    {{else}}
    <p>当前没有活跃的客户端连接。</p>
    {{end}}
    
    <p style="margin-top: 30px; color: #666; font-size: 14px;">
        页面每10秒自动刷新 | <a href="javascript:refreshPage()">手动刷新</a>
    </p>
</body>
</html>`

	keys := peers.Keys()
	var peerInfos []map[string]interface{}
	
	for _, k := range keys {
		if p, ok := peers.Get(k); ok {
			peer := p.(udppunch.Peer)
			key := fmt.Sprintf("%x", k)
			peerKey, _ := peer.Parse()
			_ = peerKey
			ip := net.IPv4(peer[32], peer[33], peer[34], peer[35]).String()
			port := int(peer[36])<<8 + int(peer[37])
			peerInfos = append(peerInfos, map[string]interface{}{
				"Key":       key[:16] + "...",
				"Address":   ip,
				"Port":      port,
				"Timestamp": time.Now().Format("2006-01-02 15:04:05"),
			})
		}
	}

	data := map[string]interface{}{
		"UDPPort":     *port,
		"WebPort":     *webPort,
		"ClientCount": len(peerInfos),
		"UpdateTime":  time.Now().Format("2006-01-02 15:04:05"),
		"Peers":       peerInfos,
	}

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	keys := peers.Keys()
	var peerInfos []PeerInfo
	
	for _, k := range keys {
		if p, ok := peers.Get(k); ok {
			peer := p.(udppunch.Peer)
			key := fmt.Sprintf("%x", k)
			_, addr := peer.Parse()
			peerInfos = append(peerInfos, PeerInfo{
				Key:       key,
				Address:   addr,
				Timestamp: time.Now(),
			})
		}
	}

	response := map[string]interface{}{
		"status":       "running",
		"udp_port":     *port,
		"web_port":     *webPort,
		"client_count": len(peerInfos),
		"clients":      peerInfos,
		"timestamp":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func startWebServer() {
	http.HandleFunc("/", requireAuth(webHandler))
	http.HandleFunc("/api", requireAuth(apiHandler))
	
	webAddr := fmt.Sprintf(":%d", *webPort)
	l.Printf("Web界面启动在端口 %d", *webPort)
	l.Printf("访问 http://localhost:%d 查看客户端状态", *webPort)
	l.Printf("Web界面用户名: admin, 密码: %s", *webPass)
	
	if err := http.ListenAndServe(webAddr, nil); err != nil {
		l.Printf("Web服务器启动失败: %v", err)
	}
}

func main() {
	if flag.Parse(); !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	if *version {
		fmt.Println(udppunch.Version)
		os.Exit(0)
	}

	peers, _ = lru.New(1024)

	// 启动Web服务器
	go startWebServer()

	// handle dump peers
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP)
		for range ch {
			ks := peers.Keys()
			l.Print("dump peers:", len(ks))
			for _, k := range ks {
				if p, ok := peers.Get(k); ok {
					l.Print(p)
				}
			}
		}
	}()

	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		l.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", addr)

	if err != nil {
		l.Fatal(err)
	}

	for {
		buf := make([]byte, 1024*8)
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		if n < 1 {
			continue
		}

		// l.Printf("\nfrom:%v\n%s", raddr, hex.Dump(buf[:n]))

		switch buf[0] {
		case udppunch.HandshakeType:
			var key udppunch.Key
			copy(key[:], buf[1:])
			peers.Add(key, udppunch.NewPeerFromAddr(key, raddr))
		case udppunch.ResolveType:
			data := make([]byte, 0, (n-1)/32*38)
			for i := 1; i < n; i += 32 {
				var key udppunch.Key
				copy(key[:], buf[i:i+32])
				if v, ok := peers.Get(key); ok {
					peer := v.(udppunch.Peer)
					data = append(data, peer[:]...)
				}
			}
			conn.WriteToUDP(data, raddr)
		}
	}
}
