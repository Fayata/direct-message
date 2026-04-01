const THEME_KEY = 'wa_theme';

function setTheme(theme) {
  if (!document.body) return;
  document.body.dataset.theme = theme;
  const btn = document.getElementById('themeToggle');
  if (btn) {
    btn.textContent = theme === 'light' ? 'Light' : 'Dark';
  }
}

const storedTheme = localStorage.getItem(THEME_KEY) || 'dark';

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    setTheme(storedTheme);
    const btn = document.getElementById('themeToggle');
    if (btn) btn.addEventListener('click', toggleTheme);
  });
} else {
  setTheme(storedTheme);
  const btn = document.getElementById('themeToggle');
  if (btn) btn.addEventListener('click', toggleTheme);
}

function toggleTheme() {
  const current = document.body?.dataset.theme || storedTheme || 'dark';
  const next = current === 'light' ? 'dark' : 'light';
  localStorage.setItem(THEME_KEY, next);
  setTheme(next);
}


