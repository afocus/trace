# trace

基于`opentelemetry`和`zerolog` 实现的链路追踪利器。 


## 初始化

全局初始化 指定名字的trace，以及数据导出方式。目前支持stdout，http，grpc

```go
func main(){

    shutdown, err := trace.InitProvider("afocus", trace.ExportStdOut())
	if err != nil {
		log.Fatal(err)
	}
    defer shutdown()
    // todo
}
```


## 开始一个trace

使用Start开启一个trace 并使用其End方法结束trace.    
trace 使用context.Context传递

```go
t, ctx := trace.Start(parentCtx, "执行10次sleep")
defer t.End()

for i:=0;i<10;i++{
    // 子任务
    ts, _ := trace.Start(ctx, fmt.Sprintf("执行第%d次",i+1))
    time.Sleep(time.Second)
    ts.End()
}
```

## 为trace设置一些属性

一般情况下trace会携带一些属性，通过属性我们可以更将详细的了解过程

```go

uid := 1234

// 一开始设置
t, ctx := trace.Start(parentCtx, "做一次请求", trace.Attribute("uid",uid))
defer t.End()

// 中途设置
t.SetAttribute(trace.Attribute("uid",uid))

```

## 初始化

```go

func main(){

    shutdown, err := trace.InitProvider("my service", trace.ExportStdOut())
    checkerr(err)
    defer shutdown()

    // todo
}


```

## 使用日志

trace的日志基于`zerolog` 并自动携带trace信息...
与日志结合起来主要是 为了通过trace可以找到日志 反之  通过一条日志 可以找到整个链路

```go
t, ctx := trace.Start(parentCtx, "做一次请求", trace.Attribute("uid",uid))
defer t.End()

t.Log().Info().Msg("xxxxx)

```


## 封装一个功能库

很多操作通过封装自动启用trace，可以有效的简化代码。如数据库操作，接口请求等   
这里已经封装了一些常用的功能

* http
* gorm
* redis
* gin





