package adapters

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

const fileLimit = 1024
const branchLimit = 512

func (g *Git) details(ctx context.Context, root string) (domain.Details, error) {
	result := domain.Details{Files: []domain.FileChange{}, Branches: []domain.Branch{}}
	output, err := g.runner.Run(ctx, root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return result, fmt.Errorf("read Git file inventory: %w", err)
	}
	result.Files, result.FilesTruncated = parseFiles(output)
	output, err = g.runner.Run(ctx, root, "for-each-ref", "--count=513", "--sort=refname", "--format=%(refname:short)%00%(objectname)%00%(HEAD)%00%(upstream:short)", "refs/heads/")
	if err != nil {
		return result, fmt.Errorf("read Git branches: %w", err)
	}
	for _, line := range bytes.Split(bytes.TrimSuffix(output, []byte{'\n'}), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		fields := strings.Split(string(line), "\x00")
		if len(fields) != 4 {
			return result, fmt.Errorf("invalid Git branch inventory")
		}
		if len(result.Branches) == branchLimit {
			result.BranchesTruncated = true
			break
		}
		result.Branches = append(result.Branches, domain.Branch{Name: fields[0], Head: fields[1], Current: fields[2] == "*", Upstream: fields[3]})
	}
	return result, nil
}

func parseFiles(output []byte) ([]domain.FileChange, bool) {
	result := []domain.FileChange{}
	parts := bytes.Split(output, []byte{0})
	for index := 0; index < len(parts); index++ {
		record := parts[index]
		if len(record) < 4 || record[2] != ' ' {
			continue
		}
		if len(result) == fileLimit {
			return result, true
		}
		change := domain.FileChange{Path: string(record[3:]), Index: string(record[0]), WorkingTree: string(record[1]), Untracked: record[0] == '?'}
		change.Conflicted = record[0] == 'U' || record[1] == 'U' || string(record[:2]) == "AA" || string(record[:2]) == "DD"
		if record[0] == 'R' || record[0] == 'C' || record[1] == 'R' || record[1] == 'C' {
			index++
			if index < len(parts) {
				change.OriginalPath = string(parts[index])
			}
		}
		result = append(result, change)
	}
	return result, false
}
