// Copyright 2026 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"bytes"
	"context"
	"os"
	"syscall"
	"testing"

	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestMemFileExtendingFallocate(t *testing.T) {
	// len < cap, with stale bytes beyond len.
	backing := []byte("hello, stale bytes")
	data := backing[:5]

	root := &Inode{}
	mntDir, _ := testMount(t, root, &Options{
		FirstAutomaticIno: 1,
		OnAdd: func(ctx context.Context) {
			n := root.EmbeddedInode()
			ch := n.NewPersistentInode(
				ctx,
				&MemRegularFile{
					Data: data,
					Attr: fuse.Attr{
						Mode: 0644,
					},
				},
				StableAttr{})
			n.AddChild("file", ch, false)
		},
	})

	fn := mntDir + "/file"
	fd, err := syscall.Open(fn, syscall.O_RDWR, 0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer syscall.Close(fd)

	for _, sz := range []int{10, 2 * len(backing)} {
		if err := syscall.Fallocate(fd, 0, 0, int64(sz)); err != nil {
			t.Fatalf("Fallocate(%d): %v", sz, err)
		}
		got, err := os.ReadFile(fn)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		want := append([]byte("hello"), make([]byte, sz-5)...)
		if !bytes.Equal(got, want) {
			t.Errorf("fallocate to %d: got %q, want %q", sz, got, want)
		}
	}
}
