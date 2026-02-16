package webassets

import _ "embed"

// IndexHTML is the embedded browser client page served at "/".
//
//go:embed index.html
var IndexHTML []byte
