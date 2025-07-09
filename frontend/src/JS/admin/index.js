const BASE_API_URL = PANEL_CONFIG.API_BASE;
const SECTION_IDS = { users: 'users-section', logs: 'logs-section' };
let currentUserId = null;
let currentPage = 1;
let currentUserRole = null;
let currentActivityPage = 1;
let totalActivity = 0;

const preloader = document.getElementById('preloader');
const adminContent = document.getElementById('admin-content');
const themeToggle = document.getElementById('theme-toggle');
const adminAvatar = document.getElementById('admin-avatar');

function toggleTheme() {
    document.body.classList.toggle('light-theme');
    localStorage.setItem('theme', document.body.classList.contains('light-theme') ? 'light' : 'dark');
    updateThemeIcon();
}

function updateThemeIcon() {
    const icon = themeToggle.querySelector('i');
    if (document.body.classList.contains('light-theme')) {
    icon.classList.replace('fa-moon', 'fa-sun');
    } else {
    icon.classList.replace('fa-sun', 'fa-moon');
    }
}

themeToggle.addEventListener('click', toggleTheme);
if (localStorage.getItem('theme') === 'light') {
    document.body.classList.add('light-theme');
}
updateThemeIcon();

function showToast(type, message) {
    const toastContainer = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    
    const icons = {
    success: 'check-circle',
    error: 'exclamation-circle',
    warning: 'exclamation-triangle',
    info: 'info-circle'
    };
    
    toast.innerHTML = `
    <i class="fas fa-${icons[type]} toast-icon"></i>
    <span>${message}</span>
    <button class="toast-close">&times;</button>
    `;
    
    toastContainer.appendChild(toast);
    
    setTimeout(() => {
    toast.classList.add('show');
    }, 100);
    
    setTimeout(() => {
    toast.classList.remove('show');
    setTimeout(() => {
        toast.remove();
    }, 300);
    }, 5000);
    
    toast.querySelector('.toast-close').addEventListener('click', () => {
    toast.classList.remove('show');
    setTimeout(() => {
        toast.remove();
    }, 300);
    });
}

async function checkAdminPermissions() {
    try {
    const token = localStorage.getItem('token');
    if (!token) {
        throw new Error('No token found');
    }

    const response = await fetch(`${BASE_API_URL}/user/me`, {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    const data = await response.json();
    if (!data.success || !data.user || data.user.role !== 'Admin') {
        throw new Error('User is not admin');
    }

    const user = data.user;
    currentUserRole = user.role;
    adminAvatar.src = user.avatar_url ? `${user.avatar_url}` : '/assets/account.png';
    
    if (user.role === 'Moderator') {
        window.location.href = '/moderator';
        return false;
    }

    return true;
    } catch (error) {
    console.error('Permission check failed:', error);
    return false;
    }
}

async function fetchUsers(search = '', page = 1) {
    showLoadingSkeleton('users');
    
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users?search=${encodeURIComponent(search)}&page=${page}`, {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    const data = await response.json();
    return data;
    } catch (error) {
    showToast('error', `Erreur lors du chargement des utilisateurs: ${error.message}`);
    console.error('Error fetching users:', error);
    return { users: [], total: 0 };
    }
}

async function fetchUserDetails(userId) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}`, {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    return await response.json();
    } catch (error) {
    showToast('error', `Erreur lors du chargement des détails: ${error.message}`);
    console.error('Error fetching user details:', error);
    return null;
    }
}

async function fetchUserActivity(userId, page = 1) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}/activity?page=${page}`, {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    return await response.json();
    } catch (error) {
    showToast('error', `Erreur lors du chargement de l'activité: ${error.message}`);
    console.error('Error fetching user activity:', error);
    return { logs: [], total: 0 };
    }
}

async function fetchUserTools(userId) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}/tools`, {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    return await response.json();
    } catch (error) {
    showToast('error', `Erreur lors du chargement des outils: ${error.message}`);
    console.error('Error fetching user tools:', error);
    return { tools: [] };
    }
}

async function fetchBanInfo(userId) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}/ban`, {
        headers: { 'Authorization': `Bearer ${token}` }
    });
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    return await response.json();
    } catch (error) {
    showToast('error', `Erreur lors du chargement de la raison: ${error.message}`);
    console.error('Error fetching ban info:', error);
    return null;
    }
}

async function fetchSystemLogs(page = 1) {
    showLoadingSkeleton('logs');
    
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/logs?page=${page}`, {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    return await response.json();
    } catch (error) {
    showToast('error', `Erreur lors du chargement des logs: ${error.message}`);
    console.error('Error fetching system logs:', error);
    return { logs: [], total: 0 };
    }
}

async function fetchStats() {
    showLoadingSkeleton('stats');
    
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/stats`, {
        headers: {
        'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    return await response.json();
    } catch (error) {
    showToast('error', `Erreur lors du chargement des statistiques: ${error.message}`);
    console.error('Error fetching stats:', error);
    return null;
    }
}

async function updateUser(userId, data) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}`, {
        method: 'PUT',
        headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    const result = await response.json();
    showToast('success', 'Utilisateur mis à jour avec succès');
    return result;
    } catch (error) {
    showToast('error', `Erreur lors de la mise à jour: ${error.message}`);
    console.error('Error updating user:', error);
    return null;
    }
}

async function banUser(userId, reason) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}/ban`, {
        method: 'POST',
        headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
        },
        body: JSON.stringify({ reason, duration_hours: parseInt(document.getElementById('ban-duration').value, 10) || 0 })
    });
    
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    
    const result = await response.json();
    showToast('success', 'Utilisateur banni avec succès');
    return result;
    } catch (error) {
    showToast('error', `Erreur lors du ban: ${error.message}`);
    console.error('Error banning user:', error);
    return null;
    }
}

async function unbanUser(userId) {
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/users/${userId}/unban`, {
        method: 'POST',
        headers: {
        'Authorization': `Bearer ${token}`
        }
    });

    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }

    const result = await response.json();
    showToast('success', 'Utilisateur débanni avec succès');
    return result;
    } catch (error) {
    showToast('error', `Erreur lors du débannissement: ${error.message}`);
    console.error('Error unbanning user:', error);
    return null;
    }
}

async function clearLogs() {
    if (!confirm('Supprimer tous les logs ?')) return;
    try {
    const token = localStorage.getItem('token');
    const response = await fetch(`${BASE_API_URL}/admin/logs/clear`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
    });
    if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
    }
    showToast('success', 'Logs supprimés');
    loadSystemLogs();
    } catch (error) {
    showToast('error', `Erreur lors de la suppression: ${error.message}`);
    console.error('Error clearing logs:', error);
    }
}

function showLoadingSkeleton(type) {
    if (type === 'stats') {
    const statsCards = document.querySelector('.stats-cards');
    if (statsCards) {
        statsCards.innerHTML = `
        <div class="stat-card">
            <h3 class="skeleton skeleton-title"></h3>
            <div class="value skeleton skeleton-value"></div>
            <div class="change skeleton skeleton-change"></div>
        </div>
        <div class="stat-card">
            <h3 class="skeleton skeleton-title"></h3>
            <div class="value skeleton skeleton-value"></div>
            <div class="change skeleton skeleton-change"></div>
        </div>
        <div class="stat-card">
            <h3 class="skeleton skeleton-title"></h3>
            <div class="value skeleton skeleton-value"></div>
            <div class="change skeleton skeleton-change"></div>
        </div>
        <div class="stat-card">
            <h3 class="skeleton skeleton-title"></h3>
            <div class="value skeleton skeleton-value"></div>
            <div class="change skeleton skeleton-change"></div>
        </div>
        `;
    }
    } else if (type === 'users') {
    const tbody = document.getElementById('users-table-body');
    if (tbody) {
        tbody.innerHTML = '';
        for (let i = 0; i < 5; i++) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>
            <div class="user-cell">
                <div class="skeleton skeleton-avatar"></div>
                <div class="user-info">
                <span class="user-name skeleton skeleton-text" style="width: 120px;"></span>
                <span class="user-email skeleton skeleton-text" style="width: 80px;"></span>
                </div>
            </div>
            </td>
            <td><div class="skeleton skeleton-text" style="width: 150px;"></div></td>
            <td><div class="skeleton skeleton-text" style="width: 80px;"></div></td>
            <td><div class="skeleton skeleton-text" style="width: 60px;"></div></td>
            <td>
            <div style="display: flex; gap: 5px;">
                <div class="skeleton" style="width: 36px; height: 36px; border-radius: 8px;"></div>
                <div class="skeleton" style="width: 36px; height: 36px; border-radius: 8px;"></div>
                <div class="skeleton" style="width: 36px; height: 36px; border-radius: 8px;"></div>
            </div>
            </td>
        `;
        tbody.appendChild(row);
        }
    }
    } else if (type === 'logs') {
    const tbody = document.getElementById('logs-table-body');
    if (tbody) {
        tbody.innerHTML = '';
        for (let i = 0; i < 5; i++) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><div class="skeleton skeleton-text" style="width: 120px;"></div></td>
            <td>
            <div class="user-cell">
                <div class="skeleton skeleton-avatar"></div>
                <span class="user-name skeleton skeleton-text" style="width: 80px;"></span>
            </div>
            </td>
            <td><div class="skeleton skeleton-text" style="width: 150px;"></div></td>
            <td><div class="skeleton skeleton-text" style="width: 200px;"></div></td>
            <td><div class="skeleton skeleton-text" style="width: 100px;"></div></td>
        `;
        tbody.appendChild(row);
        }
    }
    }
}

function renderStats(stats) {
    const statsContainer = document.createElement('div');
    statsContainer.innerHTML = `
    <h2 class="section-title">
        <i class="fas fa-chart-line"></i>
        Statistiques <span>globales</span>
    </h2>
    <div class="stats-cards">
        <div class="stat-card">
        <h3><i class="fas fa-users"></i> Utilisateurs totaux</h3>
        <div class="value">${stats.totalUsers.toLocaleString()}</div>
        <div class="change ${stats.userGrowth >= 0 ? 'positive' : 'negative'}">
            <i class="fas fa-${stats.userGrowth >= 0 ? 'arrow-up' : 'arrow-down'}"></i>
            ${Math.abs(stats.userGrowth)}% ce mois-ci
        </div>
        </div>
        <div class="stat-card">
        <h3><i class="fas fa-user-plus"></i> Nouveaux utilisateurs</h3>
        <div class="value">${stats.newUsers.toLocaleString()}</div>
        <div class="change ${stats.newUsersGrowth >= 0 ? 'positive' : 'negative'}">
            <i class="fas fa-${stats.newUsersGrowth >= 0 ? 'arrow-up' : 'arrow-down'}"></i>
            ${Math.abs(stats.newUsersGrowth)}% cette semaine
        </div>
        </div>
        <div class="stat-card">
        <h3><i class="fas fa-user-slash"></i> Utilisateurs bannis</h3>
        <div class="value">${stats.bannedUsers.toLocaleString()}</div>
        <div class="change ${stats.bannedUsersChange >= 0 ? 'positive' : 'negative'}">
            <i class="fas fa-${stats.bannedUsersChange >= 0 ? 'arrow-up' : 'arrow-down'}"></i>
            ${Math.abs(stats.bannedUsersChange)}% ce mois-ci
        </div>
        </div>
        <div class="stat-card">
        <h3><i class="fas fa-user-shield"></i> Modérateurs</h3>
        <div class="value">${stats.moderators.toLocaleString()}</div>
        <div class="change neutral">
            <i class="fas fa-equals"></i> stable
        </div>
        </div>
    </div>
    `;
    return statsContainer;
}

function renderUsersTable(users, page = 1, total = 0) {
    const container = document.createElement('div');
    container.innerHTML = `
    <h2 class="section-title">
    <i class="fas fa-users"></i>
    Gestion des <span>utilisateurs</span>
    </h2>
    <div class="admin-table-container">
    <div class="table-header">
    <h3 class="table-title"><i class="fas fa-list"></i> Liste des utilisateurs</h3>
    <div class="search-box">
        <i class="fas fa-search"></i>
        <input type="text" placeholder="Rechercher un utilisateur..." id="user-search">
    </div>
    </div>
    <table class="admin-table">
    <thead>
        <tr>
        <th>Utilisateur</th>
        <th>Email</th>
        <th>Inscription</th>
        <th>2FA</th>
        <th>Status</th>
        <th>Actions</th>
        </tr>
    </thead>
    <tbody id="users-table-body">
        <!-- Users will be loaded here via JS -->
    </tbody>
    </table>
    <div class="pagination" id="users-pagination"></div>
    </div>
    `;
    
    const tbody = container.querySelector('#users-table-body');
    
    users.forEach(user => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
    <td>
    <div class="user-cell">
        <img src="${user.avatar_url || '/assets/account.png'}" alt="${user.username}" class="user-avatar-sm">
        <div class="user-info">
        <span class="user-name" style="
        ${user.role === 'admin' ? 'color:#ef4444;font-weight:600;' : 
        user.role === 'moderator' ? 'color:#f59e0b;font-weight:600;' : ''}
        ">
        ${user.username}
        ${user.role === 'admin' ? ' (Admin)' : user.role === 'moderator' ? ' (Mod)' : ''}
        ${user.is_verified === true ? '<img src="/assets/verified.png" alt="Vérifié" title="Compte vérifié" style="width: 15px;height: 15px;vertical-align:middle;margin-bottom: 2px;">' : ''}
        </span>
        <span class="user-email">@${user.username}</span>
        </div>
    </div>
    </td>
    <td>${user.email}</td>
    <td>${new Date(user.created_at).toLocaleDateString()}</td>
    <td>${user.two_factor_enabled ? 'Oui' : 'Non'}</td>
    <td>
    <span class="status-badge 
        ${user.status === 'Good' ? 'status-active' : 
        user.status === 'Limited' ? 'status-pending' : 
        user.status === 'Very Limited' ? 'status-pending' : 
        user.status === 'At Risk' ? 'status-pending' : 
        user.status === 'Banned' ? 'status-banned' : 'status-pending'}">
        <i class="fas fa-${
        user.status === 'Good' ? 'check-circle' : 
        user.status === 'Limited' ? 'exclamation-triangle' : 
        user.status === 'Very Limited' ? 'exclamation-triangle' : 
        user.status === 'At Risk' ? 'exclamation-circle' : 
        user.status === 'Banned' ? 'ban' : 'clock'
        }"></i>
        ${
        user.status === 'Good' ? 'Actif' : 
        user.status === 'Limited' ? 'Limité' : 
        user.status === 'Very Limited' ? 'Très limité' : 
        user.status === 'At Risk' ? 'À risque' : 
        user.status === 'Banned' ? 'Banni' : 'En attente'
        }
    </span>
    </td>
    <td>
    <button class="action-btn view" data-user-id="${user.user_id}" title="Voir">
        <i class="fas fa-eye"></i>
    </button>
    <button class="action-btn edit" data-user-id="${user.user_id}" title="Modifier">
        <i class="fas fa-edit"></i>
    </button>
    ${user.status === 'Banned' ? `
    <button class="action-btn delete" data-user-id="${user.user_id}" data-action="unban" title="Débannir">
        <i class="fas fa-unlock"></i>
    </button>
    ` : `
    <button class="action-btn delete" data-user-id="${user.user_id}" data-action="ban" title="Bannir">
        <i class="fas fa-ban"></i>
    </button>
    `}
    </td>
    `;
    tbody.appendChild(tr);
    });
    
    container.querySelectorAll('.action-btn.view, .action-btn.edit').forEach(btn => {
    btn.addEventListener('click', async () => {
    const userId = btn.getAttribute('data-user-id');
    openUserModal(userId);
    });
    });
    
    container.querySelectorAll('.action-btn.delete').forEach(btn => {
    btn.addEventListener('click', () => {
    currentUserId = btn.getAttribute('data-user-id');
    const action = btn.getAttribute('data-action');
    openBanModal(action === 'unban');
    });
    });
    
    const searchInput = container.querySelector('#user-search');
    let searchTimeout;

    searchInput.addEventListener('input', () => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
    loadUsers(searchInput.value, 1);
    }, 500);
    });

    const pag = container.querySelector('#users-pagination');
    const maxPage = Math.max(1, Math.ceil(total / 10));
    pag.innerHTML = `
    <button class="page-btn" ${page<=1?'disabled':''} data-page="${page-1}"><i class="fas fa-chevron-left"></i></button>
    <span class="page-info">Page ${page}/${maxPage}</span>
    <button class="page-btn" ${page>=maxPage?'disabled':''} data-page="${page+1}"><i class="fas fa-chevron-right"></i></button>`;
    pag.querySelectorAll('button').forEach(btn=>{
    btn.addEventListener('click',()=>{
        const p=parseInt(btn.getAttribute('data-page')); if(p>=1&&p<=maxPage){ loadUsers(searchInput.value,p); }
    });
    });
    
    return container;
}

function renderSystemLogs(logs, page = 1, total = 0) {
    const container = document.createElement('div');
    container.innerHTML = `
    <h2 class="section-title">
        <i class="fas fa-clipboard-list"></i>
        Logs <span>système</span>
    </h2>
    <div class="admin-table-container">
        <div class="table-header">
        <h3 class="table-title"><i class="fas fa-history"></i> Activité récente</h3>
        <div class="header-actions">
            <div class="search-box">
            <i class="fas fa-search"></i>
            <input type="text" placeholder="Rechercher dans les logs..." id="logs-search">
            </div>
            <button class="btn btn-danger btn-sm" id="clear-logs-btn" title="Vider les logs"><i class="fas fa-trash"></i></button>
        </div>
        </div>
        <table class="admin-table">
        <thead>
            <tr>
            <th>Date</th>
            <th>Utilisateur</th>
            <th>Action</th>
            <th>Détails</th>
            <th>IP</th>
            </tr>
        </thead>
        <tbody id="logs-table-body">
            <!-- Logs will be loaded here via JS -->
        </tbody>
        </table>
        <div class="pagination" id="logs-pagination"></div>
    </div>
    `;
    
    const tbody = container.querySelector('#logs-table-body');
    
    logs.forEach(log => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
        <td>${new Date(log.timestamp).toLocaleString()}</td>
        <td>
        ${log.user ? `
        <div class="user-cell">
            <img src="${log.user.avatar || '/assets/account.png'}" alt="${log.user.username}" class="user-avatar-sm">
            <span class="user-name">${log.user.username}</span>
        </div>
        ` : 'Système'}
        </td>
        <td>${log.action}</td>
        <td>${log.details || ''}</td>
        <td>${log.ip || 'N/A'}</td>
    `;
    tbody.appendChild(tr);
    });
    
    const searchInput = container.querySelector('#logs-search');
    let searchTimeout;

    searchInput.addEventListener('input', () => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
        loadSystemLogs(1);
    }, 500);
    });

    container.querySelector('#clear-logs-btn').addEventListener('click', clearLogs);

    const pag = container.querySelector('#logs-pagination');
    const maxPage = Math.max(1, Math.ceil(total / 10));
    pag.innerHTML = `
    <button class="page-btn" ${page<=1?'disabled':''} data-page="${page-1}"><i class="fas fa-chevron-left"></i></button>
    <span class="page-info">Page ${page}/${maxPage}</span>
    <button class="page-btn" ${page>=maxPage?'disabled':''} data-page="${page+1}"><i class="fas fa-chevron-right"></i></button>`;
    pag.querySelectorAll('button').forEach(btn=>{
    btn.addEventListener('click',()=>{
        const p=parseInt(btn.getAttribute('data-page')); if(p>=1&&p<=maxPage){ loadSystemLogs(p); }
    });
    });
    
    return container;
}

function renderUserActivity(logs, page = 1, total = 0) {
    const container = document.createElement('div');
    if (!logs.length) {
    container.innerHTML = '<p>Aucune activité récente.</p>';
    return container;
    }
    const list = document.createElement('ul');
    list.className = 'activity-list';
    logs.forEach(log => {
    const li = document.createElement('li');
    li.className = 'activity-item';
    li.innerHTML = `
        <span class="activity-date">${new Date(log.timestamp).toLocaleString()}</span>
        <span class="activity-action">${log.action}${log.message ? ' - ' + log.message : ''}</span>
        <span class="${log.success ? 'activity-success' : 'activity-fail'}">
        <i class="fas fa-${log.success ? 'check-circle' : 'times-circle'}"></i>
        </span>`;
    list.appendChild(li);
    });
    container.appendChild(list);
    const maxPage = Math.max(1, Math.ceil(total / 10));
    const pag = document.createElement('div');
    pag.className = 'pagination';
    pag.innerHTML = `
    <button class="page-btn" ${page<=1?'disabled':''} data-page="${page-1}"><i class="fas fa-chevron-left"></i></button>
    <span class="page-info">Page ${page}/${maxPage}</span>
    <button class="page-btn" ${page>=maxPage?'disabled':''} data-page="${page+1}"><i class="fas fa-chevron-right"></i></button>`;
    pag.querySelectorAll('button').forEach(btn=>{
    btn.addEventListener('click',()=>{
        const p=parseInt(btn.getAttribute('data-page')); if(p>=1&&p<=maxPage){ loadActivityPage(p); }
    });
    });
    container.appendChild(pag);
    return container;
}

async function loadActivityPage(page) {
    if (!currentUserId) return;
    const activityTab = document.getElementById('activity-tab');
    activityTab.innerHTML = '<div class="skeleton" style="height: 300px; border-radius: 8px;"></div>';
    const data = await fetchUserActivity(currentUserId, page);
    currentActivityPage = page;
    totalActivity = data.total || 0;
    activityTab.innerHTML = '';
    activityTab.appendChild(renderUserActivity(data.logs || [], page, totalActivity));
}

function renderUserTools(tools) {
    const container = document.createElement('div');
    if (!tools.length) {
    container.innerHTML = '<p>Aucun outil.</p>';
    return container;
    }
    const grid = document.createElement('div');
    grid.className = 'admin-tools-grid';
    tools.forEach(t => {
    const card = document.createElement('div');
    card.className = 'admin-tool-card';
    const imgUrl = t.thumbnail_url || '/assets/default-tool.png';
    card.innerHTML = `
        <img src="${imgUrl}" alt="${t.title}">
        <div class="tool-title">${t.title}</div>
        <div class="tool-status">${t.status}</div>
    `;
    grid.appendChild(card);
    });
    container.appendChild(grid);
    return container;
}

function renderAccessDenied() {
    const container = document.createElement('div');
    container.className = 'access-denied';
    container.innerHTML = `
    <i class="fas fa-ban"></i>
    <h2>Accès refusé</h2>
    <p>Vous n'avez pas les permissions nécessaires pour accéder à cette section. Seuls les administrateurs peuvent accéder au panel admin.</p>
    <a href="/" class="btn btn-primary">
        <i class="fas fa-home"></i> Retour à l'accueil
    </a>
    `;
    return container;
}

async function openUserModal(userId) {
    currentUserId = userId;
    const modal = document.getElementById('user-modal');
    modal.classList.add('active');
    
    document.getElementById('modal-user-name').textContent = 'Chargement...';
    document.getElementById('modal-avatar').src = '/assets/account.png';
    
    const data = await fetchUserDetails(userId);
    if (!data || !data.user) {
    console.error('Failed to load user details');
    return;
    }
    const user = data.user;
    
    document.getElementById('modal-user-name').textContent = user.username;
    const verifiedIcon = document.getElementById('modal-verified');
    if (user.is_verified) {
    verifiedIcon.style.display = 'inline-block';
    } else {
    verifiedIcon.style.display = 'none';
    }
    document.getElementById('modal-avatar').src = user.avatar || '/assets/account.png';

    const banButton = document.getElementById('ban-user');
    if (user.status === 'Banned') {
    banButton.innerHTML = '<i class="fas fa-undo"></i> Débannir';
    } else {
    banButton.innerHTML = '<i class="fas fa-ban"></i> Bannir';
    }
    
    const roleBadge = document.getElementById('modal-user-role');
    roleBadge.textContent =
    user.role === 'Admin' ? 'Administrateur' :
    user.role === 'Moderator' ? 'Modérateur' : 'Utilisateur';
    roleBadge.className =
    `user-role ${user.role === 'Admin' ? 'role-admin' :
    user.role === 'Moderator' ? 'role-moderator' : 'role-user'}`;
    
    document.getElementById('tools-count').textContent = user.toolsCount || 0;
    document.getElementById('reports-count').textContent = user.reportsCount || 0;
    document.getElementById('joined-date').textContent = new Date(user.createdAt).toLocaleDateString();
    const banInfo = document.getElementById('ban-info');
    if (user.status === 'Banned') {
    let text = '';
    let intervalId = null;

    function updateBanCountdown() {
        if (user.ban_permanent) {
    text = 'Banni définitivement';
    banInfo.innerHTML = `<span id="ban-text">${text}</span> <button class="btn btn-warning btn-sm" id="view-ban-reason">Voir la raison du bannissement</button>`;
    clearInterval(intervalId);
        } else if (user.ban_until) {
    const end = new Date(user.ban_until);
    const now = Date.now();
    const diff = Math.max(0, end - now);
    if (diff <= 0) {
        text = 'Ban terminé';
        clearInterval(intervalId);
    } else {
        const totalSeconds = Math.floor(diff / 1000);
        const hours = Math.floor(totalSeconds / 3600);
        const minutes = Math.floor((totalSeconds % 3600) / 60);
        const seconds = totalSeconds % 60;
        text = `Durée de ban restant : ${hours}h ${minutes}m ${seconds}s`;
    }
    banInfo.innerHTML = `<span id="ban-text">${text}</span> <button class="btn btn-warning btn-sm" id="view-ban-reason">Voir la raison du bannissement</button>`;
        }
        // Ajoute l'event listener à chaque update
        const btn = document.getElementById('view-ban-reason');
        if (btn) {
    btn.onclick = async () => {
        const res = await fetchBanInfo(userId);
        if (res && res.ban && res.ban.reason) {
        showToast('info', `Raison du bannissement de ${user.username} : ${res.ban.reason}`);
        }
    };
        }
    }

    updateBanCountdown();
    if (!user.ban_permanent && user.ban_until) {
        intervalId = setInterval(updateBanCountdown, 1000);
    }
    banInfo.style.display = 'block';
    } else {
    banInfo.style.display = 'none';
    }
    
    document.getElementById('username').value = user.username;
    document.getElementById('email').value = user.email;
    const bioField = document.getElementById('bio');
    if (user.bio) {
    bioField.value = user.bio;
    bioField.placeholder = '';
    } else {
    bioField.value = '';
    bioField.placeholder = 'Aucune bio';
    }
    document.getElementById('user-role').value = user.role || 'User';
    document.getElementById('account-status').value = user.status || 'Good';

    const activityTab = document.getElementById('activity-tab');
    activityTab.innerHTML = '<div class="skeleton" style="height: 300px; border-radius: 8px;"></div>';
    const activityData = await fetchUserActivity(userId, 1);
    totalActivity = activityData.total || 0;
    currentActivityPage = 1;
    activityTab.innerHTML = '';
    activityTab.appendChild(renderUserActivity(activityData.logs || [], 1, totalActivity));

    const toolsTab = document.getElementById('tools-tab');
    toolsTab.innerHTML = '<div class="skeleton" style="height: 300px; border-radius: 8px;"></div>';
    const toolsData = await fetchUserTools(userId);
    toolsTab.innerHTML = '';
    toolsTab.appendChild(renderUserTools(toolsData.tools || []));
}

function closeUserModal() {
    document.getElementById('user-modal').classList.remove('active');
    currentUserId = null;
}

function setupTabs() {
    document.querySelectorAll('.user-tab').forEach(tab => {
    tab.addEventListener('click', async () => {
        document.querySelectorAll('.user-tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        
        tab.classList.add('active');
        const tabId = tab.getAttribute('data-tab');
        document.getElementById(`${tabId}-tab`).classList.add('active');
        if (tabId === 'activity') {
        loadActivityPage(currentActivityPage);
        } else if (tabId === 'tools') {
        if (!currentUserId) return;
        const toolsTab = document.getElementById('tools-tab');
        toolsTab.innerHTML = '<div class="skeleton" style="height: 300px; border-radius: 8px;"></div>';
        const toolsData = await fetchUserTools(currentUserId);
        toolsTab.innerHTML = '';
        toolsTab.appendChild(renderUserTools(toolsData.tools || []));
        }
    });
    });
}

async function loadStats() {
    const stats = await fetchStats();
    if (stats) {
    const statsElement = renderStats(stats);
    adminContent.appendChild(statsElement);
    }
}

async function loadUsers(search = '', page = 1) {
    const existing = document.getElementById(SECTION_IDS.users);
    if (existing) existing.remove();

    const placeholder = renderUsersTable([]);
    placeholder.id = SECTION_IDS.users;
    adminContent.appendChild(placeholder);
    showLoadingSkeleton('users');

    const { users, total } = await fetchUsers(search, page);
    const usersElement = renderUsersTable(users, page, total);
    usersElement.id = SECTION_IDS.users;

    if (!document.querySelector('.stats-cards')) {
    await loadStats();
    }

    placeholder.replaceWith(usersElement);
}

async function loadSystemLogs(page = 1) {
    const existing = document.getElementById(SECTION_IDS.logs);
    if (existing) existing.remove();

    const placeholder = renderSystemLogs([]);
    placeholder.id = SECTION_IDS.logs;
    adminContent.appendChild(placeholder);
    showLoadingSkeleton('logs');

    const { logs, total } = await fetchSystemLogs(page);
    const logsElement = renderSystemLogs(logs, page, total);
    logsElement.id = SECTION_IDS.logs;

    if (!document.querySelector('.stats-cards')) {
    await loadStats();
    }
    if (!document.getElementById(SECTION_IDS.users)) {
    await loadUsers();
    }

    placeholder.replaceWith(logsElement);
}

function navigate(path) {
    history.pushState({}, '', path);
    if (path.endsWith('/logs')) {
    loadSystemLogs();
    } else {
    loadUsers();
    }
}

window.addEventListener('popstate', () => {
    if (location.pathname.endsWith('/logs')) {
    loadSystemLogs();
    } else {
    loadUsers();
    }
});

async function initAdminPanel() {
    const isAdmin = await checkAdminPermissions();
    if (!isAdmin) {
    preloader.style.display = 'none';
    adminContent.appendChild(renderAccessDenied());
    return;
    }
    
    await loadStats();
    await loadUsers();
    await loadSystemLogs();
    
    setTimeout(() => {
    preloader.style.opacity = '0';
    setTimeout(() => {
        preloader.style.display = 'none';
    }, 500);
    }, 500);
    
    setupTabs();

    document.querySelectorAll('.sidebar-item[data-path]').forEach(link => {
    link.addEventListener('click', e => {
        e.preventDefault();
        document.querySelectorAll('.sidebar-item').forEach(a => a.classList.remove('active'));
        link.classList.add('active');
        navigate(link.getAttribute('data-path'));
    });
    });
    
    document.getElementById('close-modal').addEventListener('click', closeUserModal);
    document.getElementById('cancel-changes').addEventListener('click', closeUserModal);
    
    document.getElementById('save-changes').addEventListener('click', async () => {
    if (!currentUserId) return;

    const userData = {
        username: document.getElementById('username').value,
        email: document.getElementById('email').value,
        bio: document.getElementById('bio').value,
        role: document.getElementById('user-role').value,
        status: document.getElementById('account-status').value
    };

    const result = await updateUser(currentUserId, userData);
    if (result) {
        closeUserModal();
        await loadUsers();
    }
    });
    
    const banModal = document.getElementById('ban-modal');
    const closeBanModal = document.getElementById('close-ban-modal');
    const cancelBan = document.getElementById('cancel-ban');
    const confirmBan = document.getElementById('confirm-ban');
    const banReasonInput = document.getElementById('ban-reason');
    const unbanText = document.getElementById('unban-text');
    const banReasonGroup = document.querySelector('.ban-reason-group');

    function openBanModal(isUnban) {
    banModal.classList.add('active');
    if (isUnban) {
        document.getElementById('ban-modal-title').textContent = 'Débannir l\'utilisateur';
        banReasonGroup.style.display = 'none';
        unbanText.classList.remove('hidden');
        confirmBan.textContent = 'Débannir';
    } else {
        document.getElementById('ban-modal-title').textContent = 'Bannir l\'utilisateur';
        banReasonGroup.style.display = 'block';
        unbanText.classList.add('hidden');
        confirmBan.textContent = 'Bannir';
        banReasonInput.value = '';
    }
    }

    function closeBan() {
    banModal.classList.remove('active');
    banReasonInput.value = '';
    document.getElementById('ban-duration').value = '24';
    }

    closeBanModal.addEventListener('click', closeBan);
    cancelBan.addEventListener('click', closeBan);

    document.getElementById('ban-user').addEventListener('click', () => {
    if (!currentUserId) return;
    const status = document.getElementById('account-status').value;
    openBanModal(status === 'Banned');
    });

    confirmBan.addEventListener('click', async () => {
    if (!currentUserId) return;
    const status = document.getElementById('account-status').value;
    confirmBan.disabled = true;
    const spinner = document.createElement('span');
    spinner.className = 'spinner spinner-btn';
    confirmBan.prepend(spinner);
    let result = null;
    if (status === 'Banned') {
        result = await unbanUser(currentUserId);
    } else {
        const reason = banReasonInput.value.trim();
        if (!reason) {
        showToast('error', 'Vous devez préciser une raison');
        confirmBan.disabled = false;
        spinner.remove();
        return;
        }
        result = await banUser(currentUserId, reason);
    }
    if (result) {
        closeBan();
        closeUserModal();
        await loadUsers();
    }
    confirmBan.disabled = false;
    spinner.remove();
    });
}

document.addEventListener('DOMContentLoaded', initAdminPanel);