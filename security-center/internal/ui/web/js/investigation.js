const Investigation = {
    init() {},

    async search(rawQuery) {
        const query = String(rawQuery || '').trim();
        const placeholder = document.getElementById('investigationPlaceholder');
        const result = document.getElementById('investigationResult');

        placeholder.style.display = 'none';
        result.style.display = 'block';
        result.innerHTML = '<div class="grid-item col-12" style="text-align:center;padding:40px;color:var(--text-muted);">Поиск...</div>';

        try {
            const isUUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(query);
            const lookupQuery = isUUID ? query : (normalizeDeviceQuery(query) || query);
            const data = await API.lookup(lookupQuery);

            if (!data) {
                result.innerHTML = '<div class="grid-item col-12" style="text-align:center;padding:40px;color:var(--accent-amber);">Ничего не найдено</div>';
                return;
            }

            if (data.accounts && Array.isArray(data.accounts)) {
                await this.renderDeviceLookup(data);
                return;
            }

            if (data.user_id) {
                await this.renderProfile(data);
                return;
            }

            result.innerHTML = '<div class="grid-item col-12" style="text-align:center;padding:40px;color:var(--accent-amber);">Ничего не найдено</div>';
        } catch (err) {
            result.innerHTML = `<div class="grid-item col-12" style="text-align:center;padding:40px;color:var(--accent-red);">Ошибка: ${err.message}</div>`;
        }
    },

    async renderDeviceLookup(data) {
        const result = document.getElementById('investigationResult');
        const accounts = data.accounts || [];

        if (accounts.length === 0) {
            result.innerHTML = '<div class="grid-item col-12" style="text-align:center;padding:40px;color:var(--accent-amber);">Аккаунты не найдены</div>';
            return;
        }

        if (accounts.length === 1) {
            await this.renderProfile(accounts[0]);
            return;
        }

        result.innerHTML = `
            <div class="grid-item col-12 detail-panel">
                <div class="profile-card fingerprint-detail-card">
                    <div class="profile-avatar">
                        <canvas id="deviceLookupBlockie" width="64" height="64"></canvas>
                    </div>
                    <div class="profile-info">
                        <div class="flex-between">
                            <span class="widget-title">Совпадение по устройству</span>
                            <span class="verdict clone">КЛОН</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">Хеш устройства</span>
                            <span class="profile-value hash-text">${data.device_hash || '—'}</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">Тип совпадения</span>
                            <span class="profile-value">${data.match_type === 'prefix' ? 'по началу хеша' : 'точное'}</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">Аккаунтов</span>
                            <span class="profile-value">${accounts.length}</span>
                        </div>
                    </div>
                </div>
            </div>
            <div class="grid-item col-12">
                <div class="widget-header"><span class="widget-title">Аккаунты на этом устройстве</span></div>
                <div class="related-card-grid" id="deviceAccountsGrid"></div>
            </div>
        `;

        const canvas = document.getElementById('deviceLookupBlockie');
        if (canvas) renderDeviceBlockie(data.device_hash, canvas, 64);

        const grid = document.getElementById('deviceAccountsGrid');
        grid.innerHTML = '';
        accounts.forEach(account => {
            const card = document.createElement('article');
            card.className = 'related-card risk-clone';

            const avatarWrap = document.createElement('div');
            avatarWrap.className = 'fingerprint-avatar-wrap';
            const avatar = document.createElement('canvas');
            avatar.width = 40;
            avatar.height = 40;
            renderAccountBlockie(account, avatar, 40);
            avatarWrap.appendChild(avatar);

            const title = document.createElement('div');
            title.className = 'related-card-title';
            title.textContent = truncateUUID(account.user_id);

            const meta = document.createElement('div');
            meta.className = 'related-card-reason';
            meta.textContent = `Активность: ${formatDate(account.last_seen_at)}`;

            const button = document.createElement('button');
            button.className = 'btn compare';
            button.type = 'button';
            button.textContent = 'Открыть';
            button.addEventListener('click', () => this.renderProfile(account));

            card.append(avatarWrap, title, meta, button);
            grid.appendChild(card);
        });
    },

    async renderProfile(profile) {
        const result = document.getElementById('investigationResult');

        result.innerHTML = `
            <div class="grid-item col-12 detail-panel">
                <div class="profile-card fingerprint-detail-card">
                    <div class="profile-avatar">
                        <canvas id="profileBlockie" width="64" height="64"></canvas>
                    </div>
                    <div class="profile-info">
                        <div class="flex-between">
                            <span class="widget-title">Карточка аккаунта</span>
                            <span class="verdict neutral" id="profileVerdict">НЕИЗВЕСТНО</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">ID аккаунта</span>
                            <span class="profile-value" id="profileUserId">${profile.user_id || '—'}</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">Основное устройство</span>
                            <span class="profile-value hash-text" id="profileHash">${profile.primary_device_hash || '—'}</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">Первое появление</span>
                            <span class="profile-value" id="profileCreated">${formatDate(profile.first_seen_at)}</span>
                        </div>
                        <div class="profile-row">
                            <span class="profile-label">Последняя активность</span>
                            <span class="profile-value" id="profileLastSeen">${formatDate(profile.last_seen_at)}</span>
                        </div>
                    </div>
                </div>
            </div>
            <div class="grid-item col-8">
                <div class="widget-header"><span class="widget-title">Матрица связей</span></div>
                <div class="graph-container" id="investigationGraph"></div>
            </div>
            <div class="grid-item col-4">
                <div class="widget-header"><span class="widget-title">Устройства</span></div>
                <div id="componentsList"></div>
            </div>
            <div class="grid-item col-12">
                <div class="widget-header"><span class="widget-title">Связанные аккаунты</span></div>
                <div class="related-card-grid" id="relatedAccountsGrid">
                    <div class="empty-state">Загрузка связей...</div>
                </div>
            </div>
        `;

        const canvas = document.getElementById('profileBlockie');
        if (canvas) renderAccountBlockie(profile, canvas, 64);

        this.renderComponents(profile);

        let related = [];
        let verdict = null;
        try {
            const relatedResult = await API.getRelated(profile.user_id);
            related = relatedResult.related || [];
            verdict = relatedResult.risk_verdict || null;
        } catch (err) {
            console.error(err);
        }

        this.renderVerdict(verdict, related);
        this.renderRelatedCards(profile, related);
        this.renderGraph(profile, related);
    },

    renderComponents(profile) {
        const container = document.getElementById('componentsList');
        if (!container) return;

        const rows = [];
        (profile.devices || []).forEach((d, i) => {
            rows.push({ name: `Устройство ${i + 1}`, value: d.device_hash, isHash: true });
        });

        if (rows.length === 0) {
            container.innerHTML = '<div class="profile-row"><span class="profile-label">Нет данных</span></div>';
            return;
        }

        container.innerHTML = rows.map(c => `
            <div class="profile-row">
                <span class="profile-label">${c.name}</span>
                <span class="profile-value ${c.isHash ? 'hash-text' : ''}">${c.value || '—'}</span>
            </div>
        `).join('');
    },

    renderGraph(profile, related) {
        const container = document.getElementById('investigationGraph');
        if (!container) return;

        const centerId = String(profile.user_id);
        const relatedNodes = related
            .filter(r => String(r.user_id) !== centerId)
            .map(r => ({
                id: String(r.user_id),
                reason: matchReasonLabel(r.match_reason || r.link_type),
                isCenter: false,
            }));

        if (relatedNodes.length === 0) {
            container.innerHTML = '<div style="display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-muted);font-size:12px;">Нет связанных аккаунтов</div>';
            return;
        }

        container.innerHTML = '';
        const width = container.clientWidth || 600;
        const height = container.clientHeight || 360;
        const centerX = width / 2;
        const centerY = height / 2;
        const radius = Math.min(width, height) * 0.32;

        const centerNode = { id: centerId, x: centerX, y: centerY, isCenter: true };

        relatedNodes.forEach((node, index) => {
            const angle = (2 * Math.PI * index) / relatedNodes.length - Math.PI / 2;
            node.x = centerX + radius * Math.cos(angle);
            node.y = centerY + radius * Math.sin(angle);
        });

        const svg = d3.select('#investigationGraph')
            .append('svg')
            .attr('viewBox', `0 0 ${width} ${height}`)
            .attr('width', '100%')
            .attr('height', '100%');

        svg.append('defs').append('marker')
            .attr('id', 'arrowhead')
            .attr('viewBox', '0 -5 10 10')
            .attr('refX', 10)
            .attr('refY', 0)
            .attr('markerWidth', 7)
            .attr('markerHeight', 7)
            .attr('orient', 'auto')
            .append('path')
            .attr('d', 'M0,-5L10,0L0,5')
            .attr('fill', '#2D8F6E');

        relatedNodes.forEach(node => {
            const dx = node.x - centerNode.x;
            const dy = node.y - centerNode.y;
            const distance = Math.sqrt(dx * dx + dy * dy) || 1;
            const nx = dx / distance;
            const ny = dy / distance;

            svg.append('line')
                .attr('x1', centerNode.x + nx * 30)
                .attr('y1', centerNode.y + ny * 30)
                .attr('x2', node.x - nx * 22)
                .attr('y2', node.y - ny * 22)
                .attr('stroke', '#2A2E38')
                .attr('stroke-width', 2)
                .attr('marker-end', 'url(#arrowhead)');

            svg.append('text')
                .attr('x', (centerNode.x + node.x) / 2)
                .attr('y', (centerNode.y + node.y) / 2 - 8)
                .attr('text-anchor', 'middle')
                .attr('fill', '#8B95A5')
                .attr('font-size', '9px')
                .text(node.reason);
        });

        [centerNode, ...relatedNodes].forEach(node => {
            const group = svg.append('g')
                .attr('transform', `translate(${node.x},${node.y})`)
                .style('cursor', 'pointer');

            group.append('circle')
                .attr('r', node.isCenter ? 30 : 22)
                .attr('fill', node.isCenter ? '#2D8F6E' : '#1A1D24')
                .attr('stroke', node.isCenter ? '#3BAF88' : '#2D8F6E')
                .attr('stroke-width', node.isCenter ? 2.5 : 2);

            group.append('text')
                .text(truncateUUID(node.id))
                .attr('text-anchor', 'middle')
                .attr('dy', '0.35em')
                .attr('fill', '#E8EDF2')
                .attr('font-size', node.isCenter ? '10px' : '9px')
                .attr('font-family', 'JetBrains Mono, monospace')
                .style('pointer-events', 'none');

            group.on('click', () => {
                document.getElementById('investigationSearch').value = node.id;
                Investigation.search(node.id);
            });
        });
    },

    renderVerdict(verdict, related) {
        const el = document.getElementById('profileVerdict');
        if (!el) return;

        const level = verdict?.level || (related.length > 0 ? 'suspicious' : 'clean');
        el.className = `verdict ${level}`;
        el.textContent = verdictLabel(level);
    },

    renderRelatedCards(profile, related) {
        const grid = document.getElementById('relatedAccountsGrid');
        if (!grid) return;

        const items = related.filter(item => String(item.user_id) !== String(profile.user_id));
        if (items.length === 0) {
            grid.innerHTML = '<div class="empty-state">Связанные аккаунты не найдены</div>';
            return;
        }

        grid.innerHTML = '';
        items.forEach(item => {
            const card = document.createElement('article');
            card.className = 'related-card risk-clone';

            const avatarWrap = document.createElement('div');
            avatarWrap.className = 'fingerprint-avatar-wrap';
            const avatar = document.createElement('canvas');
            avatar.width = 40;
            avatar.height = 40;
            renderAccountBlockie({ user_id: item.user_id, device_hash: item.device_hash }, avatar, 40);
            avatarWrap.appendChild(avatar);

            const title = document.createElement('div');
            title.className = 'related-card-title';
            title.textContent = truncateUUID(item.user_id);

            const hash = document.createElement('div');
            hash.className = 'hash-text related-card-hash';
            hash.textContent = item.device_hash ? truncateHash(item.device_hash) : linkTypeLabel(item.link_type);

            const reason = document.createElement('div');
            reason.className = 'related-card-reason';
            reason.textContent = matchReasonLabel(item.match_reason || item.link_type);

            const button = document.createElement('button');
            button.className = 'btn compare';
            button.type = 'button';
            button.textContent = 'Сравнить';
            button.addEventListener('click', () => {
                App.navigateTo('diff');
                document.getElementById('diffIdA').value = profile.user_id;
                document.getElementById('diffIdB').value = item.user_id;
                Diff.compare();
            });

            card.append(avatarWrap, title, hash, reason, button);
            grid.appendChild(card);
        });
    },
};
