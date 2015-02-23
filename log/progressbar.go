package log

import (
	"io"

	"github.com/dustin/go-humanize"
)

func (l *logger) ReadProgress(r io.Reader, name string, size int64) io.ReadCloser {

	var sizeHuman string

	if size > 0 {
		sizeHuman = humanize.Bytes(uint64(size))
	}
	return &ProgressBar{r, name, size, 0, l, sizeHuman, 0, NewRateLimit(Rate)}
}

func (l *logger) Counter(name string, size int) *Counter {
	return &Counter{name, size, 0, "", l, NewRateLimit(Rate / 2)}
}

type ProgressBar struct {
	io.Reader

	name string
	size int64

	done int64
	l    Log

	sizeHuman string //So we don't calcuate it in every read.
	last      int64

	limit *ratelimit
}

func (p *ProgressBar) print() {

	if p.sizeHuman == "" {
		p.l.Infof("%s [UKN%%] %s of UKN\n", p.name, humanize.Bytes(uint64(p.done)))
		return
	}
	p.l.Infof("%s [%3d%%] %s of %s\n", p.name, p.done*100/p.size, humanize.Bytes(uint64(p.done)), p.sizeHuman)
}
func (p *ProgressBar) Read(b []byte) (int, error) {
	n, err := p.Reader.Read(b)
	p.done += int64(n)

	if (p.done-p.last) > (p.size/50) && !p.limit.Limit() {
		p.last = p.done
		p.print()
	}

	return n, err
}

func (p *ProgressBar) Close() error {
	p.print()
	c, ok := p.Reader.(io.Closer)
	if ok {
		return c.Close()
	}
	return nil
}

type Counter struct {
	name string
	size int

	cur  int
	last string
	l    Log

	limit *ratelimit
}

func (c *Counter) Set(s int, last string) {
	c.cur = s
	c.last = last
	if !c.limit.Limit() || c.cur == c.size {
		c.print()
	}
}

func (c *Counter) print() {
	c.l.Infof("%s [%3d%%] %d of %d %s\n", c.name, c.cur*100/c.size, c.cur, c.size, c.last)
}
