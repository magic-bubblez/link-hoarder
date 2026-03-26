/* ============================================
   BUBLET — Client-side Logic
   Wired to Go backend API
   ============================================ */

(() => {
    'use strict';

    // ── Constants ──
    const VISITED_KEY = 'bublet_visited';
    const HEADER_KEY = 'bublet_header_index';
    const HEADER_IMAGES = ['abstract-purple', 'blush', 'redflowers', 'whitebunny'];
    const BUBBLE_VARIANTS = ['variant-rose', 'variant-peach', 'variant-lavender', 'variant-sage', 'variant-sky', 'variant-gold', 'variant-coral', 'variant-teal'];
    const BLOB_COUNT = 12;

    // ── View Mode Detection ──
    const slugMatch = window.location.pathname.match(/^\/b\/([a-z0-9]+)$/);
    const viewMode = !!slugMatch;
    const viewSlug = slugMatch ? slugMatch[1] : null;

    // ── State ──
    let bubbles = [];
    let currentItems = [];
    let currentDetailBubble = null;
    let currentDetailItems = [];
    let isGuest = true;
    let publicSlug = null;

    // ── API Client ──
    async function apiCheck(res, fallback) {
        if (!res.ok) {
            let msg = fallback;
            try {
                const body = await res.json();
                if (body.error) msg = body.error;
            } catch { }
            throw new Error(msg);
        }
    }

    const API = {
        async me() {
            const res = await fetch('/api/me');
            await apiCheck(res, 'failed to get user info');
            return res.json();
        },
        async getBubbles() {
            const res = await fetch('/api/bubbles');
            await apiCheck(res, 'failed to load bubbles');
            return res.json();
        },
        async createBubble(name, tags) {
            const res = await fetch('/api/bubbles', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, tags }),
            });
            await apiCheck(res, 'failed to create bubble');
            return res.json();
        },
        async deleteBubble(bid) {
            const res = await fetch(`/api/bubbles/${bid}`, { method: 'DELETE' });
            await apiCheck(res, 'failed to delete bubble');
        },
        async updateBubble(bid, name) {
            const res = await fetch(`/api/bubbles/${bid}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name }),
            });
            await apiCheck(res, 'failed to update bubble');
        },
        async getItems(bid) {
            const res = await fetch(`/api/bubbles/${bid}/items`);
            await apiCheck(res, 'failed to load items');
            return res.json();
        },
        async addItem(bid, content) {
            const res = await fetch(`/api/bubbles/${bid}/items`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ content }),
            });
            await apiCheck(res, 'failed to add item');
            return res.json();
        },
        async deleteItem(bid, iid) {
            const res = await fetch(`/api/bubbles/${bid}/items/${iid}`, { method: 'DELETE' });
            await apiCheck(res, 'failed to delete item');
        },
        async getPublicBubbles(slug) {
            const res = await fetch(`/api/public/${slug}`);
            await apiCheck(res, 'page not found');
            return res.json();
        },
        async getPublicItems(slug, bid) {
            const res = await fetch(`/api/public/${slug}/items/${bid}`);
            await apiCheck(res, 'failed to load items');
            return res.json();
        },
        async togglePageVisibility(makePublic) {
            const res = await fetch('/api/visibility', {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ public: makePublic }),
            });
            await apiCheck(res, 'failed to update visibility');
            return res.json();
        },
    };

    // ── Visual Helpers (deterministic from bubble ID) ──
    function hashString(str) {
        let hash = 5381;
        for (let i = 0; i < str.length; i++) {
            hash = ((hash << 5) + hash) + str.charCodeAt(i);
            hash |= 0;
        }
        return Math.abs(hash);
    }

    function getVariantForBubble(id) {
        return BUBBLE_VARIANTS[hashString(id) % BUBBLE_VARIANTS.length];
    }

    function getBlobForBubble(id) {
        return (hashString(id + 'blob') % BLOB_COUNT) + 1;
    }

    // ── DOM Refs ──
    const grid = document.getElementById('bubble-grid');
    const emptyState = document.getElementById('empty-state');
    const bubbleCount = document.getElementById('bubble-count');
    const searchInput = document.getElementById('search-input');
    const tagFilter = document.getElementById('tag-filter');

    // Modals
    const policyModal = document.getElementById('policy-modal');
    const createModal = document.getElementById('create-modal');

    // Detail modal
    const detailModal = document.getElementById('detail-modal');
    const detailTitle = document.getElementById('detail-title');
    const detailItemList = document.getElementById('detail-link-list');
    const btnDetailAdd = document.getElementById('btn-detail-add');
    const btnDetailCopy = document.getElementById('btn-detail-copy');
    const btnDetailDelete = document.getElementById('btn-detail-delete');
    const confirmDeleteBar = document.getElementById('confirm-delete-bar');
    const btnDeleteConfirm = document.getElementById('btn-delete-confirm');
    const btnDeleteCancel = document.getElementById('btn-delete-cancel');

    // Add-item sub-popup
    const addItemSubModal = document.getElementById('add-link-modal');
    const detailItemInput = document.getElementById('detail-link-input');
    const btnDetailAddItem = document.getElementById('btn-detail-add-link');

    // Toast
    const toastEl = document.getElementById('toast');

    // Existing create-modal refs
    const btnDismissPolicy = document.getElementById('btn-dismiss-policy');
    const btnCreateBubble = document.getElementById('btn-create-bubble');
    const btnCancelCreate = document.getElementById('btn-cancel-create');
    const btnAddItem = document.getElementById('btn-add-link');
    const createForm = document.getElementById('create-bubble-form');
    const itemInput = document.getElementById('link-input');
    const itemList = document.getElementById('link-list');

    // Login button
    const loginBtn = document.getElementById('btn-login');

    // Share page toggle
    const btnSharePage = document.getElementById('btn-share-page');

    // ── Init ──
    async function init() {
        localStorage.removeItem(VISITED_KEY);

        if (viewMode) {
            await initViewMode();
        } else {
            await initOwnerMode();
        }
    }

    // ── View Mode (public page at /b/{slug}) ──
    async function initViewMode() {
        // Hide owner-only controls
        if (btnCreateBubble) btnCreateBubble.style.display = 'none';
        if (bubbleCount) bubbleCount.style.display = 'none';
        if (loginBtn) loginBtn.style.display = 'none';
        if (btnSharePage) btnSharePage.style.display = 'none';
        const rotateBtn = document.getElementById('btn-rotate-header');
        if (rotateBtn) rotateBtn.style.display = 'none';
        const backLink = document.querySelector('.header-back');
        if (backLink) backLink.style.display = 'none';

        setHeaderImage();

        // Load public bubbles
        try {
            const data = await API.getPublicBubbles(viewSlug);
            bubbles = data || [];
        } catch {
            bubbles = [];
        }
        renderBubbles();
        updateTagFilter();

        // Search and tag filter still work
        searchInput.addEventListener('input', filterBubbles);
        tagFilter.addEventListener('change', filterBubbles);

        // Detail modal — read-only (only copy button)
        btnDetailCopy.addEventListener('click', copyBubbleItems);

        // Close modals
        [detailModal].forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) closeModal(modal);
            });
        });
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && detailModal.classList.contains('active')) {
                closeModal(detailModal);
            }
        });
    }

    // ── Owner Mode (normal dashboard) ──
    async function initOwnerMode() {
        // Check auth status
        try {
            const user = await API.me();
            isGuest = user.is_guest;
            publicSlug = user.public_slug || null;
        } catch (err) {
            console.error('Could not reach API:', err);
        }

        setHeaderImage();
        updateAuthUI();
        updateShareButton();

        await loadBubblesFromAPI();
        renderBubbles();
        updateTagFilter();

        // First-visit policy popup (per-session for guests)
        if (isGuest && !sessionStorage.getItem(VISITED_KEY)) {
            setTimeout(() => openModal(policyModal), 400);
        }

        // Event listeners
        btnDismissPolicy.addEventListener('click', () => {
            closeModal(policyModal);
            sessionStorage.setItem(VISITED_KEY, 'true');
        });

        btnCreateBubble.addEventListener('click', () => {
            resetCreateForm();
            openModal(createModal);
        });

        btnCancelCreate.addEventListener('click', () => {
            closeModal(createModal);
        });

        btnAddItem.addEventListener('click', addItem);
        itemInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addItem();
            }
        });

        createForm.addEventListener('submit', handleCreateBubble);

        searchInput.addEventListener('input', filterBubbles);
        tagFilter.addEventListener('change', filterBubbles);

        // Detail modal
        btnDetailAdd.addEventListener('click', () => {
            detailItemInput.value = '';
            openModal(addItemSubModal);
        });
        btnDetailCopy.addEventListener('click', copyBubbleItems);
        btnDetailDelete.addEventListener('click', showDeleteConfirm);
        btnDeleteConfirm.addEventListener('click', deleteBubble);
        btnDeleteCancel.addEventListener('click', hideDeleteConfirm);

        // Editable bubble name
        detailTitle.addEventListener('click', startEditBubbleName);
        detailTitle.addEventListener('blur', finishEditBubbleName);
        detailTitle.addEventListener('keydown', (e) => {
            if (detailTitle.contentEditable !== 'true') return;
            if (e.key === 'Enter') {
                e.preventDefault();
                detailTitle.blur();
            }
            if (e.key === 'Escape') {
                e.preventDefault();
                e.stopPropagation();
                detailTitle.textContent = currentDetailBubble.name;
                detailTitle.contentEditable = 'false';
            }
        });

        // Add-item sub-popup
        btnDetailAddItem.addEventListener('click', addDetailItem);
        detailItemInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addDetailItem();
            }
        });

        // Login buttons
        if (loginBtn) {
            loginBtn.addEventListener('click', () => {
                window.location.href = '/auth/google/login';
            });
        }
        const policySignup = document.getElementById('btn-policy-signup');
        if (policySignup) {
            policySignup.addEventListener('click', () => {
                window.location.href = '/auth/google/login';
            });
        }

        // Header image rotate button (signed-in users)
        const rotateBtn = document.getElementById('btn-rotate-header');
        if (rotateBtn) {
            rotateBtn.addEventListener('click', () => {
                rotateHeaderImage();
            });
        }

        // Share page toggle
        if (btnSharePage) {
            btnSharePage.addEventListener('click', togglePageVisibility);
        }

        // Close modals on overlay click
        [policyModal, createModal, detailModal, addItemSubModal].forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) closeModal(modal);
            });
        });

        // Keyboard: Escape to close
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                if (addItemSubModal.classList.contains('active')) { closeModal(addItemSubModal); return; }
                if (detailModal.classList.contains('active')) { closeModal(detailModal); return; }
                if (createModal.classList.contains('active')) closeModal(createModal);
                if (policyModal.classList.contains('active')) {
                    closeModal(policyModal);
                    sessionStorage.setItem(VISITED_KEY, 'true');
                }
            }
        });
    }

    // ── Auth ──
    function updateAuthUI() {
        if (loginBtn) {
            loginBtn.style.display = isGuest ? '' : 'none';
        }
        const policySignup = document.getElementById('btn-policy-signup');
        if (policySignup) {
            policySignup.style.display = isGuest ? '' : 'none';
        }
        const rotateBtn = document.getElementById('btn-rotate-header');
        if (rotateBtn) {
            rotateBtn.style.display = isGuest ? 'none' : '';
        }
    }

    // ── Share Page ──
    function updateShareButton() {
        if (!btnSharePage) return;
        if (isGuest) {
            btnSharePage.style.display = 'none';
            return;
        }
        btnSharePage.style.display = '';
        if (publicSlug) {
            btnSharePage.textContent = 'public ✦';
            btnSharePage.classList.add('is-public');
            btnSharePage.title = 'Click to make private or copy link';
        } else {
            btnSharePage.textContent = 'share page';
            btnSharePage.classList.remove('is-public');
            btnSharePage.title = 'Share your page publicly';
        }
    }

    async function togglePageVisibility() {
        const isPublic = !!publicSlug;

        if (isPublic) {
            // Already public — copy link on first click, show toast with option context
            const url = `${window.location.origin}/b/${publicSlug}`;
            try {
                await navigator.clipboard.writeText(url);
                showToast('link copied! click again to make private');
            } catch {
                showToast('link: ' + url);
            }
            // Switch button to "make private" mode for the next click
            btnSharePage.textContent = 'make private';
            btnSharePage.classList.remove('is-public');
            btnSharePage.onclick = async () => {
                try {
                    await API.togglePageVisibility(false);
                    publicSlug = null;
                    updateShareButton();
                    btnSharePage.onclick = null;
                    btnSharePage.addEventListener('click', togglePageVisibility);
                    showToast('page is now private');
                } catch (err) {
                    showToast(err.message || 'failed to make private');
                }
            };
            return;
        }

        // Not public — make it public
        try {
            const result = await API.togglePageVisibility(true);
            publicSlug = result.public_slug;
            updateShareButton();

            if (publicSlug) {
                const url = `${window.location.origin}/b/${publicSlug}`;
                try {
                    await navigator.clipboard.writeText(url);
                    showToast('page is public! link copied');
                } catch {
                    showToast('page is now public');
                }
            }
        } catch (err) {
            showToast(err.message || 'failed to share page');
        }
    }

    // ── Data Loading ──
    async function loadBubblesFromAPI() {
        try {
            const data = await API.getBubbles();
            bubbles = data || [];
        } catch {
            bubbles = [];
            showToast('failed to load bubbles');
        }
    }

    // ── Modal Helpers ──
    function openModal(modal) {
        modal.classList.add('active');
        document.body.style.overflow = 'hidden';
        const firstInput = modal.querySelector('input');
        if (firstInput) setTimeout(() => firstInput.focus(), 200);
    }

    function closeModal(modal) {
        modal.classList.remove('active');
        if (!document.querySelector('.modal-overlay.active')) {
            document.body.style.overflow = '';
        }
    }

    // ── Header Image ──
    function setHeaderImage() {
        const img = document.getElementById('header-bg-img');
        if (!img) return;
        if (viewMode || isGuest) {
            img.src = 'images/redflowers-1920.webp';
            img.srcset = 'images/redflowers-600.webp 600w, images/redflowers-1200.webp 1200w, images/redflowers-1920.webp 1920w';
            img.sizes = '100vw';
        } else {
            const idx = parseInt(localStorage.getItem(HEADER_KEY) || '0', 10);
            const baseName = HEADER_IMAGES[idx % HEADER_IMAGES.length];
            img.src = `images/${baseName}-1920.webp`;
            img.srcset = `images/${baseName}-600.webp 600w, images/${baseName}-1200.webp 1200w, images/${baseName}-1920.webp 1920w`;
            img.sizes = '100vw';
        }
    }

    function rotateHeaderImage() {
        const img = document.getElementById('header-bg-img');
        if (!img) return;
        let idx = parseInt(localStorage.getItem(HEADER_KEY) || '0', 10);
        idx = (idx + 1) % HEADER_IMAGES.length;
        localStorage.setItem(HEADER_KEY, idx.toString());
        const baseName = HEADER_IMAGES[idx];
        img.src = `images/${baseName}-1920.webp`;
        img.srcset = `images/${baseName}-600.webp 600w, images/${baseName}-1200.webp 1200w, images/${baseName}-1920.webp 1920w`;
        img.sizes = '100vw';
    }

    // ── Detail Modal ──
    async function openDetailModal(bubble) {
        currentDetailBubble = bubble;
        detailTitle.textContent = bubble.name;
        hideDeleteConfirm();

        // In view mode: hide owner controls, disable editing
        if (viewMode) {
            btnDetailAdd.style.display = 'none';
            btnDetailDelete.style.display = 'none';
            detailTitle.style.cursor = 'default';
        } else {
            btnDetailAdd.style.display = '';
            btnDetailDelete.style.display = '';
            detailTitle.style.cursor = 'text';
        }

        detailItemList.innerHTML = '<li class="detail-empty"><p>loading...</p></li>';
        openModal(detailModal);

        try {
            const data = viewMode
                ? await API.getPublicItems(viewSlug, bubble.id)
                : await API.getItems(bubble.id);
            currentDetailItems = data || [];
        } catch {
            currentDetailItems = [];
            showToast('failed to load items');
        }
        renderDetailItems();
    }

    // ── URL Detection ──
    function isUrl(str) {
        return /^https?:\/\/\S+$/.test(str.trim());
    }

    function highlightUrls(text) {
        return escapeHtml(text).replace(
            /https?:\/\/[^\s]+/g,
            url => `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`
        );
    }

    function renderDetailItems() {
        if (!currentDetailBubble) return;
        const items = currentDetailItems;

        if (items.length === 0) {
            const msg = viewMode ? 'nothing here yet' : 'nothing here yet — click + to add a link or note';
            detailItemList.innerHTML = `
                <li class="detail-empty">
                    <span style="font-size:1.4rem; margin-bottom:8px; display:block;">✦</span>
                    <p>${msg}</p>
                </li>
            `;
            return;
        }

        detailItemList.innerHTML = items.map(item => {
            let contentHtml;
            if (isUrl(item.content)) {
                contentHtml = `<a href="${escapeHtml(item.content)}" target="_blank" rel="noopener noreferrer" title="${escapeHtml(item.content)}">${escapeHtml(truncateUrl(item.content))}</a>`;
            } else {
                contentHtml = `<span class="item-text">${highlightUrls(item.content)}</span>`;
            }
            const removeBtn = viewMode ? '' : `<span class="remove-detail-link" data-id="${item.id}" title="Remove item">✕</span>`;
            return `
                <li class="detail-link-item">
                    <span class="link-bullet">${isUrl(item.content) ? '🔗' : '•'}</span>
                    ${contentHtml}
                    ${removeBtn}
                </li>
            `;
        }).join('');

        if (!viewMode) {
            detailItemList.querySelectorAll('.remove-detail-link').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    deleteDetailItem(btn.dataset.id);
                });
            });
        }
    }

    async function deleteDetailItem(itemId) {
        if (!currentDetailBubble) return;
        try {
            await API.deleteItem(currentDetailBubble.id, itemId);
            currentDetailItems = currentDetailItems.filter(i => i.id !== itemId);
            renderDetailItems();
            currentDetailBubble.item_count = Math.max(0, (currentDetailBubble.item_count || 0) - 1);
            updateBubbleInGrid(currentDetailBubble);
        } catch {
            showToast('failed to delete item');
        }
    }

    async function addDetailItem() {
        if (!currentDetailBubble) return;
        const content = detailItemInput.value.trim();
        if (!content) return;

        try {
            const item = await API.addItem(currentDetailBubble.id, content);
            currentDetailItems.unshift(item);
            renderDetailItems();
            currentDetailBubble.item_count = (currentDetailBubble.item_count || 0) + 1;
            updateBubbleInGrid(currentDetailBubble);
            detailItemInput.value = '';
            closeModal(addItemSubModal);
        } catch (err) {
            showToast(err.message || 'failed to add item');
        }
    }

    function copyBubbleItems() {
        if (!currentDetailItems || currentDetailItems.length === 0) {
            showToast('nothing to copy');
            return;
        }
        const text = currentDetailItems.map(item => '• ' + item.content).join('\n');
        navigator.clipboard.writeText(text).then(() => {
            showToast('copied to clipboard!');
        }).catch(() => {
            showToast('failed to copy');
        });
    }

    function showDeleteConfirm() {
        confirmDeleteBar.style.display = 'flex';
    }

    function hideDeleteConfirm() {
        confirmDeleteBar.style.display = 'none';
    }

    async function deleteBubble() {
        if (!currentDetailBubble) return;
        try {
            await API.deleteBubble(currentDetailBubble.id);
            bubbles = bubbles.filter(b => b.id !== currentDetailBubble.id);
            closeModal(detailModal);
            currentDetailBubble = null;
            currentDetailItems = [];
            renderBubbles();
            updateTagFilter();
        } catch {
            showToast('failed to delete bubble');
        }
    }

    // ── Update Single Bubble in Grid ──
    function updateBubbleInGrid(bubble) {
        const el = grid.querySelector(`.bubble[data-id="${bubble.id}"]`);
        if (!el) return;
        const countEl = el.querySelector('.bubble-item-count');
        if (countEl) {
            const itemCount = bubble.item_count || 0;
            countEl.textContent = `${itemCount} item${itemCount !== 1 ? 's' : ''}`;
        }
        const nameEl = el.querySelector('.bubble-name');
        if (nameEl) {
            nameEl.textContent = bubble.name;
        }
        el.title = bubble.name;
    }

    // ── Editable Bubble Name ──
    function startEditBubbleName() {
        if (viewMode) return;
        if (detailTitle.contentEditable === 'true') return;
        detailTitle.contentEditable = 'true';
        detailTitle.focus();
        const range = document.createRange();
        range.selectNodeContents(detailTitle);
        const sel = window.getSelection();
        sel.removeAllRanges();
        sel.addRange(range);
    }

    async function finishEditBubbleName() {
        if (detailTitle.contentEditable !== 'true') return;
        detailTitle.contentEditable = 'false';
        if (!currentDetailBubble) return;
        const newName = detailTitle.textContent.trim();
        if (!newName) {
            detailTitle.textContent = currentDetailBubble.name;
            return;
        }
        if (newName !== currentDetailBubble.name) {
            try {
                await API.updateBubble(currentDetailBubble.id, newName);
                currentDetailBubble.name = newName;
                updateBubbleInGrid(currentDetailBubble);
            } catch {
                showToast('failed to rename bubble');
                detailTitle.textContent = currentDetailBubble.name;
            }
        }
    }

    // ── Toast ──
    function showToast(message) {
        toastEl.textContent = message;
        toastEl.classList.add('show');
        setTimeout(() => {
            toastEl.classList.remove('show');
        }, 1800);
    }

    // ── Item Management (Create Modal) ──
    function addItem() {
        const content = itemInput.value.trim();
        if (!content) return;

        currentItems.push(content);
        renderItemList();
        itemInput.value = '';
        itemInput.focus();
    }

    function removeItem(index) {
        currentItems.splice(index, 1);
        renderItemList();
    }

    function renderItemList() {
        itemList.innerHTML = currentItems.map((content, i) => {
            const display = isUrl(content) ? truncateUrl(content) : (content.length > 50 ? content.slice(0, 50) + '…' : content);
            return `
            <div class="link-item">
                <span style="flex:1; overflow:hidden; text-overflow:ellipsis;">${escapeHtml(display)}</span>
                <span class="remove-link" data-index="${i}" title="Remove item">✕</span>
            </div>
        `;
        }).join('');

        itemList.querySelectorAll('.remove-link').forEach(btn => {
            btn.addEventListener('click', () => {
                removeItem(parseInt(btn.dataset.index));
            });
        });
    }

    function truncateUrl(url) {
        try {
            const u = new URL(url);
            const path = u.pathname.length > 20 ? u.pathname.slice(0, 20) + '…' : u.pathname;
            return u.hostname + (path !== '/' ? path : '');
        } catch {
            return url.length > 40 ? url.slice(0, 40) + '…' : url;
        }
    }

    // ── Create Bubble ──
    async function handleCreateBubble(e) {
        e.preventDefault();

        const nameInput = document.getElementById('bubble-name');
        const tagsInput = document.getElementById('bubble-tags');
        const name = nameInput.value.trim();
        if (!name) return;

        const tags = tagsInput.value
            .split(',')
            .map(t => t.trim().toLowerCase())
            .filter(t => t.length > 0);

        try {
            const bubble = await API.createBubble(name, tags);

            // Add items one by one
            for (const content of currentItems) {
                try {
                    await API.addItem(bubble.id, content);
                } catch {
                    // continue adding remaining items
                }
            }

            closeModal(createModal);
            await loadBubblesFromAPI();
            renderBubbles();
            updateTagFilter();
        } catch (err) {
            showToast(err.message || 'failed to create bubble');
        }
    }

    function resetCreateForm() {
        createForm.reset();
        currentItems = [];
        itemList.innerHTML = '';
    }

    // ── Render Bubbles ──
    function renderBubbles(filtered = null) {
        const list = filtered || bubbles;

        grid.querySelectorAll('.bubble').forEach(b => b.remove());

        if (list.length === 0) {
            emptyState.style.display = 'flex';
        } else {
            emptyState.style.display = 'none';
            list.forEach((bubble, index) => {
                const el = createBubbleElement(bubble, index);
                grid.appendChild(el);
            });
        }

        updateBubbleCount(list.length);
    }

    function createBubbleElement(bubble, index) {
        const wrapper = document.createElement('div');
        wrapper.className = 'bubble-shadow-wrap';
        wrapper.style.animationDelay = `${index * 0.05}s`;

        const el = document.createElement('div');
        const variant = BUBBLE_VARIANTS[index % BUBBLE_VARIANTS.length];
        const blobNum = getBlobForBubble(bubble.id);
        el.className = `bubble ${variant} blob-${blobNum}`;
        el.dataset.id = bubble.id;
        el.title = bubble.name;

        const itemCount = bubble.item_count || 0;
        el.innerHTML = `
            <span class="bubble-name">${escapeHtml(bubble.name)}</span>
            <span class="bubble-item-count">${itemCount} item${itemCount !== 1 ? 's' : ''}</span>
        `;

        el.addEventListener('click', () => {
            wrapper.style.transform = 'translate(1px, 1px)';
            wrapper.style.filter = 'drop-shadow(1px 1px 0px var(--border-dark))';
            setTimeout(() => {
                wrapper.style.transform = '';
                wrapper.style.filter = '';
                openDetailModal(bubble);
            }, 150);
        });

        wrapper.appendChild(el);

        return wrapper;
    }

    // ── Filtering ──
    function filterBubbles() {
        const query = searchInput.value.trim().toLowerCase();
        const tag = tagFilter.value;

        let filtered = bubbles;

        if (query) {
            filtered = filtered.filter(b =>
                b.name.toLowerCase().includes(query) ||
                (b.tags && b.tags.some(t => t.name.toLowerCase().includes(query)))
            );
        }

        if (tag) {
            filtered = filtered.filter(b => b.tags && b.tags.some(t => t.name === tag));
        }

        renderBubbles(filtered);
    }

    // ── Tag Filter Dropdown ──
    function updateTagFilter() {
        const allTags = new Set();
        bubbles.forEach(b => {
            if (b.tags) b.tags.forEach(t => allTags.add(t.name));
        });

        const currentValue = tagFilter.value;
        tagFilter.innerHTML = '<option value="">all tags</option>';
        [...allTags].sort().forEach(tag => {
            const opt = document.createElement('option');
            opt.value = tag;
            opt.textContent = `# ${tag}`;
            tagFilter.appendChild(opt);
        });
        tagFilter.value = currentValue;
    }

    // ── Bubble Count ──
    function updateBubbleCount(count) {
        bubbleCount.textContent = `${count} bubble${count !== 1 ? 's' : ''}`;
    }

    // ── Utils ──
    function escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    // ── Boot ──
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();
