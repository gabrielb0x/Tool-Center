async function fetchUserInfo() {
    const token = localStorage.getItem('token');
    const baseUrl = window.API_BASE_URL;
    if (!token || token === 'undefined' || token === 'null') return;
    try {
      const response = await fetch(baseUrl + '/user/me', {
        method: 'GET',
        headers: { 'Authorization': 'Bearer ' + token }
      });
      const data = await response.json();
      if (data.success && data.user) {
        const accountIcon = document.querySelector('.account-icon img');
        if (data.user.avatar_url) {
          accountIcon.src = data.user.avatar_url;
        }
        document.querySelector('.auth-buttons').style.display = 'none';

        if (data.user.role && data.user.role === 'Admin') {
          if (!document.getElementById('admin-panel-btn')) {
            const navRight = document.querySelector('.nav-right');
            const adminBtn = document.createElement('a');
            adminBtn.href = '/admin/';
            adminBtn.id = 'admin-panel-btn';
            adminBtn.className = 'btn btn-admin';
            adminBtn.style.marginBottom = '5px';
            adminBtn.textContent = 'Admin Panel';
            navRight.insertBefore(adminBtn, navRight.children[1]);
          }
        }
      }
    } catch (error) {
      console.log('[TC LOGS] Error when fetching acc :', error);
    }
  }
  fetchUserInfo();
  function toggleNav() {
    const menu = document.getElementById("side-menu");
    const menuIcon = document.getElementById("menu-icon");
    if (menu.classList.contains("open")) {
      menu.classList.remove("open");
      menuIcon.classList.remove("open");
    } else {
      menu.classList.add("open");
      menuIcon.classList.add("open");
    }
  }

  // Expose functions used by inline handlers when scripts are loaded as modules
  window.toggleNav = toggleNav;
  const themeSwitcher = document.getElementById('theme-switcher');
  const body = document.body;
  if (localStorage.getItem('theme') === 'light') {
    body.classList.add('light-theme');
  }
  themeSwitcher.addEventListener('click', () => {
    body.classList.toggle('light-theme');
    localStorage.setItem('theme', body.classList.contains('light-theme') ? 'light' : 'dark');
  });
  function revealFeatures() {
    const features = document.querySelectorAll('.feature-item');
    for (let i = 0; i < features.length; i++) {
      const windowHeight = window.innerHeight;
      const elementTop = features[i].getBoundingClientRect().top;
      const revealPoint = 150;
      if(elementTop < windowHeight - revealPoint) {
        features[i].classList.add('active');
      } else {
        features[i].classList.remove('active');
      }
    }
  }
  window.addEventListener('scroll', () => {
    revealFeatures();
    if (window.pageYOffset > 300) {
      document.getElementById('back-to-top').style.display = 'flex';
    } else {
      document.getElementById('back-to-top').style.display = 'none';
    }
  });
  setTimeout(() => {
    const preloader = document.getElementById('preloader');
    if(preloader && getComputedStyle(preloader).opacity !== '0') {
      document.getElementById('preloader-message').style.opacity = '1';
    }
  }, 1500);
  window.addEventListener('load', () => {
    const preloader = document.getElementById('preloader');
    preloader.style.opacity = '0';
    setTimeout(() => {
      preloader.style.display = 'none';
    }, 500);
  });
  function scrollToTop() {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
  function toggleFaq(item) {
    item.classList.toggle('open');
  }

  // Export additional functions for inline usage
  window.scrollToTop = scrollToTop;
  window.toggleFaq = toggleFaq;
