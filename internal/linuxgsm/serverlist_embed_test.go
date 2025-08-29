package linuxgsm

import "testing"

func TestSupportedGamesLoad(t *testing.T) {
	games, err := SupportedGames()
	if err != nil {
		t.Fatalf("SupportedGames error: %v", err)
	}
	if len(games) == 0 {
		t.Fatalf("expected games > 0")
	}
	// spot check a couple
	if !IsSupportedGame("armar") {
		t.Errorf("expected armar to be supported")
	}
	if info, ok := games["csgo"]; ok {
		if info.GameServerName == "" || info.GameName == "" {
			t.Errorf("csgo info incomplete: %+v", info)
		}
	}
}
