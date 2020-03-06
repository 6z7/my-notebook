package main

import "fmt"

//冒泡排序
//两两比较较大者后移
func main() {
	var splice = []int{3, 1, 4, 6, 0}
	for i := 0; i < len(splice); i++ {
		for j := 1; j < len(splice)-i; j++ {
			t := splice[j-1]
			if splice[j] < t {
				splice[j-1], splice[j] = splice[j], splice[j-1]
			}
		}
	}
	fmt.Println(splice)
}
