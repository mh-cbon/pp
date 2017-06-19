package pp

import (
	"io"
)

// PP helps to pull from readers and pull to writers.
type PP struct {
	steps []steper
}

// Pull from a reader
func (p *PP) Pull(r io.Reader) *PP {
	b := readStep{r}
	p.steps = append(p.steps, b)
	return p
}

// Push to a writer
func (p *PP) Push(w io.Writer) *PP {
	b := writeStep{w}
	p.steps = append(p.steps, b)
	return p
}

type readStep struct {
	r io.Reader
}

func (s readStep) do(p []byte) (n int, err error) {
	return s.r.Read(p)
}
func (s readStep) flush() (p []byte, err error) {
	if x, ok := s.r.(Flusher); ok {
		return x.Flush()
	}
	return nil, nil
}

type writeStep struct {
	w io.Writer
}

func (s writeStep) do(p []byte) (n int, err error) {
	return s.w.Write(p)
}
func (s writeStep) flush() (p []byte, err error) {
	if x, ok := s.w.(Flusher); ok {
		return x.Flush()
	}
	return nil, nil
}

// steper does a step.
type steper interface {
	do(p []byte) (n int, err error)
	flush() (p []byte, err error)
}

// Flusher flushes its buffer, n
type Flusher interface {
	Flush() (p []byte, e error)
}

// Copy data pulled from the readers, push them to the writers, stop and returns on first error.
// on success, it returns a read error==io.EOF and a write error == nil.
func (p PP) Copy(buf []byte) (writeLen int, readWriteErr error) {
	var done bool
	var n int
	for {
		for _, s := range p.steps {
			n, readWriteErr = s.do(buf)
			buf = buf[0:n]

			if _, ok := s.(writeStep); ok {
				writeLen += n
			}

			if readWriteErr != nil {
				done = true
			}
			if done && readWriteErr != io.EOF && readWriteErr != nil {
				break
			}
		}

		if done {
			break
		}
	}
	if readWriteErr == nil || readWriteErr == io.EOF {
		for i, s := range p.steps {
			var flushBuf []byte
			flushBuf, readWriteErr = s.flush()
			if readWriteErr != io.EOF && readWriteErr != nil {
				break
			}
			for _, ss := range p.steps[i:] {
				n, readWriteErr = ss.do(flushBuf)
				if _, ok := ss.(writeStep); ok {
					writeLen += n
				}
				if readWriteErr != io.EOF && readWriteErr != nil {
					return
				}
			}
		}
	}

	return
}
