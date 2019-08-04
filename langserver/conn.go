package langserver

import "io"

type conn struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func NewConn(reader io.ReadCloser, writer io.WriteCloser) *conn {
	return &conn{
		reader: reader,
		writer: writer,
	}
}

func (c *conn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *conn) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}

func (c *conn) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.writer.Close()
}
