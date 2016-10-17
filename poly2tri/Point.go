package poly2tri

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

type Point struct {
	x             float32
	y             float32
	p2t_edge_list []*Edge
	Activation    bool
	Id            int
}

func NewPoint(x, y float32) *Point {
	return &Point{x, y, nil, false, 0}
}
func (this *Point) toString() string {
	return strconv.FormatFloat(float64(this.x), 'f', 4, 32) + "," + strconv.FormatFloat(float64(this.y), 'f', 4, 32)
}

func (this *Point) X() float32 {
	return this.x
}

func (this *Point) Y() float32 {
	return this.y
}

func (this *Point) clone() *Point {
	return NewPoint(this.x, this.y)
}
func (this *Point) set_zero() {
	this.x = 0
	this.y = 0
}
func (this *Point) set(x, y float32) {
	this.x = x
	this.y = y
}
func (this *Point) negate() {
	this.x = -this.x
	this.y = -this.y
}
func (this *Point) add(n *Point) {
	this.x += n.x
	this.y += n.y
}
func (this *Point) sub(n *Point) {
	this.x -= n.x
	this.y -= n.y
}
func (this *Point) mul(s float32) {
	this.x *= s
	this.y *= s
}
func (this *Point) length() float32 {
	return float32(math.Sqrt(float64(this.x*this.x + this.y*this.y)))
}
func (this *Point) normalize() float32 {
	var l = this.length()
	this.x /= l
	this.y /= l
	return l
}
func (this *Point) equals(p *Point) bool {
	return this.x == p.x && this.y == p.y
}
func StringConvertPoint(path string) [][]*Point {
	if db, err := ioutil.ReadFile(path); err == nil {
		str := string(db)
		strlist := strings.Split(str, "\n")
		list := [][]*Point{}
		for i := 0; i < len(strlist); i++ {
			list = append(list, []*Point{})
			ps := []string{}
			if strings.Index(strlist[i], " ") != -1 {
				ps = strings.Split(strlist[i], " ")
			} else if strings.Index(strlist[i], ",") != -1 {
				ps = strings.Split(strlist[i], ",")
			}
			for j := 0; j < len(ps)/2; j++ {
				x, _ := strconv.ParseFloat(ps[j*2], 32)
				y, _ := strconv.ParseFloat(ps[j*2+1], 32)
				list[i] = append(list[i], NewPoint(float32(x), float32(y)))
			}
		}
		return list
	} else {
		panic(err.Error())
	}
	return [][]*Point{}
}
