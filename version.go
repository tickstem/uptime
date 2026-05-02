package uptime

// Version is set at build time via -ldflags="-X github.com/tickstem/uptime.Version=vX.Y.Z".
// Falls back to "dev" when built without ldflags (local development).
var Version = "dev"
