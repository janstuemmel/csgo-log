# csgo-log

Go package for parsing csgo server logfiles.

## Usage

For more examples look at the tests.

```go
package main

import (
  "fmt"

  "github.com/janstuemmel/csgolog"
)

func main() {

  var (
    msg csgolog.Message
    err error
    jsn string
  )

  // a line from a server logfile
  line := `L 11/05/2018 - 15:44:36: "Player<12><STEAM_1:1:0101011><TERRORIST>" purchased "m4a1"`

  // parse into Message
  msg, err = csgolog.Parse(line)

  if err != nil {
    panic(err)
  }

  fmt.Println(msg)

  // get json non-htmlescaped
  jsn = csgolog.ToJSON(msg)

  fmt.Println(jsn)
}
```