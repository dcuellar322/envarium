package application

import (
	"bytes"
	"testing"
)

func TestByteRingMatchesTailAcrossWraps(t *testing.T) {
	for _, limit := range []int{1, 7, 4096} {
		ring := newByteRing(limit)
		var all []byte
		for index, size := range []int{0, 1, 3, 7, 2, 10000, 3, 4096, 0, 9} {
			chunk := bytes.Repeat([]byte{byte(index)}, size)
			all = append(all, chunk...)
			truncated := ring.Write(chunk)
			got, snapshotTruncated := ring.Snapshot()
			want := all[max(0, len(all)-limit):]
			if !bytes.Equal(got, want) || truncated != (len(all) > limit) || snapshotTruncated != truncated {
				t.Fatalf("limit %d write %d: got %v bytes, truncated %v; want %d, %v", limit, index, len(got), truncated, len(want), len(all) > limit)
			}
			if len(got) > 0 {
				got[0]++
			}
			again, _ := ring.Snapshot()
			if !bytes.Equal(again, want) {
				t.Fatal("snapshot aliases retained storage")
			}
			if cap(ring.data) > limit {
				t.Fatal("ring exceeded capacity")
			}
		}
	}
}
