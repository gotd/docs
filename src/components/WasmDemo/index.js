import React, {useEffect, useRef, useState} from 'react';
import BrowserOnly from '@docusaurus/BrowserOnly';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './styles.module.css';

// loadScript injects wasm_exec.js (which defines the global Go runtime shim)
// once and resolves when it is ready.
function loadScript(src) {
  return new Promise((resolve, reject) => {
    if (window.Go) {
      resolve();
      return;
    }
    const existing = document.querySelector('script[data-gotd-wasm-exec]');
    if (existing) {
      existing.addEventListener('load', () => resolve());
      existing.addEventListener('error', () => reject(new Error('failed to load ' + src)));
      return;
    }
    const s = document.createElement('script');
    s.src = src;
    s.async = true;
    s.dataset.gotdWasmExec = 'true';
    s.onload = () => resolve();
    s.onerror = () => reject(new Error('failed to load ' + src));
    document.head.appendChild(s);
  });
}

function Demo() {
  const wasmExecUrl = useBaseUrl('/wasm/wasm_exec.js');
  const wasmUrl = useBaseUrl('/wasm/main.wasm');

  // status: loading | ready | running | error
  const [status, setStatus] = useState('loading');
  const [error, setError] = useState('');
  // Public Telegram Desktop credentials, so the demo runs out of the box.
  const [appID, setAppID] = useState('17349');
  const [appHash, setAppHash] = useState('344583e45741c457fe1862106095a5eb');
  const [logs, setLogs] = useState([]);
  const logRef = useRef(null);

  // Instantiate the WASM module once. main() registers window.gotdConnect and
  // then blocks forever, so go.run never resolves — we deliberately do not await it.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        await loadScript(wasmExecUrl);
        if (cancelled) return;
        // eslint-disable-next-line no-undef
        const go = new window.Go();
        const result = await WebAssembly.instantiateStreaming(fetch(wasmUrl), go.importObject);
        if (cancelled) return;
        go.run(result.instance);
        setStatus('ready');
      } catch (e) {
        setError(String(e && e.message ? e.message : e));
        setStatus('error');
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [wasmExecUrl, wasmUrl]);

  // Keep the log view pinned to the bottom.
  useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight;
    }
  }, [logs]);

  const append = (line) => setLogs((prev) => [...prev, line]);

  const run = async () => {
    if (!appID.trim() || !appHash.trim()) {
      setLogs(['Enter both App ID and App Hash (from my.telegram.org/apps).']);
      return;
    }
    setLogs([]);
    setStatus('running');
    try {
      // gotdConnect(appID, appHash, onLog) streams log lines to append and
      // resolves with a short report.
      const report = await window.gotdConnect(appID.trim(), appHash.trim(), append);
      append('✓ ' + report);
    } catch (e) {
      append('✗ ' + e);
    } finally {
      setStatus('ready');
    }
  };

  const busy = status === 'loading' || status === 'running';
  const label =
    status === 'loading' ? 'Loading WASM…' : status === 'running' ? 'Running…' : 'Run';

  return (
    <div className={styles.demo}>
      <div className={styles.fields}>
        <label className={styles.field}>
          <span>App ID</span>
          <input
            type="number"
            placeholder="123456"
            value={appID}
            onChange={(e) => setAppID(e.target.value)}
            disabled={busy}
          />
        </label>
        <label className={styles.field}>
          <span>App Hash</span>
          <input
            type="text"
            placeholder="0123456789abcdef0123456789abcdef"
            value={appHash}
            onChange={(e) => setAppHash(e.target.value)}
            disabled={busy}
          />
        </label>
      </div>

      <button className={styles.run} onClick={run} disabled={busy || status === 'error'}>
        {label}
      </button>

      {status === 'error' && (
        <p className={styles.error}>Failed to load the demo: {error}</p>
      )}

      <pre className={styles.logs} ref={logRef}>
        {logs.length ? logs.join('\n') : 'Logs will appear here.'}
      </pre>

      <p className={styles.note}>
        Credentials are used only in your browser to build the client and are sent only to
        Telegram. The demo calls <code>help.getNearestDC</code>, which needs no login.
      </p>
    </div>
  );
}

export default function WasmDemo() {
  return (
    <BrowserOnly fallback={<div>Loading demo…</div>}>
      {() => <Demo />}
    </BrowserOnly>
  );
}
