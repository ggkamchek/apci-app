const Dashboard = {
    init() {
        this.refresh();
        if (App.refreshInterval) clearInterval(App.refreshInterval);
        App.refreshInterval = setInterval(() => {
            if (App.currentScreen === 'dashboard') this.refresh();
        }, 30000);
    },

    async refresh() {
        await Promise.all([
            this.loadCounters(),
            this.loadTopClones(),
            this.loadActivityChart(),
        ]);
    },

    async loadCounters() {
        try {
            const stats = await API.getStats();
            document.getElementById('counterHourly').textContent = stats.hourly || 0;
            document.getElementById('counterTotal').textContent = stats.total_accounts || 0;
            document.getElementById('counterUnique').textContent = stats.unique_devices || 0;
            document.getElementById('counterEntropy').textContent = stats.total_links || 0;
        } catch {
            document.getElementById('counterHourly').textContent = '—';
            document.getElementById('counterTotal').textContent = '—';
            document.getElementById('counterUnique').textContent = '—';
            document.getElementById('counterEntropy').textContent = '—';
        }
    },

    async loadTopClones() {
        const container = document.getElementById('topClonesCards');
        try {
            const clones = await API.getTopClones();
            if (!clones || clones.length === 0) {
                container.innerHTML = '<div class="empty-state">Нет мультиаккаунтов. Зарегистрируйте два аккаунта с одного ПК.</div>';
                return;
            }

            container.innerHTML = '';
            clones.forEach(c => {
                const card = document.createElement('article');
                card.className = 'fingerprint-card risk-clone';
                card.addEventListener('click', () => Dashboard.openInvestigation(c.device_hash));

                const avatarWrap = document.createElement('div');
                avatarWrap.className = 'fingerprint-avatar-wrap';

                const avatar = document.createElement('canvas');
                avatar.className = 'fingerprint-avatar';
                avatar.width = 48;
                avatar.height = 48;
                generateBlockie(deviceAvatarSeed(c.device_hash), avatar, 48);

                const status = document.createElement('span');
                status.className = 'avatar-status-dot status-clone';
                avatarWrap.append(avatar, status);

                const panel = document.createElement('div');
                panel.className = 'fingerprint-card-panel';

                const body = document.createElement('div');
                body.className = 'fingerprint-card-body';

                const hash = document.createElement('button');
                hash.className = 'hash-link';
                hash.type = 'button';
                hash.textContent = truncateHash(c.device_hash);
                hash.addEventListener('click', (event) => {
                    event.stopPropagation();
                    Dashboard.openInvestigation(c.device_hash);
                });

                const meta = document.createElement('div');
                meta.className = 'fingerprint-card-meta';
                meta.textContent = `Последняя активность: ${formatDate(c.last_seen)}`;

                body.append(hash, meta);

                const badge = document.createElement('div');
                badge.className = 'risk-badge badge-clone';
                badge.textContent = `Клон · ${c.count} аккаунта`;

                panel.append(body, badge);
                card.append(avatarWrap, panel);
                container.appendChild(card);
            });
        } catch {
            container.innerHTML = '<div class="empty-state">Нет данных</div>';
        }
    },

    async loadActivityChart() {
        const container = document.getElementById('entropyChart');
        if (!container) return;

        try {
            const series = await API.getActivitySeries();
            this.renderActivityChart(container, series || []);
        } catch {
            container.innerHTML = '<div class="empty-state">Не удалось загрузить график</div>';
        }
    },

    renderActivityChart(container, series) {
        container.innerHTML = '';

        const width = container.clientWidth || 560;
        const height = container.clientHeight || 300;
        const margin = { top: 18, right: 18, bottom: 34, left: 42 };
        const innerWidth = width - margin.left - margin.right;
        const innerHeight = height - margin.top - margin.bottom;

        const data = series.map(point => ({
            hour: new Date(point.hour),
            events: Number(point.total || 0),
            links: Number(point.links || 0),
            accounts: Number(point.unique_accounts || 0),
        }));

        if (data.length === 0 || data.every(point => point.events === 0 && point.links === 0)) {
            container.innerHTML = '<div class="empty-state">Нет данных для графика</div>';
            return;
        }

        const maxValue = Math.max(...data.map(p => Math.max(p.events, p.links, p.accounts)), 1);

        const svg = d3.select(container)
            .append('svg')
            .attr('viewBox', `0 0 ${width} ${height}`)
            .attr('width', '100%')
            .attr('height', '100%');

        const chart = svg.append('g')
            .attr('transform', `translate(${margin.left},${margin.top})`);

        const x = d3.scaleTime()
            .domain(d3.extent(data, point => point.hour))
            .range([0, innerWidth]);

        const y = d3.scaleLinear()
            .domain([0, maxValue])
            .nice()
            .range([innerHeight, 0]);

        const lineLinks = d3.line()
            .x(point => x(point.hour))
            .y(point => y(point.links))
            .curve(d3.curveMonotoneX);

        chart.append('path')
            .datum(data)
            .attr('d', lineLinks)
            .attr('fill', 'none')
            .attr('stroke', '#2D8F6E')
            .attr('stroke-width', 3);

        chart.append('g')
            .attr('class', 'entropy-axis')
            .attr('transform', `translate(0,${innerHeight})`)
            .call(d3.axisBottom(x).ticks(6).tickFormat(d3.timeFormat('%H:%M')));

        chart.append('g')
            .attr('class', 'entropy-axis')
            .call(d3.axisLeft(y).ticks(5));

        chart.selectAll('.entropy-dot')
            .data(data.filter(point => point.links > 0))
            .enter()
            .append('circle')
            .attr('class', 'entropy-dot')
            .attr('cx', point => x(point.hour))
            .attr('cy', point => y(point.links))
            .attr('r', 3)
            .attr('fill', '#2D8F6E')
            .append('title')
            .text(point => `${formatDate(point.hour)} · связей ${point.links} · событий ${point.events}`);
    },

    openInvestigation(hash) {
        App.navigateTo('investigation');
        document.getElementById('investigationSearch').value = hash;
        Investigation.search(hash);
    },
};
