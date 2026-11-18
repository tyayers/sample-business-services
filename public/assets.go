package public

import "embed"

//go:embed index.html styles.css app.js favicon.svg
var Assets embed.FS
