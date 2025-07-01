let apiBaseURL = window.API_BASE_URL;
let currentUserTools = [];
let currentEditingToolId = null;

document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('token');
    
    if (!token) {
        showAuthInterface();
    } else {
        showMainContent();
        
        Promise.resolve()
            .then(() => fetchUserTools())
            .then(() => {
                initTheme();
                initToolModal();
                initDeleteModal();
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

function fetchUserTools() {
    const token = localStorage.getItem('token');
    const skeletonGrid = document.getElementById('skeletonGrid');
    const noTools = document.getElementById('noTools');
    const toolsGrid = document.getElementById('toolsGrid');
    
    return fetch(`${apiBaseURL}/tools/me`, {
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
            throw new Error('Erreur lors de la récupération des tools');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            currentUserTools = data.tools;
            
            skeletonGrid.classList.add('hidden');
            
            noTools.classList.add('hidden');
            toolsGrid.classList.add('hidden');

            if (data.tools.length === 0) {
                noTools.classList.remove('hidden');
            } else {
                toolsGrid.classList.remove('hidden');
                renderTools(data.tools);
            }
                            
            return data;
        } else {
            throw new Error(data.message || "Erreur lors de la récupération des tools");
        }
    });
}

function renderTools(tools) {
    const toolsGrid = document.getElementById('toolsGrid');
    toolsGrid.innerHTML = '';
    
    tools.forEach(tool => {
        const toolCard = document.createElement('div');
        toolCard.className = 'tool-card';
        
        const createdAt = new Date(tool.created_at);
        const formattedDate = createdAt.toLocaleDateString('fr-FR', {
            day: 'numeric',
            month: 'short',
            year: 'numeric'
        });
        
        let statusText = '';
        let statusClass = '';
        
        switch (tool.status && tool.status.toLowerCase()) {
            case 'published':
                statusText = 'Publié';
                statusClass = 'status-published';
                break;
            case 'moderated':
                statusText = 'En attente';
                statusClass = 'status-pending';
                break;
            case 'hidden':
                statusText = 'REFUSE';
                statusClass = 'status-rejected';
                break;
            default:
                statusText = tool.status || 'INCONNU';
                statusClass = 'status-draft';
                break;
        }
        
        const imageUrl = tool.thumbnail_url || '/assets/default-tool.png';
        const toolId = tool.tool_id || tool.id;
        
        toolCard.innerHTML = `
            <div class="tool-status ${statusClass}">${statusText}</div>
            <img src="${imageUrl}" alt="${tool.title}" class="tool-image">
            <h3 class="tool-title">${tool.title}</h3>
            <p class="tool-description">${tool.description}</p>
            <div class="tool-meta">
                <span class="tool-date">${formattedDate}</span>
                <div class="tool-actions">
                    <button class="tool-action-btn view-tool" data-id="${toolId}">
                        <img src="/assets/show.png" alt="Voir" class="tool-action-icon view-icon">
                    </button>
                    <button class="tool-action-btn edit-tool" data-id="${toolId}">
                        <img src="/assets/edit-icon.png" alt="Éditer" class="tool-action-icon edit-icon">
                    </button>
                    <button class="tool-action-btn delete-tool" data-id="${toolId}">
                        <img src="/assets/trash.png" alt="Supprimer" class="tool-action-icon delete-icon">
                    </button>
                </div>
            </div>
        `;
        
        toolsGrid.appendChild(toolCard);
    });
    
    document.querySelectorAll('.edit-tool').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const toolId = e.currentTarget.getAttribute('data-id');
            openEditToolModal(toolId);
        });
    });
    
    document.querySelectorAll('.delete-tool').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const toolId = e.currentTarget.getAttribute('data-id');
            openDeleteModal(toolId);
        });
    });
    
    document.querySelectorAll('.view-tool').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const toolId = e.currentTarget.getAttribute('data-id');
            const tool = currentUserTools.find(t => t.id === toolId || t.tool_id === toolId);

            if (tool) {
                window.open(tool.url, '_blank');
            }
        });
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

function initToolModal() {
    const toolModal = document.getElementById('toolModal');
    const addToolBtn = document.getElementById('addToolBtn');
    const addToolBtnEmpty = document.getElementById('addToolBtnEmpty');
    const closeToolModal = document.getElementById('closeToolModal');
    const cancelToolModal = document.getElementById('cancelToolModal');
    const submitToolForm = document.getElementById('submitToolForm');
    const submitSpinner = document.getElementById('submitSpinner');
    const submitText = document.getElementById('submitText');
    const toolForm = document.getElementById('toolForm');
    const toolModalTitle = document.getElementById('toolModalTitle');
    
    addToolBtn.addEventListener('click', () => {
        currentEditingToolId = null;
        toolModalTitle.textContent = 'Publier un nouveau tool';
        submitText.textContent = 'Publier';
        toolForm.reset();
        document.getElementById('toolModalError').classList.add('hidden');
        document.getElementById('toolModalSuccess').classList.add('hidden');
        toolModal.classList.add('active');
    });
    
    addToolBtnEmpty.addEventListener('click', () => {
        currentEditingToolId = null;
        toolModalTitle.textContent = 'Publier un nouveau tool';
        submitText.textContent = 'Publier';
        toolForm.reset();
        document.getElementById('toolModalError').classList.add('hidden');
        document.getElementById('toolModalSuccess').classList.add('hidden');
        toolModal.classList.add('active');
    });
    
    closeToolModal.addEventListener('click', () => {
        toolModal.classList.remove('active');
    });
    
    cancelToolModal.addEventListener('click', () => {
        toolModal.classList.remove('active');
    });
    
    submitToolForm.addEventListener('click', () => {
        const title = document.getElementById('toolTitle').value.trim();
        const description = document.getElementById('toolDescription').value.trim();
        const category = document.getElementById('toolCategory').value;
        const url = document.getElementById('toolUrl').value.trim();
        const tags = document.getElementById('toolTags').value.trim();
        const imageFile = document.getElementById('toolImage').files[0];
        
        if (!title || !description || !category || !url) {
            showModalError('Veuillez remplir tous les champs obligatoires');
            return;
        }
        
        submitSpinner.classList.remove('hidden');
        submitText.textContent = 'Envoi en cours...';
        submitToolForm.disabled = true;
        closeToolModal.disabled = true;
        cancelToolModal.disabled = true;
        
        const token = localStorage.getItem('token');
        const formData = new FormData();
        
        formData.append('title', title);
        formData.append('description', description);
        formData.append('category', category);
        formData.append('url', url);
        if (tags) formData.append('tags', tags);
        if (imageFile) formData.append('image', imageFile);
        
        let apiEndpoint = `${apiBaseURL}/tools/add`;
        let method = 'POST';
        
        if (currentEditingToolId) {
            apiEndpoint = `${apiBaseURL}/tools/${currentEditingToolId}`;
            method = 'PUT';
        }
        
        fetch(apiEndpoint, {
            method: method,
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showModalSuccess(currentEditingToolId ? 
                    'Tool mis à jour avec succès !' : 
                    'Tool soumis avec succès ! Il sera publié après modération.');
                
                if (!currentEditingToolId) {
                    toolForm.reset();
                    document.getElementById('noTools').classList.add('hidden');
                    document.getElementById('toolsGrid').classList.remove('hidden');
                }
                
                setTimeout(() => {
                    fetchUserTools().then(() => {
                        toolModal.classList.remove('active');
                    });
                }, 1500);
            } else {
                if (typeof data.retry_after_seconds !== "undefined") {
                    let seconds = parseInt(data.retry_after_seconds, 10);
                    let hours = Math.floor(seconds / 3600);
                    let minutes = Math.floor((seconds % 3600) / 60);
                    let s = seconds % 60;
                    let timeStr = [];
                    if (hours > 0) timeStr.push(hours + "h");
                    if (minutes > 0) timeStr.push(minutes + "min");
                    if (s > 0 && hours === 0) timeStr.push(s + "s");
                    showModalError(
                        (data.message || "Vous devez attendre avant de soumettre un nouvel outil.") +
                        "<br>Veuillez réessayer dans " + timeStr.join(" ") + "."
                    );
                } else {
                    throw new Error(data.message || 'Erreur lors de la soumission du tool');
                }
            }
        })
        .catch(error => {
            showModalError(error.message);
        })
        .finally(() => {
            submitSpinner.classList.add('hidden');
            submitText.textContent = currentEditingToolId ? 'Mettre à jour' : 'Publier';
            submitToolForm.disabled = false;
            closeToolModal.disabled = false;
            cancelToolModal.disabled = false;
        });
    });
}

function openEditToolModal(toolId) {
    const tool = currentUserTools.find(t => t.id === toolId || t.tool_id === toolId);
    if (!tool) return;
    
    currentEditingToolId = toolId;
    const toolModal = document.getElementById('toolModal');
    const toolModalTitle = document.getElementById('toolModalTitle');
    const submitText = document.getElementById('submitText');
    
    toolModalTitle.textContent = 'Modifier le tool';
    submitText.textContent = 'Mettre à jour';
    document.getElementById('toolModalError').classList.add('hidden');
    document.getElementById('toolModalSuccess').classList.add('hidden');
    
    document.getElementById('toolTitle').value = tool.title;
    document.getElementById('toolDescription').value = tool.description;
    document.getElementById('toolCategory').value = tool.category || '';
    document.getElementById('toolUrl').value = tool.content_url || tool.url;
    document.getElementById('toolTags').value = tool.tags ? (Array.isArray(tool.tags) ? tool.tags.join(', ') : tool.tags) : '';
    
    toolModal.classList.add('active');
}

function initDeleteModal() {
    const deleteModal = document.getElementById('deleteModal');
    const closeDeleteModal = document.getElementById('closeDeleteModal');
    const cancelDeleteModal = document.getElementById('cancelDeleteModal');
    const confirmDeleteModal = document.getElementById('confirmDeleteModal');
    const deleteSpinner = document.getElementById('deleteSpinner');
    const deleteText = document.getElementById('deleteText');
    
    closeDeleteModal.addEventListener('click', () => {
        deleteModal.classList.remove('active');
    });
    
    cancelDeleteModal.addEventListener('click', () => {
        deleteModal.classList.remove('active');
    });
    
    confirmDeleteModal.addEventListener('click', () => {
        const toolId = currentEditingToolId;
        if (!toolId) return;
        
        deleteSpinner.classList.remove('hidden');
        deleteText.textContent = 'Suppression...';
        confirmDeleteModal.disabled = true;
        closeDeleteModal.disabled = true;
        cancelDeleteModal.disabled = true;
        
        const token = localStorage.getItem('token');
        
        fetch(`${apiBaseURL}/tools/delete/${toolId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showSuccess('Tool supprimé avec succès');
                deleteModal.classList.remove('active');
                
                fetchUserTools();
            } else {
                throw new Error(data.message || 'Erreur lors de la suppression du tool');
            }
        })
        .catch(error => {
            showError(error.message);
        })
        .finally(() => {
            deleteSpinner.classList.add('hidden');
            deleteText.textContent = 'Supprimer';
            confirmDeleteModal.disabled = false;
            closeDeleteModal.disabled = false;
            cancelDeleteModal.disabled = false;
        });
    });
}

function openDeleteModal(toolId) {
    currentEditingToolId = toolId;
    document.getElementById('deleteModal').classList.add('active');
}

function showModalError(message) {
    const errorContainer = document.getElementById('toolModalError');
    errorContainer.innerHTML = `
        <div class="error-message">
            <img src="/assets/error.png" alt="Erreur" class="error-icon">
            <span>${message}</span>
        </div>
    `;
    errorContainer.classList.remove('hidden');
}

function showModalSuccess(message) {
    const successContainer = document.getElementById('toolModalSuccess');
    successContainer.innerHTML = `
        <div class="success-message">
            <img src="/assets/success.png" alt="Succès" class="success-icon">
            <span>${message}</span>
        </div>
    `;
    successContainer.classList.remove('hidden');
}