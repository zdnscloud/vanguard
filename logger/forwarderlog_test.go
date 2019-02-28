package logger

import (
	"fmt"
	"testing"

	ut "github.com/zdnscloud/cement/unittest"
	"vanguard/config"
)

func TestForwarderLog(t *testing.T) {
	var conf config.VanguardConf
	conf.Logger.Forwarderlog.Enabled = true
	conf.Logger.Forwarderlog.Path = "forwarder.log"
	conf.Logger.Forwarderlog.FileSize = 5000000
	conf.Logger.Forwarderlog.Versions = 5

	ql, err := NewForwarderlog(&conf)
	ut.Assert(t, err == nil, "failed to create query log")

	for i := 0; i < 500000; i++ {
		ql.LogWrite(nil, fmt.Sprintf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx%d", i))
	}
}
