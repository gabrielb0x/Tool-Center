async function checkApiStatus() {
  try {
    const baseUrl = (await (await fetch('/ressources/utils/api')).text()).trim();
    const data    = await (await fetch(baseUrl + '/status')).json();
    data.show_banner ? showStatusBanner(data.message, data.link)
                     : removeStatusBanner();
  } catch (err) {
    console.log('[TC LOGS] Failed to fetch API status:', err);
  }
}

function showStatusBanner(message, link) {
  let banner = document.getElementById('tc-status-banner');

  if (!banner) {
    banner = document.createElement('div');
    banner.id  = 'tc-status-banner';
    banner.style.cssText = `
      position: fixed; top: 0; left: 0; right: 0;
      background:#e63946; color:#fff; padding:10px;
      text-align:center; font-weight:600; z-index:10000;
    `;
    document.body.appendChild(banner);
  } else {
    banner.innerHTML = '';
  }

  banner.textContent = message;
  if (link) {
    const a = document.createElement('a');
    a.href  = link;
    a.textContent = 'En savoir plus';
    a.style.cssText = 'color:#fff;margin-left:10px;text-decoration:underline;';
    banner.appendChild(a);
  }

  requestAnimationFrame(applyOffset);
  window.addEventListener('resize', applyOffset);
}

function applyOffset() {
  const banner = document.getElementById('tc-status-banner');
  if (!banner) return;

  const h = banner.offsetHeight + 'px';

  document.documentElement.style.setProperty('--tc-banner-h', h);

  if (!document.getElementById('tc-banner-style')) {
    const style = document.createElement('style');
    style.id = 'tc-banner-style';
    style.textContent = `
      body            { margin-top: var(--tc-banner-h) !important; }
      .navbar         { top: var(--tc-banner-h) !important; }
      .side-menu      { top: calc(var(--tc-banner-h) + 85px) !important; }
      /* Ajoute d’autres composants fixed ici si besoin */
    `;
    document.head.appendChild(style);
  }
}

function removeStatusBanner() {
  const banner = document.getElementById('tc-status-banner');
  if (banner) banner.remove();

  document.documentElement.style.removeProperty('--tc-banner-h');
  const style = document.getElementById('tc-banner-style');
  if (style) style.remove();
}

window.addEventListener('load', checkApiStatus);
