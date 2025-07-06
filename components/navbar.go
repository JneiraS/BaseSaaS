package components

import (
	"log"

	"github.com/JneiraS/BaseSasS/components/elements"
	"github.com/gin-contrib/sessions"
	gom "maragu.dev/gomponents"
	gomh "maragu.dev/gomponents/html"
)

func NavBar(user any, csrfToken string, session sessions.Session) gom.Node {
	log.Printf("DEBUG: NavBar called. User: %v", user != nil)
	logoElement := gomh.A(
		gomh.Class("navbar-brand"),
		gom.Text("🚀"),
	)

	themeSwitcher := gomh.I(gomh.Class("fa-solid fa-lightbulb"), gom.Attr("id", "theme-switcher"))

	// Contenu du div ctn-btn
	var ctnBtnContent []gom.Node
	if user != nil {
		ctnBtnContent = []gom.Node{
			containerButton(),
			logoutForm(csrfToken),
			themeSwitcher,
		}
	} else {
		ctnBtnContent = []gom.Node{
			elements.Button("Connexion", "btn", "/login"),
			themeSwitcher,
		}
	}

	// Construire les arguments pour gomh.Div
	var divArgs []gom.Node
	divArgs = append(divArgs, gomh.Class("ctn-btn"))
	divArgs = append(divArgs, ctnBtnContent...)

	flashNodes := FlashMessages(session)
	log.Printf("DEBUG: FlashMessages generated.")

	return gomh.Section(
		gomh.Class("navbar"),
		logoElement,
		gomh.Div(divArgs...),
		flashNodes,
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
			gomh.A(gom.Text("Mes membres"), gom.Attr("href", "/members")),
			gomh.A(gom.Text("Mes événements"), gom.Attr("href", "/events")),
			gomh.A(gom.Text("Communication"), gom.Attr("href", "/communication/email")),
			gomh.A(gom.Text("Finance"), gom.Attr("href", "/finance/transactions")),
			gomh.A(gom.Text("Documents"), gom.Attr("href", "/documents")),
			gomh.A(gom.Text("Mes favoris"), gom.Attr("href", "/favoris")),
			gomh.A(gom.Text("Mes commandes"), gom.Attr("href", "/commandes")),
		),
	)
}
func FlashMessages(session sessions.Session) gom.Node {
	log.Printf("DEBUG: FlashMessages function called.")
	successFlashes := session.Flashes("success")
	errorFlashes := session.Flashes("error")
	warningFlashes := session.Flashes("warning")

	log.Printf("DEBUG: Flashes - Success: %v, Error: %v, Warning: %v", successFlashes, errorFlashes, warningFlashes)

	// Sauvegarder la session après avoir lu les flashs
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session après lecture des flashs dans FlashMessages: %v", err)
	}

	return gomh.Div(
		gomh.Class("flash-message-container"),
		gom.Group(
			gom.Map(successFlashes, func(success interface{}) gom.Node {
				return gomh.Div(
					gomh.Class("flash-message success"),
					gom.Text(success.(string)),
				)
			}),
		),
		gom.Group(
			gom.Map(errorFlashes, func(err interface{}) gom.Node {
				return gomh.Div(
					gomh.Class("flash-message error"),
					gom.Text(err.(string)),
				)
			}),
		),
		gom.Group(
			gom.Map(warningFlashes, func(warning interface{}) gom.Node {
				return gomh.Div(
					gomh.Class("flash-message warning"),
					gom.Text(warning.(string)),
				)
			}),
		),
	)
}
