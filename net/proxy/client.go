package proxy

import (
	"bufio"
	"io"
	"time"

	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/net/libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Client struct {
	stream network.Stream
	reader *bufio.Reader
	sender *bufio.Writer
}

func NewClient(host *libp2p.Host, proxy peer.AddrInfo) (*Client, error) {
	stream, err := host.NewStream(proxy, ProtocolID)
	if err != nil {
		return nil, err
	}
	return &Client{
		stream: stream,
		reader: bufio.NewReader(stream),
		sender: bufio.NewWriter(stream),
	}, nil
}

func (c *Client) Close() error {
	return c.stream.Reset()
}

func (c *Client) Decide() (*net.Decision, error) {
	message := make([]byte, 16)
	_, err := io.ReadFull(c.reader, message)
	if err != nil {
		return nil, err
	}
	decision := DecodeDecision(message)
	decision.Timestamp = time.Now()
	return decision, nil
}

func (c *Client) Propose(value []byte) error {
	header := make([]byte, 4)
	encoding.PutUint32(header, uint32(len(value)))
	_, err := c.sender.Write(header)
	if err == nil {
		_, err = c.sender.Write(value)
		err = c.sender.Flush()
	}
	return err
}
