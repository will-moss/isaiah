package tty

import (
	"fmt"
	"io"
)

// Represent a TTY (pseudo-terminal)
type TTY struct {
	Stdin  TTYReader // Standard Input
	Stdout io.Writer // Standard Output
	Stderr io.Writer // Standard Error (may be used as stdin mirror)
	Input  TTYWriter // Writer piped to Stdin to send commands
}

func New(stdout io.Writer) TTY {
	commandReader, commandWriter := io.Pipe()
	return TTY{
		Stdin:  TTYReader{Reader: commandReader},
		Input:  TTYWriter{Writer: commandWriter},
		Stdout: stdout,
	}
}

// Send an "exit" command to the pseudo-terminal, and close Stdin
func (tty *TTY) ClearAndQuit() {
	if tty == nil {
		return
	}

	if tty.Stdin == (TTYReader{}) {
		return
	}

	if tty.Input == (TTYWriter{}) {
		return
	}

	io.WriteString(tty.Input.Writer, "exit\n")
	tty.Stdin.Reader.Close()
	tty.Input.Writer.Close()
}

// Send the given command to Stdin, with specific treatment to later
// distinguish our commands from Stdout results
func (tty *TTY) RunCommand(command string) error {
	bashCommand := fmt.Sprintf("%s #ISAIAH", command)

	_, err := io.WriteString(
		tty.Input,
		bashCommand+"\n",
	)

	return err
}

// Wrapper around io.PipeReader to be able to pass it as an io.Reader
type TTYReader struct {
	Reader *io.PipeReader
}

func (r TTYReader) Read(p []byte) (int, error) {
	return r.Reader.Read(p)
}
func (r TTYReader) Close() error {
	return r.Reader.Close()
}

// Wrapper around io.PipeWriter to be able to pass it as an io.Writer
type TTYWriter struct {
	Writer *io.PipeWriter
}

func (w TTYWriter) Write(p []byte) (int, error) {
	return w.Writer.Write(p)
}
func (w TTYWriter) Close() error {
	return w.Writer.Close()
}
