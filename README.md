# valuegraph

Package valuegraph produces a graph representation of any Go value.

See [the docs](http://godoc.org/github.com/tcard/valuegraph) or [some examples](#examples).

## Install

Add `"github.com/tcard/valuegraph"` to your import list and then `go get`.

The [OpenSVG function](http://godoc.org/github.com/tcard/valuegraph#Make) function, and some methods on [Graph](http://godoc.org/github.com/tcard/valuegraph#Make), depend on [Graphviz](http://www.graphviz.org/)'s `dot` command to be installed in your system.

## Examples

This:

```go
valuegraph.OpenSVG(&Point{1, 2})
```

Opens this in a browser:

![captura de pantalla 2015-06-19 a las 19 41 34](https://cloud.githubusercontent.com/assets/727422/8259224/3d2fd884-16bb-11e5-9725-c50aed4251ef.png)

You also have [the valuegraph.Make function](http://godoc.org/github.com/tcard/valuegraph#Make), which returns a struct you can call less side-effectful methods on.

Recursive data types are handled OK:

```go
v := &List{1, &List{2, &List{3, nil}}}
v.Next.Next.Next = v
valuegraph.OpenSVG(v)
```

![captura de pantalla 2015-06-19 a las 19 30 56](https://cloud.githubusercontent.com/assets/727422/8259057/0942ef44-16ba-11e5-8397-88555226c2d3.png)

Try it with complex data structures. For example, the `go/ast` representation of this code:

```go
package main

// Main function.
func main() {
	x := 5
	if x < 3 {
		fmt.Println("x < 3")
	}
}
```

Yields a certainly big, but navigable, representation. Here's part of it:

![captura de pantalla 2015-06-19 a las 19 27 39](https://cloud.githubusercontent.com/assets/727422/8259117/7db54782-16ba-11e5-9609-9d73d49dec74.png)
