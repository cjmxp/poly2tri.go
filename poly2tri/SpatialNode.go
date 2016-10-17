package poly2tri

import (
	"math"
)

type SpatialNode struct {
	x         float32
	y         float32
	t         *Triangle
	neighbors []*SpatialNode
	g         int
	h         int
	parent    *SpatialNode
	flags     int
}

func (this *SpatialNode) X() int {
	return int(this.x)
}
func (this *SpatialNode) Y() int {
	return int(this.y)
}
func (this *SpatialNode) distanceToSpatialNode(that *SpatialNode) int {
	return this.poly(this.x-that.x, this.y-that.y)
}
func (this *SpatialNode) pointInsideTriangle(pp *Point) bool {
	p1 := this.t.points[0]
	p2 := this.t.points[1]
	p3 := this.t.points[2]
	if this._product(p1, p2, p3) >= 0 {
		return (this._product(p1, p2, pp) >= 0) && (this._product(p2, p3, pp)) >= 0 && (this._product(p3, p1, pp) >= 0)
	} else {
		return (this._product(p1, p2, pp) <= 0) && (this._product(p2, p3, pp)) <= 0 && (this._product(p3, p1, pp) <= 0)
	}
}
func (this *SpatialNode) _product(p1, p2, p3 *Point) float32 {
	return (p1.x-p3.x)*(p2.y-p3.y) - (p1.y-p3.y)*(p2.x-p3.x)
}
func (this *SpatialNode) poly(x, y float32) int {
	return int(math.Sqrt(float64(x*x + y*y)))
}
