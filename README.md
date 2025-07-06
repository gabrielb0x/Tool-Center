# 🚀 **Tool Center**

![Tool Center Banner](./frontend/assets/Banniere-TC.png)

> **Tool Center** is the flagship project of **[@gabrielb0x](https://github.com/gabrielb0x)**.
> A mix of **code**, **passion** and **usefulness** made for people who want quality tools.

---

## 🌐 **Quick overview**

**Tool Center** is a web platform designed to:
- 🔧 **Create & publish** your own tools
- 💬 **Like, comment and share** other people's tools
- 👤 **Manage your account**: avatar, security, statistics
- 🔐 **Two-factor authentication** with Google Authenticator
- 🤖 Smooth 2FA prompt when signing in if your account requires it
- 🔑 **Password reset** via email link
- 🖥️ **Manage active sessions** in your security settings
- ✏️ **Update email and password** directly from the security page (2FA required when enabled)
- 🛡️ A clean **moderation system**
- ⏳ **Ban durations** and role restrictions for moderators
- 📜 **Comprehensive logs** for users and admins
- ⚡ A responsive design focused on usability

<br/>

![Preview Interface](./frontend/assets/demo-preview.png)

---

## 🧠 **Why does Tool Center exist?**

Because the world needed:
- An **open-source hub** for web tools, **without ads** and **without trackers**
- A place where **indie developers can shine**
- A **modern** and **fast** site not solely aimed at developers
- A project made **by a passionate developer** for other enthusiasts

---

## 🧱 **Project architecture**

| 🧩 Part       | ⚙️ Tech stack                        |
|--------------|--------------------------------------|
| **Backend API**   | Go (Golang) + MariaDB               |
| **Frontend**      | HTML, JS, CSS (vanilla)             |
| **Auth**          | Email with hashed tokens, UUIDv7 IDs, verification, sessions |
| **Hosting**       | Raspberry Pi 5                      |
| **Proxy / HTTPS** | Nginx + SSL via Cloudflare          |
| **Domains**       | [tool-center.fr](https://tool-center.fr) & [gabex.xyz](https://gabex.xyz) |

---

## ⚙️ **Quick configuration**

All API variables live in `api/example config.json`.
Adjust this file (ports, database, SMTP...) to match your environment.
A new `private_news_password` field secures access to private news articles.
Set `cors_allowed_origin` to control the `Access-Control-Allow-Origin` header.
Use the `storage` section to configure directories for avatars and tool images.
The `moderation` section now includes `auto_unban` to automatically lift temporary bans when expired.
The `status_banner` section controls the outage banner displayed on the frontend.
Update `frontend/src/utils/config.js` to change the API base URL used by the static pages.

### Useful API endpoints
- `POST /v{n}/admin/logs/clear` – clear all activity logs
- `GET /v{n}/admin/users/{id}/tools` – list tools of a specific user
- `GET /v{n}/admin/users/{id}/ban` – get last ban reason
- `GET /v{n}/auth/sessions` – list active sessions
- `DELETE /v{n}/auth/sessions` – revoke all other sessions
- `DELETE /v{n}/auth/sessions/{id}` – revoke a specific session
- `GET /v{n}/status` – check API health status
- `GET /v{n}/users/search?q=<name>&page=<n>` – search users by username
- `GET /v{n}/users/{username}` – public profile of a user

Example search request:

```bash
curl "https://api.tool-center.fr/v1/users/search?q=gab&page=1"
```

Example profile request:

```bash
curl https://api.tool-center.fr/v1/users/gabex
```

---

## 📸 **Gallery**

| 🔐 Login                             | 📊 Dashboard                        |
|-------------------------------------|------------------------------------|
| ![Login](./frontend/assets/login-preview.png)        | ![Dashboard](./frontend/assets/dashbord-preview.png) |

> _Screenshots taken on 2025‑05‑24. The real interface may have evolved since then._

---

## 🧙‍♂️ **The creator**

Developed by **[@gabrielb0x](https://github.com/gabrielb0x)**,
a young full‑stack developer passionate about **AI**, **cybersecurity** and **innovative projects**.
👉 **Tool Center** is his biggest and most ambitious project so far.

---

## ❤️ **Support & Contributions**

**Tool Center** is an **open‑source** project built with:
- 💕 Love
- ⏱️ Patience and determination
- ❤️‍🔥 Passion for computing

Want to contribute or report a bug?
→ **Reach me at gabex@gabex.xyz** (address may change)

---

## 🔮 **Coming soon**

- 🔄 Automatic update of posted tools
- 📊 Public statistics & user ranking
- ⚔️ Gamification and level system
- 🌍 Multilingual translations
- 🔐 More security tools and audits

---
## **© Gabriel B., 2024-2025 — All rights reserved.**
