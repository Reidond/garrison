package steamcmd

type Config struct {
	BinaryPath string // path to steamcmd executable; if empty, expect in PATH
	Username   string // optional
	Password   string // optional
	LoginAnon  bool   // default true if no username provided
	Retries    int    // number of retries on failure (default 3)
	Backoff    int    // base backoff milliseconds between retries (default 1000)
	// OnOutput, if set, is called for each output line from steamcmd. Useful for tee-style logging in tests.
	OnOutput     func(line string)
	OutputPrefix string // optional prefix added before each output line passed to OnOutput
}

// end
