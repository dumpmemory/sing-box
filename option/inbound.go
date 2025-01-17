package option

import (
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"

	C "github.com/sagernet/sing-box/constant"

	"github.com/goccy/go-json"
)

type _Inbound struct {
	Type               string                    `json:"type"`
	Tag                string                    `json:"tag,omitempty"`
	DirectOptions      DirectInboundOptions      `json:"-"`
	SocksOptions       SimpleInboundOptions      `json:"-"`
	HTTPOptions        SimpleInboundOptions      `json:"-"`
	MixedOptions       SimpleInboundOptions      `json:"-"`
	ShadowsocksOptions ShadowsocksInboundOptions `json:"-"`
}

type Inbound _Inbound

func (h Inbound) Equals(other Inbound) bool {
	return h.Type == other.Type &&
		h.Tag == other.Tag &&
		h.DirectOptions == other.DirectOptions &&
		h.SocksOptions.Equals(other.SocksOptions) &&
		h.HTTPOptions.Equals(other.HTTPOptions) &&
		h.MixedOptions.Equals(other.MixedOptions) &&
		h.ShadowsocksOptions.Equals(other.ShadowsocksOptions)
}

func (h Inbound) MarshalJSON() ([]byte, error) {
	var v any
	switch h.Type {
	case C.TypeDirect:
		v = h.DirectOptions
	case C.TypeSocks:
		v = h.SocksOptions
	case C.TypeHTTP:
		v = h.HTTPOptions
	case C.TypeMixed:
		v = h.MixedOptions
	case C.TypeShadowsocks:
		v = h.ShadowsocksOptions
	default:
		return nil, E.New("unknown inbound type: ", h.Type)
	}
	return MarshallObjects((_Inbound)(h), v)
}

func (h *Inbound) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, (*_Inbound)(h))
	if err != nil {
		return err
	}
	var v any
	switch h.Type {
	case C.TypeDirect:
		v = &h.DirectOptions
	case C.TypeSocks:
		v = &h.SocksOptions
	case C.TypeHTTP:
		v = &h.HTTPOptions
	case C.TypeMixed:
		v = &h.MixedOptions
	case C.TypeShadowsocks:
		v = &h.ShadowsocksOptions
	default:
		return nil
	}
	err = UnmarshallExcluded(bytes, (*_Inbound)(h), v)
	if err != nil {
		return E.Cause(err, "inbound options")
	}
	return nil
}

type ListenOptions struct {
	Listen                   ListenAddress  `json:"listen"`
	Port                     uint16         `json:"listen_port"`
	TCPFastOpen              bool           `json:"tcp_fast_open,omitempty"`
	UDPTimeout               int64          `json:"udp_timeout,omitempty"`
	SniffEnabled             bool           `json:"sniff,omitempty"`
	SniffOverrideDestination bool           `json:"sniff_override_destination,omitempty"`
	DomainStrategy           DomainStrategy `json:"domain_strategy,omitempty"`
}

type SimpleInboundOptions struct {
	ListenOptions
	Users []auth.User `json:"users,omitempty"`
}

func (o SimpleInboundOptions) Equals(other SimpleInboundOptions) bool {
	return o.ListenOptions == other.ListenOptions &&
		common.ComparableSliceEquals(o.Users, other.Users)
}

type DirectInboundOptions struct {
	ListenOptions
	Network         NetworkList `json:"network,omitempty"`
	OverrideAddress string      `json:"override_address,omitempty"`
	OverridePort    uint16      `json:"override_port,omitempty"`
}

type ShadowsocksInboundOptions struct {
	ListenOptions
	Network      NetworkList              `json:"network,omitempty"`
	Method       string                   `json:"method"`
	Password     string                   `json:"password"`
	Users        []ShadowsocksUser        `json:"users,omitempty"`
	Destinations []ShadowsocksDestination `json:"destinations,omitempty"`
}

func (o ShadowsocksInboundOptions) Equals(other ShadowsocksInboundOptions) bool {
	return o.ListenOptions == other.ListenOptions &&
		o.Network == other.Network &&
		o.Method == other.Method &&
		o.Password == other.Password &&
		common.ComparableSliceEquals(o.Users, other.Users) &&
		common.ComparableSliceEquals(o.Destinations, other.Destinations)
}

type ShadowsocksUser struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type ShadowsocksDestination struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	ServerOptions
}
