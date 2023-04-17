package libp2p

import (
	rcmgr "github.com/libp2p/go-libp2p-resource-manager"
)

func init() {
	rcmgr.DefaultLimits.TransientBaseLimit.ConnsInbound = 1024
	rcmgr.DefaultLimits.TransientBaseLimit.ConnsOutbound = 1024
	rcmgr.DefaultLimits.TransientBaseLimit.Conns = 2048

	rcmgr.DefaultLimits.SystemBaseLimit.Conns = 2048
	rcmgr.DefaultLimits.SystemBaseLimit.ConnsInbound = 1024
	rcmgr.DefaultLimits.SystemBaseLimit.ConnsOutbound = 1024
}
