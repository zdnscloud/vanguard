package echoip

import (
	"github.com/zdnscloud/g53"
	"github.com/zdnscloud/vanguard/config"
	"github.com/zdnscloud/vanguard/core"
)

type EchoIP struct {
	core.DefaultHandler
	zone *g53.Name
	soa  *g53.RRset
	ns   *g53.RRset
	glue *g53.RRset
}

func NewEchoip(conf *config.VanguardConf) core.DNSQueryHandler {
	e := &EchoIP{}
	e.ReloadConfig(conf)
	return e
}

func (e *EchoIP) ReloadConfig(conf *config.VanguardConf) {
	e.generateNSAndSOA(conf.EchoIP.Zone, conf.EchoIP.Addrs)
}

func (e *EchoIP) HandleQuery(ctx *core.Context) {
	client := &ctx.Client
	qname := client.Request.Question.Name
	qtype := client.Request.Question.Type

	if qname.IsSubDomain(e.zone) == false {
		core.PassToNext(e, ctx)
		return
	}

	if qname.Equals(e.zone) && qtype == g53.RR_NS {
		e.returnNS(client)
		return
	}

	if qname.Equals(e.zone) && qtype == g53.RR_SOA {
		e.returnSOA(client)
		return
	}

	if qtype != g53.RR_A {
		e.returnEmptyAnswer(client)
		return
	}

	if qname.Equals(e.glue.Name) {
		e.returnGlue(client)
		return
	}

	ip := e.extractIpaddress(qname)
	if ip == "" {
		e.returnEmptyAnswer(client)
		return
	}

	a, err := g53.AFromString(ip)
	if err == nil {
		e.returnA(qname, a, client)
	} else {
		e.returnEmptyAnswer(client)
	}
}

func (e *EchoIP) extractIpaddress(qname *g53.Name) string {
	relativeLabelCount := qname.LabelCount() - e.zone.LabelCount()
	if relativeLabelCount == 4 {
		ip, _ := qname.Split(0, 4)
		return ip.String(true)
	} else if relativeLabelCount == 5 {
		ip, _ := qname.Split(1, 4)
		return ip.String(true)
	}
	return ""
}

func (e *EchoIP) returnEmptyAnswer(client *core.Client) {
	response := client.Request.MakeResponse()
	response.Header.SetFlag(g53.FLAG_AA, true)
	response.Header.Rcode = g53.R_NOERROR
	response.AddRRset(g53.AuthSection, e.soa)
	client.Response = response
}

func (e *EchoIP) returnNS(client *core.Client) {
	response := client.Request.MakeResponse()
	response.Header.SetFlag(g53.FLAG_AA, true)
	response.AddRRset(g53.AnswerSection, e.ns)
	response.AddRRset(g53.AdditionalSection, e.glue)
	client.Response = response
}

func (e *EchoIP) returnSOA(client *core.Client) {
	response := client.Request.MakeResponse()
	response.Header.SetFlag(g53.FLAG_AA, true)
	response.AddRRset(g53.AnswerSection, e.soa)
	response.AddRRset(g53.AuthSection, e.ns)
	response.AddRRset(g53.AdditionalSection, e.glue)
	client.Response = response
}

func (e *EchoIP) returnGlue(client *core.Client) {
	response := client.Request.MakeResponse()
	response.Header.SetFlag(g53.FLAG_AA, true)
	response.AddRRset(g53.AnswerSection, e.glue)
	response.AddRRset(g53.AuthSection, e.ns)
	client.Response = response
}

func (e *EchoIP) returnA(qname *g53.Name, a g53.Rdata, client *core.Client) {
	response := client.Request.MakeResponse()
	response.Header.SetFlag(g53.FLAG_AA, true)
	response.AddRR(g53.AnswerSection, qname, g53.RR_A, g53.CLASS_IN, g53.RRTTL(60), a, false)
	response.AddRRset(g53.AuthSection, e.ns)
	response.AddRRset(g53.AdditionalSection, e.glue)
	client.Response = response
}

func (e *EchoIP) generateNSAndSOA(zone string, addrs []string) {
	e.zone = g53.NameFromStringUnsafe(zone)
	nsName := g53.NameFromStringUnsafe("ns." + zone)
	e.ns = &g53.RRset{
		Name:   e.zone,
		Type:   g53.RR_NS,
		Class:  g53.CLASS_IN,
		Ttl:    g53.RRTTL(3600),
		Rdatas: []g53.Rdata{&g53.NS{Name: nsName}},
	}

	var rdatas []g53.Rdata
	for _, addr := range addrs {
		rdata, err := g53.AFromString(addr)
		if err != nil {
			panic("echo ip addr isn't valid")
		}
		rdatas = append(rdatas, rdata)
	}
	e.glue = &g53.RRset{
		Name:   nsName,
		Type:   g53.RR_A,
		Class:  g53.CLASS_IN,
		Ttl:    g53.RRTTL(3600),
		Rdatas: rdatas,
	}

	e.soa = &g53.RRset{
		Name:  e.zone,
		Type:  g53.RR_SOA,
		Class: g53.CLASS_IN,
		Ttl:   g53.RRTTL(3600),
		Rdatas: []g53.Rdata{
			&g53.SOA{
				MName:   nsName,
				RName:   g53.NameFromStringUnsafe("root." + zone),
				Serial:  1000,
				Refresh: 7200,
				Retry:   3600,
				Expire:  2419200,
				Minimum: 21600,
			}},
	}
}
