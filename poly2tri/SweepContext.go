package poly2tri

const (
	kAlpha float32 = 0.3
)

type SweepContext struct {
	triangles  []*Triangle
	maps       []*Triangle
	points     []*Point
	edge_list  []*Edge
	pmin       *Point
	pmax       *Point
	front      *AdvancingFront
	head       *Point
	tail       *Point
	af_head    *Node
	af_middle  *Node
	af_tail    *Node
	basin      *Basin
	edge_event *EdgeEvent
	sweep      *Sweep
}

func (this *SweepContext) Init(contour []*Point) {
	this.sweep = &Sweep{}
	this.triangles = []*Triangle{}
	this.maps = []*Triangle{}
	this.points = contour
	this.edge_list = []*Edge{}
	this.pmin, this.pmax = nil, nil
	this.front = nil
	this.head = nil
	this.tail = nil
	this.af_head = nil
	this.af_middle = nil
	this.af_tail = nil
	this.basin = &Basin{}
	this.edge_event = &EdgeEvent{}
	this.initEdges(this.points)
}

func (this *SweepContext) AddHole(polyline []*Point) {
	this.initEdges(polyline)
	this.points = append(this.points, polyline...)
}

func (this *SweepContext) AddHoles(holes [][]*Point) {
	length := len(holes)
	for i := 0; i < length; i++ {
		this.AddHole(holes[i])
	}
}

func (this *SweepContext) AddPoint(point *Point) {
	this.points = append(this.points, point)
}

func (this *SweepContext) AddPoints(points []*Point) {
	this.points = append(this.points, points...)
}

func (this *SweepContext) Triangulate() {
	this.sweep.triangulate(this)
}

func (this *SweepContext) GetTriangles() []*Triangle {
	return this.triangles
}

func (this *SweepContext) getBoundingBox() (*Point, *Point) {
	return this.pmin, this.pmax
}

func (this *SweepContext) initTriangulation() {
	xmax := this.points[0].x
	xmin := this.points[0].x
	ymax := this.points[0].y
	ymin := this.points[0].y
	// Calculate bounds
	length := len(this.points)
	for i := 1; i < length; i++ {
		p := this.points[i]
		/* jshint expr:true */
		if p.x > xmax {
			xmax = p.x
		}
		if p.x < xmin {
			xmin = p.x
		}
		if p.y > ymax {
			ymax = p.y
		}
		if p.y < ymin {
			ymin = p.y
		}
	}
	this.pmin = NewPoint(xmin, ymin)
	this.pmax = NewPoint(xmax, ymax)
	dx := kAlpha * (xmax - xmin)
	dy := kAlpha * (ymax - ymin)
	this.head = NewPoint(xmax+dx, ymin-dy)
	this.tail = NewPoint(xmin-dx, ymin-dy)
	// Sort points along y-axis
	shot := &ISort{}
	shot.Data = this.points
	shot.Call = func(a interface{}, b interface{}) bool {
		_a := a.(*Point)
		_b := b.(*Point)
		if _a.y == _b.y {
			return _a.x-_b.x < 0
		} else {
			return _a.y-_b.y < 0
		}
		return false
	}
	shot.Sort()
	shot.Free()
}

func (this *SweepContext) initEdges(polyline []*Point) {
	lengh := len(polyline)
	for i := 0; i < lengh; i++ {
		this.edge_list = append(this.edge_list, NewEdge(polyline[i], polyline[(i+1)%lengh]))
	}
}

func (this *SweepContext) locateNode(point *Point) *Node {
	return this.front.locateNode(point.x)
}

func (this *SweepContext) createAdvancingFront() {
	// Initial triangle
	triangle := NewTriangle(this.points[0], this.tail, this.head)
	this.maps = append(this.maps, triangle)
	head := NewNode(triangle.points[1], triangle)
	middle := NewNode(triangle.points[0], triangle)
	tail := NewNode(triangle.points[2], nil)
	this.front = NewAdvancingFront(head, tail)
	head.next = middle
	middle.next = tail
	middle.prev = head
	tail.prev = middle
}

func (this *SweepContext) mapTriangleToNodes(t *Triangle) {
	for i := 0; i < 3; i++ {
		if t.neighbors[i] == nil {
			if n := this.front.locatePoint(t.pointCW(t.points[i])); n != nil {
				n.triangle = t
			}
		}
	}
}

func (this *SweepContext) removeFromMap(triangle *Triangle) {
	length := len(this.maps)
	for i := 0; i < length; i++ {
		if this.maps[i] == triangle {
			this.maps = append(this.maps[:i], this.maps[i+1:]...)
			break
		}
	}
}

func (this *SweepContext) meshClean(triangle *Triangle) {
	triangles := []*Triangle{triangle}
	for len(triangles) > 0 {
		t := triangles[0]
		triangles = triangles[1:]
		if !t.interior {
			t.interior = true
			this.triangles = append(this.triangles, t)
			for i := 0; i < 3; i++ {
				if !t.constrained_edge[i] {
					triangles = append(triangles, t.neighbors[i])
				}
			}
		}
	}
}
