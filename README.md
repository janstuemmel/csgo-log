# csgo-log

Go package for parsing csgo server logfiles. It exports types for csgo logfiles, their regular expressions, a function for parsing and a function for converting to non-html-escaped JSON. 

## Usage

For more examples look at the [tests](./csgolog_test.go) and the command-line utility in [examples folder](./example). Have also a look at [godoc](http://godoc.org/github.com/janstuemmel/csgo-log).

```go
package main

import (
  "fmt"

  "github.com/janstuemmel/csgo-log"
)

func main() {

  var msg csgolog.Message

  // a line from a server logfile
  line := `L 11/05/2018 - 15:44:36: "Player<12><STEAM_1:1:0101011><CT>" purchased "m4a1"`

  // parse into Message
  msg, err := csgolog.Parse(line)

  if err != nil {
    panic(err)
  }

  fmt.Println(msg.GetType(), msg.GetTime().String())

  // cast Message interface to PlayerPurchase type
  playerPurchase, ok := msg.(csgolog.PlayerPurchase)

  if ok != true {
    panic("casting failed")
  }

  fmt.Println(playerPurchase.Player.SteamID, playerPurchase.Item)

  // get json non-htmlescaped
  jsn := csgolog.ToJSON(msg) 

  fmt.Println(jsn)
}
```
Example JSON output:
```json
{
  "time": "2018-11-05T15:44:36Z",
  "type": "PlayerPurchase",
  "player": {
    "name": "Player",
    "id": 12,
    "steam_id": "STEAM_1:1:0101011",
    "side": "CT"
  },
  "item": "m4a1"
}
```