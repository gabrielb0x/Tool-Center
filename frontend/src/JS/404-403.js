const themeToggle = document.getElementById('theme-toggle');
const html = document.documentElement;
function setTheme(dark) {
    html.classList.toggle('dark', dark);
}
const currentTheme = localStorage.getItem('theme') || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
setTheme(currentTheme === 'dark');
themeToggle.addEventListener('click', () => {
    const isDark = !html.classList.contains('dark');
    setTheme(isDark);
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
});