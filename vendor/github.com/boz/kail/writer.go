package kail

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

var (
	prefixColor = color.New(color.FgHiWhite, color.Bold)
)

type Writer interface {
	Print(event Event) error
	Fprint(w io.Writer, event Event) error
}

func NewWriter(out io.Writer) Writer {
	return &writer{out}
}

type writer struct {
	out io.Writer
}

func (w *writer) Print(ev Event) error {
	return w.Fprint(w.out, ev)
}

func (w *writer) Fprint(out io.Writer, ev Event) error {
	prefix := w.prefix(ev)

	if _, err := prefixColor.Fprint(out, prefix); err != nil {
		return err
	}
	if _, err := prefixColor.Fprint(out, ": "); err != nil {
		return err
	}

	log := ev.Log()

	if _, err := out.Write(log); err != nil {
		return err
	}

	if sz := len(log); sz == 0 || log[sz-1] != byte('\n') {
		if _, err := out.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) prefix(ev Event) string {
	return fmt.Sprintf("%v/%v[%v]",
		ev.Source().Namespace(),
		ev.Source().Name(),
		ev.Source().Container())
}
