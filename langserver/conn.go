package langserver

import "io"

// Conn is a connection for LSP
type Conn struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

// NewConn returns a new LSP connection
func NewConn(reader io.ReadCloser, writer io.WriteCloser) *Conn {
	return &Conn{
		reader: reader,
		writer: writer,
	}
}

// Read reads data from input
func (c *Conn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// Write writes data to output
func (c *Conn) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}

// Close closes the connection
func (c *Conn) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.writer.Close()
}
