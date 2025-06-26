const themeSwitcher = document.getElementById('theme-switcher');
const body = document.body;
const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
const savedTheme = localStorage.getItem('theme');
if (savedTheme === 'light' || (!savedTheme && !prefersDark)) {
  body.classList.add('light-theme');
}
const themeSwitcherImg = themeSwitcher.querySelector('img');
function updateSwitcherIcon() {
  if(body.classList.contains('light-theme')){
    themeSwitcherImg.src = '/assets/switcher-noir.png';
  } else {
    themeSwitcherImg.src = '/assets/switcher.png';
  }
}
updateSwitcherIcon();
themeSwitcher.addEventListener('click', () => {
  body.classList.toggle('light-theme');
  const isLight = body.classList.contains('light-theme');
  localStorage.setItem('theme', isLight ? 'light' : 'dark');
  updateSwitcherIcon();
});
let apiBaseURL = "";
fetch('/ressources/utils/api').then(res => res.text()).then(url => { apiBaseURL = url.trim(); });
const forgotButton = document.getElementById('forgot-button');
const emailInput = document.querySelector('input[name="email"]');
const formError = document.getElementById('form-error');
const errorText = document.getElementById('error-text');
const forgotForm = document.getElementById('forgot-form');
let captchaToken = '';
window.onCaptchaSuccess = function(token){ captchaToken = token; };
function validateEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(String(email).toLowerCase());
}
function updateButtonState() {
  const emailValid = validateEmail(emailInput.value.trim());
  forgotButton.disabled = !emailValid;
}
emailInput.addEventListener('input', () => {
  updateButtonState();
  emailInput.classList.remove('input-error');
  formError.classList.remove('show');
});
function showError(message) {
  errorText.textContent = message;
  formError.classList.add('show');
  emailInput.classList.add('input-error');
  formError.scrollIntoView({ behavior: 'smooth', block: 'center' });
}
forgotForm.addEventListener('submit', async function(e) {
  e.preventDefault();
  if (!validateEmail(emailInput.value.trim())) {
    showError("Veuillez entrer une adresse email valide");
    return;
  }
  if (!captchaToken) {
    showError("Veuillez compléter le captcha");
    return;
  }
  forgotButton.disabled = true;
  forgotButton.innerHTML = '<span>Envoi...</span>';
  try {
    const response = await fetch(`${apiBaseURL}/auth/password_reset/request`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: emailInput.value.trim(), turnstile_token: captchaToken })
    });
    if (response.ok) {
      forgotButton.innerHTML = '<span>Email envoyé!</span>';
    } else {
      const data = await response.json();
      showError(data.message || 'Erreur');
      forgotButton.disabled = false;
      forgotButton.innerHTML = '<span>Envoyer</span>';
    }
    turnstile.reset('#forgot-turnstile');
    captchaToken = '';
  } catch (error) {
    showError('Erreur de connexion au serveur');
    forgotButton.disabled = false;
    forgotButton.innerHTML = '<span>Envoyer</span>';
    turnstile.reset('#forgot-turnstile');
    captchaToken = '';
  }
});
