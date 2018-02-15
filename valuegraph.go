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

// A Config tweaks the generation of a Graph.
type Config struct {
	// Generate up to this many child nodes per slice or array, to reduce noise. -1 means no limit.
	RangeLimit int
	// Generate up to this many child nodes per map, to reduce noise. -1 means no limit.
	MapLimit int
	// Truncate strings to this length. -1 means no limit.
	StringLimit int
	// Stop walking inside compound data structures after reaching this many levels. -1 means no limit.
	DepthLimit int
}

// Make constructs a Graph representation of any Go value, for inspection.
func (c *Config) Make(v interface{}) *Graph {
	return c.MakeReflected(reflect.ValueOf(v))
}

// MakeReflected constructs a Graph representation of any reflected Go value, for inspection.
func (c *Config) MakeReflected(v reflect.Value) *Graph {
	g := &Graph{Graph: gographviz.NewGraph(), Nodes: make(map[reflect.Value]string), cfg: c}
	g.SetName("G")
	g.SetDir(true)
	g.addValue("G", "", v, 0, nil, "v")
	return g
}

var DefaultConfig = &Config{
	RangeLimit:  5,
	MapLimit:    -1,
	StringLimit: 30,
	DepthLimit:  -1,
}

// Make constructs a Graph representation of any Go value, for inspection.
// It uses DefaultConfig.
func Make(v interface{}) *Graph {
	return DefaultConfig.Make(v)
}

// MakeReflected constructs a Graph representation of any reflected Go value, for inspection.
// It uses DefaultConfig.
func MakeReflected(v reflect.Value) *Graph {
	return DefaultConfig.MakeReflected(v)

}

// A Graph representation of some value.
type Graph struct {
	*gographviz.Graph
	Nodes map[reflect.Value]string
	cfg   *Config
	i     int
}

func (g *Graph) nextNode() string {
	s := "N" + strconv.Itoa(g.i)
	g.i += 1
	return s
}

func (g *Graph) addValue(parent string, varName string, v reflect.Value, depth int, edgeParams map[string]string, path string) {
	node := g.nextNode()
	g.Nodes[v] = node

	if depth == g.cfg.DepthLimit {
		g.AddNode(parent, node, map[string]string{
			"label": fmt.Sprintf(`"(depth limit %v reached)"`, g.cfg.DepthLimit),
			"shape": "box",
		})

		if parent != "G" {
			g.AddEdge(parent, node, true, nil)
		}
		return
	}

	nodeParams := map[string]string{"shape": "box", "tooltip": `"` + path + `"`}

	label := ""
	if varName != "" {
		label = varName + `\n`
	}

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
			label += `: ` + fmt.Sprint(v.Interface())
		case reflect.Interface:
			label += `\ninterface`
			nodeParams["style"] = "dashed"
			if v.IsNil() {
				label += ": <nil>"
			} else {
				g.addValue(node, "", v.Elem(), depth+1, map[string]string{
					"style":     "dashed",
					"arrowhead": "empty",
				}, path+fmt.Sprintf(".(%v)", v.Elem().Type()))
			}
		case reflect.String:
			label += fmt.Sprintf(" len: %v", v.Len())
			s := v.String()
			if len(s) > g.cfg.StringLimit {
				s = s[:g.cfg.StringLimit]
			}
			s = strings.Replace(s, `\`, `\\`, -1)
			s = strings.Replace(s, `"`, `\"`, -1)
			label += "\n" + s
			if v.Len() > g.cfg.StringLimit {
				label += fmt.Sprintf("\n... %v more", v.Len()-10)
			}
		case reflect.Array:
			label += `\narray`
			l := v.Len()
			label += fmt.Sprintf(" len: %v", l)
			for i := 0; i < l; i++ {
				if i == g.cfg.RangeLimit {
					g.addEllipsis(node, l-i)
				}
				idx := "[" + strconv.Itoa(i) + "]"
				g.addValue(node, idx, v.Index(i), depth+1, nil, path+idx)
			}
		case reflect.Map:
			label += `\nmap`
			if v.IsNil() {
				label += ": <nil>"
			} else {
				keys := v.MapKeys()
				i := 0
				for _, k := range keys {
					if i == g.cfg.MapLimit {
						g.addEllipsis(node, v.Len()-i)
						break
					}
					i += 1
					kn := g.nextNode()
					g.AddNode(node, kn, map[string]string{"label": `""`})
					g.AddEdge(node, kn, true, nil)

					g.addValue(kn, "key", k, depth+1, nil, "<key>")

					kpath := ""
					switch v.Kind() {
					case reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
						kpath = "k"
					case reflect.String:
						kpath = `"` + v.String() + `"`
					default:
						kpath = fmt.Sprint(v.Interface())
					}
					g.addValue(kn, "value", v.MapIndex(k), depth+1, nil, path+"["+kpath+"]")
				}
			}
		case reflect.Ptr:
			if v.IsNil() {
				label += ": <nil>"
			} else {
				ind := reflect.Indirect(v)
				params := map[string]string{"style": "dashed"}
				if n, ok := g.Nodes[ind]; ok {
					g.AddEdge(node, n, true, params)
				} else {
					g.addValue(node, "", ind, depth, params, path)
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
					if i == g.cfg.RangeLimit {
						g.addEllipsis(node, l-i)
						break
					}
					idx := "[" + strconv.Itoa(i) + "]"
					g.addValue(node, idx, v.Index(i), depth+1, nil, path+idx)
				}
			}
		case reflect.Struct:
			label += `\nstruct`
			nf := ty.NumField()
			for i := 0; i < nf; i++ {
				g.addValue(node, ty.Field(i).Name, v.Field(i), depth+1, nil, path+"."+ty.Field(i).Name)
			}
		}
	} else {
		label += `\nInvalid`
	}

	nodeParams["label"] = `"` + label + `"`
	g.AddNode(parent, node, nodeParams)

	if parent != "G" {
		g.AddEdge(parent, node, true, edgeParams)
	}
}

func (g *Graph) addLabeledChild(parent string, label string) {
	g.addChild(parent, map[string]string{
		"label": label,
		"shape": "box",
	})
}

func (g *Graph) addChild(parent string, params map[string]string) {
	kn := g.nextNode()
	g.AddNode(parent, kn, params)
	g.AddEdge(parent, kn, true, nil)
}

func (g *Graph) addEllipsis(parent string, n int) {
	g.addLabeledChild(parent, fmt.Sprintf(`"... %v more"`, n))
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
// Uses DefaultConfig.
func OpenSVG(v interface{}) error {
	return DefaultConfig.OpenSVG(v)
}

// OpenSVG is a convenience method for opening a graph visualization of the value in the system SVG visualizer.
// It is intended for debugging.
func (c *Config) OpenSVG(v interface{}) error {
	s, err := c.Make(v).SVG()
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
