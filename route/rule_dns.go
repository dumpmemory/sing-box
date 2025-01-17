package route

import (
	"strings"

	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

func NewDNSRule(router adapter.Router, logger log.Logger, options option.DNSRule) (adapter.Rule, error) {
	if common.IsEmptyByEquals(options) {
		return nil, E.New("empty rule config")
	}
	switch options.Type {
	case "", C.RuleTypeDefault:
		if !options.DefaultOptions.IsValid() {
			return nil, E.New("missing conditions")
		}
		if options.DefaultOptions.Server == "" {
			return nil, E.New("missing server field")
		}
		return NewDefaultDNSRule(router, logger, options.DefaultOptions)
	case C.RuleTypeLogical:
		if !options.LogicalOptions.IsValid() {
			return nil, E.New("missing conditions")
		}
		if options.LogicalOptions.Server == "" {
			return nil, E.New("missing server field")
		}
		return NewLogicalDNSRule(router, logger, options.LogicalOptions)
	default:
		return nil, E.New("unknown rule type: ", options.Type)
	}
}

var _ adapter.Rule = (*DefaultDNSRule)(nil)

type DefaultDNSRule struct {
	items    []RuleItem
	outbound string
}

func NewDefaultDNSRule(router adapter.Router, logger log.Logger, options option.DefaultDNSRule) (*DefaultDNSRule, error) {
	rule := &DefaultDNSRule{
		outbound: options.Server,
	}
	if len(options.Inbound) > 0 {
		item := NewInboundRule(options.Inbound)
		rule.items = append(rule.items, item)
	}
	if options.Network != "" {
		switch options.Network {
		case C.NetworkTCP, C.NetworkUDP:
			item := NewNetworkItem(options.Network)
			rule.items = append(rule.items, item)
		default:
			return nil, E.New("invalid network: ", options.Network)
		}
	}
	if len(options.Protocol) > 0 {
		item := NewProtocolItem(options.Protocol)
		rule.items = append(rule.items, item)
	}
	if len(options.Domain) > 0 || len(options.DomainSuffix) > 0 {
		item := NewDomainItem(options.Domain, options.DomainSuffix)
		rule.items = append(rule.items, item)
	}
	if len(options.DomainKeyword) > 0 {
		item := NewDomainKeywordItem(options.DomainKeyword)
		rule.items = append(rule.items, item)
	}
	if len(options.DomainRegex) > 0 {
		item, err := NewDomainRegexItem(options.DomainRegex)
		if err != nil {
			return nil, E.Cause(err, "domain_regex")
		}
		rule.items = append(rule.items, item)
	}
	if len(options.Geosite) > 0 {
		item := NewGeositeItem(router, logger, options.Geosite)
		rule.items = append(rule.items, item)
	}
	if len(options.SourceGeoIP) > 0 {
		item := NewGeoIPItem(router, logger, true, options.SourceGeoIP)
		rule.items = append(rule.items, item)
	}
	if len(options.SourceIPCIDR) > 0 {
		item, err := NewIPCIDRItem(true, options.SourceIPCIDR)
		if err != nil {
			return nil, E.Cause(err, "source_ipcidr")
		}
		rule.items = append(rule.items, item)
	}
	if len(options.SourcePort) > 0 {
		item := NewPortItem(true, options.SourcePort)
		rule.items = append(rule.items, item)
	}
	if len(options.Port) > 0 {
		item := NewPortItem(false, options.Port)
		rule.items = append(rule.items, item)
	}
	if len(options.Outbound) > 0 {
		item := NewOutboundRule(options.Outbound)
		rule.items = append(rule.items, item)
	}
	return rule, nil
}

func (r *DefaultDNSRule) Start() error {
	for _, item := range r.items {
		err := common.Start(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DefaultDNSRule) Close() error {
	for _, item := range r.items {
		err := common.Close(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DefaultDNSRule) UpdateGeosite() error {
	for _, item := range r.items {
		if geositeItem, isSite := item.(*GeositeItem); isSite {
			err := geositeItem.Update()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *DefaultDNSRule) Match(metadata *adapter.InboundContext) bool {
	for _, item := range r.items {
		if !item.Match(metadata) {
			return false
		}
	}
	return true
}

func (r *DefaultDNSRule) Outbound() string {
	return r.outbound
}

func (r *DefaultDNSRule) String() string {
	return strings.Join(common.Map(r.items, F.ToString0[RuleItem]), " ")
}

var _ adapter.Rule = (*LogicalRule)(nil)

type LogicalDNSRule struct {
	mode     string
	rules    []*DefaultDNSRule
	outbound string
}

func (r *LogicalDNSRule) UpdateGeosite() error {
	for _, rule := range r.rules {
		err := rule.UpdateGeosite()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *LogicalDNSRule) Start() error {
	for _, rule := range r.rules {
		err := rule.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *LogicalDNSRule) Close() error {
	for _, rule := range r.rules {
		err := rule.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewLogicalDNSRule(router adapter.Router, logger log.Logger, options option.LogicalDNSRule) (*LogicalDNSRule, error) {
	r := &LogicalDNSRule{
		rules:    make([]*DefaultDNSRule, len(options.Rules)),
		outbound: options.Server,
	}
	switch options.Mode {
	case C.LogicalTypeAnd:
		r.mode = C.LogicalTypeAnd
	case C.LogicalTypeOr:
		r.mode = C.LogicalTypeOr
	default:
		return nil, E.New("unknown logical mode: ", options.Mode)
	}
	for i, subRule := range options.Rules {
		rule, err := NewDefaultDNSRule(router, logger, subRule)
		if err != nil {
			return nil, E.Cause(err, "sub rule[", i, "]")
		}
		r.rules[i] = rule
	}
	return r, nil
}

func (r *LogicalDNSRule) Match(metadata *adapter.InboundContext) bool {
	if r.mode == C.LogicalTypeAnd {
		return common.All(r.rules, func(it *DefaultDNSRule) bool {
			return it.Match(metadata)
		})
	} else {
		return common.Any(r.rules, func(it *DefaultDNSRule) bool {
			return it.Match(metadata)
		})
	}
}

func (r *LogicalDNSRule) Outbound() string {
	return r.outbound
}

func (r *LogicalDNSRule) String() string {
	var op string
	switch r.mode {
	case C.LogicalTypeAnd:
		op = "&&"
	case C.LogicalTypeOr:
		op = "||"
	}
	return "logical(" + strings.Join(common.Map(r.rules, F.ToString0[*DefaultDNSRule]), " "+op+" ") + ")"
}
