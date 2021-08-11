package hashmap

//hashmap的单元数据
type HashmapItem struct {
	HKey        string
	HValue      interface{} //interface{}为引用类型，这里存储*list.Element
	LastVisitor int64       //记录上次访问的时间，用于冲突时覆盖写
}
