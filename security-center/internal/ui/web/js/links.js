const Links = {
    init() {
        this.load();
    },

    async load() {
        const tbody = document.getElementById('linksTableBody');
        const summary = document.getElementById('linksSummary');
        if (!tbody) return;

        tbody.innerHTML = '<tr><td colspan="5" class="muted">Загрузка...</td></tr>';

        try {
            const links = await API.getLinks();
            if (summary) {
                summary.textContent = `${links.length} связей`;
            }

            if (!links || links.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" class="muted">Связей пока нет</td></tr>';
                return;
            }

            tbody.innerHTML = '';
            links.forEach(link => {
                const row = document.createElement('tr');

                const a = document.createElement('td');
                a.className = 'uuid-cell';
                a.title = link.accountA;
                a.textContent = truncateUUID(link.accountA);
                a.style.cursor = 'pointer';
                a.addEventListener('click', () => Links.openAccount(link.accountA));

                const b = document.createElement('td');
                b.className = 'uuid-cell';
                b.title = link.accountB;
                b.textContent = truncateUUID(link.accountB);
                b.style.cursor = 'pointer';
                b.addEventListener('click', () => Links.openAccount(link.accountB));

                const type = document.createElement('td');
                const badge = document.createElement('span');
                badge.className = `badge-link${link.linkType === 'shared_contact' ? ' contact' : ''}`;
                badge.textContent = linkTypeLabel(link.linkType);
                type.appendChild(badge);

                const weight = document.createElement('td');
                weight.textContent = String(link.weight ?? '—');

                const date = document.createElement('td');
                date.textContent = formatDate(link.detectedAt);

                row.append(a, b, type, weight, date);
                tbody.appendChild(row);
            });
        } catch (err) {
            tbody.innerHTML = `<tr><td colspan="5" class="muted">Ошибка: ${err.message}</td></tr>`;
        }
    },

    openAccount(userId) {
        App.navigateTo('investigation');
        document.getElementById('investigationSearch').value = userId;
        Investigation.search(userId);
    },
};
