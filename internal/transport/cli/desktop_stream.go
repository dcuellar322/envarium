package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"
)

const desktopResponseLimit = 1 << 20

// This transport accepts only a fixed read, never arguments, paths or mutations.
// There is no independent poll: the native adapter retains its existing cadence.
func streamDesktopSnapshots(ctx context.Context, input io.Reader, output io.Writer, read func(context.Context) (desktopPresentation, error)) error {
	if closer, ok := input.(io.Closer); ok {
		stop := context.AfterFunc(ctx, func() { _ = closer.Close() })
		defer stop()
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 32), 32)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if scanner.Text() != "snapshot" {
			return errors.New("unsupported native snapshot request")
		}
		readCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		value, err := read(readCtx)
		if err == nil {
			err = readCtx.Err()
		}
		cancel()
		if err != nil {
			return err
		}
		frame, err := json.Marshal(outputEnvelope{SchemaVersion: cliSchemaVersion, Command: "desktop.snapshot", Data: value})
		if err != nil {
			return err
		}
		if len(frame) >= desktopResponseLimit {
			return errors.New("native snapshot exceeded the response limit")
		}
		frame = append(frame, '\n')
		written, err := output.Write(frame)
		if err != nil {
			return err
		}
		if written != len(frame) {
			return io.ErrShortWrite
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return scanner.Err()
}
