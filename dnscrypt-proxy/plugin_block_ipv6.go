package main

import (
	"strings"
	"time"

	"github.com/miekg/dns"
)

type PluginBlockIPv6 struct{}

func (plugin *PluginBlockIPv6) Name() string {
	return "block_ipv6"
}

func (plugin *PluginBlockIPv6) Description() string {
	return "Immediately return a synthetic response to AAAA queries."
}

func (plugin *PluginBlockIPv6) Init(proxy *Proxy) error {
	return nil
}

func (plugin *PluginBlockIPv6) Drop() error {
	return nil
}

func (plugin *PluginBlockIPv6) Reload() error {
	return nil
}

func (plugin *PluginBlockIPv6) Eval(pluginsState *PluginsState, msg *dns.Msg) error {
	questions := msg.Question
	if len(questions) != 1 {
		return nil
	}
	question := questions[0]
	if question.Qclass != dns.ClassINET || question.Qtype != dns.TypeAAAA {
		return nil
	}
	synth, err := EmptyResponseFromMessage(msg)
	if err != nil {
		return err
	}
	hinfo := new(dns.HINFO)
	hinfo.Hdr = dns.RR_Header{Name: question.Name, Rrtype: dns.TypeHINFO,
		Class: dns.ClassINET, Ttl: 86400}
	hinfo.Cpu = "AAAA queries have been locally blocked by dnscrypt-proxy"
	hinfo.Os = "Set block_ipv6 to false to disable this feature"
	synth.Answer = []dns.RR{hinfo}
	qName := question.Name
	i := strings.Index(qName, ".")
	parentZone := "."
	if !(i < 0 || i+1 >= len(qName)) {
		parentZone = qName[i+1:]
	}
	dotParentZone := "."
	if parentZone != "." {
		dotParentZone += parentZone
	}
	soa := new(dns.SOA)
	soa.Mbox = "h" + dotParentZone
	soa.Ns = "n" + dotParentZone
	soa.Serial = uint32(time.Now().Unix())
	soa.Refresh = 10000
	soa.Minttl = 2400
	soa.Expire = 604800
	soa.Retry = 300
	soa.Hdr = dns.RR_Header{Name: parentZone, Rrtype: dns.TypeSOA,
		Class: dns.ClassINET, Ttl: 60}
	synth.Ns = []dns.RR{soa}
	pluginsState.synthResponse = synth
	pluginsState.action = PluginsActionSynth
	return nil
}
