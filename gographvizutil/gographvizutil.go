package gographvizutil

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/awalterschulze/gographviz"
)

// A Format is a string accepted by the dot command as format.
type Format string

const (
	SVG        Format = "svg"
	PNG        Format = "png"
	GIF        Format = "gif"
	PDF        Format = "pdf"
	PostScript Format = "ps"
)

var ErrNoDot = errors.New("cannot find dot installed in the system")

// Render turns *github.com/awalterschulze/gographviz.Graph into the desired format.
// It requires the dot command to be available in the system.
func Render(g *gographviz.Graph, fmt Format) (string, error) {
	if _, err := exec.LookPath("dot"); err != nil {
		return "", ErrNoDot
	}

	cmd := exec.Command("dot", "-T"+string(fmt))
	var outBuf, errBuf bytes.Buffer
	cmd.Stdin, cmd.Stdout, cmd.Stderr = strings.NewReader(g.String()), &outBuf, &errBuf
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outBuf.String(), nil
}
