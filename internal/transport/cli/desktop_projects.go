package cli

import (
	"context"

	"envarium.dev/envarium/internal/transport/contract/generated"
)

type desktopStatusClient interface {
	StatusBatch(context.Context, []string) (generated.ProjectObservationBatch, error)
}

func desktopProjects(ctx context.Context, client desktopStatusClient, projects []generated.Project) ([]desktopProjectSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := make([]desktopProjectSnapshot, len(projects))
	ids := []string{}
	indices := map[string]int{}
	for i, project := range projects {
		result[i] = desktopProjectSnapshot{Project: project, Warnings: []string{}}
		if project.TrustState == generated.ProjectTrustStateTrusted {
			ids = append(ids, project.Id)
			indices[project.Id] = i
		}
	}
	for start := 0; start < len(ids); start += 64 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		chunk := ids[start:min(start+64, len(ids))]
		batch, err := client.StatusBatch(ctx, chunk)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err == nil {
			wanted := make(map[string]struct{}, len(chunk))
			for _, id := range chunk {
				wanted[id] = struct{}{}
			}
			for _, value := range batch.Projects {
				if _, ok := wanted[value.ProjectId]; !ok {
					continue
				}
				item := &result[indices[value.ProjectId]]
				item.Runtime, item.Health = value.Runtime, value.Health
				delete(wanted, value.ProjectId)
			}
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, id := range chunk {
			item := &result[indices[id]]
			if item.Runtime == nil {
				item.Warnings = append(item.Warnings, "runtime observation unavailable")
			}
			if item.Health == nil {
				item.Warnings = append(item.Warnings, "health observation unavailable")
			}
		}
	}
	return result, nil
}
