const Diff = {
    init() {},

    async compare() {
        const idA = document.getElementById('diffIdA').value.trim();
        const idB = document.getElementById('diffIdB').value.trim();

        if (!idA || !idB) {
            App.toast('Введите оба UUID аккаунта', 'error');
            return;
        }

        const columns = document.getElementById('diffColumns');
        columns.style.display = 'grid';
        document.getElementById('diffContentA').innerHTML = '<div class="muted" style="padding:20px;">Загрузка...</div>';
        document.getElementById('diffContentB').innerHTML = '<div class="muted" style="padding:20px;">Загрузка...</div>';

        try {
            const result = await API.diffAccounts(idA, idB);
            this.renderDiff(result);
        } catch (err) {
            App.toast('Ошибка сравнения: ' + err.message, 'error');
        }
    },

    renderDiff(result) {
        const fpA = result.account_a;
        const fpB = result.account_b;
        const sharedDevices = result.shared?.devices || [];

        const params = [
            ['ID аккаунта', fpA.user_id, fpB.user_id],
            ['Первое появление', formatDate(fpA.first_seen_at), formatDate(fpB.first_seen_at)],
            ['Последняя активность', formatDate(fpA.last_seen_at), formatDate(fpB.last_seen_at)],
            ['Основное устройство', fpA.primary_device_hash, fpB.primary_device_hash],
            ['Устройств', String((fpA.devices || []).length), String((fpB.devices || []).length)],
            ['Общие устройства', sharedDevices.join(', ') || '—', sharedDevices.join(', ') || '—'],
        ];

        let matchCount = sharedDevices.length;
        if (fpA.primary_device_hash && fpA.primary_device_hash === fpB.primary_device_hash) {
            matchCount += 1;
        }

        const colA = document.getElementById('diffColA');
        const colB = document.getElementById('diffColB');
        if (matchCount >= 1) {
            colA.classList.add('match');
            colB.classList.add('match');
        } else {
            colA.classList.remove('match');
            colB.classList.remove('match');
        }

        let htmlA = '';
        let htmlB = '';
        params.forEach(([label, valA, valB]) => {
            const isMatch = String(valA) === String(valB) && !['ID аккаунта'].includes(label) && valA !== '—';
            const matchClass = isMatch ? ' style="color:var(--accent-red);font-weight:600;"' : '';
            htmlA += `<div class="profile-row"><span class="profile-label">${label}</span><span class="profile-value"${matchClass}>${valA || '—'}</span></div>`;
            htmlB += `<div class="profile-row"><span class="profile-label">${label}</span><span class="profile-value"${matchClass}>${valB || '—'}</span></div>`;
        });

        document.getElementById('diffContentA').innerHTML = htmlA;
        document.getElementById('diffContentB').innerHTML = htmlB;

        const verdict = result.verdict || {};
        let verdictText;
        let verdictColor;
        if (verdict.level === 'clone') {
            verdictText = `Вердикт: КЛОН (${matchReasonLabel(verdict.reason)})`;
            verdictColor = 'var(--risk-clone)';
        } else if (verdict.level === 'suspicious') {
            verdictText = `Вердикт: РИСК (${matchReasonLabel(verdict.reason)})`;
            verdictColor = 'var(--risk-suspicious)';
        } else {
            verdictText = 'Вердикт: РАЗНЫЕ';
            verdictColor = 'var(--risk-clean)';
        }

        const existingVerdict = document.getElementById('diffVerdict');
        if (existingVerdict) existingVerdict.remove();

        const verdictEl = document.createElement('div');
        verdictEl.id = 'diffVerdict';
        verdictEl.style.cssText = `
            text-align: center;
            padding: 16px;
            margin-bottom: 16px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 16px;
            font-weight: 700;
            color: ${verdictColor};
            background: var(--bg-panel);
            border: 1px solid var(--border);
            border-radius: var(--radius-md);
        `;
        verdictEl.textContent = verdictText;

        const columns = document.getElementById('diffColumns');
        columns.parentNode.insertBefore(verdictEl, columns);
    },
};
