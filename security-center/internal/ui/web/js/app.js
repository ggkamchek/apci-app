const App = {
    currentScreen: 'dashboard',
    refreshInterval: null,
    initialized: false,

    init() {
        if (this.initialized) {
            this.navigateTo(this.currentScreen);
            return;
        }
        this.initialized = true;
        this.bindNavigation();
        this.bindHomeLogo();
        this.bindOmnibox();
        this.checkServerHealth();
        Dashboard.init();
    },

    bindNavigation() {
        document.querySelectorAll('.nav-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                this.navigateTo(tab.dataset.screen);
            });
        });
    },

    bindHomeLogo() {
        const logo = document.getElementById('navHome');
        if (!logo) return;
        logo.addEventListener('click', () => this.navigateTo('dashboard'));
    },

    navigateTo(screen) {
        document.querySelectorAll('.nav-tab').forEach(t => t.classList.remove('active'));
        const tab = document.querySelector(`[data-screen="${screen}"]`);
        if (tab) tab.classList.add('active');

        document.querySelectorAll('.screen').forEach(s => s.classList.remove('active'));
        const screenEl = document.getElementById(`screen-${screen}`);
        if (screenEl) screenEl.classList.add('active');

        this.currentScreen = screen;

        if (screen === 'dashboard') Dashboard.refresh();
        if (screen === 'hashes') Hashes.init();
        if (screen === 'links') Links.init();
        if (screen === 'investigation') Investigation.init();
        if (screen === 'diff') Diff.init();
    },

    bindOmnibox() {
        const input = document.getElementById('investigationSearch');
        if (!input) return;
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                const query = input.value.trim();
                if (query) Investigation.search(query);
            }
        });
    },

    async checkServerHealth() {
        const statusEl = document.getElementById('navStatus');
        const dot = statusEl?.querySelector('.dot');
        if (!dot) return;

        const ok = await API.healthCheck();

        if (ok) {
            statusEl.title = 'Сервер доступен';
            dot.style.background = 'var(--accent-green)';
            dot.style.boxShadow = '0 0 6px var(--accent-green)';
        } else {
            statusEl.title = 'Сервер недоступен';
            dot.style.background = 'var(--accent-red)';
            dot.style.boxShadow = '0 0 6px var(--accent-red)';
        }
    },

    toast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = 'toast';

        const colors = {
            success: 'var(--accent-green)',
            error: 'var(--accent-red)',
            info: 'var(--accent-cyan)',
        };

        toast.style.borderLeft = `3px solid ${colors[type] || colors.info}`;
        toast.textContent = message;
        container.appendChild(toast);

        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transition = 'opacity 0.3s ease';
            setTimeout(() => toast.remove(), 300);
        }, 4000);
    },
};

function accountAvatarSeed(userId) {
    return String(userId || '').replace(/-/g, '').toLowerCase();
}

function deviceAvatarSeed(deviceHash) {
    return String(deviceHash || '').replace(/\s/g, '').toLowerCase();
}

function normalizeDeviceQuery(query) {
    let value = String(query || '').trim().toLowerCase();
    if (value.includes('...')) {
        value = value.split('...')[0];
    }
    if (value.includes('…')) {
        value = value.split('…')[0];
    }
    return value.replace(/\s/g, '');
}

function generateBlockie(hash, canvas, size = 64) {
    const ctx = canvas.getContext('2d');
    canvas.width = size;
    canvas.height = size;

    const seed = [];
    const source = deviceAvatarSeed(hash) || '0000000000000000';
    for (let i = 0; i < 16 && i * 2 < source.length; i++) {
        seed.push(parseInt(source.substr(i * 2, 2), 16) || 0);
    }

    ctx.fillStyle = '#1A1D24';
    ctx.fillRect(0, 0, size, size);

    const cellSize = size / 8;
    const emeraldLightness = 100 + ((seed[3] || 0) % 40);
    ctx.fillStyle = `rgb(45, ${emeraldLightness}, 110)`;

    for (let y = 0; y < 8; y++) {
        for (let x = 0; x < 5; x++) {
            const idx = y * 5 + x;
            const byte = seed[6 + Math.floor(idx / 4)] || 0;
            const bit = (byte >> ((idx % 4) * 2)) & 3;
            if (bit >= 2) {
                ctx.fillRect(x * cellSize, y * cellSize, cellSize, cellSize);
                ctx.fillRect((7 - x) * cellSize, y * cellSize, cellSize, cellSize);
            }
        }
    }
}

function resolveAccountBlockieSeed(profileOrUserId) {
    if (typeof profileOrUserId === 'string') {
        return accountAvatarSeed(profileOrUserId);
    }
    const deviceHash = profileOrUserId?.primary_device_hash || profileOrUserId?.device_hash || '';
    if (deviceHash) {
        return deviceAvatarSeed(deviceHash);
    }
    return accountAvatarSeed(profileOrUserId?.user_id);
}

function renderAccountBlockie(profileOrUserId, canvas, size = 64) {
    generateBlockie(resolveAccountBlockieSeed(profileOrUserId), canvas, size);
}

function renderDeviceBlockie(deviceHash, canvas, size = 64) {
    generateBlockie(deviceAvatarSeed(deviceHash), canvas, size);
}

function truncateUUID(value) {
    if (!value || value.length < 12) return value || '—';
    return `${value.slice(0, 8)}…${value.slice(-4)}`;
}

function truncateHash(hash) {
    if (!hash) return '—';
    if (hash.length <= 18) return hash;
    return hash.substring(0, 8) + '...' + hash.substring(hash.length - 6);
}

function formatDate(dateStr) {
    if (!dateStr) return '—';
    const d = new Date(dateStr);
    return d.toLocaleString('ru-RU', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
}

function linkTypeLabel(type) {
    const labels = {
        same_device: 'одно устройство',
        shared_contact: 'общий контакт',
        communicates_with: 'переписка',
    };
    return labels[type] || type;
}

function verdictLabel(level) {
    const labels = {
        clean: 'ЧИСТЫЙ',
        suspicious: 'ПОДОЗРИТЕЛЬНЫЙ',
        clone: 'КЛОН',
        neutral: 'НЕИЗВЕСТНО',
    };
    return labels[level] || labels.neutral;
}

function matchReasonLabel(reason) {
    const labels = {
        same_device: 'общее устройство',
        shared_contact: 'общий контакт',
        communicates_with: 'переписка',
        'shared device': 'общее устройство',
        'graph link': 'связь в графе',
    };
    return labels[reason] || reason;
}
