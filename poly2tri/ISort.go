package poly2tri

import (
	"reflect"
	"sort"
)


type ISort struct {
	Data interface{}
	Call func(interface{}, interface{}) bool
}

func (this *ISort) Len() int {
	return reflect.ValueOf(this.Data).Len()
}
func (this *ISort) Swap(a, b int) {
	db := reflect.ValueOf(this.Data)
	A := db.Index(a)
	B := db.Index(b)
	AV := reflect.ValueOf(A.Interface())
	BV := reflect.ValueOf(B.Interface())
	A.Set(BV)
	B.Set(AV)
}
func (this *ISort) Less(a, b int) bool {
	db := reflect.ValueOf(this.Data)
	A := db.Index(a)
	B := db.Index(b)
	return this.Call(A.Interface(), B.Interface())
}
func (this *ISort) Sort() {
	sort.Sort(this)
}
func (this *ISort) Stable() {
	sort.Stable(this)
}
func (this *ISort) Free() {
	this.Data = nil
	this.Call = nil
}
