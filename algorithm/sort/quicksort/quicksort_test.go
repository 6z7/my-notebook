package main

import (
	"testing"
)

var arr = []int{9, 5, 6, 3, 4}

func TestSortAsc(t *testing.T) {
	quickSortAsc(0, len(arr)-1)
	t.Log(arr)
}

func TestSortDesc(t *testing.T) {
	quickSortDesc(0, len(arr)-1)
	t.Log(arr)
}

func quickSortAsc(left, right int) {
	if left > right {
		return
	}
	benchmark := arr[left]
	i := left
	j := right
	for i != j {
		//找到大于基准元素的值
		for i < j && arr[j] >= benchmark {
			j--
		}
		//找到小于基准元素的值
		for i < j && arr[i] <= benchmark {
			i++
		}
		if i < j {
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	//与基准元素交换
	arr[left] = arr[i]
	arr[i] = benchmark

	quickSortAsc(left, i-1)
	quickSortAsc(i+1, right)
}

func quickSortDesc(left, right int) {
	if left > right {
		return
	}
	benchmark := arr[left]
	i := left
	j := right
	for i != j {
		for i < j && arr[j] <= benchmark {
			j--
		}
		for i < j && arr[i] >= benchmark {
			i++
		}
		if i < j {
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	//与基准元素交换
	arr[left] = arr[i]
	arr[i] = benchmark

	quickSortDesc(left, i-1)
	quickSortDesc(i+1, right)
}
