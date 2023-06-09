/*
Package cs2 provides utilities for parsing a counter-strike 2 server logfile.
It exports types for cs2 logfiles, their regular expressions, a function
for parsing and a function for converting to non-html-escaped JSON.

Look at the examples for Parse and ToJSON for usage instructions.
*/
package cs2

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrorNoMatch error when pattern is not matching
var ErrorNoMatch = errors.New("no match")

// LogLinePattern is the regular expression to capture a line of a logfile
var LogLinePattern = regexp.MustCompile(`L (\d{2}\/\d{2}\/\d{4} - \d{2}:\d{2}:\d{2}): (.*)`)

type (

	// Player holds the information about a player known from log
	Player struct {
		Name    string `json:"name"`
		ID      int    `json:"id"`
		SteamID string `json:"steam_id"`
		Side    string `json:"side"`
	}

	// Position holds the coords for a event happend on the map
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
		Z int `json:"z"`
	}

	// PositionFloat holds more exact coords
	PositionFloat struct {
		X float32 `json:"x"`
		Y float32 `json:"y"`
		Z float32 `json:"z"`
	}

	// Velocity holds information about the velocity of a projectile
	Velocity struct {
		X float32 `json:"x"`
		Y float32 `json:"y"`
		Z float32 `json:"z"`
	}

	// Equation holds the parameters and result of a money change equation
	// in the form A + B = Result
	Equation struct {
		A      int `json:"a"`
		B      int `json:"b"`
		Result int `json:"result"`
	}

	// Message is the interface for all messages
	Message interface {
		GetType() string
		GetTime() time.Time
	}

	// Meta holds time and type of a log message
	Meta struct {
		Time time.Time `json:"time"`
		Type string    `json:"type"`
	}

	// ServerMessage is received on a server event
	ServerMessage struct {
		Meta
		Text string `json:"text"`
	}

	// FreezTimeStart is received before each round
	FreezTimeStart struct{ Meta }

	// WorldMatchStart holds the map wich will be played when match starts
	WorldMatchStart struct {
		Meta
		Map string `json:"map"`
	}

	// WorldRoundStart message is received when a new round starts
	WorldRoundStart struct{ Meta }

	// WorldRoundRestart is received when the server wants to restart a round
	WorldRoundRestart struct {
		Meta
		Timeleft int `json:"timeleft"`
	}

	// WorldRoundEnd message is received when a round ends
	WorldRoundEnd struct{ Meta }

	// WorldGameCommencing message is received when a game is commencing
	WorldGameCommencing struct{ Meta }

	// TeamScored is received at the end of each round and holds
	// the scores for a team
	TeamScored struct {
		Meta
		Side       string `json:"side"`
		Score      int    `json:"score"`
		NumPlayers int    `json:"num_players"`
	}

	// TeamNotice message is received at the end of a round and holds
	// information about which team won the round and the score
	TeamNotice struct {
		Meta
		Side    string `json:"side"`
		Notice  string `json:"notice"`
		ScoreCT int    `json:"score_ct"`
		ScoreT  int    `json:"score_t"`
	}

	// PlayerConnected message is received when a player connects and
	// holds the address from where the player is connecting
	PlayerConnected struct {
		Meta
		Player  Player `json:"player"`
		Address string `json:"address"`
	}

	// PlayerDisconnected is received when a player disconnets and
	// holds the reason why the player left
	PlayerDisconnected struct {
		Meta
		Player Player `json:"player"`
		Reason string `json:"reason"`
	}

	// PlayerEntered is received when a player enters the game
	PlayerEntered struct {
		Meta
		Player Player `json:"player"`
	}

	// PlayerBanned is received when a player gots banned from the server
	PlayerBanned struct {
		Meta
		Player   Player `json:"player"`
		Duration string `json:"duration"`
		By       string `json:"by"`
	}

	// PlayerSwitched is received when a player switches sides
	PlayerSwitched struct {
		Meta
		Player Player `json:"player"`
		From   string `json:"from"`
		To     string `json:"to"`
	}

	// PlayerSay is received when a player writes into chat
	PlayerSay struct {
		Meta
		Player Player `json:"player"`
		Text   string `json:"text"`
		Team   bool   `json:"team"`
	}

	// PlayerPurchase holds info about which player bought an item
	PlayerPurchase struct {
		Meta
		Player Player `json:"player"`
		Item   string `json:"item"`
	}

	// PlayerKill is received when a player kills another
	PlayerKill struct {
		Meta
		Attacker         Player   `json:"attacker"`
		AttackerPosition Position `json:"attacker_pos"`
		Victim           Player   `json:"victim"`
		VictimPosition   Position `json:"victim_pos"`
		Weapon           string   `json:"weapon"`
		Headshot         bool     `json:"headshot"`
		Penetrated       bool     `json:"penetrated"`
	}

	// PlayerKillAssist is received when a player assisted killing another
	PlayerKillAssist struct {
		Meta
		Attacker Player `json:"attacker"`
		Victim   Player `json:"victim"`
	}

	// PlayerAttack is recieved when a player attacks another
	PlayerAttack struct {
		Meta
		Attacker         Player   `json:"attacker"`
		AttackerPosition Position `json:"attacker_pos"`
		Victim           Player   `json:"victim"`
		VictimPosition   Position `json:"victim_pos"`
		Weapon           string   `json:"weapon"`
		Damage           int      `json:"damage"`
		DamageArmor      int      `json:"damage_armor"`
		Health           int      `json:"health"`
		Armor            int      `json:"armor"`
		Hitgroup         string   `json:"hitgroup"`
	}

	// PlayerKilledBomb is received when a player is killed by the bomb
	PlayerKilledBomb struct {
		Meta
		Player   Player   `json:"player"`
		Position Position `json:"pos"`
	}

	// PlayerKilledSuicide is received when a player commited suicide
	PlayerKilledSuicide struct {
		Meta
		Player   Player   `json:"player"`
		Position Position `json:"pos"`
		With     string   `json:"with"`
	}

	// PlayerPickedUp is received when a player picks up an item
	PlayerPickedUp struct {
		Meta
		Player Player `json:"player"`
		Item   string `json:"item"`
	}

	// PlayerDropped is recieved when a player drops an item
	PlayerDropped struct {
		Meta
		Player Player `json:"player"`
		Item   string `json:"item"`
	}

	// PlayerMoneyChange is received when a player loses or receives money
	// TODO: add before +-money
	PlayerMoneyChange struct {
		Meta
		Player   Player   `json:"player"`
		Equation Equation `json:"equation"`
		Purchase string   `json:"purchase"`
	}

	// PlayerBombGot is received when a player picks up the bomb
	PlayerBombGot struct {
		Meta
		Player Player `json:"player"`
	}

	// PlayerBombPlanted is received when a player plants the bomb
	PlayerBombPlanted struct {
		Meta
		Player Player `json:"player"`
	}

	// PlayerBombDropped is received when a player drops the bomb
	PlayerBombDropped struct {
		Meta
		Player Player `json:"player"`
	}

	// PlayerBombBeginDefuse is received when a player begins
	// defusing the bomb
	PlayerBombBeginDefuse struct {
		Meta
		Player Player `json:"player"`
		Kit    bool   `json:"kit"`
	}

	// PlayerBombDefused is received when a player defused the bomb
	PlayerBombDefused struct {
		Meta
		Player Player `json:"player"`
	}

	// PlayerThrew is received when a player threw a grenade
	PlayerThrew struct {
		Meta
		Player   Player   `json:"player"`
		Position Position `json:"pos"`
		Entindex int      `json:"entindex"`
		Grenade  string   `json:"grenade"`
	}

	// PlayerBlinded is received when a player got blinded
	PlayerBlinded struct {
		Meta
		Attacker Player  `json:"attacker"`
		Victim   Player  `json:"victim"`
		For      float32 `json:"for"`
		Entindex int     `json:"entindex"`
	}

	// ProjectileSpawned is received when a molotov spawned
	ProjectileSpawned struct {
		Meta
		Position PositionFloat `json:"pos"`
		Velocity Velocity      `json:"velocity"`
	}

	// GameOver is received when a team won and the game ends
	GameOver struct {
		Meta
		Mode     string `json:"mode"`
		MapGroup string `json:"map_group"`
		Map      string `json:"map"`
		ScoreCT  int    `json:"score_ct"`
		ScoreT   int    `json:"score_t"`
		Duration int    `json:"duration"`
	}

	// Unknown holds the raw log message of a message
	// that is not defined in patterns but starts with time
	Unknown struct {
		Meta
		Raw string `json:"raw"`
	}
)

// GetType is the getter fo Meta.Type
func (m Meta) GetType() string {
	return m.Type
}

// GetTime is the getter for Meta.Time
func (m Meta) GetTime() time.Time {
	return m.Time
}

type MessageFunc func(ti time.Time, r []string) Message

const (
	// ServerMessagePattern regular expression
	ServerMessagePattern = `server_message: "(\w+)"`
	// FreezTimeStartPattern regular expression
	FreezTimeStartPattern = `Starting Freeze period`
	// WorldMatchStartPattern regular expression
	WorldMatchStartPattern = `World triggered "Match_Start" on "(\w+)"`
	// WorldRoundStartPattern regular expression
	WorldRoundStartPattern = `World triggered "Round_Start"`
	// WorldRoundRestartPattern regular expression
	WorldRoundRestartPattern = `World triggered "Restart_Round_\((\d+)_second\)`
	// WorldRoundEndPattern regular expression
	WorldRoundEndPattern = `World triggered "Round_End"`
	// WorldGameCommencingPattern regular expression
	WorldGameCommencingPattern = `World triggered "Game_Commencing"`
	// TeamScoredPattern regular expression
	TeamScoredPattern = `Team "(CT|TERRORIST)" scored "(\d+)" with "(\d+)" players`
	// TeamNoticePattern regular expression
	TeamNoticePattern = `Team "(CT|TERRORIST)" triggered "(\w+)" \(CT "(\d+)"\) \(T "(\d+)"\)`
	// PlayerConnectedPattern regular expression
	PlayerConnectedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><>" connected, address "(.*)"`
	// PlayerDisconnectedPattern regular expression
	PlayerDisconnectedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT|Unassigned|)>" disconnected \(reason "(.+)"\)`
	// PlayerEnteredPattern regular expression
	PlayerEnteredPattern = `"(.+)<(\d+)><([\[\]\w:]+)><>" entered the game`
	// PlayerBannedPattern regular expression
	PlayerBannedPattern = `Banid: "(.+)<(\d+)><([\[\]\w:]+)><\w*>" was banned "([\w. ]+)" by "(\w+)"`
	// PlayerSwitchedPattern regular expression
	PlayerSwitchedPattern = `"(.+)<(\d+)><([\[\]\w:]+)>" switched from team <(Unassigned|Spectator|TERRORIST|CT)> to <(Unassigned|Spectator|TERRORIST|CT)>`
	// PlayerSayPattern regular expression
	PlayerSayPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" say(_team)? "(.*)"`
	// PlayerPurchasePattern regular expression
	PlayerPurchasePattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" purchased "(\w+)"`
	// PlayerKillPattern regular expression
	PlayerKillPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] killed "(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w+)" ?(\(?(headshot|penetrated|headshot penetrated)?\))?`
	// PlayerKillAssistPattern regular expression
	PlayerKillAssistPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" assisted killing "(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>"`
	// PlayerAttackPattern regular expression
	PlayerAttackPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] attacked "(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w+)" \(damage "(\d+)"\) \(damage_armor "(\d+)"\) \(health "(\d+)"\) \(armor "(\d+)"\) \(hitgroup "([\w ]+)"\)`
	// PlayerKilledBombPattern regular expression
	PlayerKilledBombPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] was killed by the bomb\.`
	// PlayerKilledSuicidePattern regular expression
	PlayerKilledSuicidePattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] committed suicide with "(.*)"`
	// PlayerPickedUpPattern regular expression
	PlayerPickedUpPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" picked up "(\w+)"`
	// PlayerDroppedPattern regular expression
	PlayerDroppedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT|Unassigned)>" dropped "(\w+)"`
	// PlayerMoneyChangePattern regular expression
	PlayerMoneyChangePattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" money change (\d+)\+?(-?\d+) = \$(\d+) \(tracked\)( \(purchase: (\w+)\))?`
	// PlayerBombGotPattern regular expression
	PlayerBombGotPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" triggered "Got_The_Bomb"`
	// PlayerBombPlantedPattern regular expression
	PlayerBombPlantedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" triggered "Planted_The_Bomb"`
	// PlayerBombDroppedPattern regular expression
	PlayerBombDroppedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" triggered "Dropped_The_Bomb"`
	// PlayerBombBeginDefusePattern regular expression
	PlayerBombBeginDefusePattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" triggered "Begin_Bomb_Defuse_With(out)?_Kit"`
	// PlayerBombDefusedPattern regular expression
	PlayerBombDefusedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" triggered "Defused_The_Bomb"`
	// PlayerThrewPattern regular expression
	PlayerThrewPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" threw (\w+) \[(-?\d+) (-?\d+) (-?\d+)\]( flashbang entindex (\d+))?\)?`
	// PlayerBlindedPattern regular expression
	PlayerBlindedPattern = `"(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" blinded for ([\d.]+) by "(.+)<(\d+)><([\[\]\w:]+)><(TERRORIST|CT)>" from flashbang entindex (\d+)`
	// ProjectileSpawnedPattern regular expression
	ProjectileSpawnedPattern = `Molotov projectile spawned at (-?\d+\.\d+) (-?\d+\.\d+) (-?\d+\.\d+), velocity (-?\d+\.\d+) (-?\d+\.\d+) (-?\d+\.\d+)`
	// GameOverPattern regular expression
	GameOverPattern = `Game Over: (\w+) (\w+) (\w+) score (\d+):(\d+) after (\d+) min`
)

var DefaultPatterns = map[*regexp.Regexp]MessageFunc{
	regexp.MustCompile(ServerMessagePattern):         NewServerMessage,
	regexp.MustCompile(FreezTimeStartPattern):        NewFreezTimeStart,
	regexp.MustCompile(WorldMatchStartPattern):       NewWorldMatchStart,
	regexp.MustCompile(WorldRoundStartPattern):       NewWorldRoundStart,
	regexp.MustCompile(WorldRoundRestartPattern):     NewWorldRoundRestart,
	regexp.MustCompile(WorldRoundEndPattern):         NewWorldRoundEnd,
	regexp.MustCompile(WorldGameCommencingPattern):   NewWorldGameCommencing,
	regexp.MustCompile(TeamScoredPattern):            NewTeamScored,
	regexp.MustCompile(TeamNoticePattern):            NewTeamNotice,
	regexp.MustCompile(PlayerConnectedPattern):       NewPlayerConnected,
	regexp.MustCompile(PlayerDisconnectedPattern):    NewPlayerDisconnected,
	regexp.MustCompile(PlayerEnteredPattern):         NewPlayerEntered,
	regexp.MustCompile(PlayerBannedPattern):          NewPlayerBanned,
	regexp.MustCompile(PlayerSwitchedPattern):        NewPlayerSwitched,
	regexp.MustCompile(PlayerSayPattern):             NewPlayerSay,
	regexp.MustCompile(PlayerPurchasePattern):        NewPlayerPurchase,
	regexp.MustCompile(PlayerKillPattern):            NewPlayerKill,
	regexp.MustCompile(PlayerKillAssistPattern):      NewPlayerKillAssist,
	regexp.MustCompile(PlayerAttackPattern):          NewPlayerAttack,
	regexp.MustCompile(PlayerKilledBombPattern):      NewPlayerKilledBomb,
	regexp.MustCompile(PlayerKilledSuicidePattern):   NewPlayerKilledSuicide,
	regexp.MustCompile(PlayerPickedUpPattern):        NewPlayerPickedUp,
	regexp.MustCompile(PlayerDroppedPattern):         NewPlayerDropped,
	regexp.MustCompile(PlayerMoneyChangePattern):     NewPlayerMoneyChange,
	regexp.MustCompile(PlayerBombGotPattern):         NewPlayerBombGot,
	regexp.MustCompile(PlayerBombPlantedPattern):     NewPlayerBombPlanted,
	regexp.MustCompile(PlayerBombDroppedPattern):     NewPlayerBombDropped,
	regexp.MustCompile(PlayerBombBeginDefusePattern): NewPlayerBombBeginDefuse,
	regexp.MustCompile(PlayerBombDefusedPattern):     NewPlayerBombDefused,
	regexp.MustCompile(PlayerThrewPattern):           NewPlayerThrew,
	regexp.MustCompile(PlayerBlindedPattern):         NewPlayerBlinded,
	regexp.MustCompile(ProjectileSpawnedPattern):     NewProjectileSpawned,
	regexp.MustCompile(GameOverPattern):              NewGameOver,
}

// Parse parses a plain log message and returns
// message type or error if there's no match
func Parse(line string) (Message, error) {
	return ParseWithPatterns(line, DefaultPatterns)
}

// Parse attempts to match a plain log message against the map of provided patterns,
// if the line matches a key from the map, the corresponding MessageFunc is called on the line to
// parse it into a Message
func ParseWithPatterns(line string, patterns map[*regexp.Regexp]MessageFunc) (Message, error) {
	// pattern for date, beginning of a log message
	result := LogLinePattern.FindStringSubmatch(line)

	// if result set is empty, parsing failed, return error
	if result == nil {
		return nil, ErrorNoMatch
	}

	// parse time
	ti, err := time.Parse("01/02/2006 - 15:04:05", result[1])

	// if parsing the date failed, return error
	if err != nil {
		return nil, err
	}

	// check all patterns, return if a pattern matches
	for re, fun := range patterns {
		if result := re.FindStringSubmatch(result[2]); result != nil {
			return fun(ti, result), nil
		}
	}

	// if there was no match above but format of the log message was correct
	// it's a valid logline but pattern is not defined, return unknown type
	return NewUnknown(ti, result[1:]), nil
}

// ToJSON marshals messages to JSON without escaping html
func ToJSON(m Message) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.Encode(m)
	return buf.String()
}

func NewMeta(ti time.Time, ty string) Meta {
	return Meta{
		Time: ti,
		Type: ty,
	}
}

func NewServerMessage(ti time.Time, r []string) Message {
	return ServerMessage{
		Meta: NewMeta(ti, "ServerMessage"),
		Text: r[1],
	}
}

func NewFreezTimeStart(ti time.Time, r []string) Message {
	return FreezTimeStart{NewMeta(ti, "FreezTimeStart")}
}

func NewWorldMatchStart(ti time.Time, r []string) Message {
	return WorldMatchStart{
		Meta: NewMeta(ti, "WorldMatchStart"),
		Map:  r[1],
	}
}

func NewWorldRoundStart(ti time.Time, r []string) Message {
	return WorldRoundStart{NewMeta(ti, "WorldRoundStart")}
}

func NewWorldRoundRestart(ti time.Time, r []string) Message {
	return WorldRoundRestart{
		Meta:     NewMeta(ti, "WorldRoundRestart"),
		Timeleft: toInt(r[1]),
	}
}

func NewWorldRoundEnd(ti time.Time, r []string) Message {
	return WorldRoundEnd{NewMeta(ti, "WorldRoundEnd")}
}

func NewWorldGameCommencing(ti time.Time, r []string) Message {
	return WorldGameCommencing{NewMeta(ti, "WorldGameCommencing")}
}

func NewTeamScored(ti time.Time, r []string) Message {
	return TeamScored{
		Meta:       NewMeta(ti, "TeamScored"),
		Side:       r[1],
		Score:      toInt(r[2]),
		NumPlayers: toInt(r[3]),
	}
}

func NewTeamNotice(ti time.Time, r []string) Message {
	return TeamNotice{
		Meta:    NewMeta(ti, "TeamNotice"),
		Side:    r[1],
		Notice:  r[2],
		ScoreCT: toInt(r[3]),
		ScoreT:  toInt(r[4]),
	}
}

func NewPlayerConnected(ti time.Time, r []string) Message {
	return PlayerConnected{
		Meta: NewMeta(ti, "PlayerConnected"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    "",
		},
		Address: r[4],
	}
}

func NewPlayerDisconnected(ti time.Time, r []string) Message {
	return PlayerDisconnected{
		Meta: NewMeta(ti, "PlayerDisconnected"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Reason: r[5],
	}
}

func NewPlayerEntered(ti time.Time, r []string) Message {
	return PlayerEntered{
		Meta: NewMeta(ti, "PlayerEntered"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    "",
		},
	}
}

func NewPlayerBanned(ti time.Time, r []string) Message {
	return PlayerBanned{
		Meta: NewMeta(ti, "PlayerBanned"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    "",
		},
		Duration: r[4],
		By:       r[5],
	}
}

func NewPlayerSwitched(ti time.Time, r []string) Message {
	return PlayerSwitched{
		Meta: NewMeta(ti, "PlayerSwitched"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    "",
		},
		From: r[4],
		To:   r[5],
	}
}

func NewPlayerSay(ti time.Time, r []string) Message {
	return PlayerSay{
		Meta: NewMeta(ti, "PlayerSay"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Team: r[5] == "_team",
		Text: r[6],
	}
}

func NewPlayerPurchase(ti time.Time, r []string) Message {
	return PlayerPurchase{
		Meta: NewMeta(ti, "PlayerPurchase"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Item: r[5],
	}
}

func NewPlayerKill(ti time.Time, r []string) Message {
	return PlayerKill{
		Meta: NewMeta(ti, "PlayerKill"),
		Attacker: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		AttackerPosition: Position{
			X: toInt(r[5]),
			Y: toInt(r[6]),
			Z: toInt(r[7]),
		},
		Victim: Player{
			Name:    r[8],
			ID:      toInt(r[9]),
			SteamID: r[10],
			Side:    r[11],
		},
		VictimPosition: Position{
			X: toInt(r[12]),
			Y: toInt(r[13]),
			Z: toInt(r[14]),
		},
		Weapon:     r[15],
		Headshot:   strings.Contains(r[17], "headshot"),
		Penetrated: strings.Contains(r[17], "penetrated"),
	}
}

func NewPlayerKillAssist(ti time.Time, r []string) Message {
	return PlayerKillAssist{
		Meta: NewMeta(ti, "PlayerKillAssist"),
		Attacker: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Victim: Player{
			Name:    r[5],
			ID:      toInt(r[6]),
			SteamID: r[7],
			Side:    r[8],
		},
	}
}

func NewPlayerAttack(ti time.Time, r []string) Message {
	return PlayerAttack{
		Meta: NewMeta(ti, "PlayerAttack"),
		Attacker: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		AttackerPosition: Position{
			X: toInt(r[5]),
			Y: toInt(r[6]),
			Z: toInt(r[7]),
		},
		Victim: Player{
			Name:    r[8],
			ID:      toInt(r[9]),
			SteamID: r[10],
			Side:    r[11],
		},
		VictimPosition: Position{
			X: toInt(r[12]),
			Y: toInt(r[13]),
			Z: toInt(r[14]),
		},
		Weapon:      r[15],
		Damage:      toInt(r[16]),
		DamageArmor: toInt(r[17]),
		Health:      toInt(r[18]),
		Armor:       toInt(r[19]),
		Hitgroup:    r[20],
	}
}

func NewPlayerKilledBomb(ti time.Time, r []string) Message {
	return PlayerKilledBomb{
		Meta: NewMeta(ti, "PlayerKilledBomb"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Position: Position{
			X: toInt(r[5]),
			Y: toInt(r[6]),
			Z: toInt(r[7]),
		},
	}
}

func NewPlayerKilledSuicide(ti time.Time, r []string) Message {
	return PlayerKilledSuicide{
		Meta: NewMeta(ti, "PlayerKilledSuicide"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Position: Position{
			X: toInt(r[5]),
			Y: toInt(r[6]),
			Z: toInt(r[7]),
		},
		With: r[8],
	}
}

func NewPlayerPickedUp(ti time.Time, r []string) Message {
	return PlayerPickedUp{
		Meta: NewMeta(ti, "PlayerPickedUp"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Item: r[5],
	}
}

func NewPlayerDropped(ti time.Time, r []string) Message {
	return PlayerDropped{
		Meta: NewMeta(ti, "PlayerDropped"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Item: r[5],
	}
}

func NewPlayerMoneyChange(ti time.Time, r []string) Message {
	return PlayerMoneyChange{
		Meta: NewMeta(ti, "PlayerMoneyChange"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Equation: Equation{
			A:      toInt(r[5]),
			B:      toInt(r[6]),
			Result: toInt(r[7]),
		},
		Purchase: r[9],
	}
}

func NewPlayerBombGot(ti time.Time, r []string) Message {
	return PlayerBombGot{
		Meta: NewMeta(ti, "PlayerBombGot"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func NewPlayerBombPlanted(ti time.Time, r []string) Message {
	return PlayerBombPlanted{
		Meta: NewMeta(ti, "PlayerBombPlanted"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func NewPlayerBombDropped(ti time.Time, r []string) Message {
	return PlayerBombDropped{
		Meta: NewMeta(ti, "PlayerBombDropped"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func NewPlayerBombBeginDefuse(ti time.Time, r []string) Message {
	return PlayerBombBeginDefuse{
		Meta: NewMeta(ti, "PlayerBombBeginDefuse"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Kit: !(r[5] == "out"),
	}
}

func NewPlayerBombDefused(ti time.Time, r []string) Message {
	return PlayerBombDefused{
		Meta: NewMeta(ti, "PlayerBombDefused"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func NewPlayerThrew(ti time.Time, r []string) Message {
	return PlayerThrew{
		Meta: NewMeta(ti, "PlayerThrew"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Grenade: r[5],
		Position: Position{
			X: toInt(r[6]),
			Y: toInt(r[7]),
			Z: toInt(r[8]),
		},
		Entindex: toInt(r[10]),
	}
}

func NewPlayerBlinded(ti time.Time, r []string) Message {
	return PlayerBlinded{
		Meta: NewMeta(ti, "PlayerBlinded"),
		Victim: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		For: toFloat32(r[5]),
		Attacker: Player{
			Name:    r[6],
			ID:      toInt(r[7]),
			SteamID: r[8],
			Side:    r[9],
		},
		Entindex: toInt(r[10]),
	}
}

func NewProjectileSpawned(ti time.Time, r []string) Message {
	return ProjectileSpawned{
		Meta: NewMeta(ti, "ProjectileSpawned"),
		Position: PositionFloat{
			X: toFloat32(r[1]),
			Y: toFloat32(r[2]),
			Z: toFloat32(r[3]),
		},
		Velocity: Velocity{
			X: toFloat32(r[4]),
			Y: toFloat32(r[5]),
			Z: toFloat32(r[6]),
		},
	}
}

func NewGameOver(ti time.Time, r []string) Message {
	return GameOver{
		Meta:     NewMeta(ti, "GameOver"),
		Mode:     r[1],
		MapGroup: r[2],
		Map:      r[3],
		ScoreCT:  toInt(r[4]),
		ScoreT:   toInt(r[5]),
		Duration: toInt(r[6]),
	}
}

func NewUnknown(ti time.Time, r []string) Message {
	return Unknown{
		Meta: NewMeta(ti, "Unknown"),
		Raw:  r[1],
	}
}

// helpers

// toInt converts string to int, assigns 0 when not convertable
func toInt(v string) int {

	i, err := strconv.Atoi(v)

	if err != nil {
		return 0
	}

	return i
}

func toFloat32(v string) float32 {

	i, err := strconv.ParseFloat(v, 32)

	if err != nil {
		return float32(0)
	}

	return float32(i)
}
