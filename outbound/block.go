package outbound

import (
	"context"
	"io"
	"net"

	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
)

var _ adapter.Outbound = (*Block)(nil)

type Block struct {
	myOutboundAdapter
}

func NewBlock(logger log.Logger, tag string) *Block {
	return &Block{
		myOutboundAdapter{
			protocol: C.TypeBlock,
			logger:   logger,
			tag:      tag,
			network:  []string{C.NetworkTCP, C.NetworkUDP},
		},
	}
}

func (h *Block) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	h.logger.WithContext(ctx).Info("blocked connection to ", destination)
	return nil, io.EOF
}

func (h *Block) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	h.logger.WithContext(ctx).Info("blocked packet connection to ", destination)
	return nil, io.EOF
}

func (h *Block) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	conn.Close()
	h.logger.WithContext(ctx).Info("blocked connection to ", metadata.Destination)
	return nil
}

func (h *Block) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	conn.Close()
	h.logger.WithContext(ctx).Info("blocked packet connection to ", metadata.Destination)
	return nil
}
