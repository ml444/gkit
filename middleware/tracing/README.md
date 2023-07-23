## 在glog使用traceId
```golang

func InitLog() error {
    err := glog.InitLog(
        config.SetFileName2Logger(name),
        func (cfg *config.Config) {
            cfg.TradeIDFunc = func (entry *message.Entry) string {
                return GetTraceIdFromCache(entry.RoutineId)
            },      
        },
    )
    if err != nil {
        log.Errorf("err: %v", err)
        return err
    }
}

```