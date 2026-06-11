const TOKEN_KEY = 'apci_security_admin_token';

const API = {
    baseURL: '',

    getToken() {
        return sessionStorage.getItem(TOKEN_KEY) || '';
    },

    authHeaders(extra = {}) {
        const headers = { ...extra };
        const token = this.getToken();
        if (token) {
            headers.Authorization = `Bearer ${token}`;
        }
        return headers;
    },

    async request(path, options = {}) {
        const response = await fetch(`${this.baseURL}${path}`, {
            ...options,
            headers: this.authHeaders(options.headers || {}),
        });

        if (response.status === 401) {
            throw new Error('Неверный token');
        }

        if (!response.ok) {
            const text = await response.text().catch(() => '');
            throw new Error(text || `HTTP ${response.status}`);
        }

        if (response.status === 204) {
            return null;
        }

        return response.json();
    },

    async getStats() {
        return this.request('/admin/stats');
    },

    async getTopClones() {
        return this.request('/admin/stats/top-clones');
    },

    async getActivitySeries() {
        return this.request('/admin/stats/activity-series');
    },

    async getDeviceRegistry(filter = 'clone', limit = 100) {
        const params = new URLSearchParams({ filter, limit: String(limit) });
        return this.request(`/admin/devices?${params.toString()}`);
    },

    async getLinks() {
        return this.request('/admin/links');
    },

    async lookup(query) {
        const params = new URLSearchParams({ q: query });
        return this.request(`/admin/lookup?${params.toString()}`);
    },

    async getAccount(userId) {
        return this.request(`/admin/account/${encodeURIComponent(userId)}`);
    },

    async getRelated(userId) {
        return this.request(`/admin/account/${encodeURIComponent(userId)}/related`);
    },

    async diffAccounts(a, b) {
        const params = new URLSearchParams({ a, b });
        return this.request(`/admin/diff?${params.toString()}`);
    },

    async healthCheck() {
        try {
            const response = await fetch(`${this.baseURL}/health`);
            return response.ok;
        } catch {
            return false;
        }
    },

    async validateToken(token) {
        const response = await fetch(`${this.baseURL}/admin/links`, {
            headers: { Authorization: `Bearer ${token}` },
        });
        if (response.status === 401) {
            throw new Error('Неверный token');
        }
        if (!response.ok) {
            throw new Error(`Ошибка сервера (${response.status})`);
        }
    },
};
