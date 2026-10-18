package process

import (
	"context"
	"sort"

	"envarium.dev/envarium/internal/runtime/domain"
)

// Bindings observes current socket addresses and revalidates their owners. A
// running process and a declared port alone are not evidence of a live binding.
func (inspector gopsutilInspector) Bindings(ctx context.Context, ports []domain.PublishedPort, owners []domain.ProcessIdentity) ([]domain.PublishedPort, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return []domain.PublishedPort{}, ctx.Err()
	}
	connections, err := inspector.readConnections(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	identities := map[int32]domain.ProcessIdentity{}
	seen := map[domain.PublishedPort]bool{}
	result := []domain.PublishedPort{}
	for _, connection := range connections {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if connection.Status != "LISTEN" || connection.Pid <= 0 {
			continue
		}
		for _, port := range ports {
			if port.Protocol != "tcp" || port.HostPort != int(connection.Laddr.Port) {
				continue
			}
			identity, ok := identities[connection.Pid]
			if !ok {
				identity, _ = inspector.Snapshot(ctx, connection.Pid)
				identities[connection.Pid] = identity
			}
			if !listenerOwned(identity, owners) {
				continue
			}
			binding := port
			binding.HostIP = connection.Laddr.IP
			if binding.HostIP == "" {
				binding.HostIP = "0.0.0.0"
			}
			if !seen[binding] {
				seen[binding] = true
				result = append(result, binding)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].HostPort != result[j].HostPort {
			return result[i].HostPort < result[j].HostPort
		}
		return result[i].HostIP < result[j].HostIP
	})
	return result, nil
}
func listenerOwned(listener domain.ProcessIdentity, owners []domain.ProcessIdentity) bool {
	if listener.PID <= 0 || listener.Fingerprint == "" {
		return false
	}
	for _, owner := range owners {
		if listener.PID == owner.PID && listener.Fingerprint == owner.Fingerprint {
			return true
		}
		if owner.ProcessGroup > 0 && listener.ProcessGroup == owner.ProcessGroup {
			return true
		}
	}
	return false
}

func (d *Driver) observedPorts(ctx context.Context, requested bool, ports []domain.PublishedPort, owners []domain.ProcessIdentity) ([]domain.PublishedPort, error) {
	if !requested {
		return ports, nil
	}
	return d.observedInspector.Bindings(ctx, ports, owners)
}
