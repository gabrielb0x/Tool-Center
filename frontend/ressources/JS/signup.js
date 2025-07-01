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
let apiBaseURL = window.API_BASE_URL;
const signupButton = document.getElementById('signup-button');
const usernameInput = document.querySelector('input[name="username"]');
const emailInput = document.querySelector('input[name="email"]');
const passwordInput = document.querySelector('input[name="password"]');
const confirmPasswordInput = document.querySelector('input[name="confirm_password"]');
const formError = document.getElementById('form-error');
const errorText = document.getElementById('error-text');
const signupForm = document.getElementById('signup-form');
const successState = document.getElementById('success-state');
let captchaToken = '';
window.onCaptchaSuccess = function(token){ captchaToken = token; };
const CAPTCHA_MISSING_MSG = 'Veuillez compléter le captcha';
function validateEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(String(email).toLowerCase());
}
function updateButtonState() {
  const usernameValid = usernameInput.value.trim().length > 0;
  const emailValid = validateEmail(emailInput.value.trim());
  const passwordValid = passwordInput.value.trim().length >= 6;
  const passwordsMatch = passwordInput.value.trim() === confirmPasswordInput.value.trim();
  signupButton.disabled = !(usernameValid && emailValid && passwordValid && passwordsMatch);
}
usernameInput.addEventListener('input', () => {
  updateButtonState();
  usernameInput.classList.remove('input-error');
  formError.classList.remove('show');
});
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
confirmPasswordInput.addEventListener('input', () => {
  updateButtonState();
  confirmPasswordInput.classList.remove('input-error');
  formError.classList.remove('show');
});
function showError(message) {
  errorText.textContent = message;
  formError.classList.add('show');
  if (message.toLowerCase().includes("email")) {
    emailInput.classList.add('input-error');
  } else if (message.toLowerCase().includes("mot de passe") && !message.toLowerCase().includes("correspond")) {
    passwordInput.classList.add('input-error');
  } else if (message.toLowerCase().includes("correspond")) {
    confirmPasswordInput.classList.add('input-error');
  } else if (message.toLowerCase().includes("utilisateur")) {
    usernameInput.classList.add('input-error');
  } else {
    usernameInput.classList.add('input-error');
    emailInput.classList.add('input-error');
    passwordInput.classList.add('input-error');
    confirmPasswordInput.classList.add('input-error');
  }
  formError.scrollIntoView({ behavior: 'smooth', block: 'center' });
}
function showSuccessState() {
  const formElements = Array.from(document.querySelectorAll('.login-form, .login-options, .login-title'));
  formElements.forEach(el => {
    el.style.animation = 'fadeOut 0.4s ease forwards';
  });
  
  setTimeout(() => {
    formElements.forEach(el => {
      el.style.display = 'none';
    });
    successState.style.display = 'flex';
  }, 400);
}
signupForm.addEventListener('submit', async function(e) {
  e.preventDefault();
  if (usernameInput.value.trim().length === 0) {
    showError("Veuillez entrer un nom d'utilisateur");
    return;
  }
  if (!validateEmail(emailInput.value.trim())) {
    showError("Veuillez entrer une adresse email valide");
    return;
  }
  if (passwordInput.value.trim().length < 6) {
    showError("Le mot de passe doit contenir au moins 6 caractères");
    return;
  }
  if (passwordInput.value.trim() !== confirmPasswordInput.value.trim()) {
    showError("Les mots de passe ne correspondent pas");
    return;
  }
  if (!captchaToken) {
    showError(CAPTCHA_MISSING_MSG);
    return;
  }
  signupButton.disabled = true;
  signupButton.innerHTML = "<span>Inscription en cours...</span>";
  try {
    const response = await fetch(`${apiBaseURL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: usernameInput.value.trim(),
        email: emailInput.value.trim(),
        password: passwordInput.value.trim(),
        turnstile_token: captchaToken
      })
    });
    const data = await response.json();
    if (response.ok) {
      signupButton.innerHTML = "<span>Inscription réussie!</span>";
      setTimeout(() => {
        showSuccessState();
      }, 1000);
    } else {
      showError(data.message || "Une erreur est survenue lors de l'inscription");
      signupButton.disabled = false;
      signupButton.innerHTML = "<span>S'inscrire</span>";
    }
    turnstile.reset('#signup-turnstile');
    captchaToken = '';
  } catch (error) {
    showError("Erreur de connexion au serveur");
    signupButton.disabled = false;
    signupButton.innerHTML = "<span>S'inscrire</span>";
    turnstile.reset('#signup-turnstile');
    captchaToken = '';
  }
});