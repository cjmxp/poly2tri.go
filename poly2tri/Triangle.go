package poly2tri

type Triangle struct {
	points           []*Point
	neighbors        []*Triangle
	interior         bool
	constrained_edge []bool
	delaunay_edge    []bool
}

func NewTriangle(p1, p2, p3 *Point) *Triangle {
	t := &Triangle{}
	t.Init(p1, p2, p3)
	return t
}

func (this *Triangle) Init(a, b, c *Point) {
	this.points = []*Point{a, b, c}
	this.neighbors = []*Triangle{nil, nil, nil}
	this.interior = false
	this.constrained_edge = []bool{false, false, false}
	this.delaunay_edge = []bool{false, false, false}
}

func (this *Triangle) GetPoints() []*Point {
	return this.points
}

func (this *Triangle) containsPoint(point *Point) bool {
	points := this.points
	return (point == points[0] || point == points[1] || point == points[2])
}

func (this *Triangle) containsEdge(edge *Edge) bool {
	return this.containsPoint(edge.p) && this.containsPoint(edge.q)
}

func (this *Triangle) containsPoints(p1, p2 *Point) bool {
	return this.containsPoint(p1) && this.containsPoint(p2)
}

func (this *Triangle) markNeighborPointers(p1 *Point, p2 *Point, t *Triangle) {
	points := this.points
	// Here we are comparing point references, not values
	if (p1 == points[2] && p2 == points[1]) || (p1 == points[1] && p2 == points[2]) {
		this.neighbors[0] = t
	} else if (p1 == points[0] && p2 == points[2]) || (p1 == points[2] && p2 == points[0]) {
		this.neighbors[1] = t
	} else if (p1 == points[0] && p2 == points[1]) || (p1 == points[1] && p2 == points[0]) {
		this.neighbors[2] = t
	} else {
		panic("poly2tri Invalid Triangle.markNeighborPointers() call")
	}
}
func (this *Triangle) markNeighbor(t *Triangle) {
	points := this.points
	if t.containsPoints(points[1], points[2]) {
		this.neighbors[0] = t
		t.markNeighborPointers(points[1], points[2], this)
	} else if t.containsPoints(points[0], points[2]) {
		this.neighbors[1] = t
		t.markNeighborPointers(points[0], points[2], this)
	} else if t.containsPoints(points[0], points[1]) {
		this.neighbors[2] = t
		t.markNeighborPointers(points[0], points[1], this)
	}
}

func (this *Triangle) clearNeighbors() {
	this.neighbors[0] = nil
	this.neighbors[1] = nil
	this.neighbors[2] = nil
}

func (this *Triangle) clearDelaunayEdges() {
	this.delaunay_edge[0] = false
	this.delaunay_edge[1] = false
	this.delaunay_edge[2] = false
}

func (this *Triangle) pointCW(p *Point) *Point {
	points := this.points
	// Here we are comparing point references, not values
	if p == points[0] {
		return points[2]
	} else if p == points[1] {
		return points[0]
	} else if p == points[2] {
		return points[1]
	} else {
		return nil
	}
}

func (this *Triangle) pointCCW(p *Point) *Point {
	points := this.points
	// Here we are comparing point references, not values
	if p == points[0] {
		return points[1]
	} else if p == points[1] {
		return points[2]
	} else if p == points[2] {
		return points[0]
	} else {
		return nil
	}
}

func (this *Triangle) neighborCW(p *Point) *Triangle {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.neighbors[1]
	} else if p == this.points[1] {
		return this.neighbors[2]
	} else {
		return this.neighbors[0]
	}
}

func (this *Triangle) neighborCCW(p *Point) *Triangle {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.neighbors[2]
	} else if p == this.points[1] {
		return this.neighbors[0]
	} else {
		return this.neighbors[1]
	}
}

func (this *Triangle) getConstrainedEdgeCW(p *Point) bool {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.constrained_edge[1]
	} else if p == this.points[1] {
		return this.constrained_edge[2]
	} else {
		return this.constrained_edge[0]
	}
}

func (this *Triangle) getConstrainedEdgeCCW(p *Point) bool {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.constrained_edge[2]
	} else if p == this.points[1] {
		return this.constrained_edge[0]
	} else {
		return this.constrained_edge[1]
	}
}

func (this *Triangle) getConstrainedEdgeAcross(p *Point) bool {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.constrained_edge[0]
	} else if p == this.points[1] {
		return this.constrained_edge[1]
	} else {
		return this.constrained_edge[2]
	}
}

func (this *Triangle) setConstrainedEdgeCW(p *Point, ce bool) {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		this.constrained_edge[1] = ce
	} else if p == this.points[1] {
		this.constrained_edge[2] = ce
	} else {
		this.constrained_edge[0] = ce
	}
}

func (this *Triangle) setConstrainedEdgeCCW(p *Point, ce bool) {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		this.constrained_edge[2] = ce
	} else if p == this.points[1] {
		this.constrained_edge[0] = ce
	} else {
		this.constrained_edge[1] = ce
	}
}

func (this *Triangle) getDelaunayEdgeCW(p *Point) bool {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.delaunay_edge[1]
	} else if p == this.points[1] {
		return this.delaunay_edge[2]
	} else {
		return this.delaunay_edge[0]
	}
}

func (this *Triangle) getDelaunayEdgeCCW(p *Point) bool {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.delaunay_edge[2]
	} else if p == this.points[1] {
		return this.delaunay_edge[0]
	} else {
		return this.delaunay_edge[1]
	}
}

func (this *Triangle) setDelaunayEdgeCW(p *Point, e bool) {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		this.delaunay_edge[1] = e
	} else if p == this.points[1] {
		this.delaunay_edge[2] = e
	} else {
		this.delaunay_edge[0] = e
	}
}

func (this *Triangle) setDelaunayEdgeCCW(p *Point, e bool) {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		this.delaunay_edge[2] = e
	} else if p == this.points[1] {
		this.delaunay_edge[0] = e
	} else {
		this.delaunay_edge[1] = e
	}
}

func (this *Triangle) neighborAcross(p *Point) *Triangle {
	// Here we are comparing point references, not values
	if p == this.points[0] {
		return this.neighbors[0]
	} else if p == this.points[1] {
		return this.neighbors[1]
	} else {
		return this.neighbors[2]
	}
}

func (this *Triangle) oppositePoint(t *Triangle, p *Point) *Point {
	cw := t.pointCW(p)
	return this.pointCW(cw)
}

func (this *Triangle) legalize(opoint, npoint *Point) {
	points := this.points
	// Here we are comparing point references, not values
	if opoint == points[0] {
		points[1] = points[0]
		points[0] = points[2]
		points[2] = npoint
	} else if opoint == points[1] {
		points[2] = points[1]
		points[1] = points[0]
		points[0] = npoint
	} else if opoint == points[2] {
		points[0] = points[2]
		points[2] = points[1]
		points[1] = npoint
	} else {
		panic("poly2tri Invalid Triangle.legalize() call")
	}
}

func (this *Triangle) index(p *Point) int {
	points := this.points
	// Here we are comparing point references, not values
	if p == points[0] {
		return 0
	} else if p == points[1] {
		return 1
	} else if p == points[2] {
		return 2
	}
	panic("poly2tri Invalid Triangle.index() call")
	return -1
}

func (this *Triangle) edgeIndex(p1, p2 *Point) int {
	points := this.points
	// Here we are comparing point references, not values
	if p1 == points[0] {
		if p2 == points[1] {
			return 2
		} else if p2 == points[2] {
			return 1
		}
	} else if p1 == points[1] {
		if p2 == points[2] {
			return 0
		} else if p2 == points[0] {
			return 2
		}
	} else if p1 == points[2] {
		if p2 == points[0] {
			return 1
		} else if p2 == points[1] {
			return 0
		}
	}
	return -1
}

func (this *Triangle) markConstrainedEdgeByEdge(edge *Edge) {
	this.markConstrainedEdgeByPoints(edge.p, edge.q)
}

func (this *Triangle) markConstrainedEdgeByPoints(p, q *Point) {
	points := this.points
	// Here we are comparing point references, not values
	if (q == points[0] && p == points[1]) || (q == points[1] && p == points[0]) {
		this.constrained_edge[2] = true
	} else if (q == points[0] && p == points[2]) || (q == points[2] && p == points[0]) {
		this.constrained_edge[1] = true
	} else if (q == points[1] && p == points[2]) || (q == points[2] && p == points[1]) {
		this.constrained_edge[0] = true
	}
}

func (this *Triangle) pointInsideTriangle(pp *Point) bool {
	p1 := this.points[0]
	p2 := this.points[1]
	p3 := this.points[2]
	if product(p1, p2, p3) >= 0 {
		return product(p1, p2, pp) >= 0 && product(p2, p3, pp) >= 0 && product(p3, p1, pp) >= 0
	} else {
		return product(p1, p2, pp) <= 0 && product(p2, p3, pp) <= 0 && product(p3, p1, pp) <= 0
	}
}

func product(p1, p2, p3 *Point) float32 {
	return (p1.x-p3.x)*(p2.y-p3.y) - (p1.y-p3.y)*(p2.x-p3.x)
}
