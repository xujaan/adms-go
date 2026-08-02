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
function tickClock() {
    var c = document.getElementById('nav-clock');
    if (!c) return;
    var now = new Date();
    c.textContent = now.toLocaleString('id-ID', { weekday: 'short', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit', timeZoneName: 'short' });
}
document.addEventListener('DOMContentLoaded', function () {
    updateToggle(document.documentElement.getAttribute('data-bs-theme') || 'dark');
    tickClock();
    setInterval(tickClock, 1000);
});
