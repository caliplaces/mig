// Package assets embeds the PNG image library shipped with the binary.
package assets

import "embed"

//go:embed images/*.png
var FS embed.FS
