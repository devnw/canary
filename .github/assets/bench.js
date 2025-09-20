// Benchmark page renderer: builds table + sparklines from bench JSON data.
;(async function () {
  const root = document.getElementById('bench-charts')
  if (!root) return
  function el(tag, attrs = {}, html = '') {
    const e = document.createElement(tag)
    for (const k in attrs) {
      e.setAttribute(k, attrs[k])
    }
    if (html) e.innerHTML = html
    return e
  }
  async function loadJSON(p) {
    try {
      const r = await fetch(p)
      if (!r.ok) return null
      return await r.json()
    } catch (_) {
      return null
    }
  }
  const summary = await loadJSON('summary.json')
  if (!summary) {
    root.textContent = 'No benchmark summary available.'
    return
  }
  if (!summary.benchmarks || !summary.benchmarks.length) {
    root.textContent = 'Benchmark summary empty.'
    return
  }
  root.textContent = ''
  function spark(values, width = 160, height = 40, pad = 3) {
    if (!values.length) return ''
    const min = Math.min(...values),
      max = Math.max(...values)
    const span = max - min || 1
    const pts = values
      .map((v, i) => {
        const x = pad + (i / (values.length - 1)) * (width - 2 * pad)
        const y = pad + (1 - (v - min) / span) * (height - 2 * pad)
        return x.toFixed(1) + ',' + y.toFixed(1)
      })
      .join(' ')
    const last = values[values.length - 1]
    const first = values[0]
    const diff = last - first
    const cls = diff < 0 ? 'better' : diff > 0 ? 'worse' : 'same'
    return `<svg viewBox="0 0 ${width} ${height}" width="${width}" height="${height}" class="spark ${cls}"><polyline fill="none" stroke="currentColor" stroke-width="1.2" points="${pts}"/></svg>`
  }
  const table = el('table')
  table.innerHTML =
    '<thead><tr><th>Name</th><th>ns/op</th><th>bytes/op</th><th>allocs/op</th><th>trend</th></tr></thead><tbody></tbody>'
  const tbody = table.querySelector('tbody')
  for (const b of summary.benchmarks) {
    const series = await loadJSON('data/' + b.file)
    if (!series || !series.length) continue
    const latest = series[series.length - 1]
    const nsSeries = series.map((r) => r.ns_per_op)
    const row = el('tr')
    row.innerHTML = `<td><code>${
      b.name
    }</code></td><td>${latest.ns_per_op.toLocaleString()}</td><td>${
      latest.bytes_per_op ?? ''
    }</td><td>${latest.allocs_per_op ?? ''}</td><td>${spark(nsSeries)}</td>`
    tbody.appendChild(row)
  }
  if (!tbody.children.length) {
    root.textContent = 'No benchmark series data.'
    return
  }
  const style = el(
    'style',
    {},
    '.spark{background:var(--md-code-bg-color);border:1px solid var(--md-default-fg-color--light);border-radius:2px;margin:0 2px}.spark.better polyline{stroke:var(--md-typeset-color-success,#00aa55)}.spark.worse polyline{stroke:var(--md-typeset-color-error,#cc3344)} table{width:100%;border-collapse:collapse;margin-top:.75rem;font-size:.8rem}th,td{padding:4px 6px;border-bottom:1px solid var(--md-default-fg-color--light);}'
  )
  root.appendChild(style)
  root.appendChild(table)
})()
