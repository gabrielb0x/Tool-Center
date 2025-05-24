class TcNavbar extends HTMLElement {
  constructor() {
    super();
    const shadow = this.attachShadow({ mode: 'open' });

    // Mets tout le CSS bien stylé ici
    shadow.innerHTML = /*html*/`
      <style>
        :host { all: initial; }
        *, *::before, *::after { box-sizing: border-box; font-family: 'Poppins', Arial, sans-serif; }
        .navbar {
            width: 100%;
            background: rgba(0,0,0,0.4);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            transition: background 0.3s ease;
            padding: 10px 0;
            position: fixed;
            top: 0;
            left: 0;
            z-index: 999;
            border-bottom: 1px solid rgb(115,115,115);
        }
        .container {
            display: flex;
            justify-content: space-between;
            align-items: center;
            width: 100%;
            padding: 0 40px;
        }
        .logo-link {
            display: flex;
            align-items: center;
            gap: 10px;
            text-decoration: none;
            padding: 10px 15px;
            border-radius: 20px;
            transition: background-color 0.3s ease;
        }
        .logo-link:hover { background-color: #1a1a1a; }
        .logo {
            width: 45px;
            height: auto;
        }
        .logo-text {
            font-weight: bold;
            color: white;
        }
        .nav-left {
            display: flex;
            align-items: center;
            gap: 15px;
        }
        .menu-icon {
            cursor: pointer;
            width: 30px;
            height: 30px;
            background: url('/assets/menu.png') no-repeat center;
            background-size: contain;
            transition: transform 0.3s ease, background-image 0.3s ease;
            display:inline-block;
        }
        .menu-icon.open {
            transform: rotate(180deg);
            background: url('/assets/croix.png') no-repeat center;
            background-size: contain;
        }
        .nav-right {
            display: flex;
            align-items: center;
        }
        .theme-switcher {
            background: none;
            border: none;
            cursor: pointer;
            margin-right: 15px;
        }
        .theme-switcher img {
            width: 30px;
            height: 30px;
            transition: 0.3s ease;
        }
        .auth-buttons {
            display: flex;
            gap: 10px;
        }
        .auth-buttons .btn {
            background: linear-gradient(135deg, #2a00d9, #3000FF);
            color: white;
            padding: 8px 16px;
            text-decoration: none;
            border-radius: 9999px;
            transition: background 0.3s ease, transform 0.3s ease;
        }
        .auth-buttons .btn:hover { transform: translateY(-5px);}
        .auth-buttons .btn-signup {
            background: linear-gradient(135deg, #777, #808080);
        }
        .auth-buttons .btn-signup:hover { transform: translateY(-5px);}
        .account-icon img {
            width: 35px;
            height: 35px;
            object-fit: cover;
            border-radius: 50%;
            cursor: pointer;
        }
        .account-icon {
            margin-left: 15px;
        }
        .account-icon:hover img {
            filter: brightness(0.8);
            transition: filter 0.3s ease;
        }
        .side-menu {
            height: 100%;
            width: 250px;
            position: fixed;
            z-index: 1000;
            top: 85px;
            left: -250px;
            background: linear-gradient(145deg, rgba(0,0,0,0.4), rgba(0,0,0,0.5));
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            overflow-x: hidden;
            transition: left 0.5s;
            padding-top: 20px;
            border-right: 1px solid rgba(115, 115, 115, 1);
            box-shadow: 2px 0 15px rgba(0,0,0,0.4);
        }
        .side-menu.open { left: 0; }
        .side-menu a {
            padding: 12px 8px 12px 32px;
            text-decoration: none;
            font-size: 22px;
            color: #fff;
            display: block;
            transition: 0.3s;
            margin: 5px 0;
            border-radius: 5px;
        }
        .side-menu a:hover {
            background: rgba(42,0,217,0.3);
            color: #2a00d9;
        }
        /* Theme light */
        :host([data-theme="light"]) .navbar {
            background: rgba(255,255,255,0.4);
            border-bottom: 1px solid rgba(74, 74, 74, 0.3);
        }
        :host([data-theme="light"]) .logo-text { color: #000;}
        :host([data-theme="light"]) .logo-link:hover { background-color: #d1d1d1;}
        :host([data-theme="light"]) .side-menu {
            background: rgba(255,255,255,0.4) !important;
            border-right: 1px solid rgba(74,74,74,0.3) !important;
            box-shadow: 2px 0 15px rgba(0,0,0,0.1);
        }
        :host([data-theme="light"]) .side-menu a { color: #333;}
        :host([data-theme="light"]) .side-menu a:hover { color: #2a00d9; background: rgba(42,0,217,0.15);}
        :host([data-theme="light"]) .theme-switcher img { content: url('/assets/switcher-noir.png'); }
        /* Responsive */
        @media (max-width: 768px) {
            .auth-buttons { display: none !important; }
            .container { padding: 0 10px;}
        }
      </style>
      <header>
        <nav class="navbar">
          <div class="container">
            <div class="nav-left">
              <span class="menu-icon" id="menu-icon"></span>
              <a href="https://tool-center.fr" class="logo-link">
                <img src="/assets/tc_logo.webp" alt="ToolCenter Logo" class="logo">
                <span class="logo-text">ToolCenter (BETA)</span>
              </a>
            </div>
            <div class="nav-right">
              <button class="theme-switcher" id="theme-switcher">
                <img src="/assets/switcher.png" alt="Switcher Icon">
              </button>
              <div class="auth-buttons">
                <a href="https://tool-center.fr/signin" class="btn btn-signin">Sign In</a>
                <a href="https://tool-center.fr/signup" class="btn btn-signup">Sign Up</a>
              </div>
              <a href="https://tool-center.fr/account" class="account-icon">
                <img src="/assets/account.png" alt="Account Icon">
              </a>
            </div>
          </div>
        </nav>
      </header>
      <div id="side-menu" class="side-menu">
        <a href="https://tool-center.fr">Home</a>
        <a href="https://tool-center.fr/tools">Tools</a>
        <a href="https://tool-center.fr/account">Account</a>
        <a href="https://tool-center.fr/about">About</a>
        <a href="https://tool-center.fr/legals">Légals</a>
      </div>
    `;
  }

  connectedCallback() {
    const $ = s => this.shadowRoot.querySelector(s);

    // --- Menu burger ---
    $('#menu-icon').addEventListener('click', () => {
      $('#side-menu').classList.toggle('open');
      $('#menu-icon').classList.toggle('open');
    });

    // --- Thème ---
    const body = document.body;
    const img = $('#theme-switcher img');
    const updateThemeAttr = () => {
      const light = body.classList.contains('light-theme');
      this.setAttribute('data-theme', light ? 'light' : 'dark');
      img.src = light ? '/assets/switcher-noir.png' : '/assets/switcher.png';
    };
    // init
    if (localStorage.getItem('theme') === 'light') body.classList.add('light-theme');
    updateThemeAttr();

    $('#theme-switcher').addEventListener('click', () => {
      body.classList.toggle('light-theme');
      localStorage.setItem('theme', body.classList.contains('light-theme') ? 'light' : 'dark');
      updateThemeAttr();
    });

    // --- Avatar user ---
    this.#fetchUserInfo().catch(console.error);

    // --- Décaler le contenu sous la navbar (enlève l’overlap) ---
    queueMicrotask(() => {
      const navH = $('nav.navbar').offsetHeight;
      document.body.style.paddingTop = navH + 'px';
    });
  }

  // Méthode privée pour récupérer info user et maj l’avatar
  async #fetchUserInfo() {
    const token = localStorage.getItem('token');
    if (!token || token === 'undefined' || token === 'null') return;
    // récupère l'URL base API (fichier texte côté serveur)
    let baseUrl;
    try {
      baseUrl = (await (await fetch('/ressources/utils/api')).text()).trim();
    } catch (e) { return; }
    try {
      const res = await fetch(baseUrl + '/user/me', {
        headers: { Authorization: 'Bearer ' + token }
      });
      const { success, user } = await res.json();
      if (success && user) {
        if (user.avatar_url) this.shadowRoot.querySelector('.account-icon img').src = user.avatar_url;
        this.shadowRoot.querySelector('.auth-buttons').style.display = 'none';
      }
    } catch (err) {
      console.log('[TC-NAV] fetch user failed :', err);
    }
  }
}

customElements.define('tc-navbar', TcNavbar);
