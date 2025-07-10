let apiBaseURL = window.API_BASE_URL;
let currentUserData = null;
let twoFactorEnabled = false;

document.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const emailVerified = urlParams.get('event') === 'email_verified';
    
    if (emailVerified) {
        showSuccessModal("Email vérifié !", "Votre adresse email a été vérifiée avec succès. Vous pouvez maintenant profiter pleinement de votre compte ToolCenter.");
    }
    
    const token = localStorage.getItem('token');
    
    if (!token) {
        showAuthInterface();
    } else {
        showMainContent();
        
        Promise.resolve()
            .then(() => fetchUserData())
            .then(() => {
                initAvatarModal();
                initEmailModal();
                initUsernameModal();
                initTheme();
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

function showSuccessModal(title, description) {
    const successModal = document.getElementById('successModal');
    successModal.classList.add('active');

    const successTitle = document.querySelector('.success-title');
    const successMessage = document.querySelector('.success-message');

    successTitle.textContent = title;
    successMessage.textContent = description;

    document.getElementById('successModalClose').addEventListener('click', () => {
    successModal.classList.remove('active');

    const url = new URL(window.location.href);
    url.searchParams.delete('event');
    window.history.replaceState({}, '', url);
    });
}

function fetchUserData() {
    const token = localStorage.getItem('token');
    
    return fetch(`${apiBaseURL}/user/me`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('token');
                throw new Error("Votre session a expiré. Veuillez vous reconnecter.");
            }
            throw new Error('Erreur lors de la récupération des données utilisateur');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            currentUserData = data.user;
            twoFactorEnabled = data.user.two_factor_enabled === true;
            displayUserData(data.user);
            fetchSanctions();
            return data;
        } else {
            throw new Error(data.message || "Erreur lors de la récupération des données utilisateur");
        }
    });
}

function checkTwoFactorStatus() {
    const token = localStorage.getItem('token');
    return fetch(`${apiBaseURL}/user/me`, { headers: { 'Authorization': `Bearer ${token}` } })
        .then(r => r.json())
        .then(d => { if (d.success) twoFactorEnabled = d.user.two_factor_enabled === true; });
}

function displayUserData(userData) {
    document.querySelectorAll('.skeleton').forEach(el => {
        el.classList.add('fade-out');
        setTimeout(() => {
            setAccountStatus(userData.account_status);
            el.classList.add('hidden');
        }, 300);
    });

    setTimeout(() => {
        const usernameEl=document.getElementById("username")
        const name=userData.username
        usernameEl.textContent=''
        ;[...name].forEach((ch,i)=>{
            const span=document.createElement('span')
            span.textContent=ch
            span.style.setProperty('--i',i)
            span.className='letter'
            usernameEl.appendChild(span)
        })
        usernameEl.classList.add('active')
        
        const createdAt = new Date(userData.created_at);
        document.getElementById("member-date").textContent = `Membre depuis ${createdAt.toLocaleDateString('fr-FR')}`;
        document.getElementById("member-date").classList.add('active');
        
        if (userData.avatar_url) {
            document.getElementById("avatar").src = userData.avatar_url;
        } else {
            document.getElementById("avatar").src = "/assets/account.png";
        }
        document.getElementById("avatar").classList.add('active');

        const email = userData.email;
        const [localPart, domain] = email.split("@");
        const hiddenEmail = localPart.replace(/./g, "*") + "@" + domain;

        document.getElementById("email").textContent = hiddenEmail;
        document.getElementById("email").dataset.fullEmail = email;
        document.getElementById("email").classList.add('active');
        
        document.getElementById("toggleEmail").classList.add('active');
        document.getElementById("editEmail").classList.add('active');

        if (userData.is_verified) {
            document.getElementById("verified-badge").classList.add('active');
        }

        document.getElementById("tools-count").textContent = userData.stats.tools_posted || 0;
        document.getElementById("likes-count").textContent = userData.stats.likes_received || 0;
        document.getElementById("views-count").textContent = "0";
        document.getElementById("followers-count").textContent = "0";
        
        document.querySelectorAll('.stat-value').forEach(el => el.classList.add('active'));
        document.querySelectorAll('.stat-label').forEach(el => el.classList.add('active'));

        animateCounters();

        document.getElementById('toggleEmail').addEventListener('click', function() {
            const emailSpan = document.getElementById('email');
            const isHidden = emailSpan.textContent.includes('*');
            
            emailSpan.textContent = isHidden ? emailSpan.dataset.fullEmail : hiddenEmail;
            this.src = isHidden ? '/assets/dont-show.png' : '/assets/show.png';
            
            emailSpan.style.animation = 'none';
            setTimeout(() => {
                emailSpan.style.animation = 'fadeIn 0.3s ease';
            }, 10);
        });
    }, 300);
}

function animateCounters() {
    const counters = document.querySelectorAll('.stat-value');
    const speed = 200;
    
    counters.forEach(counter => {
        const target = +counter.innerText;
        let count = 0;
        
        const updateCount = () => {
            if (count < target) {
                counter.innerText = Math.ceil(count);
                count += target / speed;
                requestAnimationFrame(updateCount);
            } else {
                counter.innerText = target;
            }
        };
        
        updateCount();
    });
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

function initEmailModal() {
    const emailModal = document.getElementById("emailModal");
    const closeEmailModal = document.getElementById("closeEmailModal");
    const cancelEmailChange = document.getElementById("cancelEmailChange");
    const editEmail = document.getElementById("editEmail");
    const newEmailInput = document.getElementById("newEmail");
    const currentPasswordInput = document.getElementById("currentPassword");
    const codeInput = document.getElementById("email2FACode");
    const confirmEmailChange = document.getElementById("confirmEmailChange");

    editEmail.addEventListener("click", () => {
        checkTwoFactorStatus().finally(() => {
            document.getElementById("email2FAGroup").style.display = twoFactorEnabled ? 'block' : 'none';
            emailModal.classList.add("active");
            newEmailInput.focus();
        });
    });
    closeEmailModal.addEventListener("click", () => {
        emailModal.classList.remove("active");
        newEmailInput.value = "";
        currentPasswordInput.value = "";
        codeInput.value = "";
    });
    cancelEmailChange.addEventListener("click", () => {
        emailModal.classList.remove("active");
        newEmailInput.value = "";
        currentPasswordInput.value = "";
        codeInput.value = "";
    });

    confirmEmailChange.addEventListener("click", () => {
        const newEmail = newEmailInput.value.trim();
        const currentPassword = currentPasswordInput.value.trim();
        const code = codeInput.value.trim();

        if (!newEmail || !currentPassword) {
            showError("Veuillez remplir tous les champs");
            return;
        }
        if (twoFactorEnabled && code.length !== 6) {
            showError("Code 2FA invalide");
            return;
        }

        const spinner = document.createElement("span");
        spinner.className = "spinner";
        confirmEmailChange.prepend(spinner);
        closeEmailModal.disabled = true;
        cancelEmailChange.disabled = true;
        confirmEmailChange.disabled = true;
        const token = localStorage.getItem("token");

        const payload = { new_email: newEmail, current_password: currentPassword };
        if (twoFactorEnabled) payload.two_factor_code = code;
        fetch(`${apiBaseURL}/user/update_email`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify(payload)
        })
        .then(response => response.json())
        .then(data => {
        if (data.success) {
            currentUserData.email = newEmail;
            document.getElementById("email").textContent = newEmail;
            document.getElementById("email").dataset.fullEmail = newEmail;
            emailModal.classList.remove("active");
            newEmailInput.value = "";
            currentPasswordInput.value = "";
            const emailSpan = document.getElementById("email");
            emailSpan.style.animation = "none";
            setTimeout(() => {
            emailSpan.style.animation = "pulse 0.5s ease";
            }, 10);
            showSuccessModal("Votre email a été modifié avec succès !", "Vous recevrez un email de confirmation à votre nouvelle adresse.");
        } else {
            throw new Error(data.message || "Erreur lors de la modification de l'email");
        }
        })
        .catch(error => {
            emailModal.classList.remove("active");
            console.error("Erreur :", error);
            showError(error.message || "Une erreur s'est produite lors de la modification de l'email");
        })
        .finally(() => {
            closeEmailModal.disabled = false;
            cancelEmailChange.disabled = false;
            confirmEmailChange.disabled = false;
            spinner.remove();
        });
    });
}

function initAvatarModal() {
    const avatarModal = document.getElementById("avatarModal");
    const avatarInput = document.getElementById("avatarInput");
    const avatarPreview = document.getElementById("avatarPreview");
    const avatarEditBtn = document.getElementById("avatarEditBtn");
    const closeAvatarModal = document.getElementById("closeAvatarModal");
    const cancelAvatarChange = document.getElementById("cancelAvatarChange");
    const confirmAvatarChange = document.getElementById("confirmAvatarChange");
    const MAX_SIZE = 5 * 1024 * 1024;
    const ALLOWED_TYPES = ["image/jpeg","image/png","image/webp"];
    
    function previewAvatar(event) {
        const file = event.target.files && event.target.files[0];
        if (!file) return;

        if (!ALLOWED_TYPES.includes(file.type)) {
            showError("Format d'image non supporté (jpeg, png ou webp)");
            avatarModal.classList.remove("active");
            avatarInput.value = "";
            return;
        }

        if (file.size > MAX_SIZE) {
            showError(`Image trop lourde (max ${MAX_SIZE/(1024*1024)} Mo)`);
            avatarModal.classList.remove("active");
            avatarInput.value = "";
            return;
        }

        const reader = new FileReader();

        reader.onload = function() {
            avatarPreview.src = reader.result;
            avatarModal.classList.add("active");
        };

        reader.readAsDataURL(file);
    }

    avatarEditBtn.addEventListener("click", () => {
        avatarInput.click();
    });
    avatarInput.addEventListener("change", previewAvatar);
    closeAvatarModal.addEventListener("click", () => {
        avatarModal.classList.remove("active");
        avatarInput.value = "";
    });
    cancelAvatarChange.addEventListener("click", () => {
        avatarModal.classList.remove("active");
        avatarInput.value = "";
    });

    confirmAvatarChange.addEventListener("click", () => {
        const spinner = document.createElement("span");
        spinner.className = "spinner";
        confirmAvatarChange.prepend(spinner);
        closeAvatarModal.disabled = true;
        cancelAvatarChange.disabled = true;
        confirmAvatarChange.disabled = true;
        const token = localStorage.getItem("token");
        
        fetch(avatarPreview.src)
        .then(res => res.blob())
        .then(blob => {
            const formData = new FormData();
            formData.append("avatar",blob,"avatar.jpg");
            return fetch(`${apiBaseURL}/user/avatar`, {
                method: "POST",
                headers: {
                    "Authorization": `Bearer ${token}`
                },
                body: formData
            });
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                document.getElementById("avatar").src = avatarPreview.src;
                const avatar = document.getElementById("avatar");
                avatar.style.animation = "none";
                setTimeout(() => {
                    avatar.style.animation = "pulse 0.5s ease";
                }, 10);
                showSuccessModal("Votre avatar a été modifié !", "Votre nouvel avatar sera visible sur votre profil et dans vos outils.");
            } else {
                const err = new Error(
                    data.message || "Erreur lors de la modification de l'avatar"
                );
                err.retry_at = data.retry_at;
                throw err;   
            }
        })
        .catch(error => {
            if (error.retry_at) {
                const retryAt = new Date(error.retry_at);
                const now = new Date();
                const timeDiff = Math.ceil((retryAt - now) / 1000 / 60 / 60);
                const retryMessage = `Vous ne pouvez changer votre photo de profil qu'une fois par jour. Réessayez dans ${timeDiff} heure(s).`;
                showError(retryMessage);
            } else {
                showError(error.message || "Une erreur s'est produite lors de la modification de l'avatar");
            }
        })
        .finally(() => {
            closeAvatarModal.disabled = false;
            cancelAvatarChange.disabled = false;
            confirmAvatarChange.disabled = false;
            spinner.remove();
            avatarModal.classList.remove("active");
            avatarInput.value = "";
        });
    });
}

// G0
document.addEventListener('DOMContentLoaded', () => {
    const tabMainBtn = document.getElementById('tabMainBtn');
    const tabStatusBtn = document.getElementById('tabStatusBtn');
    const tabMain = document.getElementById('tabMain');
    const tabStatus = document.getElementById('tabStatus');

    tabMainBtn.onclick = function() {
        tabMainBtn.classList.add('active');
        tabStatusBtn.classList.remove('active');
        tabMain.classList.add('active');
        tabStatus.classList.remove('active');
    }
    tabStatusBtn.onclick = function() {
        tabStatusBtn.classList.add('active');
        tabMainBtn.classList.remove('active');
        tabStatus.classList.add('active');
        tabMain.classList.remove('active');
    }
});

function initUsernameModal() {
    const usernameModal = document.getElementById('usernameModal');
    const newUsernameInput = document.getElementById('newUsername');
    const closeUsernameModal = document.getElementById('closeUsernameModal');
    const cancelUsernameChange = document.getElementById('cancelUsernameChange');
    const confirmUsernameChange = document.getElementById('confirmUsernameChange');
    const usernameEl = document.getElementById('username');
    const editBtn = document.getElementById('editUsername');

    editBtn.addEventListener('click', () => {
        newUsernameInput.value = usernameEl.textContent.trim();
        usernameModal.classList.add('active');
        newUsernameInput.focus();
    });

    function closeModal() {
        usernameModal.classList.remove('active');
        newUsernameInput.value = '';
    }

    closeUsernameModal.addEventListener('click', closeModal);
    cancelUsernameChange.addEventListener('click', closeModal);

    confirmUsernameChange.addEventListener('click', () => {
        const newUsername = newUsernameInput.value.trim();
        if (newUsername.length < 3 || newUsername.length > 50) {
            showError('Le pseudo doit contenir entre 3 et 50 caractères.');
            return;
        }

        const spinner = document.createElement('span');
        spinner.className = 'spinner';
        confirmUsernameChange.prepend(spinner);
        closeUsernameModal.disabled = true;
        cancelUsernameChange.disabled = true;
        confirmUsernameChange.disabled = true;

        const token = localStorage.getItem('token');
        fetch(`${apiBaseURL}/user/update_username`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
            body: JSON.stringify({ username: newUsername })
        })
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                usernameEl.textContent = newUsername;
                showSuccessModal('Pseudo modifié !', 'Ton pseudo a été mis à jour !');
                closeModal();
            } else {
                showError(data.message || 'Erreur lors de la modification du pseudo');
            }
        })
        .catch(() => {
            showError("Impossible de se connecter au serveur. Veuillez réessayer plus tard.");
        })
        .finally(() => {
            closeUsernameModal.disabled = false;
            cancelUsernameChange.disabled = false;
            confirmUsernameChange.disabled = false;
            spinner.remove();
        });
    });
}

function setAccountStatus(status) {
    const bar = document.getElementById('statusProgressFill');
    const statusText = document.getElementById('accountStatusText');
    const statusDesc = document.getElementById('statusDescription');
    const steps = ['Good','Limited','Very Limited','At Risk','Banned'];
    const colors = [
        'linear-gradient(90deg,#10c95a,#2fd174)', // Good
        'linear-gradient(90deg,#ffca29,#ffd66b)', // Limited
        'linear-gradient(90deg,#f88407,#ff5f00)', // Very Limited
        'linear-gradient(90deg,#e94434,#fc6a55)', // At Risk
        'linear-gradient(90deg,#a10000,#570404)'  // Banned
    ];
    const descriptions = [
        "Merci de respecter les règles de ToolCenter !",
        "Vous ne pouvez plus utiliser certaines parties de ToolCenter. Vous risquez une suspension si vous enfreignez à nouveau les règles.",
        "Vous ne pouvez plus utiliser la plupart des parties de ToolCenter. Votre compte pourrait être banni.",
        "Votre compte est à risque. Vous ne pouvez plus utiliser ToolCenter tant que vous n'avez pas réglé le problème.",
        "Votre compte est banni. Vous ne pouvez plus utiliser ToolCenter."
    ];
    const widths = ['10%','30%','50%','70%','100%'];
    const idx = steps.indexOf(status);
    if(idx === -1) return;
    bar.style.width = widths[idx];
    bar.style.background = colors[idx];
    statusText.textContent = steps[idx];
    statusDesc.textContent = descriptions[idx];
}

function fetchSanctions() {
    const token = localStorage.getItem('token');
    fetch(`${apiBaseURL}/user/sanctions`, {
        headers: { 'Authorization': `Bearer ${token}` }
    })
    .then(r => r.json())
    .then(d => {
        if (!d.success) return;
        const cont = document.getElementById('sanctionsContainer');
        cont.innerHTML = '';
        const now = Date.now();
        d.sanctions.forEach(s => {
            const end = s.end_date ? new Date(s.end_date).getTime() : null;
            const active = !end || end > now;
            const div = document.createElement('div');
            div.className = active ? 'sanction active' : 'sanction expired';
            div.textContent = `${s.type} - ${s.reason}` + (end ? ` (jusqu'au ${new Date(end).toLocaleString()})` : '');
            cont.appendChild(div);
        });
    });
}