const PANEL_CONFIG = {
  API_BASE: '/api/v1',
  ADMIN_BASE: '/api/v1/admin',
  MODERATION_BASE: '/api/v1/moderation',
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
