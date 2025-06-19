# 🚀 **Tool Center**

![Tool Center Banner](./frontend/assets/Banniere-TC.png)

> **Tool Center** is the flagship project of **[@gabrielb0x](https://github.com/gabrielb0x)**.
> A mix of **code**, **passion** and **usefulness** made for people who want quality tools.

---

## 🌐 **Quick overview**

**Tool Center** is a web platform designed to:
- 🔧 **Create & publish** your own tools
- 💬 **Like, comment and share** other people's tools
- 👤 **Manage your account**: avatar, settings, statistics
- 🛡️ A clean **moderation system**
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

---
## **© Gabriel B., 2024-2025 — All rights reserved.**
