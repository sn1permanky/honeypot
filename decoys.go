package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func startTCP(ctx context.Context, decoy DecoySpec, tel *Telemetry) {
	addr := fmt.Sprintf("0.0.0.0:%d", decoy.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("listen %s failed: %v", addr, err)
		return
	}
	defer ln.Close()
	go func() {
		<-ctx.Done()
		ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		go handleConn(conn, decoy, tel)
	}
}

func startUDP(ctx context.Context, decoy DecoySpec, tel *Telemetry) {
	addr := fmt.Sprintf("0.0.0.0:%d", decoy.Port)
	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		log.Printf("listen udp %s failed: %v", addr, err)
		return
	}
	defer pc.Close()
	go func() {
		<-ctx.Done()
		pc.Close()
	}()
	buf := make([]byte, 2048)
	for {
		n, src, err := pc.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		tel.emit("decoy", "udp_probe", map[string]any{
			"src_ip":     src.String(),
			"dst_port":   decoy.Port,
			"bytes_read": n,
		})
	}
}

func handleConn(conn net.Conn, decoy DecoySpec, tel *Telemetry) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	switch decoy.Behavior {
	case "http_admin":
		serveHTTPAdmin(conn, decoy, tel)
	case "mongo_decoy":
		serveMongoDecoy(conn, decoy, tel)
	case "kaspersky_like_tls":
		serveKasperskyTLS(conn, decoy, tel)
	case "telnet_banner":
		serveTelnetBanner(conn, decoy, tel)
	default:
		serveDefaultBanner(conn, decoy, tel)
	}
}

func serveHTTPAdmin(conn net.Conn, decoy DecoySpec, tel *Telemetry) {
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	request := string(buf[:n])
	requestLine := strings.SplitN(request, "\r\n", 2)[0]
	body := "<html><body><h1>Admin Console</h1>" +
		"<p>Authorization required</p></body></html>"
	response := "HTTP/1.1 401 Unauthorized\r\n" +
		"Server: nginx/1.24.0\r\n" +
		"WWW-Authenticate: Basic realm=\"admin\"\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"\r\n" + body
	conn.Write([]byte(response))
	tel.emit("decoy", "http_admin_probe", map[string]any{
		"src_ip":     conn.RemoteAddr().String(),
		"dst_port":   decoy.Port,
		"request":    requestLine,
		"bytes_read": n,
	})
}

func serveMongoDecoy(conn net.Conn, decoy DecoySpec, tel *Telemetry) {
	buf := make([]byte, 2048)
	n, _ := conn.Read(buf)
	tel.emit("decoy", "mongo_probe", map[string]any{
		"src_ip":     conn.RemoteAddr().String(),
		"dst_port":   decoy.Port,
		"bytes_read": n,
	})
}

func serveKasperskyTLS(conn net.Conn, decoy DecoySpec, tel *Telemetry) {
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	tel.emit("decoy", "kaspersky_tls_probe", map[string]any{
		"src_ip":     conn.RemoteAddr().String(),
		"dst_port":   decoy.Port,
		"bytes_read": n,
	})
}

func serveTelnetBanner(conn net.Conn, decoy DecoySpec, tel *Telemetry) {
	r := bufio.NewReader(conn)
	conn.Write([]byte("\r\nUbuntu 22.04.4 LTS\r\nlogin: "))
	user, _ := r.ReadString('\n')
	conn.Write([]byte("Password: "))
	pass, _ := r.ReadString('\n')
	conn.Write([]byte("\r\nLogin incorrect\r\n"))
	tel.emit("decoy", "telnet_auth_attempt", map[string]any{
		"src_ip":   conn.RemoteAddr().String(),
		"dst_port": decoy.Port,
		"username": strings.TrimSpace(user),
		"password": strings.TrimSpace(pass),
	})
}

func serveDefaultBanner(conn net.Conn, decoy DecoySpec, tel *Telemetry) {
	banner := decoy.Params["banner"]
	if banner == "" {
		banner = "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3\r\n"
	}
	conn.Write([]byte(banner))
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	tel.emit("decoy", "banner_probe", map[string]any{
		"src_ip":     conn.RemoteAddr().String(),
		"dst_port":   decoy.Port,
		"bytes_read": n,
	})
}
