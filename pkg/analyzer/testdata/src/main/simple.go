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

	abc := []int{1, 2, 3, 4}
	for i := range abc {
		_ = abc[i]
	}

	xyz := []int{1, 2, 3}
	for i := 0; i < len(xyz); i++ {
		_ = xyz[i]
	}

	s1 := []int{1, 2, 3}
	_ = s1[0:1] // want `slen: check slice s1 length before accessing`

	s2 := []int{1, 2, 3}
	if len(s2) == 0 {
		fmt.Println("bad")
	}
	_ = s2[0:]

	do2(s2)
}

func do2(a []int) {
	_ = a[0] // want `slen: check slice a length before accessing`
}

func check(a []int) bool {
	return len(a) == 0
}
