package linuxgsm

import (
	"bufio"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"sync"
)

//go:embed serverlist.csv
var serverListFS embed.FS

var (
	serverListOnce sync.Once
	serverListErr  error
	supportedGames map[GameID]ServerInfo
)

// SupportedGames returns a copy of the supported games map keyed by GameID
func SupportedGames() (map[GameID]ServerInfo, error) {
	serverListOnce.Do(func() {
		supportedGames = make(map[GameID]ServerInfo)
		f, err := serverListFS.Open("serverlist.csv")
		if err != nil {
			serverListErr = fmt.Errorf("open serverlist: %w", err)
			return
		}
		defer f.Close()

		// Use a bufio.Reader to handle possible BOM
		r := csv.NewReader(bufio.NewReader(f))
		r.FieldsPerRecord = -1
		first := true
		for {
			rec, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				serverListErr = fmt.Errorf("read serverlist: %w", err)
				return
			}
			if first {
				first = false
				continue // skip header
			}
			if len(rec) < 4 {
				// ignore short lines
				continue
			}
			short := strings.TrimSpace(rec[0])
			gsn := strings.TrimSpace(rec[1])
			gname := strings.TrimSpace(rec[2])
			os := strings.TrimSpace(rec[3])
			if short == "" || gsn == "" {
				continue
			}
			supportedGames[GameID(short)] = ServerInfo{
				ShortName:      short,
				GameServerName: gsn,
				GameName:       gname,
				OS:             os,
			}
		}
	})
	if serverListErr != nil {
		return nil, serverListErr
	}
	// Return a copy to avoid external mutation
	out := make(map[GameID]ServerInfo, len(supportedGames))
	for k, v := range supportedGames {
		out[k] = v
	}
	return out, nil
}

// IsSupportedGame returns true when id exists in serverlist.csv
func IsSupportedGame(id GameID) bool {
	games, err := SupportedGames()
	if err != nil {
		return false
	}
	_, ok := games[id]
	return ok
}

// ServerBinaryName returns the LinuxGSM script name, e.g., armarserver
func ServerBinaryName(id GameID) (string, error) {
	games, err := SupportedGames()
	if err != nil {
		return "", err
	}
	info, ok := games[id]
	if !ok {
		return "", ErrUnsupportedGame
	}
	return info.GameServerName, nil
}
