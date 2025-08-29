package linuxgsm

// Config contains container and image configuration for a game server
type Config struct {
	ContainerName string   // e.g., "arma-reforger"
	GameID        GameID   // e.g., "armar"
	Image         string   // default: "gameservermanagers/gameserver:<GameID>"
	Volumes       []string // docker bind mounts, e.g., ["/host/path:/data"]
	Network       string   // e.g., "host" or empty for default bridge
	RestartPolicy string   // e.g., "unless-stopped"
	Ports         []string // port mappings, e.g., ["7777:7777/udp"] ignored if Network=="host"
	EnvVars       []string // environment variables, e.g., ["TZ=UTC"]
	PullAlways    bool     // always pull image before create
}

// ImageOrDefault returns configured image or the default registry/tag based on GameID
func (c Config) ImageOrDefault() string {
	if c.Image != "" {
		return c.Image
	}
	if c.GameID == "" {
		return ""
	}
	return "gameservermanagers/gameserver:" + string(c.GameID)
}
