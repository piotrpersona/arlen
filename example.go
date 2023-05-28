package main

func newArray() []int {
	return []int{1, 2, 3}
}

func main() {
	// should return error
	arr1 := newArray()
	_ = arr1[0]

	// check if expression
	arr2 := newArray()
	if len(arr2) == 0 {
		return
	}
	_ = arr2[0]

	arr3 := newArray()
	if 0 == len(arr3) {
		return
	}
	_ = arr3[0]
}
