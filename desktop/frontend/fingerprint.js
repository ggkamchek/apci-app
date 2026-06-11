async function sha256Hex(text) {
    const data = new TextEncoder().encode(text);
    const hash = await crypto.subtle.digest('SHA-256', data);
    return Array.from(new Uint8Array(hash))
        .map((b) => b.toString(16).padStart(2, '0'))
        .join('');
}

function canvasHash() {
    try {
        const canvas = document.createElement('canvas');
        canvas.width = 240;
        canvas.height = 60;
        const ctx = canvas.getContext('2d');
        if (!ctx) return 'canvas-unavailable';

        ctx.textBaseline = 'top';
        ctx.font = '16px Arial';
        ctx.fillStyle = '#f60';
        ctx.fillRect(0, 0, 120, 30);
        ctx.fillStyle = '#069';
        ctx.fillText('apci-fingerprint', 2, 2);
        ctx.strokeStyle = '#ff0';
        ctx.arc(60, 30, 20, 0, Math.PI * 2);
        ctx.stroke();

        return canvas.toDataURL();
    } catch {
        return 'canvas-error';
    }
}

function webGLInfo() {
    try {
        const canvas = document.createElement('canvas');
        const gl =
            canvas.getContext('webgl') ||
            canvas.getContext('experimental-webgl');
        if (!gl) return { vendor: 'webgl-unavailable', renderer: 'webgl-unavailable' };

        const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
        if (!debugInfo) {
            return { vendor: 'unknown-vendor', renderer: 'unknown-renderer' };
        }

        return {
            vendor: gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL) || '',
            renderer: gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) || '',
        };
    } catch {
        return { vendor: 'webgl-error', renderer: 'webgl-error' };
    }
}

async function audioContextHash() {
    try {
        const AudioCtx = window.AudioContext || window.webkitAudioContext;
        if (!AudioCtx) return 'audio-unavailable';

        const ctx = new AudioCtx();
        const oscillator = ctx.createOscillator();
        const analyser = ctx.createAnalyser();
        const gain = ctx.createGain();
        const processor = ctx.createScriptProcessor(256, 1, 1);

        gain.gain.value = 0;
        oscillator.type = 'triangle';
        oscillator.connect(analyser);
        analyser.connect(processor);
        processor.connect(gain);
        gain.connect(ctx.destination);
        oscillator.start(0);

        const samples = new Float32Array(analyser.frequencyBinCount);
        analyser.getFloatFrequencyData(samples);
        oscillator.stop();
        await ctx.close();

        return await sha256Hex(samples.slice(0, 32).join(','));
    } catch {
        return 'audio-error';
    }
}

async function fontsHash() {
    const testFonts = [
        'Arial',
        'Courier New',
        'Georgia',
        'Times New Roman',
        'Verdana',
        'Segoe UI',
        'Tahoma',
        'Consolas',
    ];
    const base = ['monospace', 'sans-serif', 'serif'];
    const testString = 'APCIwmil12';
    const body = document.body;
    const span = document.createElement('span');
    span.style.position = 'absolute';
    span.style.left = '-9999px';
    span.style.fontSize = '72px';
    span.textContent = testString;
    body.appendChild(span);

    const baseWidths = base.map((font) => {
        span.style.fontFamily = font;
        return span.getBoundingClientRect().width;
    });

    const detected = [];
    for (const font of testFonts) {
        for (let i = 0; i < base.length; i++) {
            span.style.fontFamily = `"${font}", ${base[i]}`;
            if (span.getBoundingClientRect().width !== baseWidths[i]) {
                detected.push(font);
                break;
            }
        }
    }

    body.removeChild(span);
    return await sha256Hex(detected.sort().join(','));
}

async function collectFingerprintSignals() {
    const webgl = webGLInfo();
    const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone || '';

    return {
        userAgent: navigator.userAgent || '',
        platform: navigator.platform || '',
        language: navigator.language || '',
        screenWidth: window.screen?.width || 0,
        screenHeight: window.screen?.height || 0,
        colorDepth: window.screen?.colorDepth || 0,
        timeZone,
        timezoneOffset: new Date().getTimezoneOffset(),
        canvasHash: canvasHash(),
        webGLVendor: webgl.vendor,
        webGLRenderer: webgl.renderer,
        audioContextHash: await audioContextHash(),
        fontsHash: await fontsHash(),
    };
}

let fingerprintInitPromise = null;

function initFingerprint() {
    if (!fingerprintInitPromise) {
        fingerprintInitPromise = (async () => {
            if (!window.go?.main?.App?.SetFingerprintSignals) {
                console.warn('fingerprint: Go binding unavailable');
                return false;
            }
            const signals = await collectFingerprintSignals();
            await window.go.main.App.SetFingerprintSignals(signals);
            return true;
        })().catch((err) => {
            console.error('fingerprint init failed', err);
            return false;
        });
    }
    return fingerprintInitPromise;
}
