const Hashes = {
    filter: 'clone',

    init() {
        this.bindFilters();
        this.load();
    },

    bindFilters() {
        document.querySelectorAll('[data-hash-filter]').forEach(button => {
            button.addEventListener('click', () => {
                this.filter = button.dataset.hashFilter;
                document.querySelectorAll('[data-hash-filter]').forEach(item => {
                    item.classList.toggle('active', item.dataset.hashFilter === this.filter);
                });
                this.load();
            });
        });
    },

    async load() {
        const tbody = document.getElementById('hashesTableBody');
        const summary = document.getElementById('hashesSummary');
        if (!tbody) return;

        tbody.innerHTML = '<tr><td colspan="5" class="muted">Загрузка...</td></tr>';

        try {
            const result = await API.getDeviceRegistry(this.filter);
            const items = result.items || [];
            if (summary) {
                summary.textContent = `${items.length} хешей · фильтр: ${this.filterLabel(result.filter)}`;
            }

            if (items.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" class="muted">Нет данных для выбранного фильтра</td></tr>';
                return;
            }

            tbody.innerHTML = '';
            items.forEach(item => {
                const row = document.createElement('tr');
                row.className = 'hash-row';
                row.addEventListener('click', () => Hashes.openInvestigation(item.device_hash));

                const hashCell = document.createElement('td');
                const hashBtn = document.createElement('button');
                hashBtn.type = 'button';
                hashBtn.className = 'hash-link';
                hashBtn.textContent = truncateHash(item.device_hash);
                hashBtn.addEventListener('click', (event) => {
                    event.stopPropagation();
                    Hashes.openInvestigation(item.device_hash);
                });
                hashCell.appendChild(hashBtn);

                const accountsCell = document.createElement('td');
                accountsCell.textContent = String(item.account_count || 0);

                const recordsCell = document.createElement('td');
                recordsCell.textContent = String(item.record_count || 0);

                const seenCell = document.createElement('td');
                seenCell.textContent = formatDate(item.last_seen);

                const verdictCell = document.createElement('td');
                const badge = document.createElement('span');
                badge.className = Hashes.verdictClass(item.risk_level);
                badge.textContent = Hashes.verdictLabel(item.risk_level, item.account_count);
                verdictCell.appendChild(badge);

                row.append(hashCell, accountsCell, recordsCell, seenCell, verdictCell);
                tbody.appendChild(row);
            });
        } catch (err) {
            tbody.innerHTML = `<tr><td colspan="5" class="muted">Ошибка: ${err.message}</td></tr>`;
            if (summary) summary.textContent = 'Не удалось загрузить реестр';
        }
    },

    openInvestigation(hash) {
        App.navigateTo('investigation');
        document.getElementById('investigationSearch').value = hash;
        Investigation.search(hash);
    },

    filterLabel(filter) {
        switch (filter) {
            case 'clean': return 'чистые';
            case 'suspicious': return 'подозрительные';
            case 'all': return 'all';
            default: return 'клоны';
        }
    },

    verdictClass(level) {
        switch (level) {
            case 'clone': return 'risk-badge badge-clone';
            case 'suspicious': return 'risk-badge badge-suspicious';
            default: return 'risk-badge badge-clean';
        }
    },

    verdictLabel(level, accountCount) {
        switch (level) {
            case 'clone': return `Клон · ${accountCount} аккаунта`;
            case 'suspicious': return 'Подозрительный';
            default: return 'Чистый';
        }
    },
};
