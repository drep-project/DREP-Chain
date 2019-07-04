package main

import (
	"fmt"
)

func merge(data []int) []int {
	sum := len(data)
	if sum <= 1 {
		return data
	}
	left := data[0 : sum/2]
	lSize := len(left)
	if lSize >= 2 {
		left = merge(left)
	}
	right := data[sum/2:]
	rSize := len(right)
	if rSize >= 2 {
		right = merge(right)
	}
	j := 0
	t := 0
	arr := make([]int, sum)
	fmt.Println(left, right, data)
	for i := 0; i < sum; i++ {
		if j < lSize && t < rSize {
			if left[j] <= right[t] {
				arr[i] = left[j]
				j++
			} else {
				arr[i] = right[t]
				t++
			}
		}  else if j >= lSize{
			arr[i] = right[t]
			t++
		}  else if t >= rSize{
			arr[i] = left[j]
			j++
		}
	}
	return arr
}

func main() {
	var aa = []int{1000, 2, 31, 34, 5, 9, 7, 4, 6, 89, 90, 99, 99, 99, 99, 99}

	var bb = merge(aa)
	fmt.Println(bb)
}