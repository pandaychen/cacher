package hashmap

//a hashmap 实现，采用拉链法解决冲突

import (
	"fmt"
	"sync"
	"time"

	utils "github.com/pandaychen/goes-wrapper/hashalgo"
)

const (
	SHASHMAP_ROW_COUNT          = 127
	FIX_FACTOR                  = 1.2
	DEFAULT_SHASHMAP_TOTAL_SIZE = 65536
)

/*
	|<------slot------->|
-	|---|---|---|---|---|
|	|---|---|---|---|---|
row	|---|---|---|---|---|
|	|---|---|---|---|---|
|	|---|---|---|---|---|
-	|---|---|---|---|---|
*/

type Hashmap interface {
	Capacity() int
	Len() int
	Get(key string) (value interface{}, pos int, ok bool)
	Set(key string, value interface{}) bool
	Delete(key string) bool
	//Visit()
}

type SHashmap struct {
	//lock
	sync.RWMutex

	RowCount  int
	SlotCount int
	curCount  int
	ItemsPtr  []*HashmapItem //Save all items
}

func NewSHashmap(total_item_size int) *SHashmap {
	hmap := &SHashmap{}
	if total_item_size <= 0 {
		total_item_size = DEFAULT_SHASHMAP_TOTAL_SIZE
	}
	hmap.RowCount = SHASHMAP_ROW_COUNT
	hmap.SlotCount = int(float64(total_item_size)*FIX_FACTOR)/SHASHMAP_ROW_COUNT + 1 //奇数
	hmap.ItemsPtr = make([]*HashmapItem, hmap.RowCount*hmap.SlotCount)

	return hmap
}

func (m *SHashmap) Capacity() int {
	return len(m.ItemsPtr)
}

func (m *SHashmap) Len() int {
	return m.curCount
}

// 从SHashmap中获取数据
func (m *SHashmap) GetHashmapItem(key string) (*HashmapItem, int) {
	var (
		pos        int //pos返回可读或者可写的位置
		curItem    *HashmapItem
		oldestItem *HashmapItem //扫描时记住最老的item，扫描时删除
		findItem   *HashmapItem
	)
	hash := utils.Hash33(key)
	cpos := (hash % m.SlotCount /*find which slot*/) * m.RowCount

	var freeRecordFlag bool
	//遍历冲突链
	for i := 0; i < m.RowCount; i++ {
		j := cpos + i
		curItem = m.ItemsPtr[j]
		if curItem == nil {
			//当前位置无数据，继续查找
			pos = j
			freeRecordFlag = true
			continue
		}
		if key == curItem.Key {
			//找到key数据
			pos = j
			findItem = curItem
			break
		}
		if freeRecordFlag {
			//无需走下面的逻辑
			continue
		}
		if oldestItem == nil {
			oldestItem = curItem
			pos = j
		} else {
			if oldestItem.LastVisitor > curItem.LastVisitor /*OLDEST指向最老的节点*/ {
				oldestItem = curItem
				pos = j
			}
		}
	}

	//item可能为nil
	//写入时，pos的位置有两种
	//1. 空位置 2.如整个链都满了，则以最老的节点的位置返回，直接覆盖此节点
	return findItem, pos
}

func (m *SHashmap) Del(key string) bool {
	var (
		item *HashmapItem
		pos  int
	)
	m.Lock()
	item, pos = m.GetHashmapItem(key)
	if pos < 0 {
		m.Unlock()
		return false
	}
	if item != nil {
		m.ItemsPtr[pos] = nil
	}
	m.Unlock()
	return true
}

func (m *SHashmap) Get(key string) (interface{}, int, bool) {
	var (
		pos  int
		item *HashmapItem
	)
	m.RLock()
	item, pos = m.GetHashmapItem(key)
	if item != nil {
		m.RUnlock()
		return item.Value, pos, true
	} else {
		m.RUnlock()
		return nil, pos, false
	}
}

func (m *SHashmap) Set(key string, value interface{}) bool {
	var (
		item *HashmapItem
		pos  int
	)
	//check if really properly here？
	m.Lock()
	item, pos = m.GetHashmapItem(key)
	if item != nil {
		//已存在，直接更新
		item.Value = value
		item.LastVisitor = time.Now().Unix()
		m.Unlock()
	} else {
		if pos < 0 {
			m.Unlock()
			return false
		}
		m.ItemsPtr[pos] = &HashmapItem{
			Key:         key,
			Value:       value,
			LastVisitor: time.Now().Unix(),
		}
		m.Unlock()
	}
	return true
}

func main() {
	m := NewSHashmap(60000)
	m.Set("test", "value")
	fmt.Println(m.Get("test"))
}