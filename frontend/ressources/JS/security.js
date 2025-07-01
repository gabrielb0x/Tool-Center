let apiBase = window.API_BASE_URL;
const token = localStorage.getItem('token');
const currentPwd = document.getElementById('currentPwd');
const newPwd = document.getElementById('newPwd');
const saveBtn = document.getElementById('saveBtn');
const msg = document.getElementById('message');
function validate(){
  saveBtn.disabled = !(currentPwd.value.trim() && newPwd.value.trim().length >= 7);
}
currentPwd.addEventListener('input', validate);
newPwd.addEventListener('input', validate);

document.getElementById('passwordForm').addEventListener('submit', async e => {
  e.preventDefault();
  if (!token){ msg.textContent = 'Non authentifié.'; return; }
  saveBtn.disabled = true;
  msg.textContent = 'Mise à jour...';
  try {
    const res = await fetch(`${apiBase}/user/update_password`, {
      method:'POST',
      headers:{ 'Content-Type':'application/json', 'Authorization': `Bearer ${token}` },
      body: JSON.stringify({ current_password: currentPwd.value.trim(), new_password: newPwd.value.trim() })
    });
    const data = await res.json();
    if(res.ok && data.success){
      msg.textContent = 'Mot de passe mis à jour';
      currentPwd.value=''; newPwd.value='';
    } else {
      msg.textContent = data.message || 'Erreur';
    }
  } catch(err){
    msg.textContent = 'Erreur réseau';
  }
  saveBtn.disabled = false;
});
