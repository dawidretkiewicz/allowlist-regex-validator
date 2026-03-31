const summaryEl = document.getElementById('summary');
const regexTableWrap = document.getElementById('regex-table-wrap');
const matchSummaryEl = document.getElementById('match-summary');
const matrixWrap = document.getElementById('matrix-wrap');
const resultsPanel = document.querySelector('.panel.results');
const matchBlock = document.getElementById('match-block');
const validationBlock = document.getElementById('validation-block');
const showMatchesOnlyCheckbox = document.getElementById('show-matches-only');
const presetSelect = document.getElementById('preset-allowlist');
const presetURL = document.getElementById('preset-url');

let presetItems = [];
let lastData = null;

function payloadFromInputs() {
  return {
    jsonRegexObject: document.getElementById('json-input').value,
    multilineRegexes: document.getElementById('regex-lines').value,
    multilineUrls: document.getElementById('url-lines').value,
    presetAllowlist: presetSelect.value,
    decodeJson: document.getElementById('decode-json').checked,
    decodeMultiline: document.getElementById('decode-multiline').checked,
  };
}

async function loadPresetAllowlists() {
  try {
    const resp = await fetch('/api/v1/allowlists');
    if (!resp.ok) {
      throw new Error('Failed to load presets');
    }
    const items = await resp.json();
    presetItems = Array.isArray(items) ? items : [];

    const options = ['<option value="">None</option>']
      .concat(
        presetItems.map(
          (item) => `<option value="${escapeHtml(item.key)}">${escapeHtml(item.label)}</option>`
        )
      )
      .join('');
    presetSelect.innerHTML = options;
    renderPresetURL();
  } catch (err) {
    presetURL.textContent = 'Could not load preset allowlists. You can still use manual JSON or multiline regexes.';
  }
}

function renderPresetURL() {
  if (!presetSelect.value) {
    presetURL.textContent = 'Select a preset allowlist to include in validation and matching.';
    return;
  }
  const selected = presetItems.find((item) => item.key === presetSelect.value);
  if (!selected) {
    presetURL.textContent = 'Selected preset not found.';
    return;
  }
  presetURL.textContent = `${selected.label}: ${selected.url}`;
}

async function run(path) {
  const resp = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payloadFromInputs()),
  });

  const data = await resp.json();
  if (!resp.ok) {
    renderError(data.error || 'Request failed');
    return;
  }

  render(data);
}

function renderError(message) {
  lastData = null;
  matchBlock.style.display = 'none';
  if (resultsPanel.firstElementChild !== validationBlock) {
    resultsPanel.insertBefore(validationBlock, resultsPanel.firstElementChild);
  }
  summaryEl.innerHTML = `<span class="pill bad">${message}</span>`;
  regexTableWrap.innerHTML = '';
  matchSummaryEl.innerHTML = '';
  matrixWrap.innerHTML = '';
}

function render(data) {
  lastData = data;
  if (data.match) {
    matchBlock.style.display = 'block';
    if (resultsPanel.firstElementChild !== matchBlock) {
      resultsPanel.insertBefore(matchBlock, resultsPanel.firstElementChild);
    }
  } else {
    matchBlock.style.display = 'none';
    if (resultsPanel.firstElementChild !== validationBlock) {
      resultsPanel.insertBefore(validationBlock, resultsPanel.firstElementChild);
    }
  }

  renderSummary(data.summary);
  renderRegexes(data.regexes);
  if (data.match) {
    renderMatchSummary(data.match.summaries, data.match.unmatchedUrls);
    renderMatrix(data);
  } else {
    matchSummaryEl.innerHTML = '';
    matrixWrap.innerHTML = '';
  }
}

function renderSummary(summary) {
  summaryEl.innerHTML = `
    <span class="pill">Total: ${summary.totalCount}</span>
    <span class="pill ok">Valid: ${summary.validCount}</span>
    <span class="pill bad">Invalid: ${summary.invalidCount}</span>
  `;
}

function escapeHtml(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');
}

function renderRegexes(regexes) {
  const rows = regexes
    .map(
      (r) => `<tr>
        <td>${r.id}</td>
        <td>${escapeHtml(r.label)}</td>
        <td><span class="code wrap-code">${escapeHtml(r.raw)}</span></td>
        <td><span class="code wrap-code">${escapeHtml(r.pattern)}</span></td>
        <td class="${r.valid ? 'ok' : 'bad'}">${r.valid ? 'valid' : 'invalid'}</td>
        <td>${escapeHtml(r.error || '')}</td>
      </tr>`
    )
    .join('');

  regexTableWrap.innerHTML = `
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th><th>Label</th><th>Raw</th><th>Decoded/Used</th><th>Status</th><th>Error</th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    </div>
  `;
}

function renderMatchSummary(summaries, unmatched) {
  const lines = summaries
    .map((s) => {
      const first = s.firstMatchId === undefined ? 'none' : s.firstMatchId;
      return `<li><span class="code">${escapeHtml(s.url)}</span> -> first=${first}, all=[${s.allMatchIds.join(', ')}]</li>`;
    })
    .join('');

  const unmatchedHtml = unmatched.length
    ? `<div class="pill bad">Unmatched URLs: ${unmatched.length}</div>`
    : '<div class="pill ok">All URLs matched at least one regex</div>';

  matchSummaryEl.innerHTML = `
    ${unmatchedHtml}
    <ul>${lines}</ul>
  `;
}

function renderMatrix(data) {
  const validRegexes = data.regexes.filter((r) => r.valid);

  let visibleIndexes = validRegexes.map((_, idx) => idx);
  if (showMatchesOnlyCheckbox.checked) {
    visibleIndexes = visibleIndexes.filter((colIdx) =>
      data.match.rows.some((row) => row.results[colIdx] === true)
    );
  }

  const visibleRegexes = visibleIndexes.map((idx) => validRegexes[idx]);
  const headCells = visibleRegexes
    .map((r, idx) => `<th title="Regex ID ${r.id}">R${idx + 1}</th>`)
    .join('');

  const rows = data.match.rows
    .map((row) => {
      const cells = visibleIndexes
        .map((colIdx) => row.results[colIdx])
        .map((ok) => `<td class="matrix-cell ${ok ? 'ok' : 'bad'}">${ok ? 'Y' : 'N'}</td>`)
        .join('');
      const safeURL = escapeHtml(row.url);
      return `<tr><td class="matrix-url" title="${safeURL}"><span class="code">${safeURL}</span></td>${cells}</tr>`;
    })
    .join('');

  const legend = visibleRegexes
    .map((r, idx) => `<li><span class="code">R${idx + 1}</span> = regex id ${r.id}</li>`)
    .join('');

  matrixWrap.innerHTML = `
    <div class="matrix-meta">
      <span class="pill">Rows: ${data.match.rows.length}</span>
      <span class="pill">Regex columns: ${visibleRegexes.length}/${validRegexes.length}</span>
      ${showMatchesOnlyCheckbox.checked ? '<span class="pill">Mode: matches only</span>' : '<span class="pill">Mode: all valid regexes</span>'}
    </div>
    <div class="table-wrap">
      <table class="matrix-table">
        <thead>
          <tr><th class="matrix-url-head">URL</th>${headCells}</tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    </div>
    <details class="matrix-legend">
      <summary>Regex Column Legend</summary>
      <ul>${legend || '<li>No columns in current mode.</li>'}</ul>
    </details>
  `;
}

document.getElementById('validate-btn').addEventListener('click', () => run('/api/v1/validate'));
document.getElementById('match-btn').addEventListener('click', () => run('/api/v1/match'));
presetSelect.addEventListener('change', renderPresetURL);
showMatchesOnlyCheckbox.addEventListener('change', () => {
  if (lastData && lastData.match) {
    renderMatrix(lastData);
  }
});

loadPresetAllowlists();
