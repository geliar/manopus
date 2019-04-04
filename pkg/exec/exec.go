package exec

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/geliar/manopus/pkg/report"
)

func Exec(ctx context.Context, reporter report.Driver, name string, arg ...string) (result int, stdoutResult, stderrResult string) {
	l := logger(ctx)

	cmd := exec.CommandContext(ctx, name, arg...)
	s := bytes.NewBufferString("")
	cmd.Stdin = s
	cmd.Env = os.Environ()

	reportReader, reportWriter, err := os.Pipe()
	if err != nil {
		l.Error().Err(err).Msg("Cannot open reporter pipe")
		return 1, "", ""
	}

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	stdoutMultiWriter := io.MultiWriter(stdoutBuf, reportWriter)
	stderrMultiWriter := io.MultiWriter(stderrBuf, reportWriter)

	cmd.Stdout = stdoutMultiWriter
	cmd.Stderr = stderrMultiWriter

	var rep []string
	rep = append(rep, name)
	rep = append(rep, arg...)
	reporter.PushString(ctx, "# "+strings.Join(rep, " "))
	reporter.PushReader(ctx, reportReader)

	defer func() { _ = reportReader.Close() }()
	defer func() { _ = reportWriter.Close() }()

	err = cmd.Start()
	if err != nil {
		l.Error().Err(err).Msg("Cannot start the script")
		return 1, "", ""
	}

	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok {
			l.Error().Err(err).Msg("Error when executing script")
			return 1, "", ""
		} else {
			return exitErr.Sys().(syscall.WaitStatus).ExitStatus(), stdoutBuf.String(), stderrBuf.String()
		}
	}

	return 0, stdoutBuf.String(), stderrBuf.String()
}
