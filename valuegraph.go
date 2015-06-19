// Package valuegraph produces a graph representation of any Go value.
package valuegraph

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/tcard/valuegraph/gographvizutil"
)

// Make constructs a Graph representation of any Go value, for inspection.
func Make(v interface{}) *Graph {
	return MakeReflected(reflect.ValueOf(v))
}

// MakeReflected constructs a Graph representation of any reflected Go value, for inspection.
func MakeReflected(v reflect.Value) *Graph {
	g := &Graph{Graph: gographviz.NewGraph(), Nodes: make(map[reflect.Value]string)}
	g.SetName("G")
	g.SetDir(true)
	g.addValue("G", "", v)
	return g
}

// A Graph representation of some value.
type Graph struct {
	*gographviz.Graph
	Nodes map[reflect.Value]string
	i     int
}

func (g *Graph) nextNode() string {
	s := "N" + strconv.Itoa(g.i)
	g.i += 1
	return s
}

func (g *Graph) addValue(parent string, varName string, v reflect.Value) {
	label := ""
	if varName != "" {
		label = varName + `\n`
	}
	node := g.nextNode()
	g.Nodes[v] = node

	if v.Kind() != reflect.Invalid {
		ty := v.Type()
		label += ty.String()
		switch ty.Kind() {
		case reflect.Bool,
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32,
			reflect.Float64,
			reflect.Complex64,
			reflect.Complex128,
			reflect.UnsafePointer,
			reflect.Chan,
			reflect.Func:
			label += `: ` + fmt.Sprint(v)
		case reflect.Interface:
			label += `\ninterface`
			if v.IsNil() {
				label += ": <nil>"
			} else {
				g.addValue(node, "", v.Elem())
			}
		case reflect.String:
			label += fmt.Sprintf(" len: %v", v.Len())
			s := v.String()
			if len(s) > 10 {
				s = s[:10]
			}
			s = strings.Replace(s, `\`, `\\`, -1)
			s = strings.Replace(s, `"`, `\"`, -1)
			label += "\n" + s
			if v.Len() > 10 {
				label += fmt.Sprintf("\n... %v more", v.Len()-10)
			}
		case reflect.Array:
			label += `\narray`
			l := v.Len()
			label += fmt.Sprintf(" len: %v", l)
			for i := 0; i < l; i++ {
				if i == 5 {
					kn := g.nextNode()
					g.AddNode(node, kn, map[string]string{
						"label": fmt.Sprintf(`"... %v more"`, l-i),
						"shape": "box",
					})
					g.AddEdge(node, kn, true, nil)
					break
				}
				g.addValue(node, "["+strconv.Itoa(i)+"]", v.Index(i))
			}
		case reflect.Map:
			label += `\nmap`
			if v.IsNil() {
				label += ": <nil>"
			} else {
				keys := v.MapKeys()
				for _, k := range keys {
					kn := g.nextNode()
					g.AddNode(node, kn, map[string]string{"label": `""`})
					g.AddEdge(node, kn, true, nil)

					g.addValue(kn, "key", k)
					g.addValue(kn, "value", v.MapIndex(k))
				}
			}
		case reflect.Ptr:
			if v.IsNil() {
				label += ": <nil>"
			} else {
				ind := reflect.Indirect(v)
				if n, ok := g.Nodes[ind]; ok {
					g.AddEdge(node, n, true, nil)
				} else {
					g.addValue(node, "", ind)
				}
			}
		case reflect.Slice:
			label += `\nslice`
			if v.IsNil() {
				label += ": <nil>"
			} else {
				l := v.Len()
				label += fmt.Sprintf(" len: %v cap: %v", l, v.Cap())
				for i := 0; i < l; i++ {
					if i == 5 {
						kn := g.nextNode()
						g.AddNode(node, kn, map[string]string{
							"label": fmt.Sprintf(`"... %v more"`, l-i),
							"shape": "box",
						})
						g.AddEdge(node, kn, true, nil)
						break
					}
					g.addValue(node, "["+strconv.Itoa(i)+"]", v.Index(i))
				}
			}
		case reflect.Struct:
			label += `\nstruct`
			nf := ty.NumField()
			for i := 0; i < nf; i++ {
				g.addValue(node, ty.Field(i).Name, v.Field(i))
			}
		}
	} else {
		label += `\nInvalid`
	}

	g.AddNode(parent, node, map[string]string{
		"label": `"` + label + `"`,
		"shape": "box",
	})

	if parent != "G" {
		g.AddEdge(parent, node, true, nil)
	}
}

func (g *Graph) String() string {
	return fmt.Sprint(g.Nodes)
}

// Dot returns the graph in dot format, for the dot command.
func (g *Graph) Dot() string {
	return g.Graph.String()
}

// Dot returns the graph in SVG format. It requires the dot command to be available in the system.
func (g *Graph) SVG() (string, error) {
	return gographvizutil.Render(g.Graph, gographvizutil.SVG)
}

// Dot returns the graph in PNG format. It requires the dot command to be available in the system.
func (g *Graph) PNG() (string, error) {
	return gographvizutil.Render(g.Graph, gographvizutil.PNG)
}

// Dot returns the graph in GIF format. It requires the dot command to be available in the system.
func (g *Graph) GIF() (string, error) {
	return gographvizutil.Render(g.Graph, gographvizutil.GIF)
}

// Dot returns the graph in PDF format. It requires the dot command to be available in the system.
func (g *Graph) PDF() (string, error) {
	return gographvizutil.Render(g.Graph, gographvizutil.PDF)
}

// Dot returns the graph in PostScript format. It requires the dot command to be available in the system.
func (g *Graph) PostScript() (string, error) {
	return gographvizutil.Render(g.Graph, gographvizutil.PostScript)
}

// OpenSVG is a convenience function for opening a graph visualization of the value in the system SVG visualizer.
// It is intended for debugging.
func OpenSVG(v interface{}) error {
	s, err := Make(v).SVG()
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "valuegraph")
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "valuegraph.svg"))
	if err != nil {
		return err
	}

	f.Write([]byte(s))
	f.Close()

	// From go tool pprof.
	cmds := browsers()
	for _, cmd := range cmds {
		args := strings.Split(cmd, " ")
		if len(args) == 0 {
			continue
		}
		viewer := exec.Command(args[0], append(args[1:], f.Name())...)
		viewer.Stderr = os.Stderr
		if err = viewer.Start(); err == nil {
			return nil
		}
	}

	return errors.New("no command to open SVG found; temp file is at " + f.Name())
}

func browsers() []string {
	// From go tool pprof.
	cmds := []string{"chrome", "google-chrome", "firefox"}
	switch runtime.GOOS {
	case "darwin":
		cmds = append(cmds, "/usr/bin/open")
	case "windows":
		cmds = append(cmds, "cmd /c start")
	default:
		cmds = append(cmds, "xdg-open")
	}
	return cmds
}
