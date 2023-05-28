package main

import "fmt"

// respect variable position in file except name.
func do() {
	a := []int{}
	if len(a) == 0 {
		return
	}
	_ = a[0]
}

func main() {
	do()

	a := []int{}
	_ = a[0] // want `slen: check slice a length before accessing`

	a1 := []int{}
	if 0 == len(a1) {
		_ = a1[0]
	}

	a2 := []int{1}
	if len(a2) > 0 {
		_ = a2[0]
	}

	a3 := []int{1}
	if len(a3) < 1 {
		fmt.Println("bad")
	}
	_ = a3[0]

	a4 := []int{1}
	if check(a) {
		fmt.Println("bad")
	}
	_ = a4[0] // want `slen: check slice a4 length before accessing`

	abc := []int{1, 2, 3, 4}
	for i := range abc {
		_ = abc[i]
	}
}

func check(a []int) bool {
	return len(a) == 0
}
