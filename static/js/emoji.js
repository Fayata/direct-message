const emojiCategories = [
  [
    '\u{1F60A}','\u{1F600}','\u{1F603}','\u{1F604}','\u{1F601}','\u{1F606}','\u{1F605}','\u{1F602}','\u{1F923}',
    '\u{1F970}','\u{1F60D}','\u{1F929}','\u{1F618}','\u{1F617}','\u{1F619}','\u{1F61A}','\u{1F642}','\u{1F917}',
    '\u{1F60C}','\u{1F614}','\u{1F62A}','\u{1F634}','\u{1F61B}','\u{1F61C}','\u{1F61D}','\u{1F924}','\u{1F644}',
    '\u{1F612}','\u{1F610}','\u{1F611}','\u{1F636}','\u{1F60F}','\u{1F623}','\u{1F625}','\u{1F62E}','\u{1F62F}',
    '\u{1F632}','\u{1F635}','\u{1F92F}','\u{1F631}','\u{1F628}','\u{1F630}','\u{1F627}','\u{1F622}','\u{1F62D}',
    '\u{1F613}','\u{1F629}','\u{1F62B}','\u{1F971}','\u{1F620}','\u{1F621}','\u{1F624}','\u{1F47F}'
  ],
  [
    '\u{1F389}','\u{1F38A}','\u{1F388}','\u{1F381}','\u{1F382}','\u{1F380}','\u{1F3C6}','\u{1F947}','\u{1F948}',
    '\u{1F949}','\u{1F396}','\u{1F3C5}','\u{1F397}','\u{1F39F}','\u{1F3AA}','\u{1F3AB}','\u{1F3A8}','\u{1F3AD}',
    '\u{1F3AC}','\u{1F3A4}','\u{1F3A7}','\u{1F3B5}','\u{1F3B6}','\u{1F3B7}','\u{1F3B8}','\u{1F3B9}','\u{1F3BA}',
    '\u{1F3BB}','\u{1F941}','\u{1F3AE}','\u{1F579}','\u{1F3B2}','\u{1F0CF}','\u{1F004}','\u{1F3AF}','\u{1F3B3}',
    '\u{1F3BE}','\u{1F3C0}','\u{1F3C8}','\u{1F3C9}','\u{1F6A9}','\u{1F6F9}','\u{1F938}','\u{1F973}','\u{1F64C}',
    '\u{1F44F}','\u{1F932}','\u{1F64F}'
  ],
  [
    '\u{1F4BC}','\u{1F4CA}','\u{1F4C8}','\u{1F4C9}','\u{1F4CB}','\u{1F4CC}','\u{1F4CD}','\u{1F4CE}','\u{1F587}',
    '\u{1F4CF}','\u{1F4D0}','\u{2702}','\u{1F5C3}','\u{1F5C4}','\u{1F5D1}','\u{1F58A}','\u{270F}','\u{1F58B}',
    '\u{1F4DD}','\u{1F4BB}','\u{1F5A5}','\u{1F5A8}','\u{2328}','\u{1F5B1}','\u{1F4F1}','\u{1F4F2}','\u{260E}',
    '\u{1F4DE}','\u{1F4DF}','\u{1F4E0}','\u{1F4FA}','\u{1F4F7}','\u{1F4F8}','\u{1F4FD}','\u{1F3A5}','\u{1F4E1}',
    '\u{1F52D}','\u{1F52C}','\u{1F4A1}','\u{1F526}','\u{1F527}','\u{1F528}','\u{2699}','\u{1F6E0}','\u{1F529}'
  ],
  [
    '\u{2764}','\u{1F9E1}','\u{1F49B}','\u{1F49A}','\u{1F499}','\u{1F49C}','\u{1F5A4}','\u{1F90D}','\u{1F90E}',
    '\u{1F497}','\u{1F493}','\u{1F495}','\u{1F496}','\u{1F49E}','\u{1F49D}','\u{1F494}','\u{2763}','\u{1F48B}',
    '\u{1F48C}','\u{1F48D}','\u{1F48E}','\u{1F48F}','\u{1F491}','\u{1F492}','\u{1F339}','\u{1F338}','\u{1F490}',
    '\u{1F33A}','\u{1F33B}','\u{1F33C}','\u{1F31E}','\u{1F31D}','\u{1F31C}','\u{1F31B}','\u{1F31A}','\u{1F31F}'
  ],
  [
    '\u{1F31F}','\u{2B50}','\u{2728}','\u{1F4AB}','\u{1F320}','\u{1F30C}','\u{1F319}','\u{2600}','\u{1F324}',
    '\u{26C5}','\u{1F325}','\u{2601}','\u{1F326}','\u{1F327}','\u{26C8}','\u{1F329}','\u{1F328}','\u{2744}',
    '\u{2603}','\u{26C4}','\u{1F32C}','\u{1F4A8}','\u{1F4A7}','\u{1F4A6}','\u{1F30A}','\u{1F308}','\u{1F33C}',
    '\u{1F337}','\u{1F331}','\u{1F33F}','\u{1F340}','\u{1F341}','\u{1F342}','\u{1F343}','\u{1F334}','\u{1F335}',
    '\u{1F38B}','\u{1F38D}','\u{1F33E}','\u{1F344}','\u{1F330}','\u{1F98B}','\u{1F41D}'
  ],
  [
    '\u{1F91D}','\u{1F44D}','\u{1F44E}','\u{1F44C}','\u{270C}','\u{1F91E}','\u{1F91F}','\u{1F918}','\u{1F919}',
    '\u{1F448}','\u{1F449}','\u{1F446}','\u{261D}','\u{1F447}','\u{1F44B}','\u{1F91A}','\u{1F590}','\u{270B}',
    '\u{1F596}','\u{1F44F}','\u{1F64C}','\u{1F450}','\u{1F932}','\u{1F64F}','\u{270D}','\u{1F485}','\u{1F933}',
    '\u{1F4AA}','\u{1F9BE}','\u{1F9BF}','\u{1F9B5}','\u{1F9B6}','\u{1F4B0}','\u{1F4B5}','\u{1F4B6}','\u{1F4B7}',
    '\u{1F4B8}','\u{1F4B3}','\u{1F3E6}','\u{1F4E6}','\u{1F4EB}','\u{1F4EC}','\u{1F4ED}','\u{1F4EE}','\u{1F3E2}'
  ],
  [
    '\u{1F436}','\u{1F431}','\u{1F42D}','\u{1F439}','\u{1F430}','\u{1F438}','\u{1F42E}','\u{1F437}','\u{1F43A}',
    '\u{1F43B}','\u{1F428}','\u{1F43C}','\u{1F427}','\u{1F426}','\u{1F424}','\u{1F425}','\u{1F423}','\u{1F414}',
    '\u{1F40D}','\u{1F40E}','\u{1F422}','\u{1F421}','\u{1F41B}','\u{1F41C}','\u{1F41D}','\u{1F41E}','\u{1F40C}',
    '\u{1F419}','\u{1F41A}','\u{1F40B}','\u{1F40A}','\u{1F406}','\u{1F405}','\u{1F403}','\u{1F47E}','\u{1F47D}',
    '\u{1F47B}','\u{1F47A}','\u{1F479}','\u{1F478}','\u{1F475}','\u{1F474}','\u{1F472}','\u{1F46E}','\u{1F46F}',
  ],
];

let activeEmojiCat = 0;

function showEmojiCat(idx, btn) {
  activeEmojiCat = idx;
  document.querySelectorAll('.etab').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  renderEmojis();
}

function toggleEmoji() {
  const container = document.getElementById('emojiContainer');
  const icon = document.getElementById('emojiToggleIcon');
  const isHidden = container.classList.contains('hidden');
  if (isHidden) {
    container.classList.remove('hidden');
    icon.textContent = 'Sembunyikan ▲';
    renderEmojis();
  } else {
    container.classList.add('hidden');
    icon.textContent = 'Tampilkan ▼';
  }
}

function renderEmojis() {
  const grid = document.getElementById('emojiGrid');
  grid.innerHTML = '';
  emojiCategories[activeEmojiCat].forEach(e => {
    const btn = document.createElement('button');
    btn.className = 'emoji-btn';
    btn.textContent = e;
    btn.title = e;
    btn.onclick = () => insertEmoji(e);
    grid.appendChild(btn);
  });
}

function insertEmoji(e) {
  const ta = document.getElementById('message');
  const start = ta.selectionStart;
  const end = ta.selectionEnd;
  const val = ta.value;
  ta.value = val.slice(0, start) + e + val.slice(end);
  ta.selectionStart = ta.selectionEnd = start + [...e].length;
  ta.focus();
  updatePreview();
  updateCharCount();
}

