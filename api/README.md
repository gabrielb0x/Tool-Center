# 🚀 **ToolCenter API v1**  
⏱️ **Date de dernière modification : 05/06/2025**

API performante écrite en **Go** 🐹 pour gérer les **utilisateurs**, les **outils** et l'ensemble des services associés (authentification, réservations, statistiques, etc.).

---

## 🖼️ Vue d'ensemble

| ⚙️ Stack      | 📦 Base de données | 🔒 Auth      | 🏗️ Build                 |
| ------------- | ------------------ | ------------ | ------------------------- |
| Go 1.22 + Gin | MariaDB 10.11      | JWT + BCrypt | `go mod tidy && go build` |

---

## 📑 Sommaire

1. [Features](#features)
2. [Arborescence](#arborescence)
3. [Endpoints (extraits)](#endpoints)
4. [Installation & Configuration](#installation--configuration)
5. [Démarrage](#démarrage)
6. [Tests](#tests)
7. [Worker de nettoyage](#worker-de-nettoyage)
8. [Remarques](#remarques)

---

## ✨ Features <a name="features"></a>

* Auth email + mot de passe, BCrypt 🔐
* Tokens longue durée (30 j) stockés en DB 📜
* Rôles : *User* / *Moderator* / *Admin* 🥷
* Stats utilisateur intégrées (posts, likes, favoris…) 📊
* Système de reports & modération 🕵️
* Service **systemd** prêt à l'emploi 🚀
* Worker interne en Go : purge auto des comptes non vérifiés 🧹

---

## 🌳 Arborescence <a name="arborescence"></a>

```text
/var/www/toolcenter/api
├── config/           # config.go, db.go
├── scripts/          # Handlers REST
│   ├── auth/         # login.go, register.go, logout.go
│   ├── tools/        # submit_tool.go, my_tools.go, delete_tool.go
│   ├── user/         # avatar.go, me.go, profile.go,
│   │                 # update_username.go, update_email.go,
│   │                 # verify_email.go, delete_account.go
│   └── admin/        # stats.go, logs.go, user_list.go,
│                     # user_details.go, update_user.go,
│                     # ban_user.go, unban_user.go, user_activity.go
├── utils/            # check.go (middlewares & helpers)
├── worker/           # cleanup worker intégré
├── start.sh          # démarrage dev "go run main.go"
├── main.go           # point d'entrée
└── README.md         # ce fichier
```

## 📜 Scripts principaux

- **Auth** : `login.go`, `logout.go`, `register.go`
- **User** : `me.go`, `profile.go`, `update_username.go`, `update_email.go`, `update_password.go`, `avatar.go`, `delete_account.go`, `verify_email.go`
- **Tools** : `submit_tool.go`, `my_tools.go`, `delete_tool.go`
- **Admin** : `stats.go`, `logs.go`, `user_list.go`, `user_details.go`, `update_user.go`, `ban_user.go`, `unban_user.go`, `user_activity.go`

---

## 🔌 Endpoints (extraits) <a name="endpoints"></a>

### Auth

| Méthode | URL                  | Description                     |
| ------- | -------------------- | ------------------------------- |
| `POST`  | `/api/auth/login`    | Connexion + retour token        |
| `POST`  | `/api/auth/register` | Création de compte + mail vérif |

#### Exemple : Login

```bash
curl -X POST https://api.tool-center.fr/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"john@doe.io","password":"s3cr3t"}'
```

Réponse :

```json
{
    "success": true,
    "token": "<JWT or random 128 hex>"
}
```

### User

| Méthode | URL | Description |
| ------- | --- | ----------- |
| `GET`    | `/api/user/me`           | Profil complet |
| `POST`   | `/api/user/update_username` | Modifier le pseudo |
| `POST`   | `/api/user/update_email`    | Modifier l'e-mail |
| `POST`   | `/api/user/update_password` | Changer le mot de passe |
| `POST`   | `/api/user/avatar`          | Mettre à jour l'avatar |
| `DELETE` | `/api/user/delete`          | Supprimer le compte |

### Utilisateur ↦ `/api/user/me`

> Retourne le profil complet + stats (voir `scripts/user/me.go`).

```json
{
    "user_id": 42,
    "username": "gabex_749",
    "email": "gabriel@example.com",
    "avatar_url": null,
    "banner_url": null,
    "bio": "H4cker & music lover",
    "is_verified": true,
    "account_status": "Good",
    "created_at": "2025-05-08T12:00:00Z",
    "updated_at": "2025-05-08T12:00:00Z",
    "username_changed_at": null,
    "email_changed_at": null,
    "avatar_changed_at": null,
    "banner_changed_at": null,
    "password_changed_at": "2025-04-01T10:00:00Z",
    "last_login": "2025-05-08T12:05:00Z",
    "last_tool_posted": null,
    "last_tool_updated": null,
    "stats": {
        "tools_posted": 2,
        "comments": 5,
        "likes_given": 9,
        "likes_received": 7,
        "favorites": 4
    },
    "role": "User"
} 
```

### Publier un tool ↦ `/api/tools`

Envoi d'un tool (form‑data).

```bash
curl -X POST https://api.tool-center.fr/api/tools \
    -H "Authorization: Bearer <token>" \
    -F "title=Mon super tool" \
    -F "description=C'est trop cool" \
    -F "category=development" \
    -F "url=https://example.com" \
    -F "tags=cli,open-source" \
    -F "image=@/chemin/image.png"
```

Réponse :

```json
{
    "success": true,
    "tool_id": 123,
    "image_url": "https://tool-center.fr/tool_images/abcd.webp"
}
```

### Tools

| Méthode | URL | Description |
| ------- | --- | ----------- |
| `POST`   | `/api/tools`      | Publier un outil |
| `GET`    | `/api/tools`      | Liste des outils |
| `DELETE` | `/api/tools/:id`  | Supprimer un outil |

### Modifier son pseudo ↦ `/api/user/update_username`

`POST` avec un champ `username` (3-50 caractères). Limité à **une fois tous les 30 jours**.

```
curl -X POST https://api.tool-center.fr/api/user/update_username \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{"username":"nouveauPseudo"}'
```

### Modifier son email ↦ `/api/user/update_email`

`POST` avec `new_email` et `current_password`. Après changement l'email doit être revérifié. Limité à **une fois tous les 30 jours**.

```
curl -X POST https://api.tool-center.fr/api/user/update_email \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{"new_email":"exemple@mail.com","current_password":"monpass"}'
```

### Changer son mot de passe ↦ `/api/user/update_password`

`POST` avec `current_password` et `new_password` (7-30 caractères).

```
curl -X POST https://api.tool-center.fr/api/user/update_password \
    -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{"current_password":"ancien","new_password":"nouveauPass"}'
```

### Administration

| Méthode | URL | Description |
| ------- | --- | ----------- |
| `GET` | `/api/admin/user_list` | Liste des utilisateurs |
| `POST` | `/api/moderation/users/:id/ban` | Bannir un utilisateur |
| `POST` | `/api/moderation/users/:id/unban` | Débannir un utilisateur |

*(d'autres routes : `/api/reservations`, `/api/moderation`, etc. — voir le dossier `scripts/`)*

`duration` (en heures) peut être fourni lors du ban; `0` ou absence de valeur indique un bannissement permanent. La limite maximale est définie par `moderation.max_ban_hours` dans la configuration.

---

## 🛠️ Installation & Configuration <a name="installation--configuration"></a>

```bash
# 1. Clone
sudo mkdir -p /var/www/toolcenter && cd /var/www/toolcenter
git clone https://github.com/ton-org/toolcenter-api.git api && cd api

# 2. Dépendances Go
/usr/local/go/bin/go mod tidy

# 3. Config JSON (exemple)
cp config.sample.json config.json && nano config.json
```

### `config.json` (extrait)

```json
{
    "port": 8000,
    "gin_mode": "release",
    "version": "1.3.3",
    "URL_api": "https://api.tool-center.fr",
    "database": {
        "host": "localhost",
        "port": 3306,
        "user": "toolcenter",
        "password": "***",
        "dbname": "toolcenter"
    },
    "email": {
        "host": "ssl0.ovh.net",
        "port": 465,
        "username": "support@tool-center.fr",
        "password": "***"
    },
    "cleanup": {
        "check_interval": 600,
        "grace_period": 10
    },
    "cooldowns": {
        "email_change_days": 30,
        "username_change_days": 30,
        "tool_post_hours": 24,
        "avatar_change_hours": 24
    }
}
```

---

## 🚀 Démarrage <a name="démarrage"></a>

### Dev rapide

```bash
./start.sh            # alias pour `go run main.go`
```

### Prod : service **systemd**

```ini
[Unit]
Description=ToolCenter API
After=network.target

[Service]
User=root
WorkingDirectory=/var/www/toolcenter/api
ExecStartPre=/usr/local/go/bin/go mod tidy
ExecStartPre=/usr/local/go/bin/go build -o /var/www/toolcenter/api/toolcenter /var/www/toolcenter/api/main.go
ExecStart=/var/www/toolcenter/api/toolcenter
Restart=on-failure
StartLimitIntervalSec=60
StartLimitBurst=3
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now toolcenter-api.service
```

---

## 🔬 Tests <a name="tests"></a>

```bash
# health‑check
curl -s https://api.tool-center.fr/ping

# Récupérer mes infos
curl -H "Authorization: Bearer <token>" https://api.tool-center.fr/api/user/me | jq
```

---

## 🧹 Worker de nettoyage <a name="worker-de-nettoyage"></a>

Le worker Go embarqué tourne en continu et **supprime** les comptes **non vérifiés** après la période configurée. Il traite aussi la file d'attente d'emails.

---

## 🗒️ Remarques <a name="remarques"></a>

* Pense à importer le **schema SQL** (`db/schema.sql`) avant premier lancement.
* Active les logs pour savoir quand ça casse (fichier `logs/api.log`).
* Contributions bienvenues : *fork → feat branch → PR* 🚀
