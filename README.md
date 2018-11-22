[![Build Status](https://travis-ci.org/janstuemmel/csgo-log.svg?branch=master)](https://travis-ci.org/janstuemmel/csgo-log) [![Coverage Status](https://coveralls.io/repos/github/janstuemmel/csgo-log/badge.svg?branch=master)](https://coveralls.io/github/janstuemmel/csgo-log?branch=master) [![Godoc](https://godoc.org/github.com/janstuemmel/csgo-log?status.svg)](http://godoc.org/github.com/janstuemmel/csgo-log)

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