package pp

import (
	"errors"
	"io"
	"log"
	"testing"
)

func TestReadErr(t *testing.T) {
	pipe := PP{}

	pipe.
		Pull(pseudoReader{"1", errors.New("pseudoReader 1")}).
		Pull(pseudoTransform{"2", nil}).
		Push(pseudoWriter{"3", nil})

	_, err := pipe.Copy(make([]byte, 1024))
	if err == nil {
		t.Errorf("read error not returned on copy")
	} else {
		want := "pseudoReader 1"
		got := err.Error()
		if want != got {
			t.Errorf("read error incorrect, want=%q got %q", want, got)
		}
	}
}

func TestTransformErr(t *testing.T) {
	pipe := PP{}

	pipe.
		Pull(pseudoReader{"1", io.EOF}).
		Pull(pseudoTransform{"2", errors.New("pseudoTransform 2")}).
		Push(pseudoWriter{"3", nil})

	_, err := pipe.Copy(make([]byte, 1024))
	if err == nil {
		t.Errorf("read error not returned on copy")
	} else {
		want := "pseudoTransform 2"
		got := err.Error()
		if want != got {
			t.Errorf("read error incorrect, want=%q got %q", want, got)
		}
	}
}

func TestWriteErr(t *testing.T) {
	pipe := PP{}

	pipe.
		Pull(pseudoReader{"1", io.EOF}).
		Pull(pseudoTransform{"2", nil}).
		Push(pseudoWriter{"3", errors.New("pseudoWriter 1")})

	_, err := pipe.Copy(make([]byte, 1024))
	if err == nil {
		t.Errorf("write error not returned on copy")
	} else {
		want := "pseudoWriter 1"
		got := err.Error()
		if want != got {
			t.Errorf("write error incorrect, want=%q got %q", want, got)
		}
	}
}

func TestReadFlushErr(t *testing.T) {
	pipe := PP{}

	pipe.
		Pull(pseudoReader{"1", io.EOF}).
		Pull(pseudoReaderFlusher{pseudoReader{"2", nil}, errors.New("flush 2")}).
		Push(pseudoWriter{"3", nil})

	_, err := pipe.Copy(make([]byte, 1024))
	if err == nil {
		t.Errorf("read flush error not returned on copy")
	} else {
		want := "flush 2"
		got := err.Error()
		if want != got {
			t.Errorf("write error incorrect, want=%q got %q", want, got)
		}
	}
}

func TestWriteFlushErr(t *testing.T) {
	pipe := PP{}

	pipe.
		Pull(pseudoReader{"1", io.EOF}).
		Pull(pseudoReader{"2", nil}).
		Push(pseudoWriterFlusher{pseudoWriter{"3", nil}, errors.New("flush 3")})

	_, err := pipe.Copy(make([]byte, 1024))
	if err == nil {
		t.Errorf("write flush error not returned on copy")
	} else {
		want := "flush 3"
		got := err.Error()
		if want != got {
			t.Errorf("write error incorrect, want=%q got %q", want, got)
		}
	}
}

func TestPP(t *testing.T) {
	pipe := PP{}

	pipe.
		Pull(pseudoReader{"1", io.EOF}).
		Pull(pseudoTransform{"2", nil}).
		Push(pseudoWriter{"3", nil})

	pipe.Copy(make([]byte, 10))
}

type pseudoReader struct {
	n   string
	err error
}

func (r pseudoReader) Read(p []byte) (n int, err error) {
	log.Println("pseudoReader", r.n, string(p))
	for i := range p {
		p[i] = 'l'
	}
	log.Println("pseudoReader", r.n, string(p))
	return len(p), r.err
}

type pseudoReaderFlusher struct {
	pseudoReader
	f error
}

func (r pseudoReaderFlusher) Flush() (p []byte, err error) {
	log.Println("pseudoReaderFlusher", r.n)
	return make([]byte, 0), r.f
}

type pseudoTransform struct {
	n   string
	err error
}

func (r pseudoTransform) Read(p []byte) (n int, err error) {
	log.Println("pseudoTransform", r.n, string(p))
	for i := range p {
		p[i] = 'L'
	}
	log.Println("pseudoTransform", r.n, string(p))
	return len(p), r.err
}

type pseudoWriter struct {
	n   string
	err error
}

func (w pseudoWriter) Write(p []byte) (n int, err error) {
	log.Println("pseudoWriter", w.n, string(p))
	return len(p), w.err
}

type pseudoWriterFlusher struct {
	pseudoWriter
	f error
}

func (w pseudoWriterFlusher) Flush() (p []byte, err error) {
	log.Println("pseudoWriterFlusher", w.n)
	return make([]byte, 0), w.f
}
