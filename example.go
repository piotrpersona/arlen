package main

func a() []int {
	return []int{1, 2, 3}
}

func myLog() {
	a3 := a()
	_ = a3[0]
}
