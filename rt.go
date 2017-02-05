package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"sort"
)

// P3 is a 3d point
type P3 struct {
	X, Y, Z float64
}

// Origin is the 0, 0, 0 point
var Origin P3

// Epsilon is our small value
var Epsilon = 1e-6

// Add adds two vectors
func (p P3) Add(q P3) P3 {
	return P3{X: p.X + q.X, Y: p.Y + q.Y, Z: p.Z + q.Z}
}

// Sub subtracts two vectors
func (p P3) Sub(q P3) P3 {
	return P3{X: p.X - q.X, Y: p.Y - q.Y, Z: p.Z - q.Z}
}

// Len returns the length of a vector
func (p P3) Len() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
}

// Normalise returns a unit vector in the same direction
func (p P3) Normalise() P3 {
	l := p.Len()
	return P3{X: p.X / l, Y: p.Y / l, Z: p.Z / l}
}

// Scale scales the vector
func (p P3) Scale(s float64) P3 {
	return P3{X: p.X * s, Y: p.Y * s, Z: p.Z * s}
}

// Dot returns the dot product of two vectors
func (p P3) Dot(q P3) float64 {
	return (p.X*q.X + p.Y*q.Y + p.Z*q.Z)
}

// Component returns the component of q in the p direction
func (p P3) Component(q P3) P3 {
	s := p.Dot(q) / p.Len()
	return p.Normalise().Scale(s)
}

// Cross returns the cross product of two vectors
func (p P3) Cross(q P3) P3 {
	return P3{
		X: p.Y*q.Z - p.Z*q.Y,
		Y: p.Z*q.X - p.X*q.Z,
		Z: p.X*q.Y - p.Y*q.X,
	}
}

// R3 is a ray - it has a 3d location and a 3d direction
type R3 struct {
	At  P3
	Dir P3
}

// T3 is a 3d triangle
type T3 struct {
	A, B, C P3
}

// ZFront is the nearest part of the item to Z
func (t3 T3) ZFront() float64 {
	z := t3.A.Z
	if t3.B.Z < z {
		z = t3.B.Z
	}
	if t3.C.Z < z {
		z = t3.C.Z
	}
	return z
}

// Intersect returns whether a ray intersects this triangle
func (t3 T3) Intersect(r R3) (bool, color.Color) {
	// Triangle edges sharing A
	e1 := t3.B.Sub(t3.A)
	e2 := t3.C.Sub(t3.A)

	// https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm
	// Determinant
	P := r.Dir.Cross(e2)
	det := P.Dot(e1)
	if math.Abs(det) < Epsilon {
		// Parallel
		return false, nil
	}
	invDet := 1.0 / det
	//calculate distance from A to ray origin
	T := r.At.Sub(t3.A)

	//Calculate u parameter and test bound
	u := T.Dot(P) * invDet
	//The intersection lies outside of the triangle
	if u < 0 || u > 1 {
		return false, nil
	}
	//Prepare to test v parameter
	Q := T.Cross(e1)

	//Calculate V parameter and test bound
	v := r.Dir.Dot(Q) * invDet
	//The intersection lies outside of the triangle
	if v < 0 || u+v > 1 {
		return false, nil
	}

	t := e2.Dot(Q) * invDet
	if t > Epsilon {
		// Bingo
		// u and v are the AB and AC co-ordinates and sum to 1
		return true, color.NRGBA{R: uint8(u * 255), G: uint8(v * 255), B: 0, A: 255}
	}
	return false, nil
}

// Kite3 is a 2d kite
type Kite3 struct {
	TA, TB T3
}

// NewKite3 returns a kite with opposite corners A and B, and C+C' (C' reflected in AB)
func NewKite3(A, B, C P3) *Kite3 {
	CA := A.Sub(C)
	CB := B.Sub(C)
	C2 := C.Add(CA).Add(CB)
	return &Kite3{
		TA: T3{A, B, C},
		TB: T3{A, B, C2},
	}
}

// ZFront is the nearest part of the item to Z
func (k3 Kite3) ZFront() float64 {
	z := k3.TA.ZFront()
	if k3.TB.ZFront() < z {
		z = k3.TB.ZFront()
	}
	return z
}

// Intersect returns whether a ray intersects this kite
func (k3 Kite3) Intersect(r R3) (bool, color.Color) {
	yes, colour := k3.TA.Intersect(r)
	if yes {
		return yes, colour
	}
	yes, colour = k3.TB.Intersect(r)
	if yes {
		return yes, colour
	}
	return false, nil
}

// An Item is something visible which can be added to a scene
type Item interface {
	ZFront() float64
	Intersect(r R3) (bool, color.Color)
}

// An ItemSource is an object which knows how to decompose itself
// into items, suitable for adding to the scene
type ItemSource interface {
	Items() []Item
}

// Sorting interface crud
type itemSlice []Item

func (isa itemSlice) Len() int {
	is := []Item(isa)
	return len(is)
}
func (isa itemSlice) Less(i, j int) bool {
	is := []Item(isa)
	return is[i].ZFront() < is[j].ZFront()
}
func (isa itemSlice) Swap(i, j int) {
	is := []Item(isa)
	is[i], is[j] = is[j], is[i]
}

// Scene contains the items, lighting and viewport
type Scene struct {
	viewerDist  float64
	screenDist  float64
	sortedItems []Item
}

// New returns a new Scene
func New(viewerDist float64, screenDist float64) *Scene {
	return &Scene{
		viewerDist: viewerDist,
		screenDist: screenDist,
	}
}

var hits int

// Render returns the colour at the x,y co-ords of the viewport (-1, -1 -> 0, 0)
func (s *Scene) Render(x, y float64) color.Color {
	ray := R3{
		At:  P3{X: 0, Y: 0, Z: s.viewerDist},
		Dir: P3{X: x / s.screenDist, Y: y / s.screenDist, Z: 1.0},
	}
	for _, item := range s.sortedItems {
		intersects, colour := item.Intersect(ray)
		if intersects {
			hits++
			return colour
		}
	}
	return s.ambient(ray)
}

func (s *Scene) ambient(r R3) color.Color {
	//	red := uint8(10000 * (r.Dir.X / (r.Dir.Y + 1)))
	return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
}

// Add an item to the Scene
func (s *Scene) Add(i Item) {
	s.sortedItems = append(s.sortedItems, i)
	sort.Sort(itemSlice(s.sortedItems))
}

// AddItems asks an ItemSource for the items it wants to add to the scene
func (s *Scene) AddItems(is ItemSource) {
	for _, i := range is.Items() {
		s.Add(i)
	}
}

// Disc is an ItemSource
type Disc struct {
	centre    P3
	axis      P3
	inner     float64
	thickness float64
}

// NewDisc constructs a Disc
func NewDisc(centre P3, axis P3, inner float64, thickness float64) *Disc {
	return &Disc{centre: centre, axis: axis.Normalise(), inner: inner, thickness: thickness}
}

// Circle is a helper type which can calculate points on its circumference
type Circle struct {
	centre P3
	axis   P3
	radius float64
}

// Points returns the requested number of evenly-spaced points on the circle
func (c *Circle) Points(steps int) []P3 {
	v := P3{1, 0, 0}
	if v.Dot(c.axis) < Epsilon {
		// Too close to parallel, choose another starting vector
		v = P3{0, 1, 0}
	}

	// A radius vector
	rx := c.axis.Cross(v).Normalise().Scale(c.radius)
	// A perpendicular radius vector
	ry := c.axis.Cross(rx).Normalise().Scale(c.radius)
	fmt.Printf("rx %v ry %v\n", rx, ry)

	pi2 := math.Pi * 2.0
	var points []P3
	for theta := 0.0; theta < pi2; theta += pi2 / float64(steps) {
		r := rx.Scale(math.Cos(theta)).Add(ry.Scale(math.Sin(theta)))
		points = append(points, c.centre.Add(r))
	}
	return points
}

// Items returns a set of triangles to represent the disc
func (d *Disc) Items() []Item {
	var items []Item

	numSteps := 10
	innerCircle := Circle{centre: d.centre, axis: d.axis, radius: d.inner}
	innerPoints := innerCircle.Points(numSteps)
	outerCircle := Circle{centre: d.centre, axis: d.axis, radius: d.inner + d.thickness}
	outerPoints := outerCircle.Points(numSteps)

	for i := 0; i < numSteps; i++ {
		//		nextInner := innerPoints[(i+1)%numSteps]
		//		nextOuter := outerPoints[(i+1)%numSteps]
		/*		t := T3{
					A: innerPoints[i],
					B: outerPoints[i],
					C: outerPoints[i].Add(d.axis.Scale(d.thickness)),
				}
		*/
		k := NewKite3(innerPoints[i], outerPoints[i].Add(d.axis.Scale(d.thickness)), outerPoints[i])

		items = append(items, k)

	}

	return items
}

// MakeScene constructs our test scene
func MakeScene() *Scene {
	scene := New(-100, 5)

	/*
		t := T3{
			A: P3{X: 0, Y: 0, Z: 0},
			B: P3{X: 10, Y: 0, Z: 0},
			C: P3{X: 0, Y: 10, Z: 0},
		}

		scene.Add(t)
	*/
	disc := NewDisc(P3{X: 5, Y: 5, Z: 100}, P3{X: 1, Y: 1, Z: 1}, 10, 10)
	scene.AddItems(disc)

	return scene
}

func main() {
	width := 500
	height := 500
	screenSize := image.Rect(0, 0, width, height)
	i := image.NewRGBA(screenSize)
	draw.Draw(i, screenSize, image.White, image.Point{}, draw.Over)

	scene := MakeScene()

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			xf := float64(x)/float64(width)*2 - 1.0
			yf := float64(y)/float64(height)*2 - 1.0
			colour := scene.Render(xf, yf)
			i.Set(x, y, colour)
		}
	}
	fmt.Printf("HITS %d\n", hits)

	fname := "tt.png"
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Can't open %s: %s", fname, err)
	}
	defer f.Close()
	png.Encode(f, i)
}
