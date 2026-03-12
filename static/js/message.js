function wrapText(before, after) {
  const ta = document.getElementById('message');
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  const sel = ta.value.slice(start, end);
  ta.value = ta.value.slice(0, start) + before + (sel || 'teks') + after + ta.value.slice(end);
  ta.selectionStart = start + before.length;
  ta.selectionEnd = start + before.length + (sel || 'teks').length;
  ta.focus();
  updatePreview();
  updateCharCount();
}

function insertText(txt) {
  const ta = document.getElementById('message');
  const start = ta.selectionStart;
  ta.value = ta.value.slice(0, start) + txt + ta.value.slice(start);
  ta.selectionStart = ta.selectionEnd = start + txt.length;
  ta.focus();
  updatePreview();
  updateCharCount();
}

function clearMsg() {
  document.getElementById('message').value = '';
  updatePreview();
  updateCharCount();
}

function buildMessageWithLink() {
  let msg = document.getElementById('message').value;
  const useForm = document.getElementById('useFormLink')?.checked;
  if (useForm) {
    const custom = document.getElementById('formLinkText')?.value.trim();
    const base = (typeof window !== 'undefined' ? window.location.origin : '');
    const link = base + '/form';
    const prefix = custom || 'Isi data diri di sini:';
    msg = msg + '\n\n' + prefix + ' ' + link;
  }
  return msg;
}

function updatePreview() {
  const msg = buildMessageWithLink();
  if (!msg.trim()) {
    document.getElementById('preview').innerHTML = '<span style="color:var(--text-dim)">Pesan Anda akan tampil di sini...</span>';
    return;
  }
  let html = msg
    .replace(/&/g,'&amp;')
    .replace(/</g,'&lt;')
    .replace(/>/g,'&gt;')
    .replace(/\*([^*\n]+)\*/g,'<strong>$1</strong>')
    .replace(/_([^_\n]+)_/g,'<em>$1</em>')
    .replace(/~([^~\n]+)~/g,'<span style="text-decoration:line-through;opacity:0.7">$1</span>')
    .replace(/```([\s\S]*?)```/g,'<code style="background:var(--bg);padding:2px 6px;border-radius:3px;font-family:var(--mono)">$1</code>')
    .replace(/\n/g,'<br>');
  document.getElementById('preview').innerHTML = html;
}

function updateCharCount() {
  const len = document.getElementById('message').value.length;
  document.getElementById('charCount').textContent = len;
  const pct = Math.min(len / 1000 * 100, 100);
  const bar = document.getElementById('charBar');
  bar.style.width = pct + '%';
  bar.style.background = pct > 80 ? 'var(--danger)' : pct > 60 ? 'var(--warn)' : 'var(--accent)';
}

