package main

import (
	"fmt"
	"log"

	"github.com/apoydence/pubsub"
)

type someType struct {
	a string
	b string
	w *w
	x *x
}

type w struct {
	i string
	j string
}

type x struct {
	i string
	j string
}

// Here we demonstrate how powerful a TreeTraverser can be. We define a
// StructTraverser that reads each field. Fields can be left blank upon
// subscription meaning the field is optinal.
func main() {
	ps := pubsub.New()

	ps.Subscribe(Subscription("sub-0"), pubsub.WithPath([]interface{}{"a", "b", "w", "w.i", "w.j"}))
	ps.Subscribe(Subscription("sub-1"), pubsub.WithPath([]interface{}{"a", "b", "x", "x.i", "x.j"}))
	ps.Subscribe(Subscription("sub-2"), pubsub.WithPath([]interface{}{"", "b", "x", "x.i", "x.j"}))
	ps.Subscribe(Subscription("sub-3"), pubsub.WithPath([]interface{}{"", "", "x", "x.i", "x.j"}))
	ps.Subscribe(Subscription("sub-4"), pubsub.WithPath([]interface{}{""}))

	ps.Publish(&someType{a: "a", b: "b", w: &w{i: "w.i", j: "w.j"}, x: &x{i: "x.i", j: "x.j"}}, StructTraverser{}.Traverse)
	ps.Publish(&someType{a: "a", b: "b", x: &x{i: "x.i", j: "x.j"}}, StructTraverser{}.Traverse)
	ps.Publish(&someType{a: "a'", b: "b'", x: &x{i: "x.i", j: "x.j"}}, StructTraverser{}.Traverse)
	ps.Publish(&someType{a: "a", b: "b"}, StructTraverser{}.Traverse)
}

// Subscription writes any results to stderr
func Subscription(s string) func(interface{}) {
	return func(data interface{}) {
		d := data.(*someType)
		var w string
		if d.w != nil {
			w = fmt.Sprintf("w:{i:%s j:%s}", d.w.i, d.w.j)
		}

		var x string
		if d.x != nil {
			x = fmt.Sprintf("x:{i:%s j:%s}", d.x.i, d.x.j)
		}
		log.Printf("%s <- {a:%s b:%s %s %s", s, d.a, d.b, w, x)
	}
}

// StructTraverser traverses type SomeType.
type StructTraverser struct{}

// Traverse implements pubsub.TreeTraverser. It demonstrates how complex/powerful
// Paths can be. In this case, it builds new TreeTraversers for
// each part of the struct. This demonstrates how flexible a TreeTraverser
// can be.
//
// In this case, each field (e.g. a or b) are optional.
func (s StructTraverser) Traverse(data interface{}) pubsub.Paths {
	// a
	return pubsub.PathsWithTraverser([]interface{}{"", data.(*someType).a}, pubsub.TreeTraverser(s.b))
}

func (s StructTraverser) b(data interface{}) pubsub.Paths {
	return pubsub.PathAndTraversers(
		[]pubsub.PathAndTraverser{
			{
				Path:      "",
				Traverser: pubsub.TreeTraverser(s.w),
			},
			{
				Path:      data.(*someType).b,
				Traverser: pubsub.TreeTraverser(s.w),
			},
			{
				Path:      "",
				Traverser: pubsub.TreeTraverser(s.x),
			},
			{
				Path:      data.(*someType).b,
				Traverser: pubsub.TreeTraverser(s.x),
			},
		},
	)
}

func (s StructTraverser) w(data interface{}) pubsub.Paths {
	if data.(*someType).w == nil {
		return pubsub.PathsWithTraverser([]interface{}{""}, pubsub.TreeTraverser(s.done))
	}

	return pubsub.PathsWithTraverser([]interface{}{"w"}, pubsub.TreeTraverser(s.wi))
}

func (s StructTraverser) wi(data interface{}) pubsub.Paths {
	return pubsub.PathsWithTraverser([]interface{}{"", data.(*someType).w.i}, pubsub.TreeTraverser(s.wj))
}

func (s StructTraverser) wj(data interface{}) pubsub.Paths {
	return pubsub.PathsWithTraverser([]interface{}{"", data.(*someType).w.j}, pubsub.TreeTraverser(s.done))
}

func (s StructTraverser) x(data interface{}) pubsub.Paths {
	if data.(*someType).x == nil {
		return pubsub.PathsWithTraverser([]interface{}{""}, pubsub.TreeTraverser(s.done))
	}

	return pubsub.PathsWithTraverser([]interface{}{"x"}, pubsub.TreeTraverser(s.xi))
}

func (s StructTraverser) xi(data interface{}) pubsub.Paths {
	return pubsub.PathsWithTraverser([]interface{}{"", data.(*someType).x.i}, pubsub.TreeTraverser(s.xj))
}

func (s StructTraverser) xj(data interface{}) pubsub.Paths {
	return pubsub.PathsWithTraverser([]interface{}{"", data.(*someType).x.j}, pubsub.TreeTraverser(s.done))
}

func (s StructTraverser) done(data interface{}) pubsub.Paths {
	return pubsub.FlatPaths(nil)
}
