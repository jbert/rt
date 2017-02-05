package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"sort"
)

// P3 is a 3d point
type P3 struct {
	X, Y, Z float64
}

// Add returns a new point, the sum
func (p P3) Add(q P3) P3 {
	return P3{X: p.X + q.X, Y: p.Y + q.Y, Z: p.Z + q.Z}
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
	return true, color.NRGBA{R: 128, G: 0, B: 0, A: 255}
}

// An Item is something visible which can be added to a scene
type Item interface {
	ZFront() float64
	Intersect(r R3) (bool, color.Color)
}

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
	viewer      R3
	viewport    image.Rectangle
	sortedItems []Item
}

// New returns a new Scene
func New(viewer R3, viewport image.Rectangle) *Scene {
	return &Scene{
		viewer:   viewer,
		viewport: viewport,
	}
}

// Render returns the colour at the x,y co-ords of the viewport
func (s *Scene) Render(x, y int) color.Color {
	xf := float64(x) / float64(s.viewport.Max.X)
	yf := float64(y) / float64(s.viewport.Max.Y)
	red := uint8(256 * xf / (yf + 0.01))
	return color.NRGBA{R: red, G: 0, B: 0, A: 255}
}

// Add an item to the Scene
func (s *Scene) Add(i Item) {
	s.sortedItems = append(s.sortedItems, i)
	sort.Sort(itemSlice(s.sortedItems))
}

// MakeScene constructs our test scene
func MakeScene(viewport image.Rectangle) *Scene {
	viewer := R3{
		At:  P3{0, 0, -100},
		Dir: P3{0, 0, 1},
	}
	scene := New(viewer, viewport)

	t := T3{
		A: P3{X: 0, Y: 0, Z: 0},
		B: P3{X: 0, Y: 0, Z: 0},
		C: P3{X: 0, Y: 0, Z: 0},
	}

	scene.Add(t)
	return scene
}

func main() {
	width := 500
	height := 500
	screenSize := image.Rect(0, 0, width, height)
	i := image.NewRGBA(screenSize)
	draw.Draw(i, screenSize, image.White, image.Point{}, draw.Over)

	scene := MakeScene(screenSize)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			colour := scene.Render(x, y)
			i.Set(x, y, colour)
		}
	}

	fname := "tt.png"
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Can't open %s: %s", fname, err)
	}
	defer f.Close()
	png.Encode(f, i)
}
