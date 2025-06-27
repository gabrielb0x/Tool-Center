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
fetch('/ressources/utils/api').then(res => res.text()).then(url => { apiBaseURL = url; });
const loginButton = document.getElementById('login-button');
const emailInput = document.querySelector('input[name="email"]');
const passwordInput = document.querySelector('input[name="password"]');
const twoFactorGroup = document.getElementById('twofactor-group');
const twoFactorInput = document.querySelector('input[name="two_factor_code"]');
const twoFactorState = document.getElementById('twofactor-state');
const twoFactorCode = document.getElementById('twofactor-code');
const twoFactorButton = document.getElementById('twofactor-button');
const formError = document.getElementById('form-error');
const errorText = document.getElementById('error-text');
const loginForm = document.getElementById('login-form');
let storedEmail = '';
let storedPassword = '';
let captchaToken = '';
window.onCaptchaSuccess = function(token){ captchaToken = token; };
const CAPTCHA_MISSING_MSG = 'Veuillez compléter le captcha';
function validateEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(String(email).toLowerCase());
}
function updateButtonState() {
  const emailValid = validateEmail(emailInput.value.trim());
  const passwordValid = passwordInput.value.trim().length >= 6;
  let twofaValid = true;
  if (twoFactorGroup.style.display !== 'none') {
    twofaValid = twoFactorInput.value.trim().length === 6;
  }
  loginButton.disabled = !(emailValid && passwordValid && twofaValid);
}
emailInput.addEventListener('input', () => {
  updateButtonState();
  emailInput.classList.remove('input-error');
  formError.classList.remove('show');
});
passwordInput.addEventListener('input', () => {
  updateButtonState();
  passwordInput.classList.remove('input-error');
  formError.classList.remove('show');
});
if(twoFactorInput){
  twoFactorInput.addEventListener('input', () => {
    updateButtonState();
    twoFactorInput.classList.remove('input-error');
    formError.classList.remove('show');
  });
}
if(twoFactorCode){
  twoFactorCode.addEventListener('input', () => {
    twoFactorButton.disabled = twoFactorCode.value.trim().length !== 6;
    twoFactorCode.classList.remove('input-error');
    formError.classList.remove('show');
  });
}
function showError(message) {
  errorText.textContent = message;
  formError.classList.add('show');
  if (message.toLowerCase().includes('email')) {
    emailInput.classList.add('input-error');
  } else if (message.toLowerCase().includes('mot de passe')) {
    passwordInput.classList.add('input-error');
  } else if (message.toLowerCase().includes('2fa')) {
    if(twoFactorState.style.display === 'flex'){
      twoFactorCode.classList.add('input-error');
    } else {
      if(twoFactorGroup.style.display==='none'){ twoFactorGroup.style.display='block'; }
      twoFactorInput.classList.add('input-error');
    }
  } else {
    emailInput.classList.add('input-error');
    passwordInput.classList.add('input-error');
  }
  formError.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function showTwoFactorState() {
  const elements = Array.from(document.querySelectorAll('.login-form, .login-options, .login-title'));
  elements.forEach(el => {
    el.style.animation = 'fadeOut 0.4s ease forwards';
  });
  setTimeout(() => {
    elements.forEach(el => {
      el.style.display = 'none';
    });
    twoFactorState.style.display = 'flex';
    twoFactorButton.disabled = twoFactorCode.value.trim().length !== 6;
    twoFactorCode.focus();
  }, 400);
}
loginForm.addEventListener('submit', async function(e) {
  e.preventDefault();
  if (!validateEmail(emailInput.value.trim())) {
    showError("Veuillez entrer une adresse email valide");
    return;
  }
  if (passwordInput.value.trim().length < 6) {
    showError("Le mot de passe doit contenir au moins 6 caractères");
    return;
  }
  if (!captchaToken) {
    showError(CAPTCHA_MISSING_MSG);
    return;
  }
  loginButton.disabled = true;
  loginButton.innerHTML = '<span>Connexion en cours...</span>';
  try {
    const response = await fetch(`${apiBaseURL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: emailInput.value.trim(),
        password: passwordInput.value.trim(),
        turnstile_token: captchaToken,
        two_factor_code: twoFactorInput.value.trim()
      })
    });
    const data = await response.json();
    if (response.ok) {
      if (data.token) {
        localStorage.setItem('token', data.token);
      }
      loginButton.innerHTML = '<span>Connexion réussie!</span>';
      setTimeout(() => {
        window.location.href = '/account/';
      }, 1000);
    } else if (data.two_factor_required) {
      storedEmail = emailInput.value.trim();
      storedPassword = passwordInput.value.trim();
      showTwoFactorState();
      loginButton.disabled = false;
      loginButton.innerHTML = '<span>Connexion</span>';
    } else {
      showError(data.message || "Email ou mot de passe incorrect");
      loginButton.disabled = false;
      loginButton.innerHTML = '<span>Connexion</span>';
    }
    turnstile.reset('#signin-turnstile');
    captchaToken = '';
  } catch (error) {
    showError("Erreur de connexion au serveur");
    loginButton.disabled = false;
    loginButton.innerHTML = '<span>Connexion</span>';
    turnstile.reset('#signin-turnstile');
    captchaToken = '';
  }
});

if(twoFactorButton){
  twoFactorButton.addEventListener('click', async () => {
    if(twoFactorCode.value.trim().length !== 6){
      showError('Veuillez entrer le code 2FA');
      return;
    }
    twoFactorButton.disabled = true;
    twoFactorButton.innerHTML = '<span>Vérification...</span>';
    try {
      const response = await fetch(`${apiBaseURL}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          email: storedEmail,
          password: storedPassword,
          turnstile_token: captchaToken,
          two_factor_code: twoFactorCode.value.trim()
        })
      });
      const data = await response.json();
      if(response.ok){
        if (data.token) {
          localStorage.setItem('token', data.token);
        }
        twoFactorButton.innerHTML = '<span>Connexion réussie!</span>';
        setTimeout(() => { window.location.href = '/account/'; }, 1000);
      } else {
        showError(data.message || 'Code 2FA invalide');
        twoFactorButton.disabled = false;
        twoFactorButton.innerHTML = '<span>Valider</span>';
      }
    } catch(err){
      showError('Erreur de connexion au serveur');
      twoFactorButton.disabled = false;
      twoFactorButton.innerHTML = '<span>Valider</span>';
    }
  });
}
