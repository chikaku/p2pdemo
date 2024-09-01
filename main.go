package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"log/slog"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/chat/1.0.0"

func main() {
	h := Must(libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelayWithStaticRelays(convertPeers(dht.DefaultBootstrapPeers)),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
		libp2p.ForceReachabilityPrivate()))
	defer h.Close()

	if len(os.Args) > 1 {
		if err := connect(h, os.Args[1]); err != nil {
			slog.Error(color.RedString("failed to connect to peer: %s", err))
		}
		return
	}

	h.SetStreamHandler(protocolID, handleStream)

	Must(dht.New(context.Background(), h))

	for _, addr := range dht.DefaultBootstrapPeers {
		pi := Must(peer.AddrInfoFromP2pAddr(addr))
		h.Connect(context.Background(), *pi)
	}

	slog.Info("Waiting for NAT mapping and address discovery...")
	time.Sleep(10 * time.Second)

	directAddrs, relayAddrs := categorizeAddrs(h.Addrs())
	slog.Info("Direct IPv4 Addresses:\n" +
		strings.Join(Map(directAddrs, func(addr multiaddr.Multiaddr) string {
			return color.GreenString(addr.String())
		}), "\n"))

	slog.Info("Relay Addresses:\n" +
		strings.Join(Map(relayAddrs, func(addr multiaddr.Multiaddr) string {
			return color.GreenString(addr.String() + "/p2p/" + h.ID().String())
		}), "\n"))

	slog.Info("Waiting for connection...")
	select {}
}

func connect(h host.Host, addrStr string) error {
	maddr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return fmt.Errorf("error parsing multiaddr: %w", err)
	}

	peerAddr, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return fmt.Errorf("error parsing peer address: %w", err)
	}

	if err := h.Connect(context.Background(), *peerAddr); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}
	slog.Info("Connected to " + color.GreenString(peerAddr.ID.String()))

	s, openErr := h.NewStream(context.Background(), peerAddr.ID, protocolID)
	if openErr != nil {
		return fmt.Errorf("failed to open stream to peer, peer=%s, error=%w", peerAddr.ID, openErr)
	}

	slog.Info("Successfully opened stream with protocol", slog.String("protocol", protocolID))
	handleStream(s)
	return nil
}

func convertPeers(peers []multiaddr.Multiaddr) []peer.AddrInfo {
	var pinfos []peer.AddrInfo
	for _, addr := range peers {
		if info, err := peer.AddrInfoFromP2pAddr(addr); err == nil {
			pinfos = append(pinfos, *info)
		}
	}
	return pinfos
}

func categorizeAddrs(addrs []multiaddr.Multiaddr) ([]multiaddr.Multiaddr, []multiaddr.Multiaddr) {
	var directAddrs, relayAddrs []multiaddr.Multiaddr
	for _, addr := range addrs {
		addrStr := addr.String()
		if strings.Contains(addrStr, "/ip4/") {
			if strings.Contains(addrStr, "/p2p-circuit") {
				relayAddrs = append(relayAddrs, addr)
			} else {
				directAddrs = append(directAddrs, addr)
			}
		}
	}
	return directAddrs, relayAddrs
}
