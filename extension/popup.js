const API_BASE = 'https://bublet.app';
const BUBBLE_COLORS = ['#d4829a', '#e8b4a0', '#b8a9d4', '#9bc4a8', '#8cb8d4', '#d4a65a', '#d9968a', '#7bb8b0'];

let currentTab = null;
let bubbles = [];

// ── Init ──
document.addEventListener('DOMContentLoaded', async () => {
  // Get current tab info
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  currentTab = tab;

  // Show what we're saving
  document.getElementById('link-title').textContent = tab.title || 'Untitled';
  document.getElementById('link-url').textContent = tab.url;

  // Open bublet website
  document.getElementById('open-site').addEventListener('click', () => {
    chrome.tabs.create({ url: API_BASE + '/dashboard.html' });
    window.close();
  });

  // Load bubbles
  await loadBubbles();
});

async function loadBubbles() {
  try {
    // Check if user is a guest — extension is only for signed-up users
    const meRes = await fetch(API_BASE + '/api/me', { credentials: 'include' });
    if (!meRes.ok) {
      showNotLoggedIn();
      return;
    }
    const user = await meRes.json();
    if (user.is_guest) {
      showNotLoggedIn();
      return;
    }

    const res = await fetch(API_BASE + '/api/bubbles', { credentials: 'include' });
    if (!res.ok) throw new Error('Failed to load bubbles');

    bubbles = await res.json();
    renderBubbleList();
  } catch (err) {
    showNotLoggedIn();
  }
}

function renderBubbleList() {
  const main = document.getElementById('main-content');

  if (!bubbles || bubbles.length === 0) {
    main.innerHTML = `
      <div class="empty-state">no bubbles yet — create one below</div>
      ${newBubbleHTML()}
    `;
    setupNewBubble();
    return;
  }

  main.innerHTML = `
    <div class="search-bar">
      <input type="text" id="search-input" placeholder="search bubbles..." />
    </div>
    <div class="bubble-list" id="bubble-list">
      ${renderBubbleItems(bubbles)}
    </div>
    ${newBubbleHTML()}
  `;

  // Search filtering
  document.getElementById('search-input').addEventListener('input', (e) => {
    const query = e.target.value.toLowerCase();
    const filtered = bubbles.filter(b => b.name.toLowerCase().includes(query));
    document.getElementById('bubble-list').innerHTML = renderBubbleItems(filtered);
    attachBubbleClicks();
  });

  attachBubbleClicks();
  setupNewBubble();
}

function renderBubbleItems(list) {
  return list.map((b, i) => `
    <button class="bubble-item" data-bid="${b.id}">
      <div class="bubble-dot" style="background: ${BUBBLE_COLORS[i % BUBBLE_COLORS.length]}"></div>
      <div class="bubble-info">
        <div class="bubble-name">${escapeHtml(b.name)}</div>
        <div class="bubble-count">${b.item_count} link${b.item_count !== 1 ? 's' : ''}</div>
      </div>
    </button>
  `).join('');
}

function attachBubbleClicks() {
  document.querySelectorAll('.bubble-item').forEach(el => {
    el.addEventListener('click', () => saveLink(el.dataset.bid));
  });
}

function newBubbleHTML() {
  return `
    <div class="new-bubble-section">
      <button class="new-bubble-btn" id="new-bubble-btn">+ new bubble</button>
      <div class="new-bubble-form" id="new-bubble-form">
        <input type="text" id="new-bubble-name" placeholder="bubble name..." />
        <button id="new-bubble-save">save</button>
      </div>
    </div>
  `;
}

function setupNewBubble() {
  const btn = document.getElementById('new-bubble-btn');
  const form = document.getElementById('new-bubble-form');
  const input = document.getElementById('new-bubble-name');
  const save = document.getElementById('new-bubble-save');

  btn.addEventListener('click', () => {
    btn.style.display = 'none';
    form.classList.add('active');
    input.focus();
  });

  save.addEventListener('click', () => createAndSave());
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') createAndSave();
  });

  async function createAndSave() {
    const name = input.value.trim();
    if (!name) return;

    save.disabled = true;
    save.textContent = '...';

    try {
      // Create bubble
      const res = await fetch(API_BASE + '/api/bubbles', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, tags: [] }),
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to create bubble');
      }

      const bubble = await res.json();
      // Save the link to the new bubble
      await saveLink(bubble.id);
    } catch (err) {
      showToast(err.message, 'error');
      save.disabled = false;
      save.textContent = 'save';
    }
  }
}

async function saveLink(bid) {
  const content = `${currentTab.title || 'Untitled'}:- ${currentTab.url}`;

  try {
    const res = await fetch(`${API_BASE}/api/bubbles/${bid}/items`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || 'Failed to save link');
    }

    const name = bubbles.find(b => b.id === bid)?.name || 'new bubble';
    showToast(`saved to "${name}"`, 'success');

    // Update the item count locally so the UI stays fresh
    const bubble = bubbles.find(b => b.id === bid);
    if (bubble) bubble.item_count++;
    const listEl = document.getElementById('bubble-list');
    if (listEl) {
      listEl.innerHTML = renderBubbleItems(bubbles);
      attachBubbleClicks();
    }
  } catch (err) {
    showToast(err.message, 'error');
  }
}

function showNotLoggedIn() {
  document.getElementById('main-content').innerHTML = `
    <div class="not-logged-in">
      <p>visit bublet.app first to get started</p>
      <a href="${API_BASE}" target="_blank">open bublet</a>
    </div>
  `;
}

function showToast(msg, type) {
  const toast = document.getElementById('toast');
  toast.textContent = msg;
  toast.className = 'toast ' + type;
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
