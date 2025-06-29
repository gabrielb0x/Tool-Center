async function checkApiStatus() {
  try {
    const baseUrlResponse = await fetch('/ressources/utils/api');
    const baseUrl = (await baseUrlResponse.text()).trim();
    const res = await fetch(baseUrl + '/status');
    const data = await res.json();
    if (data.show_banner) {
      showStatusBanner(data.message, data.link);
    }
  } catch (err) {
    console.log('[TC LOGS] Failed to fetch API status:', err);
  }
}

function showStatusBanner(message, link) {
  let banner = document.getElementById('tc-status-banner');
  if (!banner) {
    banner = document.createElement('div');
    banner.id = 'tc-status-banner';
    banner.style.position = 'fixed';
    banner.style.top = '0';
    banner.style.left = '0';
    banner.style.right = '0';
    banner.style.zIndex = '1000';
    banner.style.background = '#e63946';
    banner.style.color = '#fff';
    banner.style.padding = '10px';
    banner.style.textAlign = 'center';
    banner.style.fontWeight = '600';
    document.body.prepend(banner);
  } else {
    banner.innerHTML = '';
  }
  banner.textContent = message;
  if (link) {
    const a = document.createElement('a');
    a.href = link;
    a.textContent = 'En savoir plus';
    a.style.color = '#fff';
    a.style.marginLeft = '10px';
    a.style.textDecoration = 'underline';
    banner.appendChild(a);
  }
}

window.addEventListener('load', checkApiStatus);
