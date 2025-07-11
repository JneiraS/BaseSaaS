package elements

import (
	gom "maragu.dev/gomponents"
	gomh "maragu.dev/gomponents/html"
)

func Button(text, class, href string, children ...gom.Node) gom.Node {
	return gomh.A(gomh.Class(class),
		gom.Text(text),
		gomh.Href(href),
		gom.Group(children),
	)
}
