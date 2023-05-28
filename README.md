# slen

Linter which reports if slice element was accessed without checking it's length.

## Examples

### Accessing index of slice

Bad
```go
a := []int{1,2,3}
_ = a[0] // want `slen: check slice a length before accessing`
```

Good
```go
a := []int{1,2,3}
if len(a) == 0 {
	// handle
}
_ = a[0]
```

### Accessing slice

Bad
```go
a := []int{1,2,3}
_ = a[0:]
```

### Check if different function

> TODO: My be a feature

```go
func check(a []int) bool {
	return len(a) == 0
}

func main() {
	a := []int{1,2,3}
	if check(a) {
		// handle
	}
	_ = a // want `slen: check slice a length before accessing`
}
```

### Positives

The following statements will not report error.

```go
_ = len(a) == 0
_ = 0 == len(a) // Yoda
_ = len(a) > 0
for i := range a {
	_ = a[i]
}
for i := 0; i < len(a); i++ {
	_ = a[i]
}
```

## TODO

* [] Check in different function

