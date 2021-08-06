package hashmap

//hashmap的单元数据
type HashmapItem struct {
	Key         string
	Value       interface{}
	LastVisitor int64 //记录上次访问的时间，用于冲突时覆盖写
}
