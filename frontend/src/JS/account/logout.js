const token = localStorage.getItem("token");
if (!token || token === "undefined" || token === "null") {
    window.location.href = "/signin";
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

let apiBaseURL = window.API_BASE_URL;

document.addEventListener('DOMContentLoaded', async () => {
    initTheme();

    document.getElementById('logoutBtn').addEventListener('click', async () => {
        try {
            await fetch(`${apiBaseURL}/auth/logout`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                }
            });
        } catch (e) {
            aleryt("Erreur lors de la déconnexion. Veuillez réessayer.");
        }
        localStorage.removeItem('token');
        window.location.href = '/signin';
    });

    document.getElementById('cancelBtn').addEventListener('click', () => {
        window.history.length > 1 ? window.history.back() : window.location.href = '/account';
    });
});