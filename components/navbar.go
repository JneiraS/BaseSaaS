package components

import (
	"github.com/JneiraS/BaseSasS/components/elements"
	gom "maragu.dev/gomponents"
	gomh "maragu.dev/gomponents/html"
)

func NavBar(user any) gom.Node {
	if user != nil {
		return gomh.Section(
			gomh.Div(
				gomh.Class("navbar"),
				gomh.A(
					gomh.Class("navbar-brand"),
					gom.Text("Logo"),
				),
				connButton(),
			),
		)
	}
	return gomh.Section(
		gomh.Div(
			gomh.Class("navbar"),
			gomh.A(
				gomh.Class("navbar-brand"),
				gom.Text("Logo"),
			),
			elements.Button("Connexion", "btn", "/login"),
		),
	)
}

func connButton() gom.Node {
	return gomh.Div(
		gomh.Class("ctn-btn"),
		elements.Button("Mon profil", "btn", "/profile"),
		elements.Button("Déconnexion", "btn", "/logout"),
	)
}
