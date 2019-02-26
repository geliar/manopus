package fs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/geliar/manopus/pkg/log"
	"github.com/geliar/manopus/pkg/report"
)

func init() {
	report.Register(log.Logger.WithContext(context.Background()), serviceName, New)
}

type FS struct {
	path     string
	id       string
	step     int
	reports  []*bytes.Buffer
	mu       sync.Mutex
	finished sync.RWMutex
	closed   bool
}

func New(config map[string]interface{}, id string, step int) report.Driver {
	i := new(FS)
	i.path, _ = config["path"].(string)
	if i.path == "" {
		return nil
	}
	i.id = id
	i.step = step
	return i
}

func (d *FS) Type() string {
	return serviceName
}

func (d *FS) PushString(ctx context.Context, report string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}
	d.reports = append(d.reports, bytes.NewBufferString(report))
}

func (d *FS) PushReader(ctx context.Context, report io.Reader) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}
	d.reports = append(d.reports, new(bytes.Buffer))
	i := len(d.reports)
	go func() {
		d.finished.RLock()
		defer d.finished.RUnlock()
		_, _ = io.Copy(d.reports[i], report)
	}()
}

func (d *FS) Close(ctx context.Context) {
	d.closed = true
	//Wait for pushes
	d.mu.Lock()
	d.mu.Unlock()
	//Wait for readers
	d.finished.Lock()
	d.finished.Unlock()
	if len(d.reports) == 0 {
		return
	}
	//Saving report
	l := logger(ctx)
	filename := path.Join(d.path, d.id+".report")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		l.Error().
			Err(err).
			Str("report_filename", filename).
			Msg("Cannot open report file")
		return
	}
	buffs := append([]*bytes.Buffer{bytes.NewBufferString(fmt.Sprintf("====== Step %d ======", d.step))}, d.reports...)
	for i := range buffs {
		buffs[i].WriteString("\n")
		_, err = buffs[i].WriteTo(f)
		if err != nil {
			l.Error().
				Err(err).
				Str("report_filename", filename).
				Int("report_part", i).
				Msg("Cannot write report file")
		}
	}
	err = f.Close()
	if err != nil {
		l.Error().
			Err(err).
			Str("report_filename", filename).
			Msg("Cannot close report file")
	}
}
