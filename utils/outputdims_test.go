package utils

import (
	"fmt"
	"testing"
)

func TestFindMaxOutput(t *testing.T) {
	maxoutput, err := FindMaxOutput([]int32{10, 3, 32, 32}, []int32{30, 3, 10, 10}, true)
	if err != nil {
		t.Error(err)
	}
	maxout := []int32{10, 30, 41, 41}
	for i := range maxoutput.Output {
		if maxoutput.Output[i] != maxout[i] {
			t.Error(maxoutput.Output[i], maxout[i], "Values not same")
		}
	}
}
func TestFindMinOutputs(t *testing.T) {

	minouts, err := FindMinOutputs([]int32{10, 3, 32, 32}, []int32{30, 3, 5, 5}, true, 2, 5, 1, 4, -1, -1)
	if err != nil {
		t.Error(err)
	}
	// /if len(minouts)
	t.Error("length of minouts is", len(minouts))
	for i := range minouts {
		fmt.Println(minouts[i])
	}

}
func TestFindAllCombos(t *testing.T) {

	allouts := FindAllCombos([]int32{10, 3, 32, 32}, []int32{30, 3, 5, 5}, true, 2, 5, 1, 4, -1, -1)

	if len(allouts) == 0 {
		t.Error("length of minouts is", len(allouts))
	}

	t.Error("length of minouts is", len(allouts))
	for i := range allouts {
		fmt.Println(allouts[i])
	}

}
