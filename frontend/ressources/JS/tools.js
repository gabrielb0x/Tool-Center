async function fetchUserInfo() {
    const token = localStorage.getItem('token');
    const baseUrl = window.API_BASE_URL;
    if (!token || token === 'undefined' || token === 'null') return;
    try {
      const response = await fetch(baseUrl + '/get_acc_info.php', {
        method: 'GET',
        headers: { 'Authorization': 'Bearer ' + token }
      });
      const data = await response.json();
      if (data.success) {
        const accountIcon = document.querySelector('.account-icon img');
        if (data.user_info.profile_picture) {
          accountIcon.src = data.user_info.profile_picture;
          document.querySelector('.auth-buttons').style.display = 'none';
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

  // Make function accessible globally when using ES modules
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
  window.addEventListener('scroll', () => {
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

  // Export for inline event handler usage
  window.scrollToTop = scrollToTop;

  // Filtrage des outils
  const filterItems = document.querySelectorAll('.filter-item');
  filterItems.forEach(item => {
    item.addEventListener('click', () => {
      filterItems.forEach(i => i.classList.remove('active'));
      item.classList.add('active');
      
      // Ici vous ajouteriez la logique pour filtrer les outils
      // Par exemple, faire une requête AJAX ou filtrer en local
      console.log(`Filtrer par: ${item.textContent}`);
    });
  });

  // Recherche d'outils
  const searchInput = document.querySelector('.search-input');
  const searchBtn = document.querySelector('.search-btn');
  
  searchBtn.addEventListener('click', performSearch);
  searchInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') performSearch();
  });
  
  function performSearch() {
    const query = searchInput.value.trim();
    if (query) {
      console.log(`Recherche de: ${query}`);
      // Ici vous ajouteriez la logique pour rechercher les outils
    }
  }

