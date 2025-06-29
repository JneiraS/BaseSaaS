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
				gomh.Div(
					gomh.Class("ctn-btn"),
					containerButton(),
					elements.Button("Déconnexion", "btn", "/logout"),
				),
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

func containerButton() gom.Node {
	return gomh.Div(
		gomh.Class("dropdown"),
		gomh.A(gom.Text("Menu ▼")),
		gomh.Ul(
			gomh.A(gom.Text("Mon profil"), gom.Attr("href", "/profile")),
			gomh.A(gom.Text("Mes favoris"), gom.Attr("href", "/favoris")),
			gomh.A(gom.Text("Mes commandes"), gom.Attr("href", "/commandes")),
		),
	)
}
