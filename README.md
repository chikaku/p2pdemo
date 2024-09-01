# p2pdemo

Peer to peer communication demo using libp2p and go-libp2p-noise.

## Usage

```bash
# client 1
go run .
...
/ip4/x.x.x.x/udp/4001/quic-v1/webtransport/certhash/peer-id/p2p/p2p-circuit/p2p/peer-id
...

# client 2
go run . /ip4/x.x.x.x/udp/4001/quic-v1/webtransport/certhash/peer-id/p2p/p2p-circuit/p2p/peer-id
```
