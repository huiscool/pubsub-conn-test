package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

var (
	port         = "18070"
	testProtocol = protocol.ID("/conn-test/1.0.0")
	topic        = "conn"
)

var (
	localHost host.Host
)

//go:generate GOOS=linux GOARCH=amd64 go build -o conn .

func main() {
	// we need two host:
	// one is the server, another is the client.
	serverAddr := flag.String("s", "", "-s /ip4/x.x.x.x/tcp/18070/p2p/xxx")
	flag.Parse()
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

	localHost = host

	host.SetStreamHandler(testProtocol, streamHandler)

	fmt.Printf("run \n./conn -s /ip4/%s/tcp/%s/p2p/%s\nin another computer\n", getPublicIP().String(), port, host.ID().Pretty())
	// hang forever
	select {}
}

func getPublicIP() (public net.IP) {
	resp, err := http.Get("http://bot.whatismyipaddress.com")
	if err != nil {
		panic(err)
	}
	bin, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return net.ParseIP(string(bin))
}
func streamHandler(s network.Stream) {
	s.Write([]byte("hello"))
	s.Close()
}

func pubsubTest() {
	psub, err := pubsub.NewGossipSub(context.Background(), localHost)
	if err != nil {
		panic(err)
	}
	topicItem, err := psub.Join(topic)
	if err != nil {
		panic(err)
	}
	// wait for our peer join
	var peers = topicItem.ListPeers()
	var retry = 10
	for len(peers) == 0 && retry > 0 {
		fmt.Println("not peers, waiting")
		time.Sleep(1 * time.Second)
		peers = topicItem.ListPeers()
		retry--
	}
	if retry == 0 {
		panic("cannot find peers")
	}
	fmt.Println(peers)
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
	localHost = host

	host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	host.SetStreamHandler(testProtocol, streamHandler)

	s, err := host.NewStream(context.Background(), info.ID, testProtocol)
	if err != nil {
		panic(err)
	}

	// handle stream
	bin, err := ioutil.ReadAll(s)
	if err != nil {
		panic(bin)
	}
	s.Close()
}
