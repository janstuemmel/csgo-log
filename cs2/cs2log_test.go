package cs2

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

func ExampleParse() {

	var msg Message

	// a line from a server logfile
	line := `L 11/05/2018 - 15:44:36: "Player-Name<12><[U:1:29384012]><CT>" purchased "m4a1"`

	// parse Message
	msg, _ = Parse(line)

	fmt.Println(msg.GetType())
	fmt.Println(msg.GetTime().String())
	// Output:
	// PlayerPurchase
	// 2018-11-05 15:44:36 +0000 UTC
}

func ExampleToJSON() {

	// parse Message
	msg, _ := Parse(`L 11/05/2018 - 15:44:36: "Player-Name<12><[U:1:29384012]><CT>" purchased "m4a1"`)

	// cast Message interface type to PlayerPurchase type
	playerPurchase, _ := msg.(PlayerPurchase)

	fmt.Println(playerPurchase.Player.SteamID)
	fmt.Println(playerPurchase.Item)

	// get json non-html-escaped
	jsn := ToJSON(msg)

	fmt.Println(jsn)
	// Output:
	// [U:1:29384012]
	// m4a1
	// {"time":"2018-11-05T15:44:36Z","type":"PlayerPurchase","player":{"name":"Player-Name","id":12,"steam_id":"[U:1:29384012]","side":"CT"},"item":"m4a1"}
}

func TestMessages(t *testing.T) {

	t.Run("Unknown", func(t *testing.T) {

		// given
		l := line(`foo`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "Unknown", m.GetType())

		// when
		u, ok := m.(Unknown)

		// then
		assert(t, true, ok)
		assert(t, u.Raw, "foo")
	})

	t.Run("ServerMessage", func(t *testing.T) {

		// given
		l := line(`server_message: "quit"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "ServerMessage", m.GetType())

		// when
		sm, ok := m.(ServerMessage)

		// then
		assert(t, true, ok)
		assert(t, "quit", sm.Text)
	})

	t.Run("FreezTimeStart", func(t *testing.T) {

		// given
		l := line(`Starting Freeze period`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "FreezTimeStart", m.GetType())

		// when
		_, ok := m.(FreezTimeStart)

		// then
		assert(t, true, ok)
	})

	t.Run("WorldMatchStart", func(t *testing.T) {

		// given
		l := line(`World triggered "Match_Start" on "de_cache"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "WorldMatchStart", m.GetType())

		// when
		ms, ok := m.(WorldMatchStart)

		// then
		assert(t, true, ok)
		assert(t, "de_cache", ms.Map)
	})

	t.Run("WorldRoundRestart", func(t *testing.T) {

		// given
		l := line(`World triggered "Restart_Round_(1_second)`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "WorldRoundRestart", m.GetType())

		// when
		mr, ok := m.(WorldRoundRestart)

		// then
		assert(t, true, ok)
		assert(t, 1, mr.Timeleft)
	})

	t.Run("WorldRoundStart", func(t *testing.T) {

		// given
		l := line(`World triggered "Round_Start"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "WorldRoundStart", m.GetType())
	})

	t.Run("WorldRoundEnd", func(t *testing.T) {

		// given
		l := line(`World triggered "Round_End"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "WorldRoundEnd", m.GetType())
	})

	t.Run("WorldGameCommencing", func(t *testing.T) {

		// given
		l := line(`World triggered "Game_Commencing"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "WorldGameCommencing", m.GetType())
	})

	t.Run("PlayerPurchase", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" purchased "m4a1"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerPurchase", m.GetType())

		// when
		pp, ok := m.(PlayerPurchase)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pp.Player.Name)
	})

	t.Run("TeamScored TERRORIST", func(t *testing.T) {

		// given
		l := line(`Team "TERRORIST" scored "1" with "5" players`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "TeamScored", m.GetType())

		// when
		ts, ok := m.(TeamScored)

		// then
		assert(t, true, ok)
		assert(t, "TERRORIST", ts.Side)
		assert(t, 1, ts.Score)
		assert(t, 5, ts.NumPlayers)
	})

	t.Run("TeamScored CT", func(t *testing.T) {

		// given
		l := line(`Team "CT" scored "1" with "5" players`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "TeamScored", m.GetType())

		// when
		ts, ok := m.(TeamScored)

		// then
		assert(t, true, ok)
		assert(t, "CT", ts.Side)
		assert(t, 1, ts.Score)
		assert(t, 5, ts.NumPlayers)
	})

	t.Run("TeamNotice", func(t *testing.T) {

		// given
		l := line(`Team "CT" triggered "SFUI_Notice_CTs_Win" (CT "1") (T "0")`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "TeamNotice", m.GetType())

		// when
		tn, ok := m.(TeamNotice)

		// then
		assert(t, true, ok)
		assert(t, "CT", tn.Side)
		assert(t, "SFUI_Notice_CTs_Win", tn.Notice)
		assert(t, 1, tn.ScoreCT)
		assert(t, 0, tn.ScoreT)
	})

	t.Run("PlayerConnected", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><>" connected, address "foo"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerConnected", m.GetType())

		// when
		pc, ok := m.(PlayerConnected)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pc.Player.Name)
		assert(t, 12, pc.Player.ID)
		assert(t, "[U:1:29384012]", pc.Player.SteamID)
		assert(t, "foo", pc.Address)
	})

	t.Run("PlayerDisconnected", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" disconnected (reason "Kicked by Console : For killing a teammate at round start")`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerDisconnected", m.GetType())

		// when
		pd, ok := m.(PlayerDisconnected)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pd.Player.Name)
		assert(t, 12, pd.Player.ID)
		assert(t, "[U:1:29384012]", pd.Player.SteamID)
		assert(t, "Kicked by Console : For killing a teammate at round start", pd.Reason)
	})

	t.Run("PlayerEntered", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><>" entered the game`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerEntered", m.GetType())

		// when
		pe, ok := m.(PlayerEntered)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pe.Player.Name)
		assert(t, 12, pe.Player.ID)
		assert(t, "[U:1:29384012]", pe.Player.SteamID)
	})

	t.Run("PlayerSwitched", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]>" switched from team <TERRORIST> to <Spectator>`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerSwitched", m.GetType())

		// when
		ps, ok := m.(PlayerSwitched)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", ps.Player.Name)
		assert(t, 12, ps.Player.ID)
		assert(t, "[U:1:29384012]", ps.Player.SteamID)
		assert(t, "TERRORIST", ps.From)
		assert(t, "Spectator", ps.To)
	})

	t.Run("PlayerSay", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" say_team ".ready"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerSay", m.GetType())

		// when
		ps, ok := m.(PlayerSay)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", ps.Player.Name)
		assert(t, 12, ps.Player.ID)
		assert(t, "[U:1:29384012]", ps.Player.SteamID)
		assert(t, ".ready", ps.Text)
		assert(t, true, ps.Team)
	})

	t.Run("PlayerKill", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" [-225 -1829 -168] killed "Zim<20><BOT><CT>" [-476 -1709 -110] with "glock"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerKill", m.GetType())

		// when
		pk, ok := m.(PlayerKill)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pk.Attacker.Name)
		assert(t, 12, pk.Attacker.ID)
		assert(t, "[U:1:29384012]", pk.Attacker.SteamID)

		assert(t, -225, pk.AttackerPosition.X)
		assert(t, -1829, pk.AttackerPosition.Y)
		assert(t, -168, pk.AttackerPosition.Z)

		assert(t, "Zim", pk.Victim.Name)
		assert(t, 20, pk.Victim.ID)
		assert(t, "BOT", pk.Victim.SteamID)

		assert(t, -476, pk.VictimPosition.X)
		assert(t, -1709, pk.VictimPosition.Y)
		assert(t, -110, pk.VictimPosition.Z)

		assert(t, "glock", pk.Weapon)
		assert(t, false, pk.Headshot)
		assert(t, false, pk.Penetrated)
	})

	t.Run("PlayerKill Headshot Penetrated", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" [-225 -1829 -168] killed "Zim<20><BOT><CT>" [-476 -1709 -110] with "glock" (headshot penetrated)`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerKill", m.GetType())

		// when
		pk, ok := m.(PlayerKill)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pk.Attacker.Name)
		assert(t, 12, pk.Attacker.ID)
		assert(t, "[U:1:29384012]", pk.Attacker.SteamID)

		assert(t, -225, pk.AttackerPosition.X)
		assert(t, -1829, pk.AttackerPosition.Y)
		assert(t, -168, pk.AttackerPosition.Z)

		assert(t, "Zim", pk.Victim.Name)
		assert(t, 20, pk.Victim.ID)
		assert(t, "BOT", pk.Victim.SteamID)

		assert(t, -476, pk.VictimPosition.X)
		assert(t, -1709, pk.VictimPosition.Y)
		assert(t, -110, pk.VictimPosition.Z)

		assert(t, "glock", pk.Weapon)
		assert(t, true, pk.Headshot)
		assert(t, true, pk.Penetrated)
	})

	t.Run("PlayerKillAssist", func(t *testing.T) {

		// given
		l := line(`"Player-Name<10><STEAM_1:1:0101010><CT>" assisted killing "Player-Name<12><[U:1:29384012]><TERRORIST>"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerKillAssist", m.GetType())

		// when
		pk, ok := m.(PlayerKillAssist)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pk.Attacker.Name)
		assert(t, 10, pk.Attacker.ID)
		assert(t, "STEAM_1:1:0101010", pk.Attacker.SteamID)

		assert(t, "Player-Name", pk.Victim.Name)
		assert(t, 12, pk.Victim.ID)
		assert(t, "[U:1:29384012]", pk.Victim.SteamID)
	})

	t.Run("PlayerAttack", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" [480 -67 1782] attacked "Jon<9><BOT><CT>" [-134 362 1613] with "ak47" (damage "27") (damage_armor "3") (health "73") (armor "96") (hitgroup "chest")`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerAttack", m.GetType())

		// when
		pa, ok := m.(PlayerAttack)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pa.Attacker.Name)
		assert(t, 2, pa.Attacker.ID)
		assert(t, "[U:1:29384012]", pa.Attacker.SteamID)

		assert(t, 480, pa.AttackerPosition.X)
		assert(t, -67, pa.AttackerPosition.Y)
		assert(t, 1782, pa.AttackerPosition.Z)

		assert(t, "Jon", pa.Victim.Name)
		assert(t, 9, pa.Victim.ID)
		assert(t, "BOT", pa.Victim.SteamID)

		assert(t, -134, pa.VictimPosition.X)
		assert(t, 362, pa.VictimPosition.Y)
		assert(t, 1613, pa.VictimPosition.Z)

		assert(t, "ak47", pa.Weapon)
		assert(t, 27, pa.Damage)
		assert(t, 3, pa.DamageArmor)
		assert(t, 73, pa.Health)
		assert(t, 96, pa.Armor)
		assert(t, "chest", pa.Hitgroup)
	})

	t.Run("PlayerKilledBomb", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" [480 -67 1782] was killed by the bomb.`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerKilledBomb", m.GetType())

		// when
		pk, ok := m.(PlayerKilledBomb)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pk.Player.Name)
		assert(t, 2, pk.Player.ID)
		assert(t, "[U:1:29384012]", pk.Player.SteamID)

		assert(t, 480, pk.Position.X)
		assert(t, -67, pk.Position.Y)
		assert(t, 1782, pk.Position.Z)
	})

	t.Run("PlayerKilledSuicide", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" [480 -67 1782] committed suicide with "hegrenade"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerKilledSuicide", m.GetType())

		// when
		pk, ok := m.(PlayerKilledSuicide)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pk.Player.Name)
		assert(t, 2, pk.Player.ID)
		assert(t, "[U:1:29384012]", pk.Player.SteamID)

		assert(t, 480, pk.Position.X)
		assert(t, -67, pk.Position.Y)
		assert(t, 1782, pk.Position.Z)

		assert(t, "hegrenade", pk.With)
	})

	t.Run("PlayerPickedUp", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" picked up "ump45"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerPickedUp", m.GetType())

		// when
		pp, ok := m.(PlayerPickedUp)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pp.Player.Name)
		assert(t, 2, pp.Player.ID)
		assert(t, "[U:1:29384012]", pp.Player.SteamID)

		assert(t, "ump45", pp.Item)
	})

	t.Run("PlayerDropped", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" dropped "knife"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerDropped", m.GetType())

		// when
		pd, ok := m.(PlayerDropped)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pd.Player.Name)
		assert(t, 2, pd.Player.ID)
		assert(t, "[U:1:29384012]", pd.Player.SteamID)

		assert(t, "knife", pd.Item)
	})

	t.Run("PlayerMoneyChange Sub", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" money change 2050-1000 = $1050 (tracked) (purchase: item_assaultsuit)`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerMoneyChange", m.GetType())

		// when
		pm, ok := m.(PlayerMoneyChange)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pm.Player.Name)
		assert(t, 2, pm.Player.ID)
		assert(t, "[U:1:29384012]", pm.Player.SteamID)

		assert(t, 2050, pm.Equation.A)
		assert(t, -1000, pm.Equation.B)
		assert(t, 1050, pm.Equation.Result)
		assert(t, "item_assaultsuit", pm.Purchase)
	})

	t.Run("PlayerMoneyChange Add", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" money change 7700+300 = $8000 (tracked)`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerMoneyChange", m.GetType())

		// when
		pm, ok := m.(PlayerMoneyChange)

		// then
		assert(t, true, ok)

		assert(t, "Player-Name", pm.Player.Name)
		assert(t, 2, pm.Player.ID)
		assert(t, "[U:1:29384012]", pm.Player.SteamID)

		assert(t, 7700, pm.Equation.A)
		assert(t, 300, pm.Equation.B)
		assert(t, 8000, pm.Equation.Result)
	})

	t.Run("PlayerBomb TERRORIST", func(t *testing.T) {

		// given
		lines := map[string]string{
			"PlayerBombGot":     line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" triggered "Got_The_Bomb"`),
			"PlayerBombPlanted": line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" triggered "Planted_The_Bomb"`),
			"PlayerBombDropped": line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" triggered "Dropped_The_Bomb"`),
		}

		for msg, l := range lines {

			// when
			m, err := Parse(l)

			// then
			assert(t, nil, err)
			assert(t, msg, m.GetType())
		}
	})

	t.Run("PlayerBomb CT", func(t *testing.T) {

		// given
		lines := map[string]string{
			"PlayerBombBeginDefuse": line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" triggered "Begin_Bomb_Defuse_Without_Kit"`),
			"PlayerBombDefused":     line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" triggered "Defused_The_Bomb"`),
		}

		for msg, l := range lines {

			// when
			m, err := Parse(l)

			// then
			assert(t, nil, err)
			assert(t, msg, m.GetType())
		}
	})

	t.Run("PlayerBomb CT With Kit", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><CT>" triggered "Begin_Bomb_Defuse_With_Kit"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerBombBeginDefuse", m.GetType())

		// when
		pb, ok := m.(PlayerBombBeginDefuse)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pb.Player.Name)
		assert(t, 2, pb.Player.ID)
		assert(t, "[U:1:29384012]", pb.Player.SteamID)
		assert(t, "CT", pb.Player.Side)
		assert(t, true, pb.Kit)
	})

	t.Run("PlayerBomb CT Without Kit", func(t *testing.T) {

		// given
		l := line(`"Player-Name<2><[U:1:29384012]><CT>" triggered "Begin_Bomb_Defuse_Without_Kit"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerBombBeginDefuse", m.GetType())

		// when
		pb, ok := m.(PlayerBombBeginDefuse)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pb.Player.Name)
		assert(t, 2, pb.Player.ID)
		assert(t, "[U:1:29384012]", pb.Player.SteamID)
		assert(t, "CT", pb.Player.Side)
		assert(t, false, pb.Kit)
	})

	t.Run("PlayerThrew", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" threw smokegrenade [-716 -1636 -170]`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerThrew", m.GetType())

		// when
		pt, ok := m.(PlayerThrew)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pt.Player.Name)
		assert(t, 12, pt.Player.ID)
		assert(t, "[U:1:29384012]", pt.Player.SteamID)
		assert(t, "TERRORIST", pt.Player.Side)

		assert(t, "smokegrenade", pt.Grenade)
		assert(t, 0, pt.Entindex)

		assert(t, -716, pt.Position.X)
		assert(t, -1636, pt.Position.Y)
		assert(t, -170, pt.Position.Z)
	})

	t.Run("PlayerThrew Flashbang", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" threw flashbang [-716 -1636 -170] flashbang entindex 163)`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerThrew", m.GetType())

		// when
		pt, ok := m.(PlayerThrew)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pt.Player.Name)
		assert(t, 12, pt.Player.ID)
		assert(t, "[U:1:29384012]", pt.Player.SteamID)
		assert(t, "TERRORIST", pt.Player.Side)

		assert(t, "flashbang", pt.Grenade)
		assert(t, 163, pt.Entindex)

		assert(t, -716, pt.Position.X)
		assert(t, -1636, pt.Position.Y)
		assert(t, -170, pt.Position.Z)
	})

	t.Run("PlayerBlinded", func(t *testing.T) {

		// given
		l := line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" blinded for 3.45 by "Player-Name<10><STEAM_1:1:0101010><CT>" from flashbang entindex 163`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerBlinded", m.GetType())

		// when
		pb, ok := m.(PlayerBlinded)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pb.Victim.Name)
		assert(t, 12, pb.Victim.ID)
		assert(t, "[U:1:29384012]", pb.Victim.SteamID)
		assert(t, "TERRORIST", pb.Victim.Side)

		assert(t, float32(3.45), pb.For)
		assert(t, 163, pb.Entindex)

		assert(t, "Player-Name", pb.Attacker.Name)
		assert(t, 10, pb.Attacker.ID)
		assert(t, "STEAM_1:1:0101010", pb.Attacker.SteamID)
		assert(t, "CT", pb.Attacker.Side)
	})

	t.Run("ProjectileSpawned", func(t *testing.T) {

		// given
		l := line(`Molotov projectile spawned at -539.715820 -2332.986572 -100.142113, velocity -77.150497 824.855957 175.574585`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "ProjectileSpawned", m.GetType())

		// when
		ps, ok := m.(ProjectileSpawned)

		// then
		assert(t, true, ok)

		assert(t, float32(-539.715820), ps.Position.X)
		assert(t, float32(-2332.986572), ps.Position.Y)
		assert(t, float32(-100.142113), ps.Position.Z)

		assert(t, float32(-77.150497), ps.Velocity.X)
		assert(t, float32(824.855957), ps.Velocity.Y)
		assert(t, float32(175.574585), ps.Velocity.Z)
	})

	t.Run("GameOver", func(t *testing.T) {

		// given
		l := line(`Game Over: competitive mg_de_cache de_cache score 16:1 after 21 min`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "GameOver", m.GetType())

		// when
		g, ok := m.(GameOver)

		// then
		assert(t, true, ok)
		assert(t, "competitive", g.Mode)
		assert(t, "mg_de_cache", g.MapGroup)
		assert(t, "de_cache", g.Map)
		assert(t, 16, g.ScoreCT)
		assert(t, 1, g.ScoreT)
		assert(t, 21, g.Duration)
	})

	t.Run("PlayerBanned", func(t *testing.T) {

		// given
		l := line(`Banid: "Player-Name<12><[U:1:29384012]><>" was banned "for 15.00 minutes" by "Console"`)

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerBanned", m.GetType())

		// when
		pb, ok := m.(PlayerBanned)

		// then
		assert(t, true, ok)
		assert(t, "Player-Name", pb.Player.Name)
		assert(t, 12, pb.Player.ID)
		assert(t, "[U:1:29384012]", pb.Player.SteamID)
		assert(t, "for 15.00 minutes", pb.Duration)
		assert(t, "Console", pb.By)
	})
}

func TestToJSON(t *testing.T) {

	t.Run("Message to json", func(t *testing.T) {

		// given
		m, _ := Parse(line(`"Player-Name<12><[U:1:29384012]><TERRORIST>" purchased "m4a1"`))
		expected := strip(`{
				"time": "2018-11-05T15:44:36Z",
				"type": "PlayerPurchase",
				"player": {
					"name": 		"Player-Name",
					"id": 			12,
					"steam_id":		"[U:1:29384012]",
					"side": 		"TERRORIST"
				},
				"item": "m4a1"
			}`)

		// when
		jsn := strip(ToJSON(m))

		// then
		assert(t, expected, jsn)
	})

	t.Run("nil Message to json", func(t *testing.T) {

		// given
		var m Message

		// when
		jsn := ToJSON(m)

		// then
		assert(t, "null", strip(jsn))
	})

}

func TestParse(t *testing.T) {

	t.Run("time and type", func(t *testing.T) {

		// given
		l := `L 11/05/2018 - 15:44:36: "Player-Name<12><[U:1:29384012]><TERRORIST>" purchased "m4a1"`

		// when
		m, err := Parse(l)

		// then
		assert(t, nil, err)
		assert(t, "PlayerPurchase", m.GetType())
		assert(t, time.Date(2018, time.November, 5, 15, 44, 36, 0, time.UTC), m.GetTime())
	})

	t.Run("error", func(t *testing.T) {

		// given
		l := `foo`

		// when
		m, err := Parse(l)

		// then
		assert(t, ErrorNoMatch, err)
		assert(t, nil, m)
	})

	t.Run("error parse date", func(t *testing.T) {

		// given
		// day 50 out of range
		l := `L 11/50/2018 - 15:44:36: "Player-Name<12><[U:1:29384012]><TERRORIST>" purchased "m4a1"`

		// when
		m, err := Parse(l)

		// then
		assert(t, `parsing time "11/50/2018 - 15:44:36": day out of range`, err.Error())
		assert(t, nil, m)
	})

	t.Run("parse with patterns", func(t *testing.T) {

		l := `L 11/05/2018 - 15:44:36: "Player-Name<12><[U:1:29384012]><TERRORIST>" purchased "m4a1"`

		patterns := map[*regexp.Regexp]MessageFunc{
			regexp.MustCompile(PlayerPurchasePattern): NewPlayerPurchase,
		}

		// parse Message
		m, err := ParseWithPatterns(l, patterns)

		// then
		assert(t, nil, err)
		assert(t, "PlayerPurchase", m.GetType())
	})
}

func TestHelpers(t *testing.T) {

	t.Run("toInt", func(t *testing.T) {

		// when
		i1 := toInt("1337")
		i2 := toInt("hello")

		// then
		assert(t, 1337, i1)
		assert(t, 0, i2)
	})

	t.Run("toFloat", func(t *testing.T) {

		// when
		f1 := toFloat32("1337.1337")
		f2 := toFloat32("hello")

		// then
		assert(t, float32(1337.1337), f1)
		assert(t, float32(0), f2)
	})
}

func BenchmarkFirstEntry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// best case: type match steart
		Parse(line(`World triggered "Match_Start" on "de_cache"`))
	}
}

func BenchmarkMidEntry(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// worst case: type unknown
		Parse(line(`"Player-Name<2><[U:1:29384012]><TERRORIST>" [480 -67 1782] attacked "Jon<9><BOT><CT>" [-134 362 1613] with "ak47" (damage "27") (damage_armor "3") (health "73") (armor "96") (hitgroup "chest")`))
	}
}

func BenchmarkUnknown(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// worst case: type unknown
		Parse(line(`"Player-Name<12><STEAM_1:1:0101010><CT>" [-854 396 -286] does FOO BAR BAZ`))
	}
}

// helper

func line(line string) string {
	return fmt.Sprintf("L 11/05/2018 - 15:44:36: %s\n", line)
}

func assert(t *testing.T, want interface{}, have interface{}) {

	// mark as test helper function
	t.Helper()

	if want != have {
		t.Error("Assertion failed for", t.Name(), "\n\twanted:\t", want, "\n\thave:\t", have)
	}
}

func strip(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return strings.Replace(s, " ", "", -1)
}
