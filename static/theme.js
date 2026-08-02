function toggleTheme() {
    var t = document.documentElement.getAttribute('data-bs-theme'),
        n = t === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-bs-theme', n);
    localStorage.setItem('adms-theme', n);
    updateToggle(n);
}
function updateToggle(t) {
    var b = document.querySelector('.theme-toggle');
    if (b) b.textContent = t === 'dark' ? '☀️' : '🌙';
}
document.addEventListener('DOMContentLoaded', function () {
    updateToggle(document.documentElement.getAttribute('data-bs-theme') || 'dark');
});
