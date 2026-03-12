let templates = JSON.parse(localStorage.getItem('wa_templates') || '[]');

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function renderTemplates() {
  const list = document.getElementById('templateList');
  const empty = document.getElementById('tplEmpty');
  const count = templates.length;
  document.getElementById('tplCount').textContent = count + ' saved';
  document.getElementById('tplCountVal').textContent = count;

  list.innerHTML = '';
  if (count === 0) {
    empty.style.display = 'block';
    return;
  }
  empty.style.display = 'none';

  templates.forEach((tpl, i) => {
    const card = document.createElement('div');
    card.className = 'tpl-card';
    card.innerHTML = `
      <div class="tpl-card-name">
        <span>${escHtml(tpl.name)}</span>
        <div class="tpl-actions">
          <button class="tpl-act-btn" onclick="loadTemplate(${i})" title="Load">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="5 12 19 12"/><polyline points="12 5 19 12 12 19"/></svg>
          </button>
          <button class="tpl-act-btn del" onclick="deleteTemplate(${i})" title="Delete">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/></svg>
          </button>
        </div>
      </div>
      <div class="tpl-card-preview">${escHtml(tpl.body.slice(0, 60))}${tpl.body.length > 60 ? '…' : ''}</div>
    `;
    list.appendChild(card);
  });
}

function saveTemplate() {
  const name = document.getElementById('tplName').value.trim();
  const body = document.getElementById('message').value.trim();
  if (!name) { showToast('Masukkan nama template', 'error'); return; }
  if (!body) { showToast('Pesan tidak boleh kosong', 'error'); return; }
  templates.unshift({ name, body });
  localStorage.setItem('wa_templates', JSON.stringify(templates));
  document.getElementById('tplName').value = '';
  renderTemplates();
  showToast('Template "' + name + '" tersimpan', 'success');
}

function loadTemplate(i) {
  const tpl = templates[i];
  document.getElementById('message').value = tpl.body;
  updatePreview();
  updateCharCount();
  showToast('Template "' + tpl.name + '" dimuat', 'success');
}

function deleteTemplate(i) {
  const name = templates[i].name;
  templates.splice(i, 1);
  localStorage.setItem('wa_templates', JSON.stringify(templates));
  renderTemplates();
  showToast('Template "' + name + '" dihapus', 'error');
}

