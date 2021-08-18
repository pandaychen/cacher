# cacher
a simple cache storage with redis/local-cache


### Build && Test
1. example
```golang
package main

import (
        "fmt"
        "time"
        "github.com/pandaychen/cacher"
)

func main() {
        lru := cacher.NewLruCacher(60000)
        go lru.LruCacheScheduler()
        insert_data := &cacher.UserData{
                UKey:   "test",
                UValue: "test",
        }

        fmt.Println(lru.Set(insert_data))
        time.Sleep(1 * time.Second)
        fmt.Println(lru.Get("test"))
}
```

2. Build
```bash
go mod init cache
go mod tidy
go build main.go
```


