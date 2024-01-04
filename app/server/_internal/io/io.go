package _io

// Represent a default io.Writer but using a custom WriteFunction
// provided by the developer. It enables us to create "any" type of
// io.Writer, without having to create new interfaces for every new
// implementation we need.
type CustomWriter struct {
	WriteFunction func(p []byte)
}

func (cw CustomWriter) Write(p []byte) (int, error) {
	cw.WriteFunction(p)
	return len(p), nil
}
