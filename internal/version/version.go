package version

// Version is overridable via -ldflags "-X .../internal/version.Version=x.y.z".
var Version = "0.3.1"

const Name = "nacos-cli"

// UserAgent returns Name/Version (picks up ldflags overrides).
func UserAgent() string { return Name + "/" + Version }
