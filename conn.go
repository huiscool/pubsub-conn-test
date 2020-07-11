package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
)

var (
	port         = "18070"
	testProtocol = protocol.ID("/conn-test/1.0.0")
)

//go:generate go build -o conn .

func main() {
	// we need two host:
	// one is the server, another is the client.
	serverAddr := flag.String("-s", "", "-c /ip4/x.x.x.x/tcp/18070/p2p/xxx")

	if *serverAddr == "" {
		server()
	} else {
		client(*serverAddr)
	}
}

func server() {
	// listen for all incoming connections
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port))
	if err != nil {
		panic(err)
	}
	host, err := libp2p.New(
		context.TODO(),
		libp2p.ListenAddrs(addr),
	)
	if err != nil {
		panic(err)
	}

	host.SetStreamHandler(testProtocol, serverStreamHandler)

	fmt.Printf("run \n./conn -s /ip4/%s/tcp/%s/p2p/%s\nin another computer\n", getPublicIP().String(), port, host.ID().Pretty())
	// hang forever
	select {}
}

func getPublicIP() (public net.IP) {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		// handle err
		if err != nil {
			panic(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address,
			if isPublicIP(ip) {
				return ip
			}
		}
	}
	// just panic
	panic("cannot find a public IP")
}

func isPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func serverStreamHandler(s network.Stream) {
	//
	fmt.Println("a new client comming:", s.Conn().RemotePeer().Pretty())
	s.Write([]byte("hello"))
	s.Close()
}

func client(serverAddr string) {

	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(serverAddr)
	if err != nil {
		panic(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		panic(err)
	}

	host, err := libp2p.New(
		context.TODO(),
	)
	if err != nil {
		panic(err)
	}

	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	host.SetStreamHandler(testProtocol, serverStreamHandler)

	s, err := host.NewStream(context.Background(), info.ID, testProtocol)
	if err != nil {
		panic(err)
	}

	// handle stream
	bin, err := ioutil.ReadAll(s)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bin))
}
