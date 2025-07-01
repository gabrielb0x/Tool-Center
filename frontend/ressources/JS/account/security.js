let apiBaseURL = window.API_BASE_URL;
let twoFactorEnabled = false;

document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    
    if (!token) {
        showAuthInterface();
    } else {
        showMainContent();
        
        Promise.resolve()
            .then(() => {
                initTheme();
                initSecurityButtons();
                checkTwoFactor();
                document.getElementById('closeEmailModal').onclick = closeEmailModal;
                document.getElementById('cancelEmailModal').onclick = closeEmailModal;
                document.getElementById('confirmEmailModal').onclick = submitEmailUpdate;
                document.getElementById('closePasswordModal').onclick = closePasswordModal;
                document.getElementById('cancelPasswordModal').onclick = closePasswordModal;
                document.getElementById('confirmPasswordModal').onclick = submitPasswordUpdate;
            })
            .catch(error => {
                console.error('Erreur:', error);
                showError("Impossible de se connecter au serveur. Veuillez réessayer plus tard.");
                setTimeout(() => {
                    hideMainContent();
                    showAuthInterface();
                }, 300);
            });
    }
    
    const url = new URL(window.location.href);
    const tok = url.searchParams.get("token");
    if (tok) {
        localStorage.setItem("token", tok);
        url.searchParams.delete("token");
        history.replaceState({}, "", url);
        window.location.reload();
    }
});

function showAuthInterface() {
    document.getElementById('authContainer').classList.add('active');
    document.getElementById('mainContentContainer').classList.remove('active');
    document.getElementById('mainHeader').classList.remove('active');
    document.getElementById('mainContainer').classList.remove('active');
}

function showMainContent() {
    document.getElementById('authContainer').classList.remove('active');
    document.getElementById('mainContentContainer').classList.add('active');
    document.getElementById('mainHeader').classList.add('active');
    document.getElementById('mainContainer').classList.add('active');
}

function hideMainContent() {
    document.getElementById('mainContentContainer').classList.remove('active');
    document.getElementById('mainHeader').classList.remove('active');
    document.getElementById('mainContainer').classList.remove('active');
}

function showError(message) {
    const errorContainer = document.getElementById('errorContainer');
    errorContainer.innerHTML = `
        <div class="error-message">
            <img src="/assets/error.png" alt="Erreur" class="error-icon">
            <span>${message}</span>
        </div>
    `;
    errorContainer.classList.remove('hidden');
}

function showSuccess(message) {
    const successContainer = document.getElementById('errorContainer');
    successContainer.innerHTML = `
        <div class="success-message">
            <img src="/assets/success.png" alt="Succès" class="success-icon">
            <span>${message}</span>
        </div>
    `;
    successContainer.classList.remove('hidden');
}

function initTheme() {
    const themeSwitcher = document.getElementById("theme-switcher");
    const themeIcon = document.getElementById("theme-icon");
    const savedTheme = localStorage.getItem("theme");

    if (savedTheme === "light") {
        document.body.classList.add("light-mode");
        themeIcon.src = "/assets/switcher-noir.png";
    } else {
        themeIcon.src = "/assets/switcher.png";
    }

    themeSwitcher.addEventListener("click", () => {
        const isDarkMode = !document.body.classList.contains("light-mode");
        const newTheme = isDarkMode ? "light" : "dark";

        document.body.classList.toggle("light-mode", newTheme === "light");

        themeIcon.src = newTheme === "dark"
            ? "/assets/switcher.png"
            : "/assets/switcher-noir.png";

        localStorage.setItem("theme", newTheme);
        
        themeIcon.style.animation = 'none';
        setTimeout(() => {
            themeIcon.style.animation = 'float 0.5s ease';
        }, 10);
    });
}

function initSecurityButtons() {
    document.getElementById('changeEmailBtn').addEventListener('click', openEmailModal);

    document.getElementById('changePasswordBtn').addEventListener('click', openPasswordModal);

    document.getElementById('enable2faBtn').addEventListener('click', handle2FA);

    document.getElementById('logoutAllBtn').addEventListener('click', () => {
        if (confirm('Êtes-vous sûr de vouloir déconnecter toutes les sessions ?')) {
            logoutAllSessions();
        }
    });

    fetchSessions();
}

function checkTwoFactor() {
    const token = localStorage.getItem('token');
    fetch(`${apiBaseURL}/user/me`, { headers: { 'Authorization': `Bearer ${token}` } })
    .then(r => r.json())
    .then(data => {
        if (data.success) {
            twoFactorEnabled = data.user.two_factor_enabled === true;
            document.getElementById('enable2faBtn').textContent = twoFactorEnabled ? 'Désactiver' : 'Activer';
        }
    });
}

function openEmailModal() {
    document.getElementById('emailError').style.display = 'none';
    document.getElementById('newEmail').value = '';
    document.getElementById('currentEmailPassword').value = '';
    document.getElementById('email2FACode').value = '';
    checkTwoFactor();
    document.getElementById('email2FAGroup').style.display = twoFactorEnabled ? 'block' : 'none';
    document.getElementById('emailModal').classList.add('active');
}

function closeEmailModal() {
    document.getElementById('emailModal').classList.remove('active');
    document.getElementById('email2FAGroup').style.display = 'none';
}

function openPasswordModal() {
    document.getElementById('passwordError').style.display = 'none';
    document.getElementById('currentPassword').value = '';
    document.getElementById('newPassword').value = '';
    document.getElementById('confirmPassword').value = '';
    document.getElementById('password2FACode').value = '';
    checkTwoFactor();
    document.getElementById('password2FAGroup').style.display = twoFactorEnabled ? 'block' : 'none';
    document.getElementById('passwordModal').classList.add('active');
}

function closePasswordModal() {
    document.getElementById('passwordModal').classList.remove('active');
}

function validateEmail(email) {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(String(email).toLowerCase());
}

function handle2FA() {
    const token = localStorage.getItem('token');
    const modal = document.getElementById('twoFAModal');
    const qrContainer = document.getElementById('qrCodeContainer');
    const codeInput = document.getElementById('twoFACodeInput');
    const errorBox = document.getElementById('twoFAError');
    const errorText = document.getElementById('twoFAErrorText');
    const modalBody = modal.querySelector('.modal-body');

    function closeModal() {
    modal.classList.remove('active');
    codeInput.value = '';
    errorBox.style.display = 'none';
    }

    document.getElementById('close2FAModal').onclick = closeModal;
    document.getElementById('cancel2FAModal').onclick = closeModal;

    if (!twoFactorEnabled) {
    fetch(`${apiBaseURL}/user/2fa/setup`, { headers: { 'Authorization': `Bearer ${token}` } })
    .then(r => r.json())
    .then(data => {
        if (data.success) {
        qrContainer.innerHTML = `<img src="data:image/png;base64,${data.qr_code}" alt="QR Code">`;
        // Ajoute la clé secrète dans le <p>
        modalBody.querySelector('p').innerHTML = `Scannez ce QR Code avec votre application d'authentification puis entrez le code généré.<br><br><strong>Ou entrez cette clé manuellement&nbsp;:</strong><br><code style="font-size:1.1em;word-break:break-all;">${data.secret}</code>`;
        modal.classList.add('active');
        document.getElementById('confirm2FAModal').onclick = () => {
            const code = codeInput.value.trim();
            if (!code) return;
            fetch(`${apiBaseURL}/user/enable_2fa`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
            body: JSON.stringify({ secret: data.secret, code })
            }).then(resp => resp.json()).then(res => {
            if (res.success) {
                twoFactorEnabled = true;
                document.getElementById('enable2faBtn').textContent = 'Désactiver';
                closeModal();
            } else {
                errorText.textContent = res.message || 'Erreur';
                errorBox.style.display = 'flex';
            }
            });
        };
        } else {
        showError(data.message || 'Erreur');
        }
    });
    } else {
    const code = prompt('Code 2FA pour désactiver');
    if (!code) return;
    fetch(`${apiBaseURL}/user/disable_2fa`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: code.trim() })
    }).then(r => r.json()).then(res => {
        if (res.success) {
        twoFactorEnabled = false;
        document.getElementById('enable2faBtn').textContent = 'Activer';
        } else { alert(res.message || 'Erreur'); }
    });
    }
}

function submitEmailUpdate() {
    const newEmail = document.getElementById('newEmail').value.trim();
    const password = document.getElementById('currentEmailPassword').value;
    const code = document.getElementById('email2FACode').value.trim();
    const btn = document.getElementById('confirmEmailModal');
    const errorBox = document.getElementById('emailError');
    const errorText = document.getElementById('emailErrorText');
    errorBox.style.display = 'none';

    if (!validateEmail(newEmail)) {
        errorText.textContent = "Email invalide";
        errorBox.style.display = 'flex';
        return;
    }
    if (password.length < 6) {
        errorText.textContent = "Mot de passe incorrect";
        errorBox.style.display = 'flex';
        return;
    }

    if (twoFactorEnabled && code.length !== 6) {
        errorText.textContent = 'Code 2FA invalide';
        errorBox.style.display = 'flex';
        return;
    }

    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span>';

    const token = localStorage.getItem('token');
    const payload = { new_email: newEmail, current_password: password };
    if (twoFactorEnabled) payload.two_factor_code = code;
    fetch(`${apiBaseURL}/user/update_email`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
    }).then(r => r.json()).then(res => {
        if (res.success) {
            closeEmailModal();
            showSuccess('Email mis à jour. Vérifiez votre boîte de réception.');
        } else {
            errorText.textContent = res.message || 'Erreur';
            errorBox.style.display = 'flex';
        }
    }).catch(() => {
        errorText.textContent = 'Erreur de connexion au serveur';
        errorBox.style.display = 'flex';
    }).finally(() => {
        btn.disabled = false;
        btn.textContent = 'Valider';
    });
}

function submitPasswordUpdate() {
    const currentPass = document.getElementById('currentPassword').value;
    const newPass = document.getElementById('newPassword').value;
    const confirmPass = document.getElementById('confirmPassword').value;
    const code = document.getElementById('password2FACode').value.trim();
    const btn = document.getElementById('confirmPasswordModal');
    const errorBox = document.getElementById('passwordError');
    const errorText = document.getElementById('passwordErrorText');
    errorBox.style.display = 'none';

    if (currentPass.length < 6) {
        errorText.textContent = 'Mot de passe actuel invalide';
        errorBox.style.display = 'flex';
        return;
    }
    if (newPass.length < 7 || newPass.length > 30) {
        errorText.textContent = 'Le nouveau mot de passe doit faire 7 à 30 caractères';
        errorBox.style.display = 'flex';
        return;
    }
    if (newPass !== confirmPass) {
        errorText.textContent = 'Les mots de passe ne correspondent pas';
        errorBox.style.display = 'flex';
        return;
    }
    if (twoFactorEnabled && code.length !== 6) {
        errorText.textContent = 'Code 2FA invalide';
        errorBox.style.display = 'flex';
        return;
    }

    btn.disabled = true;
    btn.innerHTML = '<span class="spinner"></span>';

    const token = localStorage.getItem('token');
    const payload = { current_password: currentPass, new_password: newPass };
    if (twoFactorEnabled) payload.two_factor_code = code;
    fetch(`${apiBaseURL}/user/update_password`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(payload)
    }).then(r => r.json()).then(res => {
        if (res.success) {
            closePasswordModal();
            showSuccess('Mot de passe mis à jour');
        } else {
            errorText.textContent = res.message || 'Erreur';
            errorBox.style.display = 'flex';
        }
    }).catch(() => {
        errorText.textContent = 'Erreur de connexion au serveur';
        errorBox.style.display = 'flex';
    }).finally(() => {
        btn.disabled = false;
        btn.textContent = 'Valider';
    });
}

function fetchSessions() {
    const token = localStorage.getItem('token');
    const sessionsList = document.getElementById('sessionsList');
    const skeletonSession = document.getElementById('skeletonSession');

    fetch(`${apiBaseURL}/auth/sessions`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            renderSessions(data.sessions);
            if (skeletonSession) skeletonSession.classList.add('hidden');
        } else {
            throw new Error(data.message || 'Erreur lors de la récupération des sessions');
        }
    })
    .catch(error => {
        showError(error.message);
        if (skeletonSession) skeletonSession.classList.add('hidden');
    });
}

function renderSessions(sessions) {
    const sessionsList = document.getElementById('sessionsList');
    const currentSessionId = localStorage.getItem('sessionId');
    
    sessionsList.innerHTML = '';
    
    sessions.forEach(session => {
        const sessionItem = document.createElement('div');
        sessionItem.className = 'session-item';
        
        const createdAt = new Date(session.created_at);
        const formattedDate = createdAt.toLocaleDateString('fr-FR', {
            day: 'numeric',
            month: 'short',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
        
        const isCurrent = session.id === currentSessionId;
        const deviceIcon = session.device_type === 'mobile' ? 
            '/assets/mobile-icon.png' : '/assets/desktop-icon.png';
        
        sessionItem.innerHTML = `
            <div class="session-info">
                <div class="session-device">
                    <img src="${deviceIcon}" alt="Appareil" class="session-device-icon">
                    <div>
                        <h4>${session.os} - ${session.browser}</h4>
                        <p>${session.location || 'Localisation inconnue'} • ${formattedDate}</p>
                    </div>
                </div>
                <div class="session-status ${isCurrent ? 'current' : ''}">
                    ${isCurrent ? 'Cette session' : 'Active'}
                </div>
            </div>
            <button class="session-action-btn" data-id="${session.id}" ${isCurrent ? 'disabled' : ''}>
                <img src="/assets/logout.png" alt="Déconnecter" class="session-action-icon">
            </button>
        `;
        
        sessionsList.appendChild(sessionItem);
    });
    
    document.querySelectorAll('.session-action-btn:not([disabled])').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const sessionId = e.currentTarget.getAttribute('data-id');
            if (confirm('Déconnecter cette session ?')) {
                logoutSession(sessionId);
            }
        });
    });
}

function logoutSession(sessionId) {
    const token = localStorage.getItem('token');
    
    fetch(`${apiBaseURL}/auth/sessions/${sessionId}`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showSuccess('Session déconnectée avec succès');
            fetchSessions();
        } else {
            throw new Error(data.message || 'Erreur lors de la déconnexion de la session');
        }
    })
    .catch(error => {
        showError(error.message);
    });
}

function logoutAllSessions() {
    const token = localStorage.getItem('token');
    
    fetch(`${apiBaseURL}/auth/sessions`, {
        method: 'DELETE',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showSuccess('Toutes les sessions ont été déconnectées');
            fetchSessions();
        } else {
            throw new Error(data.message || 'Erreur lors de la déconnexion des sessions');
        }
    })
    .catch(error => {
        showError(error.message);
    });
}
