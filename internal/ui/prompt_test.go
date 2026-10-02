package ui

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/huh"
)

// syncBuffer lets the test read what the program has drawn so far.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestRunForm_EscClearsThePrompt(t *testing.T) {
	const title = "Your skillshare repo"
	in, keys := io.Pipe()
	out := &syncBuffer{}
	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for !strings.Contains(out.String(), title) && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		keys.Write([]byte("\x1b"))
	}()

	err := runForm(huh.NewInput().Title(title), in, out)

	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("err = %v, want ErrCancelled", err)
	}
	rendered := out.String()
	last := strings.LastIndex(rendered, title)
	if last < 0 {
		t.Fatalf("prompt never rendered: %q", rendered)
	}
	if !strings.Contains(rendered[last:], "\x1b[J") {
		t.Errorf("prompt was not erased after cancel; output tail = %q", rendered[last:])
	}
}
