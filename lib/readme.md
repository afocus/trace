# lib

基于 github.com/afocus/trace 封装的一些常用库

* git
* http handler/client
* redis
* gorm
* 更多稍后添加




## gin

```go

import tracegin "github.com/afocus/trace/lib/gin"
import tracehttp "github.com/afocus/trace/lib/http"

router := gin.New()
router.Use(tracegin.Middlewave(), gin.Recovery())

router.GET("/", func(c *gin.Context){
    var (
		ctx      = c.Request.Context()
		log      = trace.GetLog(ctx)
	)

    log.Info().Msg("request in")
    resp,err:=tracehttp.Get(ctx,"http://www.baidu.com")
    if err!=nil{
        log.Error().Err(err).Msg("request faild")
        c.String(500, err.Error())
        return
    }


    c.String(200,"ok")


})

```



## http

```go

import tracehttp "github.com/afocus/trace/lib/http"


var client = &http.Client{
    Timeout:time.Second*10,
    Transport: tracehttp.NewTransport(http.DefaultTransport), // 注入trace
}


req,err:=http.NewRequestWithContext(parentCtx, "GET", "http://www.badiu.com", nil)
if err!=nil{

}

client.Do(req)

```


## gorm

```go

import tracegorm "github.com/afocus/trace/lib/gorm"

var db *gorm.DB

func initDatabase(dsn string) error {
	var err error
	db, err = gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{
			Logger: tracegorm.Logger(0), // 注入日志
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		},
	)
	if err != nil {
		return err
	}

	db.Use(tracegorm.Plugin()) // 注入trace
	return nil
}

// 再model中使用

type NParkArm struct {
	ParkCode    int64  `gorm:"park_code"`
	SN          string `gorm:"sn"`
	SoftVersion string `gorm:"soft_version"`
}

func GetArms(ctx context.Context, arms *[]NParkArm) error {
	trace.GetLog(ctx).Debug().Msg("进入 model方法 GetArms")
	return db.WithContext(ctx).Where("park_code = 10000000").Find(arms).Error
}
```

## amqp
```go
import traceamqp "github.com/afocus/trace/lib/amqp"
for msg := range sub.GetMessages() {
	switch e := msg.(type) {
	case *amqp.Delivery:
		ctx, err := traceamqp.SubHeader("sub msg", &e.Delivery)
		if err != nil {
			log.Println(err)
			e.Accpet(true)
		}
		log.Printf("Received a message: %s", e.Body)
		fn(ctx)
	}
}
```

