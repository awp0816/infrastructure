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
    logger.Logger.Info("message......")
```