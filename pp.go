package pp

import "io"

// PP helps to pull from readers and pull to writers.
type PP struct {
	bridges []bridge
}

// Pull from a reader
func (p *PP) Pull(r io.Reader) *PP {
	b := bridge{r: r}
	p.bridges = append(p.bridges, b)
	return p
}

// Push to a writer
func (p *PP) Push(w io.Writer) *PP {
	if len(p.bridges) == 0 {
		panic("must pull data from somewhere")
	}
	b := p.bridges[len(p.bridges)-1]
	b.w = w
	p.bridges[len(p.bridges)-1] = b
	return p
}

type Flusher interface {
	Flush() (n int, e error)
}

// Copy data pulled from the readers, push them to the writers, stop and returns on first error.
// on success, it returns a read error==io.EOF and a write error == nil.
func (p *PP) Copy(buf []byte) (res Copy) {
	for _, b := range p.bridges {
		if b.r != nil {
			n, readErr := b.Pull(buf)
			res.read(n, readErr)
			if readErr != nil {
				break
			}
		}
		if b.w != nil {
			n, writeErr := b.Push(buf)
			res.wrote(n, writeErr)
			if writeErr != nil {
				break
			}
		}
	}
	if res.IsSuccess() {
		for _, b := range p.bridges {
			if b.r != nil {
				if x, ok := b.r.(Flusher); ok {
					res.read(x.Flush())
				}
			}
			if b.w != nil {
				if x, ok := b.w.(Flusher); ok {
					res.wrote(x.Flush())
				}
			}
		}
	}
	return
}

// Copy provides status of the copy.
type Copy struct {
	ReadLen  int
	WriteLen int
	ReadErr  error
	WriteErr error
}

func (c *Copy) read(n int, e error) {
	c.ReadErr = e
	c.ReadLen += n
}

func (c *Copy) wrote(n int, e error) {
	c.WriteErr = e
	c.WriteLen += n
}

// IsSuccess if reader returned io.EOF and the writer returned nil error.
func (c Copy) IsSuccess() bool {
	return (c.ReadErr == io.EOF || c.ReadErr == nil) && c.WriteErr == nil
}

// ReadError returns nil if err==io.EOF, returns error otherwise.
func (c Copy) ReadError() error {
	if c.ReadErr == io.EOF {
		return nil
	}
	return c.ReadErr
}

// WriteError returns the write error.
func (c Copy) WriteError() error {
	return c.WriteErr
}

// Error returns one of read or write error message, excluding reader io.EOF error.
func (c Copy) Error() string {
	if c.ReadError() != nil {
		return c.ReadError().Error()
	}
	return c.WriteError().Error()
}

// Wrote returns length of wrote bytes.
func (c Copy) Wrote() int {
	return c.WriteLen
}

// Read returns length of read bytes.
func (c Copy) Read() int {
	return c.ReadLen
}

type bridge struct {
	r io.Reader
	w io.Writer
}

func (b bridge) Pull(p []byte) (n int, readErr error) {
	return b.r.Read(p)
}

func (b bridge) Push(p []byte) (n int, writeErr error) {
	return b.w.Write(p)
}
