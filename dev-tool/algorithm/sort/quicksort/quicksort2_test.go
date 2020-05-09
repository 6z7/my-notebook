package main

import (
	"container/list"
	"testing"
)

var splice = []int{10, 9, 5, 6, 3, 4, 99}

func TestQuickSortNonRecursive(t *testing.T) {
	s := stack{}
	s.push(0)
	s.push(len(splice) - 1)
	for s.Len() > 0 {
		right := s.pop().(int)
		left := s.pop().(int)

		if right <= left {
			break
		}
		i := partition(left, right)
		s.push(left)
		s.push(i - 1)
		if i+1 < right {
			s.push(i + 1)
			s.push(right)
		}
	}
	t.Log(splice)
}

func partition(left, right int) int {
	benchmark := splice[left]
	i := left
	j := right
	for i != j {
		//找到大于基准元素的值
		for i < j && splice[j] >= benchmark {
			j--
		}
		//找到小于基准元素的值
		for i < j && splice[i] <= benchmark {
			i++
		}
		if i < j {
			splice[i], splice[j] = splice[j], splice[i]
		}
	}
	//与基准元素交换
	splice[left] = splice[i]
	splice[i] = benchmark
	return i
}

type stack struct {
	list.List
}

func (s *stack) push(item interface{}) {
	s.PushBack(item)
}
func (s *stack) pop() interface{} {
	element := s.Back()
	return s.Remove(element)
}
