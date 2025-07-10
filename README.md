# 🚀✨ **Tool Center**

![Banner](./frontend/public/assets/Banniere-TC.png)

<p align="center">
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/badge/License-MIT-green.svg"></a>
  <a href="https://github.com/gabrielb0x/tool-center/actions"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/gabrielb0x/tool-center/ci.yml?label=CI&logo=github"></a>
  <a href="https://github.com/gabrielb0x/tool-center/releases"><img alt="GitHub release" src="https://img.shields.io/github/v/release/gabrielb0x/tool-center?include_prereleases&sort=semver&color=brightgreen"></a>
  <a href="https://github.com/gabrielb0x/tool-center/stargazers"><img alt="Stars" src="https://img.shields.io/github/stars/gabrielb0x/tool-center?style=social"></a>
  <img alt="Lines of code" src="https://img.shields.io/tokei/lines/github/gabrielb0x/tool-center?color=blueviolet">
  <img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/gabrielb0x/tool-center?logo=go&logoColor=white">
  <img alt="Node Version" src="https://img.shields.io/node/v/rollup?color=orange&label=node">
</p>

> **Tool Center** is the 🏡 **tracker‑free** & **ad‑free** playground where indie devs can **🚀 create**, **📦 ship** & **🔍 discover** tiny web tools ‑ all running happily on a Raspberry Pi 5.

---

## 📑 Table of Contents

* [🧠 About](#about)
* [✨ Features](#features)
* [🚀 Live Demo](#live-demo)
* [⚡ Quick Start](#quick-start)
* [🧱 Project Structure](#project-structure)
* [🔧 Configuration](#configuration)
* [🌍 Deployment](#deployment)
* [📡 API Overview](#api-overview)
* [🧪 Testing](#testing)
* [🤝 Contributing](#contributing)
* [👥 Community](#community)
* [🔮 Roadmap](#roadmap)
* [📜 Changelog](#changelog)
* [📝 License](#license)

---

## 🧠 About

<a id="about"></a>

**Tool Center** started as the late‑night idea of [@gabrielb0x](https://github.com/gabrielb0x) and quickly snowballed into a community‑driven platform. Mission 🔭: offer a **⚡ fast**, **🔒 privacy‑first** and **✨ fun** alternative to ad‑infested “online tool” sites.

* 🏎️ **Runs** on a Raspberry Pi 5 (ARM64)
* 🕵️ **Zero trackers** – your data 👉 *yours*
* 🛠️ **One‑click publish** workflow (UI & REST API)
* 🛡️ **Transparent moderation** – public JSON audit log

> **Status:** *Beta* – solid, but you might still find sharp edges.

---

## ✨ Features

<a id="features"></a>

| 📂 Area            | 🌟 Highlights                                                                       |
| ------------------ | ----------------------------------------------------------------------------------- |
| 👤 **Accounts**    | Sign‑up / login, avatar upload, profile stats, social links                         |
| 🔐 **Security**    | Password reset, **TOTP 2FA**, brute‑force shield, rate‑limit, anti‑spam             |
| 🛠️ **Tools**      | Create, edit, publish, like, comment, share, **versioning** *(soon)*                |
| 🛡️ **Moderation** | Role‑based perms, temp/perma bans, auto‑unban, sanction appeals, exportable logs    |
| 🎨 **UX**          | Responsive layout, dark‑mode, keyboard shortcuts, accessible components, PWA splash |
| ⚙️ **DevOps**      | OpenAPI 3 docs, GitHub Actions, Docker/Compose, semantic releases, Dependabot       |
| 📈 **Analytics**   | Self‑hosted [Plausible](https://plausible.io/) (💡 opt‑in)                          |

---

## 🚀 Live Demo

<a id="live-demo"></a>

👉 **[https://tool-center.fr](https://tool-center.fr)** — come poke it!

|                       🔐 Log in                      |                         📊 Dashboard                        |
| :--------------------------------------------------: | :---------------------------------------------------------: |
| ![Login](./frontend/public/assets/login-preview.png) | ![Dashboard](./frontend/public/assets/dashbord-preview.png) |

*UI snapshots from **2025‑05‑24**.*

---

## ⚡ Quick Start

<a id="quick-start"></a>

### 🐳 Docker (recommended)

```bash
# 1 › clone & cd
git clone https://github.com/gabrielb0x/tool-center.git && cd tool-center

# 2 › config
cp api/example\ config.json api/config.json
cp deploy/.env.example deploy/.env

# 3 › run 🚀
docker compose up -d --build
```

Stacks included 🧩: `api` (Go), `db` (MariaDB 11) & `frontend` (Nginx).

### 🛠️ Manual

```bash
# API (Go 1.22+)
cd api && go mod tidy && go run .

# Front (Node 20+)
cd ../frontend && npm i && npm run build  # → ./dist
```

---

## 🧱 Project Structure

<a id="project-structure"></a>

```text
📦 tool-center
 ┣ api/            # Go source, config, mail templates
 ┣ frontend/       # Vanilla JS + Vite static site
 ┣ deploy/         # Docker, Nginx, systemd, k8s (WIP)
 ┣ docs/           # Diagrams, ADRs, threat‑model
 ┣ scripts/        # Helpers & seeders
 ┗ tests/          # Go + JS suites
```

---

## 🔧 Configuration

<a id="configuration"></a>

All settings live in **`api/config.json`** 🔒 (`api/config.secrets.json` overrides 🔑).

| 🗝️ Key               | 📓 Description | 🧩 Example                                 |
| --------------------- | -------------- | ------------------------------------------ |
| `port`                | API port       | `8080`                                     |
| `database.dsn`        | MariaDB DSN    | `user:pass@tcp(db:3306)/toolcenter`        |
| `smtp`                | Mail server    | `{ "host":"smtp.gmx.net", "port":587, … }` |
| `cors_allowed_origin` | CORS origins   | `"https://tool-center.fr"`                 |
| `rate_limit.limit`    | Requests / IP  | `200`                                      |
| `status_banner`       | UI banner      | "Maintenance 22:00‑23:00 UTC"              |

---

## 🌍 Deployment

<a id="deployment"></a>

| 🌐 Where           | ⚙️ How                                        | 📝 Notes                             |
| ------------------ | --------------------------------------------- | ------------------------------------ |
| **Raspberry Pi 5** | Systemd (`deploy/systemd/`)                   | Needs < 1 GB RAM                     |
| **Cloudflare**     | Free SSL + WAF                                | orange‑cloud CNAME + *Full (strict)* |
| **Docker Hub**     | `docker pull gabrielb0x/tool-center` *(soon)* | Version‑tagged images                |
| **Kubernetes**     | Helm chart *(WIP)*                            | autoscale & GitOps ready             |
| **Backup**         | `scripts/backup.sh`                           | mysqldump + rclone                   |

---

## 📡 API Overview

<a id="api-overview"></a>

RESTful JSON, versioned (`/v1`). Docs auto‑generated:

* **Swagger UI** → `/docs/swagger/`
* **OpenAPI 3** JSON → `/docs/openapi.json`

Example 🔑:

```bash
curl -X POST https://api.tool-center.fr/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@site.com","password":"hunter2"}'
```

---

## 🧪 Testing

<a id="testing"></a>

```bash
# Go unit tests
cd api && go test ./...

# Frontend lint + tests
cd ../frontend && npm run lint && npm test
```

CI (GitHub Actions) runs both.

---

## 🤝 Contributing

<a id="contributing"></a>

1. **Fork** & `git checkout -b feat/cool dev` 🛠️
2. **Commit**: `type(scope): subject` ✅ (Conventional Commits)
3. **Lint / test** before PR 🔍
4. **Open PR** & fill template ✍️

First‑timer? Check **good first issue** label.

---

## 👥 Community

<a id="community"></a>

* 💬 **Discord** → [https://discord.gg/toolcenter](https://discord.gg/toolcenter)
* 🐦 **Twitter/X** → [https://x.com/toolcenter](https://x.com/toolcenter)
* 📝 **Blog** → *soon™*

---

## 🔮 Roadmap

<a id="roadmap"></a>

* [ ] 🔄 Auto‑update tools (webhooks)
* [ ] 📊 Public stats & leaderboard
* [ ] ⚔️ Gamification (XP, badges)
* [ ] 🌐 Multi‑language UI (i18n)
* [ ] 🐳 Docker Hub image
* [ ] 🔐 Security audit guide
* [ ] 🦾 AI‑powered code snippets
* [ ] 📱 Installable PWA

Vote / suggest in **Discussions**!

---

## 📜 Changelog

<a id="changelog"></a>
See **CHANGELOG.md** for semantic‑versioned release notes.

---

## 📝 License

<a id="license"></a>

© **2024‑2025 Gabriel B.** — Released under the **[MIT License](LICENSE)**.

---

> *Made with 🩶, insomnia & way too many cups of coffee.*
