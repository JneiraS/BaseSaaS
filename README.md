# BaseSasS 🚀

BaseSasS est un modèle de démarrage pour la création d'applications SaaS (Software-as-a-Service) en Go. Il fournit une base solide avec des fonctionnalités essentielles telles que l'authentification des utilisateurs, une structure de projet claire et une approche front-end basée sur des composants.

Le projet est construit avec une **Architecture Hexagonale** pour garantir une séparation claire des préoccupations entre la logique métier de base et les services externes comme la base de données ou le framework web.

## ✨ Fonctionnalités

- **Authentification Utilisateur Sécurisée** : OAuth2/OIDC pré-configuré pour la connexion, la déconnexion et la gestion de session des utilisateurs en utilisant [Zitadel](https://zitadel.com/) comme fournisseur d'identité.
- **Framework Web Moderne** : Utilise le framework web haute performance [Gin](https://gin-gonic.com/).
- **Frontend Basé sur les Composants** : Rendu HTML côté serveur avec des composants d'interface utilisateur réutilisables grâce à [Gomponents](https://www.gomponents.com/).
- **Architecture Claire** : Suit les principes de l'Architecture Hexagonale (Ports and Adapters) pour la maintenabilité et la testabilité.
- **Configuration Facile** : Configuration simplifiée à l'aide d'un fichier `.env`.
- **Gestion des Membres** : Fonctionnalités complètes pour ajouter, modifier, supprimer et lister les membres de l'association, y compris le suivi des paiements.
- **Gestion des Événements** : Création, modification, suppression et affichage des événements de l'association.
- **Gestion Financière** : Suivi des transactions (revenus et dépenses) et calcul du solde net.
- **Gestion Documentaire** : Téléchargement, téléchargement et suppression sécurisés de documents.
- **Sondages** : Création et gestion de sondages pour les membres.
- **Communication** : Envoi d'e-mails aux membres de l'association.
- **Tableau de Bord** : Vue d'ensemble des statistiques clés (membres, finances, documents).

## 🏗️ Architecture

Le projet suit une Architecture Hexagonale.

- `internal/domain/models/` : Contient les modèles métier de base (ex: `User`).
- `internal/services/` : Implémente la logique métier et les cas d'utilisation.
- `internal/adapters/handlers/` : Contient les gestionnaires HTTP (l'adaptateur "pilote") qui connectent le framework web aux services de l'application.
- `components/` : Définit les composants d'interface utilisateur réutilisables avec Gomponents.
- `templates/` : Contient les modèles HTML.
- `main.go` : Le point d'entrée de l'application, responsable de lier tous les éléments.

## 📁 Structure du Projet

- `components/`: Composants HTML réutilisables construits avec Gomponents.
- `data/`: Stockage des données non-base de données, comme les documents téléchargés.
- `internal/`: Code interne de l'application, suivant l'architecture hexagonale.
  - `adapters/`: Implémentations des adaptateurs (handlers HTTP, middleware).
  - `config/`: Gestion de la configuration de l'application.
  - `database/`: Initialisation de la base de données et migrations.
  - `domain/`: Cœur de la logique métier (modèles, interfaces de dépôts).
  - `services/`: Implémentations des services métier.
- `static/`: Fichiers statiques (CSS, JavaScript, images).
- `templates/`: Fichiers de modèles HTML pour le rendu des pages.
- `go.mod`, `go.sum`: Fichiers de gestion des dépendances Go.
- `main.go`: Point d'entrée principal de l'application.
- `.env.example`, `.gitignore`, `README.md`: Fichiers de configuration et de documentation du projet.

## 🚀 Démarrage Rapide

### Prérequis

- [Go](https://go.dev/doc/install) (version 1.23 ou plus récente)
- Une instance [Zitadel](https://zitadel.com/docs/guides/start/quickstart) en cours d'exécution (ou tout autre fournisseur OIDC).

### 1. Cloner le dépôt

```bash
git clone <url-du-depot>
cd BaseSasS
```

### 2. Configurer les Variables d'Environnement

Créez un fichier `.env` à la racine du projet en copiant le fichier d'exemple :

```bash
cp .env.example .env
```

Ouvrez ensuite le fichier `.env` et renseignez les valeurs requises :

- `OIDC_PROVIDER_URL` : L'URL de votre fournisseur OIDC (ex: `http://localhost:8080`).
- `CLIENT_ID` : L'ID client de votre fournisseur OIDC.
- `CLIENT_SECRET` : Le secret client de votre fournisseur OIDC.
- `SESSION_SECRET` : Une chaîne de caractères aléatoire pour sécuriser les sessions.
- `SMTP_HOST` : L'hôte de votre serveur SMTP (ex: `smtp.mailtrap.io`).
- `SMTP_PORT` : Le port de votre serveur SMTP (ex: `2525`).
- `SMTP_USERNAME` : Le nom d'utilisateur pour l'authentification SMTP.
- `SMTP_PASSWORD` : Le mot de passe pour l'authentification SMTP.
- `EMAIL_SENDER` : L'adresse e-mail de l'expéditeur (ex: `no-reply@yourdomain.com`).
- `DOCUMENT_STORAGE_PATH` : Le chemin où les documents téléchargés seront stockés (ex: `./data/documents`).

### 3. Installer les Dépendances

```bash
go mod tidy
```

## 🛠️ Utilisation

### Lancer le Serveur de Développement

Pour démarrer le serveur, exécutez :

```bash
go run main.go
```

L'application sera disponible à l'adresse `http://localhost:3000`.

### Compiler l'Application

Pour compiler l'application en un seul binaire :

```bash
go build -o basesass
```

Vous pouvez ensuite lancer l'application avec `./basesass`.

### Lancer les Tests

Pour exécuter tous les tests du projet :

```bash
go test ./...
```

Pour tester uniquement les paquets internes :
```bash
go test ./internal/...
```

## 📦 Dépendances Clés

- [Gin](https://github.com/gin-gonic/gin) : Framework web.
- [go-oidc](https://github.com/coreos/go-oidc) : Bibliothèque cliente OIDC.
- [Gomponents](https://github.com/maragudk/gomponents) : Génération de HTML basée sur les composants.
- [godotenv](https://github.com/joho/godotenv) : Chargement des variables d'environnement.

## 🤝 Contribution

Les contributions sont les bienvenues ! Veuillez suivre ces étapes :

1.  Fork le dépôt.
2.  Créez une branche pour votre fonctionnalité (`git checkout -b feature/AmazingFeature`).
3.  Commitez vos modifications (`git commit -m 'Add some AmazingFeature'`).
4.  Poussez vers la branche (`git push origin feature/AmazingFeature`).
5.  Ouvrez une Pull Request.

Assurez-vous que votre code respecte les conventions de style existantes et que tous les tests passent.