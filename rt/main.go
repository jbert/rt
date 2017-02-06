package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/jbert/rt"
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "Write the cpu profile to `filename`")
	memprofile := flag.String("memprofile", "", "Write the mem profile to `filename`")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	width := 500
	height := 500
	screenSize := image.Rect(0, 0, width, height)
	i := image.NewRGBA(screenSize)
	draw.Draw(i, screenSize, image.White, image.Point{}, draw.Over)

	scene := MakeScene()

	var totalIntersections int64
	before := time.Now()
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			xf := float64(x)/float64(width)*2 - 1.0
			yf := float64(y)/float64(height)*2 - 1.0
			colour, numIntersections := scene.Render(xf, yf)
			i.Set(x, y, colour)
			totalIntersections += numIntersections
		}
	}
	dur := time.Since(before)
	fmt.Printf("%d Intersections in %s: %f ints/sec\n", totalIntersections, dur, float64(totalIntersections)/dur.Seconds())

	fname := "tt.png"
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Can't open %s: %s", fname, err)
	}
	defer f.Close()
	png.Encode(f, i)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}

}

// MakeScene constructs our test scene
func MakeScene() *rt.Scene {
	scene := rt.New(-100, 5)

	f, err := os.Open("zac.png")
	if err != nil {
		log.Fatalf("Failed to open zac png")
	}
	defer f.Close()
	zac, err := png.Decode(f)
	if err != nil {
		log.Fatalf("Failed to decode zac png")
	}
	defer f.Close()

	/*
		t := T3{
			A: P3{X: 0, Y: 0, Z: 0},
			B: P3{X: 10, Y: 0, Z: 0},
			C: P3{X: 0, Y: 10, Z: 0},
		}

		scene.Add(t)
	*/
	blue := color.NRGBA{R: 0, G: 0, B: 128, A: 255}
	torus := rt.NewTorus(rt.P3{X: 0, Y: 0, Z: 100}, rt.P3{X: 1, Y: 1, Z: 1}, 30, 5, image.NewUniform(blue))
	scene.AddItems(torus)

	red := color.NRGBA{R: 128, G: 0, B: 0, A: 255}
	scene.AddItems(rt.NewTorus(rt.P3{X: 10, Y: 0, Z: 100}, rt.P3{X: 1, Y: -1, Z: 1}, 40, 10, image.NewUniform(red)))

	//	ppImg := image.NewUniform(red),
	ppImg := zac
	ppiped := rt.PPiped{
		Corner: rt.P3{-10, 0, 0},
		E1:     rt.P3{2, 2, 2},
		E2:     rt.P3{1, -10, 0},
		E3:     rt.P3{-10, -3, 2},
		Image:  ppImg,
	}
	scene.AddItems(ppiped)
	scene.AddLight(rt.Light{At: rt.P3{-50, 50, 50}, Colour: color.White})

	return scene
}
