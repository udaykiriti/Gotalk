package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net"
	"fmt"
	"strings"

	"github.com/grandcat/zeroconf"
	"github.com/skip2/go-qrcode"
	"gotalk/internal/handlers"
	"gotalk/internal/storage"
	"gotalk/internal/ws"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	// Initialize Storage
	store, err := storage.NewStore("data/gotalk.db")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	hub := ws.NewHub(store)
	go hub.Run()

	h := handlers.NewHandler(hub)
	http.HandleFunc("/", h.ServeHome)
	http.HandleFunc("/ws", h.ServeWs)
	http.HandleFunc("/api/rooms", h.HandleRooms)

	// Start mDNS Service Discovery
	mdnsServer, err := zeroconf.Register("GoTalk-Server", "_gotalk._tcp", "local.", 8080, []string{"txtv=0"}, nil)
	if err != nil {
		log.Printf("Failed to register mDNS service: %v", err)
	} else {
		defer mdnsServer.Shutdown()
		log.Println(" [GoTalk] mDNS Auto-Discovery service started (_gotalk._tcp)")
	}

	localIP := getLocalIP()
	port := strings.TrimPrefix(*addr, ":")
	serverURL := fmt.Sprintf("http://%s:%s", localIP, port)

	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf(" [GoTalk] Server is starting...\n")
	fmt.Printf(" Internal : http://localhost:%s\n", port)
	fmt.Printf(" Network  : %s\n", serverURL)
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	// Generate and print QR code to terminal
	qr, err := qrcode.New(serverURL, qrcode.Medium)
	if err == nil {
		fmt.Println("\n Scan to join from your phone:")
		fmt.Print(qr.ToSmallString(false))
	}
	fmt.Printf(strings.Repeat("=", 60) + "\n\n")

	srv := &http.Server{Addr: *addr}

	// Graceful shutdown channel
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Printf("Server started on %s", *addr)

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	hub.Stop()
	if store != nil {
		_ = store.Close()
	}
	log.Print("Server Exited Properly")
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	var fallback string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}

			ipStr := ip.String()
			// Prioritize common local network ranges
			if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "172.") {
				return ipStr
			}
			fallback = ipStr
		}
	}

	if fallback != "" {
		return fallback
	}
	return "localhost"
}
