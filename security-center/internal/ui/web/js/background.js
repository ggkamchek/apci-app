(function () {
    const canvas = document.getElementById('bg-canvas');
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    const colors = ['#2D8F6E', '#3BAF88', '#5A6B7C', 'rgba(255,255,255,0.03)'];

    let width = 0;
    let height = 0;
    let dpr = 1;
    let lines = [];
    let nodes = [];
    let rings = [];
    let glitches = [];
    let radar = [];
    let animationId = null;
    let running = true;

    function random(min, max) {
        return min + Math.random() * (max - min);
    }

    function pick(array) {
        return array[Math.floor(Math.random() * array.length)];
    }

    function alphaColor(color, alpha) {
        if (color.startsWith('rgba')) {
            return color.replace(/rgba\(([^)]+),\s*[\d.]+\)/, `rgba($1, ${alpha})`);
        }

        const hex = color.replace('#', '');
        const r = parseInt(hex.slice(0, 2), 16);
        const g = parseInt(hex.slice(2, 4), 16);
        const b = parseInt(hex.slice(4, 6), 16);
        return `rgba(${r}, ${g}, ${b}, ${alpha})`;
    }

    function resize() {
        dpr = Math.min(window.devicePixelRatio || 1, 2);
        width = window.innerWidth;
        height = window.innerHeight;
        canvas.width = Math.floor(width * dpr);
        canvas.height = Math.floor(height * dpr);
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
        ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
        generate();
    }

    function generate() {
        const lineCount = Math.floor(random(15, 26));
        const nodeCount = Math.floor(random(30, 51));
        const ringCount = Math.floor(random(3, 6));
        const glitchCount = Math.floor(random(5, 11));

        lines = Array.from({ length: lineCount }, () => ({
            x1: random(0, width),
            y1: random(0, height),
            x2: random(0, width),
            y2: random(0, height),
            width: random(0.5, 1.5),
            alpha: random(0.1, 0.3),
            color: pick(colors.slice(0, 3))
        }));

        nodes = Array.from({ length: nodeCount }, (_, index) => {
            const anchor = lines[index % lines.length];
            const useLineAnchor = Math.random() > 0.35 && anchor;
            const t = Math.random();
            return {
                x: useLineAnchor ? anchor.x1 + (anchor.x2 - anchor.x1) * t + random(-12, 12) : random(0, width),
                y: useLineAnchor ? anchor.y1 + (anchor.y2 - anchor.y1) * t + random(-12, 12) : random(0, height),
                radius: random(1, 3),
                alpha: random(0.3, 0.6),
                pulse: Math.random() > 0.55,
                phase: random(0, Math.PI * 2),
                duration: random(3000, 5000),
                color: pick(colors)
            };
        });

        rings = Array.from({ length: ringCount }, () => ({
            x: random(width * 0.1, width * 0.9),
            y: random(height * 0.1, height * 0.9),
            radius: random(80, 200),
            alpha: random(0.04, 0.08),
            color: pick(colors.slice(0, 3))
        }));

        glitches = Array.from({ length: glitchCount }, () => ({
            x: random(0, width),
            y: random(0, height),
            length: random(20, 60),
            alpha: random(0.05, 0.15),
            color: pick(colors)
        }));

        radar = Array.from({ length: 2 }, (_, index) => ({
            x: random(width * 0.2, width * 0.8),
            y: random(height * 0.2, height * 0.8),
            maxRadius: random(150, 260),
            duration: random(4500, 6500),
            delay: index * 1800,
            color: pick(colors.slice(0, 2))
        }));
    }

    function draw(timestamp) {
        ctx.clearRect(0, 0, width, height);

        lines.forEach(line => {
            ctx.beginPath();
            ctx.moveTo(line.x1, line.y1);
            ctx.lineTo(line.x2, line.y2);
            ctx.lineWidth = line.width;
            ctx.strokeStyle = alphaColor(line.color, line.alpha);
            ctx.stroke();
        });

        rings.forEach(ring => {
            ctx.beginPath();
            ctx.arc(ring.x, ring.y, ring.radius, 0, Math.PI * 2);
            ctx.lineWidth = 1;
            ctx.strokeStyle = alphaColor(ring.color, ring.alpha);
            ctx.stroke();
        });

        radar.forEach(ring => {
            const elapsed = Math.max(0, (timestamp - ring.delay) % ring.duration);
            const progress = elapsed / ring.duration;
            const radius = ring.maxRadius * progress;
            const alpha = 0.1 * (1 - progress);

            ctx.beginPath();
            ctx.arc(ring.x, ring.y, radius, 0, Math.PI * 2);
            ctx.lineWidth = 1;
            ctx.strokeStyle = alphaColor(ring.color, alpha);
            ctx.stroke();
        });

        glitches.forEach(glitch => {
            ctx.beginPath();
            ctx.moveTo(glitch.x, glitch.y);
            ctx.lineTo(glitch.x + glitch.length, glitch.y);
            ctx.lineWidth = 1;
            ctx.strokeStyle = alphaColor(glitch.color, glitch.alpha);
            ctx.stroke();
        });

        nodes.forEach(node => {
            const pulse = node.pulse
                ? 0.35 + 0.15 * Math.sin((timestamp / node.duration) * Math.PI * 2 + node.phase)
                : node.alpha;

            ctx.beginPath();
            ctx.arc(node.x, node.y, node.radius, 0, Math.PI * 2);
            ctx.fillStyle = alphaColor(node.color, Math.max(0.2, Math.min(0.6, pulse)));
            ctx.fill();
        });

        if (running) {
            animationId = requestAnimationFrame(draw);
        }
    }

    function start() {
        if (running) return;
        running = true;
        animationId = requestAnimationFrame(draw);
    }

    function stop() {
        running = false;
        if (animationId) {
            cancelAnimationFrame(animationId);
            animationId = null;
        }
    }

    document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
            stop();
        } else {
            start();
        }
    });

    window.addEventListener('resize', resize);
    resize();
    animationId = requestAnimationFrame(draw);
})();
