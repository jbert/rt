package rt

import (
	"image"
	"image/color"
	"math"
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

func (t3 T3) normal() P3 {
	// Triangle edges sharing A
	e1 := t3.B.Sub(t3.A)
	e2 := t3.C.Sub(t3.A)
	return e1.Cross(e2)
}

// Hit represents a ray hit on an item
type Hit struct {
	At     P3
	Normal P3
	Colour color.Color
}

// Intersect returns whether a ray intersects this triangle (and if so, the colour, position and normal)
func (t3 T3) Intersect(r R3) (bool, Hit, int64) {
	//	hit, u, v, p := t3.IntersectUV(r)
	hit, _, _, p := t3.IntersectUV(r)
	if !hit {
		return false, Hit{}, 1
	}
	//	return true, color.NRGBA{R: uint8(u * 255), G: uint8(v * 255), B: 0, A: 255}, p, t3.normal()
	return true, Hit{p, t3.normal(), color.NRGBA{R: 128, G: 0, B: 0, A: 255}}, 1
}

// IntersectUV returns true/false for an intercept
// and the the U-V co-ords in the triangle if true (and position)
func (t3 T3) IntersectUV(r R3) (bool, float64, float64, P3) {
	// Triangle edges sharing A
	e1 := t3.A.Sub(t3.C)
	e2 := t3.B.Sub(t3.C)

	// https://en.wikipedia.org/wiki/M%C3%B6ller%E2%80%93Trumbore_intersection_algorithm
	// Determinant
	P := r.Dir.Cross(e2)
	det := P.Dot(e1)
	if math.Abs(det) < Epsilon {
		// Parallel
		return false, 0, 0, Origin
	}
	invDet := 1.0 / det
	//calculate distance from A to ray origin
	T := r.At.Sub(t3.C)

	//Calculate u parameter and test bound
	u := T.Dot(P) * invDet
	//The intersection lies outside of the triangle
	if u < 0 || u > 1 {
		return false, 0, 0, Origin
	}
	//Prepare to test v parameter
	Q := T.Cross(e1)

	//Calculate V parameter and test bound
	v := r.Dir.Dot(Q) * invDet
	//The intersection lies outside of the triangle
	if v < 0 || u+v > 1 {
		return false, 0, 0, Origin
	}

	t := e2.Dot(Q) * invDet
	if t > Epsilon {
		// Bingo
		// u and v are the AB and AC co-ordinates and sum to 1
		return true, u, v, t3.C.Add(e1.Scale(u)).Add(e2.Scale(v))
	}
	return false, 0, 0, Origin
}

// Kite3 is a 2d kite
type Kite3 struct {
	TA, TB T3
	Image  image.Image
}

// NewKite3 returns a kite with opposite corners A and B, and C+C' (C' reflected in AB)
func NewKite3(A, B, C P3, img image.Image) *Kite3 {
	CA := A.Sub(C)
	CB := B.Sub(C)
	C2 := C.Add(CA).Add(CB)
	return &Kite3{
		TA:    T3{A, B, C},
		TB:    T3{A, B, C2},
		Image: img,
	}
}

// Intersect returns whether a ray intersects this kite
func (k3 Kite3) Intersect(r R3) (bool, Hit, int64) {
	bounds := k3.Image.Bounds()
	w := float64(bounds.Max.X - bounds.Min.X)
	h := float64(bounds.Max.Y - bounds.Min.Y)
	uvToColor := func(u, v float64) color.Color {
		//		if u > 0.5 || v > 0.5 {
		//		return color.NRGBA{R: uint8(u * 255), G: 0, B: uint8(v * 255), A: 255}
		//return color.NRGBA{R: uint8(u * 255), G: 0, B: uint8(v * 255), A: 255}
		//return color.NRGBA{R: 128, G: 0, B: 0, A: 255}
		//		return k3.Image.At(0, 0)
		x := bounds.Min.X + int(w*u)
		y := bounds.Min.Y + int(h*v)
		return k3.Image.At(x, y)
		//		}
		//		return color.Black
	}

	// Should only hit one triangle, so don't need to consider
	// z-orderin/
	hit, u, v, p := k3.TA.IntersectUV(r)
	if hit {
		return hit, Hit{p, k3.TA.normal(), uvToColor(u, v)}, 1
	}
	hit, u, v, p = k3.TB.IntersectUV(r)
	u, v = 1-v, 1-u
	if hit {
		return hit, Hit{p, k3.TB.normal(), uvToColor(u, v)}, 2
	}
	return false, Hit{}, 2
}

// An Item is something visible which can be added to a scene
type Item interface {
	Intersect(r R3) (bool, Hit, int64)
}

type CompositeItem struct {
	children []Item
}

func (ci *CompositeItem) AddItem(i Item) {
	ci.children = append(ci.children, i)
}

func (ci *CompositeItem) Intersect(ray R3) (bool, Hit, int64) {
	var hits []Hit
	if ci.children == nil {
		panic("Null children")
	}
	var totalTests int64
	for i := range ci.children {
		intersects, h, numTests := ci.children[i].Intersect(ray)
		totalTests += numTests
		if intersects {
			hits = append(hits, h)
		}
	}

	if len(hits) == 0 {
		return false, Hit{}, totalTests
	}

	nearestHit := hits[0]
	for _, h := range hits {
		if h.At.Len() < nearestHit.At.Len() {
			nearestHit = h
		}
	}
	return true, nearestHit, totalTests
}

// Scene contains the items, lighting and viewport
type Scene struct {
	CompositeItem
	viewerDist float64
	screenDist float64
	lights     []Light
}

// Light represents a light source
type Light struct {
	At     P3
	Colour color.Color
}

// New returns a new Scene
func New(viewerDist float64, screenDist float64) *Scene {
	return &Scene{
		viewerDist: viewerDist,
		screenDist: screenDist,
	}
}

func (s *Scene) illumination(hit Hit) color.Color {
	// In the range 0->0xffff
	r, g, b, _ := hit.Colour.RGBA()

	incidence := 0.0
	for _, light := range s.lights {
		// TODO: check intersection with other scene items for shadows
		vlight := hit.At.Sub(light.At)
		incidence += math.Abs(vlight.Dot(hit.Normal) / vlight.Len() / hit.Normal.Len())
		// TODO incorporate color of the light
	}

	incidence /= float64(len(s.lights))
	//	fmt.Printf("incidence %f\n", incidence)

	toUint := func(v uint32) uint8 { return uint8(float64(v) * incidence / 0xffff * 256) }
	return color.NRGBA{R: toUint(r), G: toUint(g), B: toUint(b), A: 255}
}

func (s *Scene) Render(x, y float64) (color.Color, int64) {
	ray := R3{
		At:  P3{X: 0, Y: 0, Z: s.viewerDist},
		Dir: P3{X: x / s.screenDist, Y: y / s.screenDist, Z: 1.0},
	}

	intersects, h, numTests := s.Intersect(ray)

	if !intersects {
		return s.ambient(ray), numTests
	}

	return s.illumination(h), numTests
}

func (s *Scene) ambient(r R3) color.Color {
	//	red := uint8(10000 * (r.Dir.X / (r.Dir.Y + 1)))
	return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
}

// AddItem adds an item to the Scene
func (s *Scene) AddItem(i Item) {
	s.children = append(s.children, i)
}

// AddLight adds a light to the scene
func (s *Scene) AddLight(l Light) {
	s.lights = append(s.lights, l)
}

// Torus is an composite item
type Torus struct {
	CompositeItem

	centre    P3
	axis      P3
	radius    float64
	thickness float64
	image     image.Image
}

// NewTorus constructs a Torus
func NewTorus(centre P3, axis P3, radius float64, thickness float64, img image.Image) *Torus {
	t := Torus{centre: centre, axis: axis.Normalise(), radius: radius, thickness: thickness, image: img}
	t.buildItems()
	return &t
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

	pi2 := math.Pi * 2.0
	var points []P3
	for theta := 0.0; theta < pi2; theta += pi2 / float64(steps) {
		r := rx.Scale(math.Cos(theta)).Add(ry.Scale(math.Sin(theta)))
		points = append(points, c.centre.Add(r))
	}
	return points
}

// buildItems sets up the child items to represent the Torus
func (d *Torus) buildItems() {
	numMajorSteps := 32
	numMinorSteps := 8
	circle := Circle{centre: d.centre, axis: d.axis, radius: d.radius}

	pts := circle.Points(numMajorSteps)

	grid := make([][]P3, numMajorSteps)

	for i := 0; i < numMajorSteps; i++ {
		pt := pts[i]
		rv := pt.Sub(d.centre).Normalise()
		interiorAxis := rv.Cross(d.axis)
		minorCircle := Circle{centre: pt, axis: interiorAxis, radius: d.thickness / 2}
		grid[i] = minorCircle.Points(numMinorSteps)
	}

	d.children = gridToKites(grid, d.image)
}

func gridToKites(grid [][]P3, img image.Image) []Item {

	var items []Item

	numRows := len(grid)
	numCols := len(grid[0])
	for i := range grid {
		row := grid[i]
		next := grid[(i+1)%numRows]
		for j := range row {
			items = append(items, NewKite3(row[j], next[(j+1)%numCols], row[(j+1)%numCols], img))
		}
	}

	return items
}

// PPiped is a parallelopipd
type PPiped struct {
	CompositeItem

	Corner     P3
	E1, E2, E3 P3
	Image      image.Image
}

// Items returns a set of items to represent the PPiped
func (ppp *PPiped) Build() {
	var items []Item

	blf := ppp.Corner

	items = append(items, NewKite3(blf, blf.Add(ppp.E1).Add(ppp.E2), blf.Add(ppp.E1), ppp.Image))
	items = append(items, NewKite3(blf, blf.Add(ppp.E2).Add(ppp.E3), blf.Add(ppp.E2), ppp.Image))
	items = append(items, NewKite3(blf, blf.Add(ppp.E3).Add(ppp.E1), blf.Add(ppp.E3), ppp.Image))

	trb := ppp.Corner.Add(ppp.E1).Add(ppp.E2).Add(ppp.E3)

	items = append(items, NewKite3(trb, trb.Sub(ppp.E1).Sub(ppp.E2), trb.Sub(ppp.E1), ppp.Image))
	items = append(items, NewKite3(trb, trb.Sub(ppp.E2).Sub(ppp.E3), trb.Sub(ppp.E2), ppp.Image))
	items = append(items, NewKite3(trb, trb.Sub(ppp.E3).Sub(ppp.E1), trb.Sub(ppp.E3), ppp.Image))

	ppp.children = items
}
