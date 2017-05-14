package generate

import "github.com/gobuffalo/packr"

var templates packr.Box

func init() {
	templates = packr.NewBox("./templates")
}
