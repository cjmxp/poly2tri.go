package poly2tri

import "math"

const (
	EPSILON   float32 = 1e-12
	CW        int     = 1
	CCW       int     = -1
	COLLINEAR int     = 0
)

type Sweep struct {
}

func (this *Sweep) triangulate(tcx *SweepContext) {
	tcx.initTriangulation()
	tcx.createAdvancingFront()
	// Sweep points; build mesh
	this.sweepPoints(tcx)
	// Clean up
	this.finalizationPolygon(tcx)
}

func (this *Sweep) sweepPoints(tcx *SweepContext) {
	length := len(tcx.points)
	for i := 1; i < length; i++ {
		point := tcx.points[i]
		node := this.pointEvent(tcx, point)
		if edges := point.p2t_edge_list; edges != nil {
			for j := 0; j < len(edges); j++ {
				this.edgeEventByEdge(tcx, edges[j], node)
			}
		}
	}
}

func (this *Sweep) finalizationPolygon(tcx *SweepContext) {
	// Get an Internal triangle to start with
	t := tcx.front.head.next.triangle
	p := tcx.front.head.next.point
	for !t.getConstrainedEdgeCW(p) {
		t = t.neighborCCW(p)
	}
	// Collect interior triangles constrained by edges
	tcx.meshClean(t)
}

func (this *Sweep) pointEvent(tcx *SweepContext, point *Point) *Node {
	node := tcx.locateNode(point)
	new_node := this.newFrontTriangle(tcx, point, node)

	// Only need to check +epsilon since point never have smaller
	// x value than node due to how we fetch nodes from the front
	if point.x <= node.point.x+(EPSILON) {
		this.fill(tcx, node)
	}
	//tcx.AddNode(new_node);
	this.fillAdvancingFront(tcx, new_node)
	return new_node
}

func (this *Sweep) edgeEventByEdge(tcx *SweepContext, edge *Edge, node *Node) {
	tcx.edge_event.constrained_edge = edge
	tcx.edge_event.right = (edge.p.x > edge.q.x)
	if this.isEdgeSideOfTriangle(node.triangle, edge.p, edge.q) {
		return
	}

	// For now we will do all needed filling
	// TODO: integrate with flip process might give some better performance
	//       but for now this avoid the issue with cases that needs both flips and fills
	this.fillEdgeEvent(tcx, edge, node)
	this.edgeEventByPoints(tcx, edge.p, edge.q, node.triangle, edge.q)
}

func (this *Sweep) edgeEventByPoints(tcx *SweepContext, ep *Point, eq *Point, triangle *Triangle, point *Point) {
	if this.isEdgeSideOfTriangle(triangle, ep, eq) {
		return
	}
	p1 := triangle.pointCCW(point)
	o1 := this.orient2d(eq, p1, ep)
	if o1 == COLLINEAR {
		// TODO integrate here changes from C++ version
		// (C++ repo revision 09880a869095 dated March 8, 2011)
		panic("poly2tri EdgeEvent: Collinear not supported! " + eq.toString() + " " + p1.toString() + " " + ep.toString())
	}
	p2 := triangle.pointCW(point)
	o2 := this.orient2d(eq, p2, ep)
	if o2 == COLLINEAR {
		// TODO integrate here changes from C++ version
		// (C++ repo revision 09880a869095 dated March 8, 2011)
		panic("poly2tri EdgeEvent: Collinear not supported! " + eq.toString() + " " + p2.toString() + " " + ep.toString())
	}
	if o1 == o2 {
		// Need to decide if we are rotating CW or CCW to get to a triangle
		// that will cross edge
		if o1 == CW {
			triangle = triangle.neighborCCW(point)
		} else {
			triangle = triangle.neighborCW(point)
		}
		this.edgeEventByPoints(tcx, ep, eq, triangle, point)
	} else {
		// This triangle crosses constraint so lets flippin start!
		this.flipEdgeEvent(tcx, ep, eq, triangle, point)
	}
}

func (this *Sweep) isEdgeSideOfTriangle(triangle *Triangle, ep *Point, eq *Point) bool {
	index := triangle.edgeIndex(ep, eq)
	if index != -1 {
		triangle.constrained_edge[index] = true
		if t := triangle.neighbors[index]; t != nil {
			t.markConstrainedEdgeByPoints(ep, eq)
		}
		return true
	}
	return false
}

func (this *Sweep) newFrontTriangle(tcx *SweepContext, point *Point, node *Node) *Node {
	triangle := NewTriangle(point, node.point, node.next.point)
	triangle.markNeighbor(node.triangle)
	tcx.maps = append(tcx.maps, triangle)
	new_node := NewNode(point, nil)
	new_node.next = node.next
	new_node.prev = node
	node.next.prev = new_node
	node.next = new_node
	if !this.legalize(tcx, triangle) {
		tcx.mapTriangleToNodes(triangle)
	}
	return new_node
}

func (this *Sweep) fill(tcx *SweepContext, node *Node) {
	triangle := NewTriangle(node.prev.point, node.point, node.next.point)
	triangle.markNeighbor(node.prev.triangle)
	triangle.markNeighbor(node.triangle)
	tcx.maps = append(tcx.maps, triangle)
	// Update the advancing front
	node.prev.next = node.next
	node.next.prev = node.prev
	// If it was legalized the triangle has already been mapped
	if !this.legalize(tcx, triangle) {
		tcx.mapTriangleToNodes(triangle)
	}
}

func (this *Sweep) fillAdvancingFront(tcx *SweepContext, n *Node) {
	node := n.next
	for node.next != nil {
		// TODO integrate here changes from C++ version
		// (C++ repo revision acf81f1f1764 dated April 7, 2012)
		if this.isAngleObtuse(node.point, node.next.point, node.prev.point) {
			break
		}
		this.fill(tcx, node)
		node = node.next
	}
	// Fill left holes
	node = n.prev
	for node.prev != nil {
		// TODO integrate here changes from C++ version
		// (C++ repo revision acf81f1f1764 dated April 7, 2012)
		if this.isAngleObtuse(node.point, node.next.point, node.prev.point) {
			break
		}
		this.fill(tcx, node)
		node = node.prev
	}
	// Fill right basins
	if n.next != nil && n.next.next != nil {
		if this.isBasinAngleRight(n) {
			this.fillBasin(tcx, n)
		}
	}
}

func (this *Sweep) isBasinAngleRight(node *Node) bool {
	ax := node.point.x - node.next.next.point.x
	ay := node.point.y - node.next.next.point.y
	if ay < 0 {
		panic("unordered y")
	}
	return (ax >= 0 || float32(math.Abs(float64(ax))) < ay)
}

func (this *Sweep) legalize(tcx *SweepContext, t *Triangle) bool {
	for i := 0; i < 3; i++ {
		if t.delaunay_edge[i] {
			continue
		}
		if ot := t.neighbors[i]; ot != nil {
			p := t.points[i]
			op := ot.oppositePoint(t, p)
			oi := ot.index(op)
			// If this is a Constrained Edge or a Delaunay Edge(only during recursive legalization)
			// then we should not try to legalize
			if ot.constrained_edge[oi] || ot.delaunay_edge[oi] {
				t.constrained_edge[i] = ot.constrained_edge[oi]
				continue
			}
			if this.inCircle(p, t.pointCCW(p), t.pointCW(p), op) {
				// Lets mark this shared edge as Delaunay
				t.delaunay_edge[i] = true
				ot.delaunay_edge[oi] = true
				// Lets rotate shared edge one vertex CW to legalize it
				this.rotateTrianglePair(t, p, ot, op)
				// We now got one valid Delaunay Edge shared by two triangles
				// This gives us 4 new edges to check for Delaunay
				// Make sure that triangle to node mapping is done only one time for a specific triangle
				if !this.legalize(tcx, t) {
					tcx.mapTriangleToNodes(t)
				}
				if !this.legalize(tcx, ot) {
					tcx.mapTriangleToNodes(ot)
				}
				// Reset the Delaunay edges, since they only are valid Delaunay edges
				// until we add a new triangle or point.
				// XXX: need to think about this. Can these edges be tried after we
				//      return to previous recursive level?
				t.delaunay_edge[i] = false
				ot.delaunay_edge[oi] = false
				// If triangle have been legalized no need to check the other edges since
				// the recursive legalization will handles those so we can end here.
				return true
			}
		}
	}
	return false
}

/**
 * <b>Requirement</b>:<br>
 * 1. a,b and c form a triangle.<br>
 * 2. a and d is know to be on opposite side of bc<br>
 * <pre>
 *                a
 *                +
 *               / \
 *              /   \
 *            b/     \c
 *            +-------+
 *           /    d    \
 *          /           \
 * </pre>
 * <b>Fact</b>: d has to be in area B to have a chance to be inside the circle formed by
 *  a,b and c<br>
 *  d is outside B if orient2d(a,b,d) or orient2d(c,a,d) is CW<br>
 *  This preknowledge gives us a way to optimize the incircle test
 * @param pa - triangle point, opposite d
 * @param pb - triangle point
 * @param pc - triangle point
 * @param pd - point opposite a
 * @return {boolean} true if d is inside circle, false if on circle edge
 */
func (this *Sweep) inCircle(pa, pb, pc, pd *Point) bool {
	adx := pa.x - pd.x
	ady := pa.y - pd.y
	bdx := pb.x - pd.x
	bdy := pb.y - pd.y
	adxbdy := adx * bdy
	bdxady := bdx * ady
	oabd := adxbdy - bdxady
	if oabd <= 0 {
		return false
	}
	cdx := pc.x - pd.x
	cdy := pc.y - pd.y
	cdxady := cdx * ady
	adxcdy := adx * cdy
	ocad := cdxady - adxcdy
	if ocad <= 0 {
		return false
	}
	bdxcdy := bdx * cdy
	cdxbdy := cdx * bdy
	alift := adx*adx + ady*ady
	blift := bdx*bdx + bdy*bdy
	clift := cdx*cdx + cdy*cdy
	det := alift*(bdxcdy-cdxbdy) + blift*ocad + clift*oabd
	return det > 0
}

/**
 * Rotates a triangle pair one vertex CW
 *<pre>
 *       n2                    n2
 *  P +-----+             P +-----+
 *    | t  /|               |\  t |
 *    |   / |               | \   |
 *  n1|  /  |n3           n1|  \  |n3
 *    | /   |    after CW   |   \ |
 *    |/ oT |               | oT \|
 *    +-----+ oP            +-----+
 *       n4                    n4
 * </pre>
 */
func (this *Sweep) rotateTrianglePair(t *Triangle, p *Point, ot *Triangle, op *Point) {
	n1 := t.neighborCCW(p)
	n2 := t.neighborCW(p)
	n3 := ot.neighborCCW(op)
	n4 := ot.neighborCW(op)
	ce1 := t.getConstrainedEdgeCCW(p)
	ce2 := t.getConstrainedEdgeCW(p)
	ce3 := ot.getConstrainedEdgeCCW(op)
	ce4 := ot.getConstrainedEdgeCW(op)
	de1 := t.getDelaunayEdgeCCW(p)
	de2 := t.getDelaunayEdgeCW(p)
	de3 := ot.getDelaunayEdgeCCW(op)
	de4 := ot.getDelaunayEdgeCW(op)
	t.legalize(p, op)
	ot.legalize(op, p)
	// Remap delaunay_edge
	ot.setDelaunayEdgeCCW(p, de1)
	t.setDelaunayEdgeCW(p, de2)
	t.setDelaunayEdgeCCW(op, de3)
	ot.setDelaunayEdgeCW(op, de4)
	// Remap constrained_edge
	ot.setConstrainedEdgeCCW(p, ce1)
	t.setConstrainedEdgeCW(p, ce2)
	t.setConstrainedEdgeCCW(op, ce3)
	ot.setConstrainedEdgeCW(op, ce4)
	// Remap neighbors
	// XXX: might optimize the markNeighbor by keeping track of
	//      what side should be assigned to what neighbor after the
	//      rotation. Now mark neighbor does lots of testing to find
	//      the right side.
	t.clearNeighbors()
	ot.clearNeighbors()
	if n1 != nil {
		ot.markNeighbor(n1)
	}
	if n2 != nil {
		t.markNeighbor(n2)
	}
	if n3 != nil {
		t.markNeighbor(n3)
	}
	if n4 != nil {
		ot.markNeighbor(n4)
	}
	t.markNeighbor(ot)
}

func (this *Sweep) fillBasin(tcx *SweepContext, node *Node) {
	if this.orient2d(node.point, node.next.point, node.next.next.point) == CCW {
		tcx.basin.left_node = node.next.next
	} else {
		tcx.basin.left_node = node.next
	}
	// Find the bottom and right node
	tcx.basin.bottom_node = tcx.basin.left_node
	for tcx.basin.bottom_node.next != nil && tcx.basin.bottom_node.point.y >= tcx.basin.bottom_node.next.point.y {
		tcx.basin.bottom_node = tcx.basin.bottom_node.next
	}
	if tcx.basin.bottom_node == tcx.basin.left_node {
		// No valid basin
		return
	}
	tcx.basin.right_node = tcx.basin.bottom_node
	for tcx.basin.right_node.next != nil && tcx.basin.right_node.point.y < tcx.basin.right_node.next.point.y {
		tcx.basin.right_node = tcx.basin.right_node.next
	}
	if tcx.basin.right_node == tcx.basin.bottom_node {
		// No valid basins
		return
	}
	tcx.basin.width = tcx.basin.right_node.point.x - tcx.basin.left_node.point.x
	tcx.basin.left_highest = tcx.basin.left_node.point.y > tcx.basin.right_node.point.y
	this.fillBasinReq(tcx, tcx.basin.bottom_node)
}
func (this *Sweep) fillBasinReq(tcx *SweepContext, node *Node) {
	// if shallow stop filling
	if this.isShallow(tcx, node) {
		return
	}
	this.fill(tcx, node)
	if node.prev == tcx.basin.left_node && node.next == tcx.basin.right_node {
		return
	} else if node.prev == tcx.basin.left_node {
		o := this.orient2d(node.point, node.next.point, node.next.next.point)
		if o == CW {
			return
		}
		node = node.next
	} else if node.next == tcx.basin.right_node {
		o := this.orient2d(node.point, node.prev.point, node.prev.prev.point)
		if o == CCW {
			return
		}
		node = node.prev
	} else {
		// Continue with the neighbor node with lowest Y value
		if node.prev.point.y < node.next.point.y {
			node = node.prev
		} else {
			node = node.next
		}
	}
	this.fillBasinReq(tcx, node)
}
func (this *Sweep) isShallow(tcx *SweepContext, node *Node) bool {
	height := float32(0)
	if tcx.basin.left_highest {
		height = tcx.basin.left_node.point.y - node.point.y
	} else {
		height = tcx.basin.right_node.point.y - node.point.y
	}

	// if shallow stop filling
	if tcx.basin.width > height {
		return true
	}
	return false
}

func (this *Sweep) fillEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	if tcx.edge_event.right {
		this.fillRightAboveEdgeEvent(tcx, edge, node)
	} else {
		this.fillLeftAboveEdgeEvent(tcx, edge, node)
	}
}

func (this *Sweep) fillRightAboveEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	for node.next.point.x < edge.p.x {
		// Check if next node is below the edge
		if this.orient2d(edge.q, node.next.point, edge.p) == CCW {
			this.fillRightBelowEdgeEvent(tcx, edge, node)
		} else {
			node = node.next
		}
	}
}
func (this *Sweep) fillRightBelowEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	if node.point.x < edge.p.x {
		if this.orient2d(node.point, node.next.point, node.next.next.point) == CCW {
			// Concave
			this.fillRightConcaveEdgeEvent(tcx, edge, node)
		} else {
			// Convex
			this.fillRightConvexEdgeEvent(tcx, edge, node)
			// Retry this one
			this.fillRightBelowEdgeEvent(tcx, edge, node)
		}
	}
}
func (this *Sweep) fillRightConcaveEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	this.fill(tcx, node.next)
	if node.next.point != edge.p {
		// Next above or below edge?
		if this.orient2d(edge.q, node.next.point, edge.p) == CCW {
			// Below
			if this.orient2d(node.point, node.next.point, node.next.next.point) == CCW {
				// Next is concave
				this.fillRightConcaveEdgeEvent(tcx, edge, node)
			} else {
				// Next is convex
				/* jshint noempty:false */
			}
		}
	}
}
func (this *Sweep) fillRightConvexEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	// Next concave or convex?
	if this.orient2d(node.next.point, node.next.next.point, node.next.next.next.point) == CCW {
		// Concave
		this.fillRightConcaveEdgeEvent(tcx, edge, node.next)
	} else {
		// Convex
		// Next above or below edge?
		if this.orient2d(edge.q, node.next.next.point, edge.p) == CCW {
			// Below
			this.fillRightConvexEdgeEvent(tcx, edge, node.next)
		} else {
			// Above
			/* jshint noempty:false */
		}
	}
}
func (this *Sweep) fillLeftAboveEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	for node.prev.point.x > edge.p.x {
		// Check if next node is below the edge
		if this.orient2d(edge.q, node.prev.point, edge.p) == CW {
			this.fillLeftBelowEdgeEvent(tcx, edge, node)
		} else {
			node = node.prev
		}
	}
}
func (this *Sweep) fillLeftBelowEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	if node.point.x > edge.p.x {
		if this.orient2d(node.point, node.prev.point, node.prev.prev.point) == CW {
			// Concave
			this.fillLeftConcaveEdgeEvent(tcx, edge, node)
		} else {
			// Convex
			this.fillLeftConvexEdgeEvent(tcx, edge, node)
			// Retry this one
			this.fillLeftBelowEdgeEvent(tcx, edge, node)
		}
	}
}
func (this *Sweep) fillLeftConvexEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	// Next concave or convex?
	if this.orient2d(node.prev.point, node.prev.prev.point, node.prev.prev.prev.point) == CW {
		// Concave
		this.fillLeftConcaveEdgeEvent(tcx, edge, node.prev)
	} else {
		// Convex
		// Next above or below edge?
		if this.orient2d(edge.q, node.prev.prev.point, edge.p) == CW {
			// Below
			this.fillLeftConvexEdgeEvent(tcx, edge, node.prev)
		} else {
			// Above
			/* jshint noempty:false */
		}
	}
}
func (this *Sweep) fillLeftConcaveEdgeEvent(tcx *SweepContext, edge *Edge, node *Node) {
	this.fill(tcx, node.prev)
	if node.prev.point != edge.p {
		// Next above or below edge?
		if this.orient2d(edge.q, node.prev.point, edge.p) == CW {
			// Below
			if this.orient2d(node.point, node.prev.point, node.prev.prev.point) == CW {
				// Next is concave
				this.fillLeftConcaveEdgeEvent(tcx, edge, node)
			} else {
				// Next is convex
				/* jshint noempty:false */
			}
		}
	}
}
func (this *Sweep) flipEdgeEvent(tcx *SweepContext, ep *Point, eq *Point, t *Triangle, p *Point) {
	ot := t.neighborAcross(p)
	if ot == nil {
		panic("FLIP failed due to missing triangle!")
	}
	op := ot.oppositePoint(t, p)
	// Additional check from Java version (see issue #88)
	if t.getConstrainedEdgeAcross(p) {
		index := t.index(p)
		panic("poly2tri Intersecting Constraints " + p.toString() + " " + op.toString() + " " + t.points[(index+1)%3].toString() + " " + t.points[(index+2)%3].toString())
	}
	if this.inScanArea(p, t.pointCCW(p), t.pointCW(p), op) {
		// Lets rotate shared edge one vertex CW
		this.rotateTrianglePair(t, p, ot, op)
		tcx.mapTriangleToNodes(t)
		tcx.mapTriangleToNodes(ot)

		// XXX: in the original C++ code for the next 2 lines, we are
		// comparing point values (and not pointers). In this JavaScript
		// code, we are comparing point references (pointers). This works
		// because we can't have 2 different points with the same values.
		// But to be really equivalent, we should use "Point.equals" here.
		if p == eq && op == ep {
			if eq == tcx.edge_event.constrained_edge.q && ep == tcx.edge_event.constrained_edge.p {
				t.markConstrainedEdgeByPoints(ep, eq)
				ot.markConstrainedEdgeByPoints(ep, eq)
				this.legalize(tcx, t)
				this.legalize(tcx, ot)
			} else {
				// XXX: I think one of the triangles should be legalized here?
				/* jshint noempty:false */
			}
		} else {
			o := this.orient2d(eq, op, ep)
			t = this.nextFlipTriangle(tcx, o, t, ot, p, op)
			this.flipEdgeEvent(tcx, ep, eq, t, p)
		}
	} else {
		newP := this.nextFlipPoint(ep, eq, ot, op)
		this.flipScanEdgeEvent(tcx, ep, eq, t, ot, newP)
		this.edgeEventByPoints(tcx, ep, eq, t, p)
	}
}

func (this *Sweep) nextFlipTriangle(tcx *SweepContext, o int, t *Triangle, ot *Triangle, p *Point, op *Point) *Triangle {
	if o == CCW {
		// ot is not crossing edge after flip
		edge_index := ot.edgeIndex(p, op)
		ot.delaunay_edge[edge_index] = true
		this.legalize(tcx, ot)
		ot.clearDelaunayEdges()
		return t
	}
	// t is not crossing edge after flip
	edge_index := t.edgeIndex(p, op)

	t.delaunay_edge[edge_index] = true
	this.legalize(tcx, t)
	t.clearDelaunayEdges()
	return ot
}

func (this *Sweep) nextFlipPoint(ep *Point, eq *Point, ot *Triangle, op *Point) *Point {
	o2d := this.orient2d(eq, op, ep)
	if o2d == CW {
		// Right
		return ot.pointCCW(op)
	} else if o2d == CCW {
		// Left
		return ot.pointCW(op)
	} else {
		panic("poly2tri [Unsupported] nextFlipPoint: opposing point on constrained edge! " + eq.toString() + " " + op.toString() + " " + ep.toString())
	}
}

func (this *Sweep) flipScanEdgeEvent(tcx *SweepContext, ep *Point, eq *Point, flip_triangle *Triangle, t *Triangle, p *Point) {
	ot := t.neighborAcross(p)
	if ot == nil {
		panic("FLIP failed due to missing triangle")
	}
	op := ot.oppositePoint(t, p)
	if this.inScanArea(eq, flip_triangle.pointCCW(eq), flip_triangle.pointCW(eq), op) {
		// flip with new edge op.eq
		this.flipEdgeEvent(tcx, eq, op, ot, op)
	} else {
		newP := this.nextFlipPoint(ep, eq, ot, op)
		this.flipScanEdgeEvent(tcx, ep, eq, flip_triangle, ot, newP)
	}
}

func (this *Sweep) orient2d(pa, pb, pc *Point) int {
	detleft := (pa.x - pc.x) * (pb.y - pc.y)
	detright := (pa.y - pc.y) * (pb.x - pc.x)
	val := detleft - detright
	if val > -(EPSILON) && val < (EPSILON) {
		return COLLINEAR
	} else if val > 0 {
		return CCW
	} else {
		return CW
	}
}

func (this *Sweep) inScanArea(pa, pb, pc, pd *Point) bool {
	oadb := (pa.x-pb.x)*(pd.y-pb.y) - (pd.x-pb.x)*(pa.y-pb.y)
	if oadb >= -EPSILON {
		return false
	}
	oadc := (pa.x-pc.x)*(pd.y-pc.y) - (pd.x-pc.x)*(pa.y-pc.y)
	if oadc <= EPSILON {
		return false
	}
	return true
}

func (this *Sweep) isAngleObtuse(pa, pb, pc *Point) bool {
	ax := pb.x - pa.x
	ay := pb.y - pa.y
	bx := pc.x - pa.x
	by := pc.y - pa.y
	return (ax*bx + ay*by) < 0
}
