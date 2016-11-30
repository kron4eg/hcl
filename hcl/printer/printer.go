// Package printer implements printing of AST nodes to HCL format.
package printer

import (
	"bytes"
	"io"
	"text/tabwriter"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
)

var DefaultConfig = Config{
	SpacesWidth: 2,
}

// A Config node controls the output of Fprint.
type Config struct {
	SpacesWidth int // if set, it will use spaces instead of tabs for alignment
}

func (c *Config) Fprint(output io.Writer, node ast.Node) error {
	p := &printer{
		cfg:                *c,
		comments:           make([]*ast.CommentGroup, 0),
		standaloneComments: make([]*ast.CommentGroup, 0),
		// enableTrace:        true,
	}

	p.collectComments(node)

	if _, err := output.Write(p.unindent(p.output(node))); err != nil {
		return err
	}

	// flush tabwriter, if any
	var err error
	if tw, _ := output.(*tabwriter.Writer); tw != nil {
		err = tw.Flush()
	}

	return err
}

// Fprint "pretty-prints" an HCL node to output
// It calls Config.Fprint with default settings.
func Fprint(output io.Writer, node ast.Node) error {
	return DefaultConfig.Fprint(output, node)
}

func useCRLF(src []byte) bool {
	// find the first newline.
	idx := bytes.IndexByte(src, '\n')

	// if the first newline was `\r\n`, use that for the rest of the file.
	if idx > 0 && src[idx-1] == '\r' {
		return true
	}
	return false
}

// Format formats src HCL and returns the result.
func Format(src []byte) ([]byte, error) {
	// record if we want the output to keep CRLF line endings
	crlf := useCRLF(src)

	// normalize all line endings
	// since the scanner and output only work with "\n" line endings, we may
	// end up with dangling "\r" characters in the parsed data.
	src = bytes.Replace(src, []byte("\r\n"), []byte("\n"), -1)

	node, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := DefaultConfig.Fprint(&buf, node); err != nil {
		return nil, err
	}

	// Add trailing newline to result
	buf.WriteString("\n")

	out := buf.Bytes()
	if crlf {
		out = bytes.Replace(out, []byte("\n"), []byte("\r\n"), -1)
	}

	return out, nil
}
