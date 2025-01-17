package route

import (
	"strings"

	"github.com/sagernet/sing/common"
	F "github.com/sagernet/sing/common/format"

	"github.com/sagernet/sing-box/adapter"
)

var _ RuleItem = (*PortItem)(nil)

type PortItem struct {
	ports    []uint16
	portMap  map[uint16]bool
	isSource bool
}

func NewPortItem(isSource bool, ports []uint16) *PortItem {
	portMap := make(map[uint16]bool)
	for _, port := range ports {
		portMap[port] = true
	}
	return &PortItem{
		ports:    ports,
		portMap:  portMap,
		isSource: isSource,
	}
}

func (r *PortItem) Match(metadata *adapter.InboundContext) bool {
	if r.isSource {
		return r.portMap[metadata.Source.Port]
	} else {
		return r.portMap[metadata.Destination.Port]
	}
}

func (r *PortItem) String() string {
	var description string
	if r.isSource {
		description = "source_port="
	} else {
		description = "port="
	}
	pLen := len(r.ports)
	if pLen == 1 {
		description += F.ToString(r.ports[0])
	} else {
		description += "[" + strings.Join(common.Map(r.ports, F.ToString0[uint16]), " ") + "]"
	}
	return description
}
