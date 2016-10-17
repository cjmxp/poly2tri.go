package poly2tri

type Node struct {
	point    *Point
	triangle *Triangle
	next     *Node
	prev     *Node
	value    float32
}

func NewNode(p *Point, t *Triangle) *Node {
	n := &Node{}
	n.Init(p, t)
	return n
}
func (this *Node) Init(p *Point, t *Triangle) {
	/** @type {XY} */
	this.point = p

	/** @type {Triangle|null} */
	this.triangle = t

	/** @type {Node|null} */
	this.next = nil
	/** @type {Node|null} */
	this.prev = nil

	/** @type {number} */
	this.value = p.x
}
