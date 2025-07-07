const API_BASE = import.meta.env.VITE_API_BASE_URL || 'https://api.tool-center.fr/v2';
const PANEL_CONFIG = {
  API_BASE,
  ADMIN_BASE: API_BASE + '/admin',
  MODERATION_BASE: API_BASE + '/moderation',
  COLORS: {
    primary: '#6366f1',
    success: '#10b981',
    error: '#ef4444',
    warning: '#f59e0b',
    info: '#3b82f6'
  }
};

for (const [name, value] of Object.entries(PANEL_CONFIG.COLORS)) {
  document.documentElement.style.setProperty(`--${name}-color`, value);
}
window.PANEL_CONFIG = PANEL_CONFIG;
