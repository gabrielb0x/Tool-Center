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
const resetButton = document.getElementById('reset-button');
const passwordInput = document.querySelector('input[name="password"]');
const formError = document.getElementById('form-error');
const errorText = document.getElementById('error-text');
const resetForm = document.getElementById('reset-form');
const params = new URLSearchParams(window.location.search);
const token = params.get('token') || '';
function updateButtonState(){
  resetButton.disabled = passwordInput.value.trim().length < 7 || !token;
}
passwordInput.addEventListener('input', () => {
  updateButtonState();
  passwordInput.classList.remove('input-error');
  formError.classList.remove('show');
});
function showError(message){
  errorText.textContent = message;
  formError.classList.add('show');
  passwordInput.classList.add('input-error');
  formError.scrollIntoView({behavior:'smooth', block:'center'});
}
resetForm.addEventListener('submit', async e => {
  e.preventDefault();
  if(passwordInput.value.trim().length < 7){
    showError('Le mot de passe doit contenir au moins 7 caractères');
    return;
  }
  if(!token){
    showError('Token manquant');
    return;
  }
  resetButton.disabled = true;
  resetButton.innerHTML = '<span>Réinitialisation...</span>';
  try{
    const response = await fetch(`${apiBaseURL}/auth/password_reset/confirm`, {
      method:'POST',
      headers:{'Content-Type':'application/json'},
      body: JSON.stringify({ token: token, new_password: passwordInput.value.trim() })
    });
    if(response.ok){
      resetButton.innerHTML = '<span>Mot de passe mis à jour !</span>';
      setTimeout(()=>{ window.location.href = '/signin'; }, 1500);
    } else {
      const data = await response.json();
      showError(data.message || 'Erreur');
      resetButton.disabled = false;
      resetButton.innerHTML = '<span>Réinitialiser</span>';
    }
  }catch(err){
    showError('Erreur de connexion au serveur');
    resetButton.disabled = false;
    resetButton.innerHTML = '<span>Réinitialiser</span>';
  }
});
updateButtonState();
