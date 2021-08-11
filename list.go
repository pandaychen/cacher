package cacher

import (
	"container/list"
	"sync"
)

type LruList struct {
	//sync.RWMutex
	sync.Mutex            //for thread-safe
	List       *list.List //list
	LocalQueue chan interface{}
}

// 将元素插入到首部
// 封装 container/list 的方法：func (l *List) PushFront(v interface{}) *Element
// 在 list l 的首部插入值为 v 的元素，并返回该元素
func (l *LruList) Push2Front(v interface{}) *list.Element {
	l.LocalQueue <- v
	return nil
}

// 将元素移动到首部
// 封装 container/list 的方法：func (l *List) MoveToFront(e *Element)
// 将元素 e 移动到 list l 的首部，如果 e 不属于 list l，则 list 不改变
func (l *LruList) Move2Front(e *list.Element) {
	l.LocalQueue <- e
	return
}

func (l *LruList) Close() {
	close(l.LocalQueue)
	l.List.Init()
}
