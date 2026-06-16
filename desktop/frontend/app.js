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

// E2E UI
const e2eStatus = document.getElementById('e2e-status');
const e2eInitBtn = document.getElementById('e2e-init-btn');
const e2eRecipient = document.getElementById('e2e-recipient');
const e2ePlaintext = document.getElementById('e2e-plaintext');
const e2eSendBtn = document.getElementById('e2e-send-btn');
const e2eInboxBtn = document.getElementById('e2e-inbox-btn');
const e2eClearBtn = document.getElementById('e2e-clear-btn');
const e2eMessages = document.getElementById('e2e-messages');
const e2eRecipientsDatalist = document.getElementById('e2e-recipients');
const e2eRecent = document.getElementById('e2e-recent');

const RECENT_RECIPIENTS_KEY = 'apci.e2e.recentRecipients.v1';

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

function setE2EStatus(text) {
    if (e2eStatus) e2eStatus.textContent = text || '—';
}

function escapeHtml(s) {
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function renderMessages(items) {
    if (!e2eMessages) return;
    if (!items || items.length === 0) {
        e2eMessages.innerHTML = '<div class="meta">Сообщений нет</div>';
        return;
    }

    e2eMessages.innerHTML = items
        .map((m) => {
            const when = m.sentAtUnix ? new Date(m.sentAtUnix * 1000).toLocaleString() : '';
            const who = `${when}  from ${m.senderId || ''}  chat ${m.chatId || ''}`;
            const text = escapeHtml(m.plaintext || '');
            return `<div class="msg"><div class="who">${escapeHtml(who)}</div><div class="text">${text}</div></div>`;
        })
        .join('');
}

function loadRecentRecipients() {
    try {
        const raw = localStorage.getItem(RECENT_RECIPIENTS_KEY);
        const parsed = raw ? JSON.parse(raw) : [];
        return Array.isArray(parsed) ? parsed.filter(Boolean) : [];
    } catch {
        return [];
    }
}

function saveRecentRecipients(ids) {
    try {
        localStorage.setItem(RECENT_RECIPIENTS_KEY, JSON.stringify(ids.slice(0, 8)));
    } catch {
        // ignore
    }
}

function addRecentRecipient(id) {
    const cur = loadRecentRecipients();
    const next = [id, ...cur.filter((x) => x !== id)].slice(0, 8);
    saveRecentRecipients(next);
    renderRecentRecipients();
}

function renderRecentRecipients() {
    const ids = loadRecentRecipients();

    if (e2eRecipientsDatalist) {
        e2eRecipientsDatalist.innerHTML = ids.map((id) => `<option value="${escapeHtml(id)}"></option>`).join('');
    }

    if (e2eRecent) {
        if (ids.length === 0) {
            e2eRecent.innerHTML = `<span class="meta" style="margin-bottom:0;">пусто</span>`;
        } else {
            e2eRecent.innerHTML = ids
                .map((id) => {
                    const short = id.length > 12 ? `${id.slice(0, 8)}…${id.slice(-4)}` : id;
                    return `<button class="chip" type="button" data-id="${escapeHtml(id)}">${escapeHtml(short)}</button>`;
                })
                .join('');
        }
    }
}

function isBundleAlreadyExistsError(message) {
    const m = String(message || '');
    return m.includes('SQLSTATE 23505') || m.includes('duplicate key value') || m.includes('e2e_one_time_prekeys_user_id_key_id_key');
}

async function handleE2EInit() {
    if (!window.go?.main?.App?.E2EInit) {
        setE2EStatus('E2EInit недоступен — перезапусти wails dev');
        return;
    }
    if (e2eInitBtn) e2eInitBtn.disabled = true;
    setE2EStatus('Загрузка E2E bundle...');
    try {
        const res = await window.go.main.App.E2EInit();
        if (res?.success) {
            setE2EStatus(res.message || 'E2EInit: ok');
            return;
        }

        if (isBundleAlreadyExistsError(res?.message)) {
            setE2EStatus('E2E bundle уже загружен (one-time prekeys уже есть). Можно отправлять сообщения.');
            return;
        }

        setE2EStatus(res?.message || 'E2EInit: ошибка');
    } catch (e) {
        if (isBundleAlreadyExistsError(e)) {
            setE2EStatus('E2E bundle уже загружен (one-time prekeys уже есть). Можно отправлять сообщения.');
        } else {
            setE2EStatus(String(e));
        }
    } finally {
        if (e2eInitBtn) e2eInitBtn.disabled = false;
    }
}

async function handleE2ESend() {
    if (!window.go?.main?.App?.E2ESend) {
        setE2EStatus('E2ESend недоступен — перезапусти wails dev');
        return;
    }
    const recipientId = (e2eRecipient?.value || '').trim();
    const plaintext = (e2ePlaintext?.value || '').trim();
    if (!recipientId) {
        setE2EStatus('Введи recipient user_id (UUID)');
        return;
    }
    if (!plaintext) {
        setE2EStatus('Введи сообщение');
        return;
    }

    if (e2eSendBtn) e2eSendBtn.disabled = true;
    setE2EStatus('Отправка...');
    try {
        const res = await window.go.main.App.E2ESend(recipientId, '', plaintext);
        if (res?.success) {
            setE2EStatus(`Отправлено. messageId=${res.messageId || ''}`);
            if (e2ePlaintext) e2ePlaintext.value = '';
            addRecentRecipient(recipientId);
        } else {
            setE2EStatus(res?.message || 'Send: ошибка');
        }
    } catch (e) {
        setE2EStatus(String(e));
    } finally {
        if (e2eSendBtn) e2eSendBtn.disabled = false;
    }
}

async function handleE2EInbox() {
    if (!window.go?.main?.App?.E2EInbox) {
        setE2EStatus('E2EInbox недоступен — перезапусти wails dev');
        return;
    }
    if (e2eInboxBtn) e2eInboxBtn.disabled = true;
    setE2EStatus('Загрузка inbox...');
    try {
        const res = await window.go.main.App.E2EInbox('', 50);
        if (res?.success) {
            setE2EStatus('Inbox: ok');
            renderMessages(res.messages || []);
        } else {
            setE2EStatus(res?.message || 'Inbox: ошибка');
        }
    } catch (e) {
        setE2EStatus(String(e));
    } finally {
        if (e2eInboxBtn) e2eInboxBtn.disabled = false;
    }
}

function handleE2EClear() {
    if (e2eRecipient) e2eRecipient.value = '';
    if (e2ePlaintext) e2ePlaintext.value = '';
    renderMessages([]);
    setE2EStatus('Очищено');
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

if (e2eInitBtn) e2eInitBtn.addEventListener('click', handleE2EInit);
if (e2eSendBtn) e2eSendBtn.addEventListener('click', handleE2ESend);
if (e2eInboxBtn) e2eInboxBtn.addEventListener('click', handleE2EInbox);
if (e2eClearBtn) e2eClearBtn.addEventListener('click', handleE2EClear);

if (e2eRecent) {
    e2eRecent.addEventListener('click', (e) => {
        const btn = e.target?.closest?.('button[data-id]');
        const id = btn?.getAttribute?.('data-id');
        if (id && e2eRecipient) e2eRecipient.value = id;
    });
}

updateModeUI();
showAuthScreen();
initFingerprint();
renderRecentRecipients();
