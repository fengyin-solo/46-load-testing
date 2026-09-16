const API_KEY = 'loadtest-secret-key';

function headers() {
  return { 'X-API-Key': API_KEY, 'Content-Type': 'application/json' };
}

async function fetchJSON(url) {
  const res = await fetch(url, { headers: headers() });
  const body = await res.json();
  if (body.code !== 0) {
    throw new Error(body.message || '请求失败');
  }
  return body.data;
}

function pct(v) {
  if (v === null || v === undefined) return '-';
  return (v * 100).toFixed(2) + '%';
}

function fmtTime(s) {
  if (!s) return '-';
  return new Date(s).toLocaleString();
}

async function loadStats() {
  const stats = await fetchJSON('/api/stats/overview');
  document.getElementById('st-targets').textContent = stats.target_count;
  document.getElementById('st-scenarios').textContent = stats.scenario_count;
  document.getElementById('st-plans').textContent = stats.plan_count;
  document.getElementById('st-executors').textContent = stats.executor_count;
  document.getElementById('st-executions').textContent = stats.execution_count;
  document.getElementById('st-running').textContent = stats.running_executions;
  document.getElementById('st-requests').textContent = stats.total_requests;
  document.getElementById('st-error-rate').textContent = pct(stats.global_error_rate);
}

async function loadExecutions() {
  const data = await fetchJSON('/api/executions?size=50');
  const tbody = document.querySelector('#exec-list tbody');
  tbody.innerHTML = '';
  const items = data.items || [];
  items.forEach(e => {
    const tr = document.createElement('tr');
    tr.innerHTML =
      `<td>${e.id}</td>` +
      `<td>${e.plan_id}</td>` +
      `<td>${e.executor_id}</td>` +
      `<td class="status-${e.status}">${e.status}</td>` +
      `<td>${e.total_requests}</td>` +
      `<td>${e.error_count}</td>` +
      `<td>${e.total_requests > 0 ? ((e.error_count / e.total_requests) * 100).toFixed(2) + '%' : '-'}</td>` +
      `<td>${fmtTime(e.created_at)}</td>`;
    tbody.appendChild(tr);
  });
  if (items.length === 0) {
    tbody.innerHTML = '<tr><td colspan="8" class="empty">暂无执行记录</td></tr>';
  }
}

async function loadMetrics() {
  const data = await fetchJSON('/api/metrics?size=50');
  const tbody = document.querySelector('#metric-list tbody');
  tbody.innerHTML = '';
  const items = data.items || [];
  items.forEach(m => {
    const tr = document.createElement('tr');
    tr.innerHTML =
      `<td>${m.id}</td>` +
      `<td>${m.execution_id}</td>` +
      `<td>${m.tps.toFixed(2)}</td>` +
      `<td>${m.avg_latency_ms.toFixed(2)}</td>` +
      `<td>${m.p99_latency_ms.toFixed(2)}</td>` +
      `<td>${pct(m.error_rate)}</td>`;
    tbody.appendChild(tr);
  });
  if (items.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">暂无指标样本</td></tr>';
  }
}

async function load() {
  try {
    await Promise.all([loadStats(), loadExecutions(), loadMetrics()]);
  } catch (err) {
    document.getElementById('st-targets').textContent = 'ERR';
    console.error(err);
  }
}

document.getElementById('btn-refresh').addEventListener('click', load);
load();
