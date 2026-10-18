package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"envarium.dev/envarium/internal/plugins/domain"
	pluginsdk "envarium.dev/envarium/sdk/plugin"
)

func (s *Service) text(value string) string {
	value, _ = s.redactor.RedactText(value)
	if len(value) > 2048 {
		value = value[:2048]
	}
	return value
}

func (s *Service) call(ctx context.Context, invocation Invocation) ([]domain.LogEntry, error) {
	logs, err := s.runner.Call(ctx, invocation)
	for i := range logs {
		logs[i].Message = s.text(logs[i].Message)
		logs[i].Level = s.text(logs[i].Level)
	}
	if err != nil {
		return logs, sanitizedError{cause: err, message: s.text(err.Error())}
	}
	switch result := invocation.Result.(type) {
	case *pluginsdk.HealthResult:
		if !s.safeIdentifier(result.Status) {
			return logs, errors.New("plugin returned a sensitive health status")
		}
		result.Message = s.text(result.Message)
	case *pluginsdk.InspectResult:
		result.Summary = s.text(result.Summary)
		for i := range result.Warnings {
			result.Warnings[i] = s.text(result.Warnings[i])
		}
		for i := range result.Facts {
			f := &result.Facts[i]
			if !s.safeIdentifier(f.ID) {
				return logs, errors.New("plugin returned a sensitive fact identifier")
			}
			f.Label, f.Value, f.Source, f.Warning = s.text(f.Label), s.text(f.Value), s.text(f.Source), s.text(f.Warning)
		}
		for i := range result.Actions {
			if !s.safeIdentifier(result.Actions[i].ID) || !s.safeIdentifier(result.Actions[i].Risk) {
				return logs, errors.New("plugin returned a sensitive action identifier or risk")
			}
			result.Actions[i].Name = s.text(result.Actions[i].Name)
			result.Actions[i].Description = s.text(result.Actions[i].Description)
		}
	case *pluginsdk.OperateResult:
		if !s.safeIdentifier(result.Status) {
			return logs, errors.New("plugin returned a sensitive operation status")
		}
		result.Summary = s.text(result.Summary)
		if len(result.Output) > 0 {
			var value any
			decoder := json.NewDecoder(bytes.NewReader(result.Output))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				return logs, fmt.Errorf("invalid plugin output: %w", err)
			}
			result.Output, err = json.Marshal(s.jsonValue(value))
		}
	}
	return logs, err
}

func (s *Service) jsonValue(value any) any {
	switch value := value.(type) {
	case string:
		return s.text(value)
	case []any:
		for i := range value {
			value[i] = s.jsonValue(value[i])
		}
		return value
	case map[string]any:
		result := make(map[string]any, len(value))
		for key, item := range value {
			result[s.text(key)] = s.jsonValue(item)
		}
		return result
	default:
		return value
	}
}

type sanitizedError struct {
	cause   error
	message string
}

func (e sanitizedError) Error() string { return e.message }
func (e sanitizedError) Unwrap() error { return e.cause }

// Preserve identifiers verbatim for valid actions; reject any identifier that
// the canonical policy recognizes as secret rather than changing its meaning.
func (s *Service) safeIdentifier(value string) bool {
	redacted, _ := s.redactor.RedactText(value)
	return value == redacted && len(value) <= 2048
}
