package sniff

import (
	"context"
	"encoding/binary"
	"io"
	"os"
	"time"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/task"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"

	"golang.org/x/net/dns/dnsmessage"
)

func StreamDomainNameQuery(readCtx context.Context, reader io.Reader) (*adapter.InboundContext, error) {
	var length uint16
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	if length > 512 {
		return nil, os.ErrInvalid
	}
	_buffer := buf.StackNewSize(int(length))
	defer common.KeepAlive(_buffer)
	buffer := common.Dup(_buffer)
	defer buffer.Release()

	readCtx, cancel := context.WithTimeout(readCtx, time.Millisecond*100)
	err = task.Run(readCtx, func() error {
		return common.Error(buffer.ReadFullFrom(reader, buffer.FreeLen()))
	})
	cancel()
	if err != nil {
		return nil, err
	}
	return DomainNameQuery(readCtx, buffer.Bytes())
}

func DomainNameQuery(ctx context.Context, packet []byte) (*adapter.InboundContext, error) {
	var parser dnsmessage.Parser
	_, err := parser.Start(packet)
	if err != nil {
		return nil, err
	}
	question, err := parser.Question()
	if err != nil {
		return nil, os.ErrInvalid
	}
	domain := question.Name.String()
	if question.Class == dnsmessage.ClassINET && (question.Type == dnsmessage.TypeA || question.Type == dnsmessage.TypeAAAA) && IsDomainName(domain) {
		return &adapter.InboundContext{Protocol: C.ProtocolDNS, Domain: domain}, nil
	}
	return nil, os.ErrInvalid
}
