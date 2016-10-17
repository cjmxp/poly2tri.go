package poly2tri

type AdvancingFront struct {
	head        *Node
	tail        *Node
	search_node *Node
}

func NewAdvancingFront(head, tail *Node) *AdvancingFront {
	af := &AdvancingFront{}
	af.Init(head, tail)
	return af
}
func (this *AdvancingFront) Init(head, tail *Node) {
	/** @type {Node} */
	this.head = head
	/** @type {Node} */
	this.tail = tail
	/** @type {Node} */
	this.search_node = head
}

func (this *AdvancingFront) locateNode(x float32) *Node {
	node := this.search_node
	if x < node.value {
		for node = node.prev; node != nil; node = node.prev {
			if x >= node.value {
				this.search_node = node
				return node
			}
		}
	} else {
		for node = node.next; node != nil; node = node.next {
			if x < node.value {
				this.search_node = node.prev
				return node.prev
			}
		}
	}
	return nil
}
func (this *AdvancingFront) locatePoint(point *Point) *Node {
	px := point.x
	node := this.search_node
	nx := node.point.x
	if px == nx {
		// Here we are comparing point references, not values
		if point != node.point {
			// We might have two nodes with same x value for a short time
			if point == node.prev.point {
				node = node.prev
			} else if point == node.next.point {
				node = node.next
			} else {
				panic("poly2tri Invalid AdvancingFront.locatePoint() call")
			}
		}
	} else if px < nx {
		/* jshint boss:true */
		for node = node.prev; node != nil; node = node.prev {
			if point == node.point {
				break
			}
		}
	} else {
		for node = node.next; node != nil; node = node.next {
			if point == node.point {
				break
			}
		}
	}
	if node != nil {
		this.search_node = node
	}
	return node
}
