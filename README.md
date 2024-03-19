# infrastructure
example:
```
logger: 
    import(
        "github.com/awp0816/infrastructure/logger"
    )
    if err := logger.SetupLogger(&logger.LogConfig{
        ...
    });err != nil{
        return
    }
    defer logger.Logger.Sync()
    
    logger.Logger.Info("message......")
```