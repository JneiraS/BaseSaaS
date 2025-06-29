package components

import (
	"github.com/JneiraS/BaseSasS/components/elements"
	gom "maragu.dev/gomponents"
	gomh "maragu.dev/gomponents/html"
)

func NavBar() gom.Node {
	return gomh.Section(
		gomh.Div(
			gomh.Class("navbar"),
			gomh.A(
				gomh.Class("navbar-brand"),
				gom.Text("Logo"),
			),
			elements.Button("Déconnexion", "btn", "/logout"),
		),
	)
}
