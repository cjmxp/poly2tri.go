package poly2tri

import (
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

func SaveOBJ(path string, triangles []*Triangle) {
	id := 0
	fstr := ""
	maps := make(map[int]string)
	for i := 0; i < len(triangles); i++ {
		t := triangles[i]
		ps := t.GetPoints()
		fstr += "f"
		for p := 0; p < len(ps); p++ {
			if ps[p].Id == 0 {
				id++
				ps[p].Id = id
			}
			fstr += " " + strconv.Itoa(ps[p].Id)
			if !ps[p].Activation {
				ps[p].Activation = true
				maps[ps[p].Id] = ("v " + strconv.FormatFloat(float64(ps[p].X())/20, 'f', 4, 32) + " 0.0 " + strconv.FormatFloat(-float64(ps[p].Y())/20, 'f', 4, 32) + " \n")
			}
		}
		fstr += "\n"
	}
	vstr := ""
	list := []int{}
	for k, _ := range maps {
		list = append(list, k)
	}
	sort.IntSlice(list).Sort()
	for i := 0; i < len(list); i++ {
		vstr += maps[list[i]]
	}
	ioutil.WriteFile(path, []byte(vstr+fstr), os.ModePerm)
}
