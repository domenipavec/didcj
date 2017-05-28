package templates

import "github.com/gobuffalo/packr"

var Box packr.Box

func init() {
	Box = packr.NewBox("./templates")
}
