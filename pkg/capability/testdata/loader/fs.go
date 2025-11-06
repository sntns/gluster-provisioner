package loader

import "embed"

//go:embed *.yaml
//go:embed sub.folder/*.yaml
var FS embed.FS
