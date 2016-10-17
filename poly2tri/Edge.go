package poly2tri

type Edge struct {
	p *Point
	q *Point
}

func NewEdge(p, q *Point) *Edge {
	e := &Edge{}
	e.Init(p, q)
	return e
}

func (this *Edge) Init(p1, p2 *Point) {
	this.p = p1
	this.q = p2

	if p1.y > p2.y {
		this.q = p1
		this.p = p2
	} else if p1.y == p2.y {
		if p1.x > p2.x {
			this.q = p1
			this.p = p2
		} else if p1.x == p2.x {
			panic("poly2tri Invalid Edge constructor: repeated points!")
		}
	}

	if this.q.p2t_edge_list == nil {
		this.q.p2t_edge_list = []*Edge{}
	}
	this.q.p2t_edge_list = append(this.q.p2t_edge_list, this)
}
func (this *Edge) hasPoint(point *Point) bool {
	return (this.p.x == point.x && this.p.y == point.y) || (this.q.x == point.x && this.q.y == point.y)
}
