package steamcmd

type Config struct {
	BinaryPath string // path to steamcmd executable; if empty, expect in PATH
	Username   string // optional
	Password   string // optional
	LoginAnon  bool   // default true if no username provided
	Retries    int    // number of retries on failure (default 3)
	Backoff    int    // base backoff milliseconds between retries (default 1000)
}

// end
