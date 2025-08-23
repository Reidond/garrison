package steamcmd

type Config struct {
	BinaryPath string // path to steamcmd executable; if empty, expect in PATH
	Username   string // optional
	Password   string // optional
	LoginAnon  bool   // default true if no username provided
}

// end
