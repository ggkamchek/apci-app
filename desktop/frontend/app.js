let isRegisterMode = false;

const authScreen = document.getElementById('auth-screen');
const mainScreen = document.getElementById('main-screen');
const usernameInput = document.getElementById('username');
const submitBtn = document.getElementById('submit-btn');
const toggleModeBtn = document.getElementById('toggle-mode');
const authSubtitle = document.getElementById('auth-subtitle');
const errorEl = document.getElementById('error');
const displayUsername = document.getElementById('display-username');
const displayUserId = document.getElementById('display-userid');
const logoutBtn = document.getElementById('logout-btn');

function showError(message) {
    if (!message) {
        errorEl.classList.add('hidden');
        errorEl.textContent = '';
        return;
    }
    errorEl.textContent = message;
    errorEl.classList.remove('hidden');
}

function showAuthScreen() {
    authScreen.classList.remove('hidden');
    mainScreen.classList.add('hidden');
    showError('');
}

function showMainScreen(user) {
    displayUsername.textContent = user.username;
    displayUserId.textContent = user.userId;
    authScreen.classList.add('hidden');
    mainScreen.classList.remove('hidden');
    showError('');
}

function updateModeUI() {
    if (isRegisterMode) {
        authSubtitle.textContent = 'Создание аккаунта';
        submitBtn.textContent = 'Зарегистрироваться';
        toggleModeBtn.textContent = 'Уже есть аккаунт? Войти';
    } else {
        authSubtitle.textContent = 'Вход в аккаунт';
        submitBtn.textContent = 'Войти';
        toggleModeBtn.textContent = 'Нет аккаунта? Зарегистрироваться';
    }
}

async function handleSubmit() {
    const username = usernameInput.value.trim();
    if (!username) {
        showError('Введите имя пользователя');
        return;
    }

    submitBtn.disabled = true;
    showError('');

    try {
        await initFingerprint();

        const result = isRegisterMode
            ? await window.go.main.App.Register(username)
            : await window.go.main.App.Login(username);

        if (result.success) {
            showMainScreen(result);
        } else {
            showError(result.message || 'Ошибка');
        }
    } catch (err) {
        showError(String(err));
    } finally {
        submitBtn.disabled = false;
    }
}

async function handleLogout() {
    await window.go.main.App.Logout();
    usernameInput.value = '';
    showAuthScreen();
}

toggleModeBtn.addEventListener('click', () => {
    isRegisterMode = !isRegisterMode;
    updateModeUI();
    showError('');
});

submitBtn.addEventListener('click', handleSubmit);
usernameInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') handleSubmit();
});
logoutBtn.addEventListener('click', handleLogout);

updateModeUI();
showAuthScreen();
initFingerprint();
