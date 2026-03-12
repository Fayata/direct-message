const THEME_KEY = 'wa_theme';

function setTheme(theme) {
  document.body.dataset.theme = theme;
  const btn = document.getElementById('themeToggle');
  if (btn) {
    btn.textContent = theme === 'light' ? 'Light' : 'Dark';
  }
}

const storedTheme = localStorage.getItem(THEME_KEY) || 'dark';
setTheme(storedTheme);

function toggleTheme() {
  const next = document.body.dataset.theme === 'light' ? 'dark' : 'light';
  localStorage.setItem(THEME_KEY, next);
  setTheme(next);
}

