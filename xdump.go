package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/smallnest/ringbuffer"
)

var (
	pcapFile      = flag.String("f", "", "offline pcap file. If it is empty xdump will uses live capture")
	captureAddr   = flag.String("c", "127.0.0.1:8972", "captured address and port, for example, 127.0.0.1:8972")
	outputPayload = flag.Bool("p", false, " print payload or not")
	withColor     = flag.Bool("color", true, "output data with color")
)

var (
	handle *pcap.Handle
	conns  map[string]*connection
	mu     sync.RWMutex
)

// xdump is a tcpdump like tool only to capture rpcx protocol package and  print them in console.
// If you want to capture all packages into files, please use tcpdump.
func main() {
	flag.Parse()

	if *captureAddr == "" {
		log.Fatal("captured address must not be empty")

	}
	host, port, err := net.SplitHostPort(*captureAddr)
	if err != nil {
		log.Fatalf("failed to find device for %s: %v", *captureAddr, err)
	}

	if *pcapFile != "" {
		handle, err = pcap.OpenOffline(*pcapFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		device, err := findDeviceByPcap(host)
		if err != nil {
			log.Fatalf("failed to find device for %s: %v", host, err)
		}

		log.Printf("find device %s for %s", device, host)

		inactive, err := pcap.NewInactiveHandle(device)
		if err != nil {
			log.Fatalf("could not create: %v", err)
		}
		defer inactive.CleanUp()
		if err = inactive.SetSnapLen(1522); err != nil {
			log.Fatalf("could not set snap length: %v", err)
		} else if err = inactive.SetPromisc(false); err != nil {
			log.Fatalf("could not set promisc mode: %v", err)
		} else if err = inactive.SetTimeout(30 * time.Second); err != nil {
			log.Fatalf("could not set timeout: %v", err)
		}

		if handle, err = inactive.Activate(); err != nil {
			log.Fatal("PCAP Activate error:", err)
		}
		//handle, err = pcap.OpenLive(device, 1522, false, 30*time.Second)
		if err != nil {
			log.Fatalf("failed to open live for %s: %v", device, err)
		}
	}
	defer handle.Close()

	conns = make(map[string]*connection)
	go dump(host, port)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exited")
}

func dump(host, port string) {
	var filter = "tcp and port " + port + " and host " + host
	err := handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatalf("failed to set filter: %v", err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		var fromIP, toIP string
		var fromPort, toPort int
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			fromIP = ip.SrcIP.String()
			toIP = ip.DstIP.String()
		}
		if fromIP == "" {
			ipLayer = packet.Layer(layers.LayerTypeIPv6)
			if ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv6)
				fromIP = ip.SrcIP.String()
				toIP = ip.DstIP.String()
			}
		}

		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			fromPort = int(tcp.SrcPort)
			toPort = int(tcp.DstPort)

		}
		applicationLayer := packet.ApplicationLayer()

		key := fmt.Sprintf("%s:%d -> %s:%d", fromIP, fromPort, toIP, toPort)
		mu.RLock()
		c := conns[key]
		mu.RUnlock()

		if tcpLayer != nil && c != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			if tcp.FIN {
				mu.Lock()
				c.Close()
				delete(conns, key)
				mu.Unlock()
				continue
			}
		}

		if applicationLayer != nil {
			data := applicationLayer.Payload()
			if len(data) > 0 {
				if c == nil {
					c = &connection{
						key: key,
						buf: ringbuffer.New(1024 * 1024),
						closeCallback: func(err error) {
							mu.Lock()
							c.Close()
							delete(conns, key)
							mu.Unlock()
						},
						parseCallBack: output,
						done:          make(chan struct{}),
					}
					mu.Lock()
					conns[key] = c
					mu.Unlock()
					go c.Start()
				}

				c.buf.Write(data)
			}

		}
	}
}

// you can't capture windows loopback address such as 127.0.0.1
func findDeviceByPcap(ip string) (string, error) {
	ifaces, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		for _, addr := range iface.Addresses {
			if addr.IP.String() == ip {
				return iface.Name, nil
			}
		}
	}

	return "", fmt.Errorf("device for %s not found", ip)
}

func findDevice(ip string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ipaddr net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ipaddr = v.IP
			case *net.IPAddr:
				ipaddr = v.IP
			}

			if ipaddr.String() == ip {
				return iface.Name, nil
			}
		}
	}
	return "", errors.New("no available device")
}
