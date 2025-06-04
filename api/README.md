# 🚀 **ToolCenter API v1**  
⏱️ **Date de dernière modification : 04/06/2025**

API **ultra‑rapide** en **Go** 🐹 pour gérer tes **utilisateurs**, tes **outils** et tout le bazar autour (auth, réservations, stats, etc.).

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
* Script **`worker.py`** : purge auto des comptes non vérifiés 🧹

---

## 🌳 Arborescence <a name="arborescence"></a>

```text
/var/www/toolcenter/api
├── config/           # config.go, db.go
├── scripts/          # Handlers REST
│   ├── auth/         # login.go, register.go
│   ├── tools/        # my_tools.go
│   └── user/         # avatar.go, me.go, profile.go, verify_email.go
├── utils/            # check.go (middlewares & helpers)
├── worker.py         # purge comptes non vérifiés
├── start.sh          # démarrage dev "go run main.go"
├── main.go           # point d'entrée
└── README.md         # ce fichier
```

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

*(d'autres routes : `/api/tools`, `/api/reservations`, `/api/moderation`, etc. — check le dossier `scripts/`)*

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

`worker.py` tourne chaque nuit (via cron ou `systemd.timer`) et **supprime** les comptes **non vérifiés** après X jours, ainsi que toutes les données liées (tokens, tools, comments…).

---

## 🗒️ Remarques <a name="remarques"></a>

* Pense à importer le **schema SQL** (`db/schema.sql`) avant premier lancement.
* Active les logs pour savoir quand ça casse (fichier `logs/api.log`).
* Contributions bienvenues : *fork → feat branch → PR* 🚀
