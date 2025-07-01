package components

import (
	"github.com/JneiraS/BaseSasS/components/elements"
	gom "maragu.dev/gomponents"
	gomh "maragu.dev/gomponents/html"
)

func NavBar(user any, csrfToken string) gom.Node {
	logoElement := gomh.A(
		gomh.Class("navbar-brand"),
		gom.Text("🚀"),
	)

	if user != nil {
		return gomh.Section(
			gomh.Class("navbar"),
			logoElement,
			gomh.Div(
				gomh.Class("ctn-btn"),
				containerButton(),
				logoutForm(csrfToken),
			),
		)
	}
	return gomh.Section(
		gomh.Class("navbar"),
		logoElement,
		gomh.Div(
			gomh.Class("ctn-btn"),
			elements.Button("Connexion", "btn", "/login"),
		),
	)
}

func logoutForm(csrfToken string) gom.Node {
	return gomh.Form(
		gomh.Action("/logout"),
		gomh.Method("POST"),
		gomh.Input(gomh.Type("hidden"), gomh.Name("_csrf"), gomh.Value(csrfToken)),
		gomh.Button(gomh.Type("submit"), gom.Text("Déconnexion"), gomh.Class("btn")),
	)
}

func containerButton() gom.Node {
	return gomh.Div(
		gomh.Class("dropdown"),
		gomh.Ul(
			gomh.A(gom.Text("Mon profil"), gom.Attr("href", "/profile")),
			gomh.A(gom.Text("Mes favoris"), gom.Attr("href", "/favoris")),
			gomh.A(gom.Text("Mes commandes"), gom.Attr("href", "/commandes")),
		),
	)
}
