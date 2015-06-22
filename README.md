# valuegraph

Package valuegraph produces a graph representation of any Go value.

See [the docs](http://godoc.org/github.com/tcard/valuegraph) or [some examples](#examples).

## Install

Add `"github.com/tcard/valuegraph"` to your import list and then `go get`.

The [OpenSVG function](http://godoc.org/github.com/tcard/valuegraph#OpenSVG) function, and some methods on [Graph](http://godoc.org/github.com/tcard/valuegraph#Graph), depend on [Graphviz](http://www.graphviz.org/)'s `dot` command to be installed in your system.

## Examples

This:

```go
valuegraph.OpenSVG(&Point{1, 2})
```

Opens this in a browser:

![captura de pantalla 2015-06-22 a las 14 07 21](https://cloud.githubusercontent.com/assets/727422/8281412/0ba2bfc8-18e8-11e5-9002-9ee0c381dde8.png)

You also have [the valuegraph.Make function](http://godoc.org/github.com/tcard/valuegraph#Make), which returns a struct you can call less side-effectful methods on.

Recursive data types are handled OK:

```go
v := &List{1, &List{2, &List{3, nil}}}
v.Next.Next.Next = v
valuegraph.OpenSVG(v)
```

![captura de pantalla 2015-06-22 a las 14 05 41](https://cloud.githubusercontent.com/assets/727422/8281389/d7cd40c4-18e7-11e5-90fb-ff8f1dbfb750.png)

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

![captura de pantalla 2015-06-22 a las 16 01 05](https://cloud.githubusercontent.com/assets/727422/8283526/13b65606-18f8-11e5-96b5-b9e6abaac309.png)

Note the path to each element as tooltip (the text that appears when hovering over a node).
