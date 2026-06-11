const Auth = {
    loginScreen: null,
    consoleApp: null,
    tokenInput: null,
    loginError: null,

    init() {
        this.loginScreen = document.getElementById('login-screen');
        this.consoleApp = document.getElementById('console-app');
        this.tokenInput = document.getElementById('token-input');
        this.loginError = document.getElementById('login-error');

        document.getElementById('login-btn').addEventListener('click', () => this.login());
        this.tokenInput.addEventListener('keydown', (event) => {
            if (event.key === 'Enter') this.login();
        });
        document.getElementById('logout-btn').addEventListener('click', () => this.logout());

        this.bootstrap();
    },

    showLogin(message) {
        this.loginScreen.classList.remove('hidden');
        this.consoleApp.classList.add('hidden');
        if (message) {
            this.loginError.textContent = message;
            this.loginError.classList.remove('hidden');
        } else {
            this.loginError.classList.add('hidden');
        }
    },

    showConsole() {
        this.loginScreen.classList.add('hidden');
        this.consoleApp.classList.remove('hidden');
        this.loginError.classList.add('hidden');
    },

    setToken(token) {
        sessionStorage.setItem(TOKEN_KEY, token);
    },

    clearToken() {
        sessionStorage.removeItem(TOKEN_KEY);
    },

    async bootstrap() {
        const token = API.getToken();
        if (!token) {
            this.showLogin('');
            return;
        }

        try {
            await API.validateToken(token);
            this.showConsole();
            App.init();
        } catch (err) {
            this.clearToken();
            this.showLogin(String(err.message || err));
        }
    },

    async login() {
        const token = this.tokenInput.value.trim();
        if (!token) {
            this.showLogin('Введите токен');
            return;
        }

        try {
            await API.validateToken(token);
            this.setToken(token);
            this.showConsole();
            App.init();
        } catch (err) {
            this.clearToken();
            this.showLogin(String(err.message || err));
        }
    },

    logout() {
        this.clearToken();
        this.tokenInput.value = '';
        if (App.refreshInterval) {
            clearInterval(App.refreshInterval);
            App.refreshInterval = null;
        }
        this.showLogin('');
    },
};

document.addEventListener('DOMContentLoaded', () => Auth.init());
