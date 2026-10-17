//go:build !darwin

package sockets

import (
	"context"
	gnet "github.com/shirou/gopsutil/v4/net"
)

// Connections reads fresh socket evidence through the platform's portable API.
func Connections(ctx context.Context, kind string) ([]gnet.ConnectionStat, error) {
	return gnet.ConnectionsWithContext(ctx, kind)
}
