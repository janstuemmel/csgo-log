/*

Package csgolog provides utilities for parsing a csgo server logfile.
It exports types for csgo logfiles, their regular expressions, a function
for parsing and a function for converting to non-html-escaped JSON.

Look at the examples for Parse and ToJSON for usage instructions.

You will find a command-line utility in examples folder as well as an
example logfile with ~3000 lines.
*/
package csgolog

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

type messageFunc func(ti time.Time, r []string) Message

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
	PlayerConnectedPattern = `"(\w+)<(\d+)><([\w:]+)><>" connected, address "(.*)"`
	// PlayerDisconnectedPattern regular expression
	PlayerDisconnectedPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT|Unassigned|)>" disconnected \(reason "(.+)"\)`
	// PlayerEnteredPattern regular expression
	PlayerEnteredPattern = `"(\w+)<(\d+)><([\w:]+)><>" entered the game`
	// PlayerBannedPattern regular expression
	PlayerBannedPattern = `Banid: "(\w+)<(\d+)><([\w:]+)><\w*>" was banned "([\w. ]+)" by "(\w+)"`
	// PlayerSwitchedPattern regular expression
	PlayerSwitchedPattern = `"(\w+)<(\d+)><([\w:]+)>" switched from team <(Unassigned|TERRORIST|CT)> to <(Unassigned|TERRORIST|CT)>`
	// PlayerSayPattern regular expression
	PlayerSayPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" say(_team)? "(.*)"`
	// PlayerPurchasePattern regular expression
	PlayerPurchasePattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" purchased "(\w+)"`
	// PlayerKillPattern regular expression
	PlayerKillPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] killed "(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w+)" ?(\(?(headshot|penetrated|headshot penetrated)?\))?`
	// PlayerKillAssistPattern regular expression
	PlayerKillAssistPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" assisted killing "(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>"`
	// PlayerAttackPattern regular expression
	PlayerAttackPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] attacked "(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] with "(\w+)" \(damage "(\d+)"\) \(damage_armor "(\d+)"\) \(health "(\d+)"\) \(armor "(\d+)"\) \(hitgroup "([\w ]+)"\)`
	// PlayerKilledBombPattern regular expression
	PlayerKilledBombPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] was killed by the bomb\.`
	// PlayerKilledSuicidePattern regular expression
	PlayerKilledSuicidePattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" \[(-?\d+) (-?\d+) (-?\d+)\] committed suicide with "(.*)"`
	// PlayerPickedUpPattern regular expression
	PlayerPickedUpPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" picked up "(\w+)"`
	// PlayerDroppedPattern regular expression
	PlayerDroppedPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT|Unassigned)>" dropped "(\w+)"`
	// PlayerMoneyChangePattern regular expression
	PlayerMoneyChangePattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" money change (\d+)\+?(-?\d+) = \$(\d+) \(tracked\)( \(purchase: (\w+)\))?`
	// PlayerBombGotPattern regular expression
	PlayerBombGotPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" triggered "Got_The_Bomb"`
	// PlayerBombPlantedPattern regular expression
	PlayerBombPlantedPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" triggered "Planted_The_Bomb"`
	// PlayerBombDroppedPattern regular expression
	PlayerBombDroppedPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" triggered "Dropped_The_Bomb"`
	// PlayerBombBeginDefusePattern regular expression
	PlayerBombBeginDefusePattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" triggered "Begin_Bomb_Defuse_With(out)?_Kit"`
	// PlayerBombDefusedPattern regular expression
	PlayerBombDefusedPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" triggered "Defused_The_Bomb"`
	// PlayerThrewPattern regular expression
	PlayerThrewPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" threw (\w+) \[(-?\d+) (-?\d+) (-?\d+)\]( flashbang entindex (\d+))?\)?`
	// PlayerBlindedPattern regular expression
	PlayerBlindedPattern = `"(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" blinded for ([\d.]+) by "(\w+)<(\d+)><([\w:]+)><(TERRORIST|CT)>" from flashbang entindex (\d+)`
	// ProjectileSpawnedPattern regular expression
	ProjectileSpawnedPattern = `Molotov projectile spawned at (-?\d+\.\d+) (-?\d+\.\d+) (-?\d+\.\d+), velocity (-?\d+\.\d+) (-?\d+\.\d+) (-?\d+\.\d+)`
	// GameOverPattern regular expression
	GameOverPattern = `Game Over: (\w+) (\w+) (\w+) score (\d+):(\d+) after (\d+) min`
)

var patterns = map[*regexp.Regexp]messageFunc{
	regexp.MustCompile(ServerMessagePattern):         newServerMessage,
	regexp.MustCompile(FreezTimeStartPattern):        newFreezTimeStart,
	regexp.MustCompile(WorldMatchStartPattern):       newWorldMatchStart,
	regexp.MustCompile(WorldRoundStartPattern):       newWorldRoundStart,
	regexp.MustCompile(WorldRoundRestartPattern):     newWorldRoundRestart,
	regexp.MustCompile(WorldRoundEndPattern):         newWorldRoundEnd,
	regexp.MustCompile(WorldGameCommencingPattern):   newWorldGameCommencing,
	regexp.MustCompile(TeamScoredPattern):            newTeamScored,
	regexp.MustCompile(TeamNoticePattern):            newTeamNotice,
	regexp.MustCompile(PlayerConnectedPattern):       newPlayerConnected,
	regexp.MustCompile(PlayerDisconnectedPattern):    newPlayerDisconnected,
	regexp.MustCompile(PlayerEnteredPattern):         newPlayerEntered,
	regexp.MustCompile(PlayerBannedPattern):          newPlayerBanned,
	regexp.MustCompile(PlayerSwitchedPattern):        newPlayerSwitched,
	regexp.MustCompile(PlayerSayPattern):             newPlayerSay,
	regexp.MustCompile(PlayerPurchasePattern):        newPlayerPurchase,
	regexp.MustCompile(PlayerKillPattern):            newPlayerKill,
	regexp.MustCompile(PlayerKillAssistPattern):      newPlayerKillAssist,
	regexp.MustCompile(PlayerAttackPattern):          newPlayerAttack,
	regexp.MustCompile(PlayerKilledBombPattern):      newPlayerKilledBomb,
	regexp.MustCompile(PlayerKilledSuicidePattern):   newPlayerKilledSuicide,
	regexp.MustCompile(PlayerPickedUpPattern):        newPlayerPickedUp,
	regexp.MustCompile(PlayerDroppedPattern):         newPlayerDropped,
	regexp.MustCompile(PlayerMoneyChangePattern):     newPlayerMoneyChange,
	regexp.MustCompile(PlayerBombGotPattern):         newPlayerBombGot,
	regexp.MustCompile(PlayerBombPlantedPattern):     newPlayerBombPlanted,
	regexp.MustCompile(PlayerBombDroppedPattern):     newPlayerBombDropped,
	regexp.MustCompile(PlayerBombBeginDefusePattern): newPlayerBombBeginDefuse,
	regexp.MustCompile(PlayerBombDefusedPattern):     newPlayerBombDefused,
	regexp.MustCompile(PlayerThrewPattern):           newPlayerThrew,
	regexp.MustCompile(PlayerBlindedPattern):         newPlayerBlinded,
	regexp.MustCompile(ProjectileSpawnedPattern):     newProjectileSpawned,
	regexp.MustCompile(GameOverPattern):              newGameOver,
}

// Parse parses a plain log message and returns
// message type or error if there's no match
func Parse(line string) (Message, error) {

	// pattern for date, beginning of a log message
	result := regexp.MustCompile(`L (\d{2}\/\d{2}\/\d{4} - \d{2}:\d{2}:\d{2}): (.*)`).FindStringSubmatch(line)

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
	return newUnknown(ti, result[1:]), nil
}

// ToJSON marshals messages to JSON without escaping html
func ToJSON(m Message) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.Encode(m)
	return buf.String()
}

func newMeta(ti time.Time, ty string) Meta {
	return Meta{
		Time: ti,
		Type: ty,
	}
}

func newServerMessage(ti time.Time, r []string) Message {
	return ServerMessage{
		Meta: newMeta(ti, "ServerMessage"),
		Text: r[1],
	}
}

func newFreezTimeStart(ti time.Time, r []string) Message {
	return FreezTimeStart{newMeta(ti, "FreezTimeStart")}
}

func newWorldMatchStart(ti time.Time, r []string) Message {
	return WorldMatchStart{
		Meta: newMeta(ti, "WorldMatchStart"),
		Map:  r[1],
	}
}

func newWorldRoundStart(ti time.Time, r []string) Message {
	return WorldRoundStart{newMeta(ti, "WorldRoundStart")}
}

func newWorldRoundRestart(ti time.Time, r []string) Message {
	return WorldRoundRestart{
		Meta:     newMeta(ti, "WorldRoundRestart"),
		Timeleft: toInt(r[1]),
	}
}

func newWorldRoundEnd(ti time.Time, r []string) Message {
	return WorldRoundEnd{newMeta(ti, "WorldRoundEnd")}
}

func newWorldGameCommencing(ti time.Time, r []string) Message {
	return WorldGameCommencing{newMeta(ti, "WorldGameCommencing")}
}

func newTeamScored(ti time.Time, r []string) Message {
	return TeamScored{
		Meta:       newMeta(ti, "TeamScored"),
		Side:       r[1],
		Score:      toInt(r[2]),
		NumPlayers: toInt(r[3]),
	}
}

func newTeamNotice(ti time.Time, r []string) Message {
	return TeamNotice{
		Meta:    newMeta(ti, "TeamNotice"),
		Side:    r[1],
		Notice:  r[2],
		ScoreCT: toInt(r[3]),
		ScoreT:  toInt(r[4]),
	}
}

func newPlayerConnected(ti time.Time, r []string) Message {
	return PlayerConnected{
		Meta: newMeta(ti, "PlayerConnected"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    "",
		},
		Address: r[4],
	}
}

func newPlayerDisconnected(ti time.Time, r []string) Message {
	return PlayerDisconnected{
		Meta: newMeta(ti, "PlayerDisconnected"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Reason: r[5],
	}
}

func newPlayerEntered(ti time.Time, r []string) Message {
	return PlayerEntered{
		Meta: newMeta(ti, "PlayerEntered"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    "",
		},
	}
}

func newPlayerBanned(ti time.Time, r []string) Message {
	return PlayerBanned{
		Meta: newMeta(ti, "PlayerBanned"),
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

func newPlayerSwitched(ti time.Time, r []string) Message {
	return PlayerSwitched{
		Meta: newMeta(ti, "PlayerSwitched"),
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

func newPlayerSay(ti time.Time, r []string) Message {
	return PlayerSay{
		Meta: newMeta(ti, "PlayerSay"),
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

func newPlayerPurchase(ti time.Time, r []string) Message {
	return PlayerPurchase{
		Meta: newMeta(ti, "PlayerPurchase"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Item: r[5],
	}
}

func newPlayerKill(ti time.Time, r []string) Message {
	return PlayerKill{
		Meta: newMeta(ti, "PlayerKill"),
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

func newPlayerKillAssist(ti time.Time, r []string) Message {
	return PlayerKillAssist{
		Meta: newMeta(ti, "PlayerKillAssist"),
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

func newPlayerAttack(ti time.Time, r []string) Message {
	return PlayerAttack{
		Meta: newMeta(ti, "PlayerAttack"),
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

func newPlayerKilledBomb(ti time.Time, r []string) Message {
	return PlayerKilledBomb{
		Meta: newMeta(ti, "PlayerKilledBomb"),
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

func newPlayerKilledSuicide(ti time.Time, r []string) Message {
	return PlayerKilledSuicide{
		Meta: newMeta(ti, "PlayerKilledSuicide"),
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

func newPlayerPickedUp(ti time.Time, r []string) Message {
	return PlayerPickedUp{
		Meta: newMeta(ti, "PlayerPickedUp"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Item: r[5],
	}
}

func newPlayerDropped(ti time.Time, r []string) Message {
	return PlayerDropped{
		Meta: newMeta(ti, "PlayerDropped"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Item: r[5],
	}
}

func newPlayerMoneyChange(ti time.Time, r []string) Message {
	return PlayerMoneyChange{
		Meta: newMeta(ti, "PlayerMoneyChange"),
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

func newPlayerBombGot(ti time.Time, r []string) Message {
	return PlayerBombGot{
		Meta: newMeta(ti, "PlayerBombGot"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func newPlayerBombPlanted(ti time.Time, r []string) Message {
	return PlayerBombPlanted{
		Meta: newMeta(ti, "PlayerBombPlanted"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func newPlayerBombDropped(ti time.Time, r []string) Message {
	return PlayerBombDropped{
		Meta: newMeta(ti, "PlayerBombDropped"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func newPlayerBombBeginDefuse(ti time.Time, r []string) Message {
	return PlayerBombBeginDefuse{
		Meta: newMeta(ti, "PlayerBombBeginDefuse"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
		Kit: !(r[5] == "out"),
	}
}

func newPlayerBombDefused(ti time.Time, r []string) Message {
	return PlayerBombDefused{
		Meta: newMeta(ti, "PlayerBombDefused"),
		Player: Player{
			Name:    r[1],
			ID:      toInt(r[2]),
			SteamID: r[3],
			Side:    r[4],
		},
	}
}

func newPlayerThrew(ti time.Time, r []string) Message {
	return PlayerThrew{
		Meta: newMeta(ti, "PlayerThrew"),
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

func newPlayerBlinded(ti time.Time, r []string) Message {
	return PlayerBlinded{
		Meta: newMeta(ti, "PlayerBlinded"),
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

func newProjectileSpawned(ti time.Time, r []string) Message {
	return ProjectileSpawned{
		Meta: newMeta(ti, "ProjectileSpawned"),
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

func newGameOver(ti time.Time, r []string) Message {
	return GameOver{
		Meta:     newMeta(ti, "GameOver"),
		Mode:     r[1],
		MapGroup: r[2],
		Map:      r[3],
		ScoreCT:  toInt(r[4]),
		ScoreT:   toInt(r[5]),
		Duration: toInt(r[6]),
	}
}

func newUnknown(ti time.Time, r []string) Message {
	return Unknown{
		Meta: newMeta(ti, "Unknown"),
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
