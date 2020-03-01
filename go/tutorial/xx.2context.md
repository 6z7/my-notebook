## context

`context`用于不同go协程之间的协调与数据传递。

常用的方法有

* WithCancel:主动取消来通知其它协程

* WithTimeout:定时到期或主动取消来通知其它协程

* WithDeadline:定时到期或主动取消来通知其它协程

* WithValue：传递kv

`context`需要关联一个root context,常用的有`context.Background()`和`context.TODO()`。

创建`context`时会将新建的contxt添加到父级别的context中做为它的子项，当取消时会从父级别context中移除context。

demo：

```go
package context

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDemo1A(t *testing.T) {
	exifFlag := make(chan bool)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			time.Sleep(time.Microsecond)
			select {
			case <-ctx.Done():
				fmt.Println("结束")
				exifFlag <- true
				return
			default:
				fmt.Println("运行中")
			}
		}

	}(ctx)
	time.Sleep(50 * time.Microsecond)
	cancelFunc()
	fmt.Println("main")
	<-exifFlag
}

func TestDemo1B(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Microsecond)
	go func(ctx context.Context) {
		for {
			time.Sleep(time.Microsecond)
			fmt.Println("运行中")
		}

	}(ctx)
	<-ctx.Done()
	fmt.Println("main")
}

func TestDemo1C(t *testing.T) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(150*time.Microsecond))
	go func(ctx context.Context) {
		for {
			time.Sleep(time.Microsecond)
			fmt.Println("运行中")
		}
	}(ctx)
	<-ctx.Done()
	fmt.Println("main")
}

func TestDemo1D(t *testing.T) {
	ctx := context.WithValue(context.Background(), "name", "abc")
	go func(ctx context.Context) {
		name := ctx.Value("name").(string)
		fmt.Println(name)
	}(ctx)
	time.Sleep(1 * time.Second)
	fmt.Println("main")
}
```