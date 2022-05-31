# go-env

1. Provides an env struct field tag to unmarshal environment variables.
2. Export env vars from a .env file
3. Environment variables substitution for Go

## Installation

```shell
go get github.com/ichaly/go-env
```

## Usage

Add your application configuration to your `.env` file in the root of your project:

```shell
S3_BUCKET=YOURS3BUCKET
SECRET_KEY=YOURSECRETKEYGOESHERE
```

Then in your Go app you can do something like

```go
package main

import (
    "github.com/ichaly/go-env"
    "log"
    "os"
)

func main() {
    err := env.Export()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    s3Bucket := os.Getenv("S3_BUCKET")
    secretKey := os.Getenv("SECRET_KEY")
    // now do something with s3 or whatever
}
```

If you're even lazier than that, you can just take advantage of the autoload package which will read in `.env` on import

```go
import _ "github.com/ichaly/go-env/auto"
```

Load environment variables with default value

```go
package main

import (
    "log"
    "github.com/ichaly/go-env"
)

func main() {
    res, _ := env.String("Server ip is :${ip:=127.0.0.1},port is ${port:=8080}")
    log.Printf("----->>>%v", res)
}
```

Use env struct field tag to unmarshal environment variables

```go
import (
    "log"
    "github.com/ichaly/go-env"
)

type Redis struct {
    Password string `env:"PASSWORD,default=redis123"`
}
type Mysql struct {
    Password string `env:"PASSWORD,default=mysql123"`
}
type Config struct {
    Port     int     `env:"PORT,default=8080"`
    Username string  `env:"USERNAME,required=true"`
    Cache    bool    `env:"CACHE1"`
    Price    float32 `env:"PRICE,default=0.0"`
    Redis    *Redis
    Mysql    Mysql
    Keys     []string
    Maps     map[int]int
}

func main() {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        panic(err)
        return
    }
    log.Printf("*****>>>%v", cfg)
}
```


## Thanks

[https://github.com/sethvargo/go-envconfig](https://github.com/sethvargo/go-envconfig)

[https://github.com/joho/godotenv](https://github.com/joho/godotenv)

[https://github.com/a8m/envsubst](https://github.com/a8m/envsubst)

[https://github.com/Netflix/go-env](https://github.com/Netflix/go-env)
