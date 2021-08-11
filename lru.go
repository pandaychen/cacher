package cacher

import (
	"container/list" // 解决 LRU 的问题
	"fmt"
	"time"

	hmp "github.com/pandaychen/cacher/hashmap"
	utils "github.com/pandaychen/goes-wrapper/zaplog"
	"go.uber.org/zap"
)

const (
	DEFAULT_PROMOTE_NUM       = 32
	DEFAULT_CHANNEL_SIZE      = 1024
	DEFAULT_CHECK_EXPIRED_LEN = 128
)

type LruCacher struct {
	DlistZone         *LruList
	HashmapZone       *hmp.SHashmap
	Logger            *zap.Logger
	Cap               int
	Promote           uint32
	schedulerDuration time.Duration
}

func NewLruCacher(cap int) *LruCacher {
	list := &LruList{
		LocalQueue: make(chan interface{}, DEFAULT_CHANNEL_SIZE),
		List:       list.New(),
	}
	hashmap := hmp.NewSHashmap(cap)
	logger, _ := utils.ZapLoggerInit("lrucache")
	return &LruCacher{
		DlistZone:   list,
		HashmapZone: hashmap,
		Cap:         cap,
		Logger:      logger,
		Promote:     DEFAULT_PROMOTE_NUM,
	}
}

func (lc *LruCacher) Get(skey string) (bool, *UserData) {
	var (
		data *UserData
		item interface{}
		find bool
	)
	item, _, find = lc.HashmapZone.Get(skey)
	if find {
		val, ok := item.(*list.Element)
		if ok {
			//return true, val.Value
		} else {
			lc.Logger.Error("LruCacher Get error: Get hashmap Value type error[NOT *list.Element]")
			return false, nil
		}
		data, ok = val.Value.(*UserData)
		if !ok {
			lc.Logger.Error("LruCacher Get error: Get hashmap Value type error[NOT *UserData]")
			return false, nil
		}
		data.Frequency++
		if data.Frequency%lc.Promote == 0 {
			lc.DlistZone.Move2Front(val)
		}
		return true, data
	} else {
		return false, nil
	}
}

//writer data
func (lc *LruCacher) Set(data *UserData) error {
	skey := data.UKey
	item, _, ok := lc.HashmapZone.Get(skey)
	if ok {
		v, ok := item.(*list.Element)
		// 更新
		if ok {
			v.Value = data //UserData 类型存储在 hashmap 的单元节点的 Value 中，直接更新
		} else {
			lc.Logger.Error("LruCacher Set error: Get hashmap Value type error[NOT *list.Element]")
			return nil
		}
	} else {
		//add new element，push to local channel
		lc.DlistZone.Push2Front(data)
	}
	return nil
}

func (lc *LruCacher) Del(skey string) {
	var (
		//data *UserData
		item interface{}
		find bool
	)
	item, _, find = lc.HashmapZone.Get(skey)
	if find {
		//hashmap中保存的是list.Element
		listdata, ok := item.(*list.Element)
		if !ok {
			lc.Logger.Error("LruCacher Del error: Get hashmap Value type error[NOT *list.Element]")
			return
		}
		userdata, ok := listdata.Value.(*UserData)
		if !ok {
			lc.Logger.Error("LruCacher Del error: Get hashmap Value type error[NOT *UserData]")
			return
		}
		userdata.Expired = true
	}
	return
}

// 用来处理无锁的功能实现
func (lc *LruCacher) LruCacheScheduler() {
	clean_ticker := time.Tick(lc.schedulerDuration)

	for {
		select {
		case item, ok := <-lc.DlistZone.LocalQueue:
			if !ok {
				//panic?
				lc.Logger.Error("LruCacheScheduler get LocalQueue chan data error")
				return
			}
			switch v := item.(type) {
			case *list.Element:
				//LRU promotion
				lc.DlistZone.List.MoveToFront(v) //func (l *List) MoveToFront(e *Element)
				break
			case *UserData:
				//insert LRU list()
				element := lc.DlistZone.List.PushFront(v) //func (l *List) PushFront(v interface{}) *Element
				//insert hashmap(*list.Element)
				lc.HashmapZone.Set(v.UKey, element) //element is  *list.Element type
				break
			default:

			}
		case <-clean_ticker:
			for i := 0; i < DEFAULT_CHECK_EXPIRED_LEN && lc.DlistZone.List.Len() > 0; i++ {
				//clean elements from list back
				del_item := lc.DlistZone.List.Back() //*list.Element
				if del_item == nil {
					break
				}
				userdata, ok := del_item.Value.(*UserData)
				if !ok {
					lc.Logger.Error("LruCacher Del error: Get hashmap Value type error[NOT *UserData]")
					continue
				}
				//todo: 设置expire时间
				if userdata.Expired {
					//remove list node
					lc.DlistZone.List.Remove(del_item)
					//remove hashmap node(set nil)
					lc.HashmapZone.Del(userdata.UKey)
				} else {
					break
				}
			}

			//当容量即将超限时清理部分数据
			if lc.DlistZone.List.Len() > lc.HashmapZone.Total /*todo:90%*/ {
				lc.Logger.Info("LruCacher Size Overflow,start gc ...")
				for lc.DlistZone.List.Len() > lc.HashmapZone.Total {
					//clean node from back
					list_item := lc.DlistZone.List.Back()
					lc.DlistZone.List.Remove(list_item)
					val, ok := list_item.Value.(*UserData)
					if !ok {
						continue
					}
					lc.HashmapZone.Del(val.UKey)
				}
				lc.Logger.Info("LruCacher Size Overflow,end gc ...")
			}

		}
	}
}

func main() {
	lru := NewLruCacher(60000)
	go lru.LruCacheScheduler()
	insert_data := &UserData{
		UKey:   "test",
		UValue: "test",
	}

	fmt.Println(lru.Set(insert_data))
	time.Sleep(1 * time.Second)
	fmt.Println(lru.Get("test"))
}
