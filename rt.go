package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

// P3 is a 3d point
type P3 struct {
	X, Y, Z float64
}

// Add returns a new point, the sum
func (p P3) Add(q P3) P3 {
	return P3{X: p.X + q.X, Y: p.Y + q.Y, Z: p.Z + q.Z}
}

// T3 is a 3d triangle
type T3 struct {
	A, B, C P3
}

// Scene contains the models, lighting and viewport
type Scene struct {
	viewport image.Rectangle
}

// New returns a new Scene
func New(viewport image.Rectangle) *Scene {
	return &Scene{
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

func main() {
	width := 500
	height := 500
	screenSize := image.Rect(0, 0, width, height)
	i := image.NewRGBA(screenSize)
	draw.Draw(i, screenSize, image.White, image.Point{}, draw.Over)

	scene := New(screenSize)

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
