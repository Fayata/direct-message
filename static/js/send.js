let sentToday = 0;
let toastTimer;

function showToast(msg, type = 'success') {
  const t = document.getElementById('toast');
  document.getElementById('toastMsg').textContent = msg;
  t.className = 'toast ' + type + ' show';
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => t.className = 'toast ' + type, 2800);
}

function sendMessage() {
  const code = document.getElementById('countryCode').value;
  let phone = document.getElementById('phoneNumber').value.replace(/\D/g, '');
  const fullMsg = buildMessageWithLink().trim();

  if (!phone) { showToast('Masukkan nomor tujuan', 'error'); return; }
  if (!fullMsg) { showToast('Pesan tidak boleh kosong', 'error'); return; }

  if (phone.startsWith('0')) phone = phone.slice(1);

  const fullPhone = code + phone;
  const encoded = encodeURIComponent(fullMsg);
  const url = `https://wa.me/${fullPhone}?text=${encoded}`;

  window.open(url, '_blank', 'noopener');

  fetch('/api/send', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json; charset=utf-8',
    },
    body: JSON.stringify({
      phone_number: fullPhone,
      message: fullMsg,
    }),
  }).catch(() => {});

  sentToday++;
  document.getElementById('sentCount').textContent = sentToday;
  showToast('Membuka WhatsApp...', 'success');
}

// init
renderEmojis();
renderTemplates();

