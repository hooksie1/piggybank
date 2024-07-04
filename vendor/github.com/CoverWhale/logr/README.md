# Logr

Logr is a stupid simple logging package for Go. It embeds and extends the standard library logger which means it's very simple to use.

### Example

```go
package main

import (
    "github.com/CoverWhale/logr"
)

func main() {
    logger := logr.NewLogger()

    // set level directly or with env vars
    logger.Level = logr.DebugLevel

    logger.Info("Info log")
    logger.Debug("Debug log")

    logger.Infof("logging with %s", "formatting")

    logger.Error("uh oh spaghettios")

    // Calling without the instantiated logger (relies on env vars: LOG_LEVEL=debug)
    logr.Debugf("debug level %s", "stuff")

    // You can crate a new logger with a context message 
    // Get caller returns the name of the current function
    functionContext := map[string]string{"function", logr.GetCaller()}
    ctx := logger.WithContext(functionContext)
    ctx.Info("another message")
}
```

### Output

Since this uses the default Go logger under the hood, you can easily set an output file instead of stdout.

```go
package main

import (
    "log"
    "github.com/CoverWhale/logr"
)

func main() {
	f, err := os.OpenFile("test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		log.Fatal("error")
	}
	defer f.Close()


	logger := logr.NewLogger()
	// use SetOutput from std logger
	logger.Logger.SetOutput(f)
}
```
