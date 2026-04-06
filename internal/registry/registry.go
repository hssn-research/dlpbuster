// Package registry provides compile-time registration of all exfil channels.
// It lives in a separate package to avoid the import cycle that would arise
// if sub-packages (e.g. channels/dns) imported the parent channels package
// for its types AND the parent imported them for registration.
package registry

import (
	"fmt"

	"github.com/hssn-research/dlpbuster/internal/channels"
	"github.com/hssn-research/dlpbuster/internal/channels/azure"
	"github.com/hssn-research/dlpbuster/internal/channels/dns"
	"github.com/hssn-research/dlpbuster/internal/channels/gcs"
	"github.com/hssn-research/dlpbuster/internal/channels/https"
	"github.com/hssn-research/dlpbuster/internal/channels/icmp"
	"github.com/hssn-research/dlpbuster/internal/channels/s3"
	"github.com/hssn-research/dlpbuster/internal/channels/smtp"
	"github.com/hssn-research/dlpbuster/internal/channels/webhook"
)

// All returns every registered channel in display order.
func All() []channels.Channel {
	return []channels.Channel{
		dns.New(),
		https.New(),
		icmp.New(),
		s3.New(),
		gcs.New(),
		azure.New(),
		smtp.New(),
		webhook.New(),
	}
}

// Lookup returns a channel by name, or an error if not found.
func Lookup(name string) (channels.Channel, error) {
	for _, ch := range All() {
		if ch.Name() == name {
			return ch, nil
		}
	}
	return nil, fmt.Errorf("unknown channel %q", name)
}
