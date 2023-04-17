package libp2p

import (
	"github.com/libp2p/go-libp2p-core/network"
)

// NotifyConnections adds to the returned channel all connections accepted.
// When the returned notification channel is full, notifications are dropped.
func (h *Host) NotifyConnections(channelSize int) <-chan network.Conn {
	notifee := notifee{
		queue: make(chan network.Conn, channelSize),
	}
	notifier := &network.NotifyBundle{
		ConnectedF: notifee.connectedF,
	}
	h.Host.Network().Notify(notifier)
	return notifee.queue
}

type notifee struct {
	queue chan network.Conn
}

func (notif *notifee) connectedF(n network.Network, conn network.Conn) {
	select {
	case notif.queue <- conn:
	default:
		// Notification is dropped to prevent the host from blocking
	}
}
