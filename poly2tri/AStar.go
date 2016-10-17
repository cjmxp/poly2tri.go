package poly2tri

import "math"

const (
	DT_NODE_OPEN   int = 0x01
	DT_NODE_CLOSED int = 0x02
)

type AStar struct {
	spatials       []*SpatialNode
	openedList     []*SpatialNode
	isort          *ISort
	spatialNodeMap map[*Triangle]*SpatialNode
}

func (this *AStar) Init(ts []*Triangle) {
	this.isort = &ISort{}
	this.spatials = []*SpatialNode{}
	this.openedList = []*SpatialNode{}
	this.spatialNodeMap = make(map[*Triangle]*SpatialNode)
	for i := 0; i < len(ts); i++ {
		this.spatials = append(this.spatials, this.getNodeFromTriangle(ts[i]))
	}
}

func (this *AStar) GetTriangleAtPoint(p *Point) *SpatialNode {
	for i := 0; i < len(this.spatials); i++ {
		v := this.spatials[i]
		if v.pointInsideTriangle(p) {
			return v
		}
	}
	return nil
}

func (this *AStar) Find(startNode, endNode *SpatialNode) []*SpatialNode {
	this.reset()
	currentNode := startNode
	this.addToOpenedList(startNode)
	startNode.flags = DT_NODE_OPEN
	if startNode != nil && endNode != nil {
		for (currentNode != endNode) && len(this.openedList) > 0 {
			currentNode = this.getAndRemoveFirstFromOpenedList()
			currentNode.flags &= ^DT_NODE_OPEN
			currentNode.flags |= DT_NODE_CLOSED
			list := this.getNodeNeighbors(currentNode)
			for i := 0; i < len(list); i++ {
				neighborNode := list[i]
				// Ignore invalid paths and the ones on the closed list.
				if neighborNode == nil {
					continue
				}
				if (neighborNode.flags & DT_NODE_CLOSED) != 0 {
					continue
				}
				g := currentNode.g + neighborNode.distanceToSpatialNode(currentNode)
				//neighborNode.flags = neighborNode.flags & ^DT_NODE_CLOSED
				// Not in opened list yet.
				if (neighborNode.flags & DT_NODE_OPEN) == 0 {
					neighborNode.g = g
					neighborNode.h = neighborNode.distanceToSpatialNode(endNode)
					neighborNode.parent = currentNode
					neighborNode.flags |= DT_NODE_OPEN
					this.addToOpenedList(neighborNode)
				} else if g < neighborNode.g { // In opened list but with a worse G than this one.
					neighborNode.g = g
					neighborNode.parent = currentNode
					this.Sort()
				}
			}
		}
	}
	if currentNode != endNode {
		panic("Can't find a path")
	}
	path := []*SpatialNode{}
	for currentNode != startNode {
		path = append(path, currentNode)
		currentNode = currentNode.parent
	}
	path = append(path, currentNode)
	return path
}

func (this *AStar) ToPath(startPoint *Point, endPoint *Point, channel []*SpatialNode) []float32 {
	points := []*Point{}
	points = append(points, startPoint, startPoint)
	if len(channel) >= 2 {
		firstTriangle := channel[len(channel)-1].t
		secondTriangle := channel[len(channel)-2].t
		lastTriangle := channel[0].t
		if !firstTriangle.pointInsideTriangle(startPoint) {
			panic("Assert error")
		}
		if !lastTriangle.pointInsideTriangle(endPoint) {
			panic("Assert error")
		}
		startVertex := this.getNotCommonVertex(firstTriangle, secondTriangle)
		vertexCW0 := startVertex
		vertexCCW0 := startVertex
		for n := len(channel) - 1; n > 1; n-- {
			triangleCurrent := channel[n].t
			triangleNext := channel[n-1].t
			commonEdge := this.getCommonEdge(triangleCurrent, triangleNext)
			vertexCW1 := triangleCurrent.pointCW(vertexCW0)
			vertexCCW1 := triangleCurrent.pointCCW(vertexCCW0)
			if !commonEdge.hasPoint(vertexCW0) {
				vertexCW0 = vertexCW1
			}
			if !commonEdge.hasPoint(vertexCCW0) {
				vertexCCW0 = vertexCCW1
			}
			points = append(points, vertexCW0, vertexCCW0)
		}
	}
	points = append(points, endPoint, endPoint)
	return this.stringPull(points)
}
func (this *AStar) stringPull(_portals []*Point) []float32 {
	pts := []*Point{}
	apexIndex := 0
	leftIndex := 0
	rightIndex := 0
	portalApex := _portals[0]
	portalLeft := _portals[0]
	portalRight := _portals[1]
	// Add start point.
	pts = append(pts, portalApex)
	for i := 1; i < len(_portals)/2; i++ {
		left := _portals[i*2]
		right := _portals[i*2+1]
		// Update right vertex.
		if this.triarea2(portalApex, portalRight, right) <= 0.0 {
			if this.vequal(portalApex, portalRight) || this.triarea2(portalApex, portalLeft, right) > 0.0 {

				portalRight = right
				rightIndex = i
			} else {
				// Right over left, insert left to path and restart scan from portal left point.
				if !portalLeft.Activation {
					pts = append(pts, portalLeft)
					portalLeft.Activation = true
				}
				// Make current left the new apex.
				portalApex = portalLeft
				apexIndex = leftIndex
				// Reset portal
				portalLeft = portalApex
				portalRight = portalApex
				leftIndex = apexIndex
				rightIndex = apexIndex
				// Restart scan
				i = apexIndex
				continue
			}
		}
		// Update left vertex.
		if this.triarea2(portalApex, portalLeft, left) >= 0.0 {
			if this.vequal(portalApex, portalLeft) || this.triarea2(portalApex, portalRight, left) < 0.0 {
				// Tighten the funnel.
				portalLeft = left
				leftIndex = i
			} else {
				// Left over right, insert right to path and restart scan from portal right point.

				if !portalRight.Activation {
					pts = append(pts, portalRight)
					portalRight.Activation = true
				}
				// Make current right the new apex.
				portalApex = portalRight
				apexIndex = rightIndex
				// Reset portal
				portalLeft = portalApex
				portalRight = portalApex
				leftIndex = apexIndex
				rightIndex = apexIndex
				// Restart scan
				i = apexIndex
				continue
			}
		}
	}
	if (len(pts) == 0) || (!this.vequal(pts[len(pts)-1], _portals[len(_portals)-1])) {
		// Append last point to path.
		pts = append(pts, _portals[len(_portals)-1])
	}
	path := []float32{}
	for i := 0; i < len(pts); i++ {
		pts[i].Activation = false
		path = append(path, pts[i].x, pts[i].y)
	}
	return path
}
func (this *AStar) vequal(a, b *Point) bool {
	return this.vdistsqr(a, b) < float32(0.001*0.001)
}
func (this *AStar) vdistsqr(a, b *Point) float32 {
	x := b.x - a.x
	y := b.y - a.y
	return float32(math.Sqrt(float64(x*x + y*y)))
}
func (this *AStar) triarea2(a, b, c *Point) float32 {
	ax := b.x - a.x
	ay := b.y - a.y
	bx := c.x - a.x
	by := c.y - a.y
	return bx*ay - ax*by
}
func (this *AStar) getCommonEdge(t1, t2 *Triangle) *Edge {
	commonIndexes := []*Point{}
	for i := 0; i < len(t1.points); i++ {
		point := t1.points[i]
		if t2.containsPoint(point) {
			commonIndexes = append(commonIndexes, point)
		}
	}
	if len(commonIndexes) != 2 {
		panic("Triangles are not contiguous")
	}
	return &Edge{commonIndexes[0], commonIndexes[1]}
}
func (this *AStar) getNotCommonVertex(t1, t2 *Triangle) *Point {
	return t1.points[this.getNotCommonVertexIndex(t1, t2)]
}
func (this *AStar) getNotCommonVertexIndex(t1, t2 *Triangle) int {
	sum := 0
	index := -1
	if !t2.containsPoint(t1.points[0]) {
		index = 0
		sum++
	}
	if !t2.containsPoint(t1.points[1]) {
		index = 1
		sum++
	}
	if !t2.containsPoint(t1.points[2]) {
		index = 2
		sum++
	}
	if sum != 1 {
		panic("Triangles are not contiguous")
	}
	return index
}
func (this *AStar) getNodeFromTriangle(triangle *Triangle) *SpatialNode {
	if v, ok := this.spatialNodeMap[triangle]; !ok {
		tp := triangle.points
		v = &SpatialNode{}
		this.spatialNodeMap[triangle] = v
		v.x = (tp[0].x + tp[1].x + tp[2].x) / 3
		v.y = (tp[0].y + tp[1].y + tp[2].y) / 3
		v.t = triangle
		v.neighbors = []*SpatialNode{}
		if !triangle.constrained_edge[0] {
			v.neighbors = append(v.neighbors, this.getNodeFromTriangle(triangle.neighbors[0]))
		}
		if !triangle.constrained_edge[1] {
			v.neighbors = append(v.neighbors, this.getNodeFromTriangle(triangle.neighbors[1]))
		}
		if !triangle.constrained_edge[2] {
			v.neighbors = append(v.neighbors, this.getNodeFromTriangle(triangle.neighbors[2]))
		}
		return v
	} else {
		return v
	}
	return nil
}
func (this *AStar) reset() {
	for i := 0; i < len(this.spatials); i++ {
		spatial := this.spatials[i]
		spatial.parent = nil
		spatial.g = 0
		spatial.h = 0
		spatial.flags = 0
	}
	this.openedList = this.openedList[:0]
}
func (this *AStar) addToOpenedList(node *SpatialNode) {
	this.openedList = append(this.openedList, node)
	this.Sort()
}
func (this *AStar) getAndRemoveFirstFromOpenedList() *SpatialNode {
	spatial := this.openedList[0]
	this.openedList = this.openedList[1:]
	return spatial
}
func (this *AStar) getNodeNeighbors(node *SpatialNode) []*SpatialNode {
	return node.neighbors
}
func (this *AStar) Sort() {
	this.isort.Data = this.openedList
	this.isort.Call = func(a, b interface{}) bool {
		_a := a.(*SpatialNode)
		_b := b.(*SpatialNode)
		if _a.g+_a.h < _b.g+_b.h {
			return true
		}
		return false
	}
	this.isort.Sort()
}
