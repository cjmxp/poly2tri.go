package poly2tri

type Basin struct {
	left_node    *Node
	bottom_node  *Node
	right_node   *Node
	width        float32
	left_highest bool
}

func (this *Basin) clear() {
	this.left_node = nil
	this.bottom_node = nil
	this.right_node = nil
	this.width = 0.0
	this.left_highest = false
}
