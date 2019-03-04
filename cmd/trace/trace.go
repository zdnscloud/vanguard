package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/zdnscloud/g53"
	"github.com/zdnscloud/vanguard/config"
	"github.com/zdnscloud/vanguard/core"
	"github.com/zdnscloud/vanguard/logger"
	"github.com/zdnscloud/vanguard/resolver"
	"github.com/zdnscloud/vanguard/resolver/querysource"
	"github.com/zdnscloud/vanguard/resolver/recursor"
	view "github.com/zdnscloud/vanguard/viewselector"
)

var (
	name string
	typ  string
)

func init() {
	flag.StringVar(&name, "n", "www.zdns.cn.", "query name")
	flag.StringVar(&typ, "t", "a", "query type")
}

func main() {
	flag.Parse()

	logger.UseDefaultLogger("info")
	conf := &config.VanguardConf{}
	conf.Recursor = []config.RecursorInView{
		config.RecursorInView{
			Enable: true,
			View:   "default",
		},
	}
	view.NewSelectorMgr(conf)
	querysource.NewQuerySourceManager(conf)

	r := resolver.NewCNameHandler(recursor.NewRecursor(conf))
	r.ReloadConfig(conf)

	qname, err := g53.NameFromString(name)
	if err != nil {
		fmt.Printf("name isn't valid")
		return
	}

	qtype, err := g53.TypeFromString(typ)
	if err != nil {
		fmt.Printf("qtype isn't valid")
		return
	}

	var client core.Client
	client.Request = g53.MakeQuery(qname, qtype, 1024, false)
	client.Addr, _ = net.ResolveUDPAddr("udp", "127.0.0.1:0")
	client.View = "default"
	r.Resolve(&client)
	if client.Response != nil {
		fmt.Printf("%s\n", client.Response.String())
	}
}
