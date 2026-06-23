/* ─── ReadMeAI App.js ─────────────────────────────────────────────
   Vanilla JS SPA logic:
   - Theme toggle (localStorage)
   - URL validation + real-time feedback
   - Style selector
   - API fetch with animated loading steps
   - Markdown preview (Marked.js + Highlight.js)
   - Download as README.md
   - Copy to clipboard
   - Toast notifications
   - Repo info card + health score animation
────────────────────────────────────────────────────────────────── */

'use strict';

/* ─── State ─────────────────────────────────────────────────────── */
const state = {
  selectedStyle: 'developer',
  generatedReadme: '',
  repoData: null,
  isLoading: false,
  activeTab: 'preview',
};

/* ─── DOM References ─────────────────────────────────────────────── */
const $ = id => document.getElementById(id);

const els = {
  themeToggle:     $('theme-toggle'),
  repoUrl:         $('repo-url'),
  clearUrl:        $('clear-url'),
  urlValidation:   $('url-validation'),
  styleGrid:       $('style-grid'),
  generateBtn:     $('generate-btn'),
  repoInfoCard:    $('repo-info-card'),
  healthCard:      $('health-card'),
  healthScoreBadge:$('health-score-badge'),
  healthBarFill:   $('health-bar-fill'),
  suggestionsList: $('suggestions-list'),
  tabPreview:      $('tab-preview'),
  tabRaw:          $('tab-raw'),
  copyBtn:         $('copy-btn'),
  downloadBtn:     $('download-btn'),
  emptyState:      $('empty-state'),
  loadingState:    $('loading-state'),
  previewContent:  $('preview-content'),
  rawContent:      $('raw-content'),
  markdownPreview: $('markdown-preview'),
  rawCode:         $('raw-code'),
  // Repo card fields
  repoCardName:    $('repo-card-name'),
  repoCardLang:    $('repo-card-lang'),
  repoCardDesc:    $('repo-card-desc'),
  rcStars:         $('rc-stars'),
  rcForks:         $('rc-forks'),
  rcIssues:        $('rc-issues'),
  repoOwnerAvatar: $('repo-owner-avatar'),
  // Loading steps
  step1: $('step-1'),
  step2: $('step-2'),
  step3: $('step-3'),
  step4: $('step-4'),
};

/* ─── GitHub URL Validation ──────────────────────────────────────── */
const GITHUB_URL_REGEX = /^https?:\/\/github\.com\/([a-zA-Z0-9_.-]+)\/([a-zA-Z0-9_.-]+?)(\/.*)?(\?.*)?$/;

function validateURL(url) {
  url = url.trim();
  if (!url) return { valid: false, message: '' };
  if (!url.startsWith('http')) return { valid: false, message: 'URL must start with https://github.com/...' };
  if (!GITHUB_URL_REGEX.test(url)) return { valid: false, message: 'Enter a valid GitHub repository URL' };
  return { valid: true, message: '✓ Valid GitHub repository URL' };
}

/* ─── Theme ──────────────────────────────────────────────────────── */
function initTheme() {
  const saved = localStorage.getItem('readmeai-theme') || 'dark';
  document.documentElement.setAttribute('data-theme', saved);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute('data-theme');
  const next = current === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', next);
  localStorage.setItem('readmeai-theme', next);
}

/* ─── Style Selector ─────────────────────────────────────────────── */
function initStyleSelector() {
  els.styleGrid.addEventListener('click', e => {
    const card = e.target.closest('.style-card');
    if (!card) return;
    document.querySelectorAll('.style-card').forEach(c => {
      c.classList.remove('active');
      c.setAttribute('aria-checked', 'false');
    });
    card.classList.add('active');
    card.setAttribute('aria-checked', 'true');
    state.selectedStyle = card.dataset.style;
  });
}

/* ─── URL Input ──────────────────────────────────────────────────── */
function initURLInput() {
  let debounceTimer;

  els.repoUrl.addEventListener('input', () => {
    const val = els.repoUrl.value;
    els.clearUrl.style.display = val ? 'flex' : 'none';

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      const { valid, message } = validateURL(val);
      els.urlValidation.textContent = message;
      els.urlValidation.className = 'validation-msg ' + (valid ? 'valid' : (message ? 'error' : ''));
      els.repoUrl.classList.toggle('input-valid', valid);
      els.repoUrl.classList.toggle('input-error', !valid && !!message);
      els.generateBtn.disabled = !valid || state.isLoading;
    }, 300);
  });

  els.clearUrl.addEventListener('click', () => {
    els.repoUrl.value = '';
    els.clearUrl.style.display = 'none';
    els.urlValidation.textContent = '';
    els.urlValidation.className = 'validation-msg';
    els.repoUrl.classList.remove('input-valid', 'input-error');
    els.generateBtn.disabled = true;
    els.repoUrl.focus();
  });

  els.repoUrl.addEventListener('keydown', e => {
    if (e.key === 'Enter' && !els.generateBtn.disabled) {
      generate();
    }
  });

  // Sample URL buttons
  document.querySelectorAll('.sample-url').forEach(btn => {
    btn.addEventListener('click', () => {
      els.repoUrl.value = btn.dataset.url;
      els.repoUrl.dispatchEvent(new Event('input'));
      els.repoUrl.focus();
    });
  });
}

/* ─── Loading Steps Animation ────────────────────────────────────── */
let stepTimers = [];

function startLoadingSteps() {
  const steps = [els.step1, els.step2, els.step3, els.step4];
  steps.forEach(s => { s.classList.remove('active', 'done'); });
  steps[0].classList.add('active');

  const delays = [0, 4000, 10000, 20000];
  stepTimers.forEach(clearTimeout);
  stepTimers = [];

  delays.forEach((delay, i) => {
    if (i === 0) return;
    const timer = setTimeout(() => {
      if (i < steps.length) {
        steps[i - 1].classList.remove('active');
        steps[i - 1].classList.add('done');
        steps[i].classList.add('active');
      }
    }, delay);
    stepTimers.push(timer);
  });
}

function stopLoadingSteps() {
  stepTimers.forEach(clearTimeout);
  stepTimers = [];
}

/* ─── Generate ───────────────────────────────────────────────────── */
async function generate() {
  if (state.isLoading) return;
  const url = els.repoUrl.value.trim();
  const { valid } = validateURL(url);
  if (!valid) return;

  state.isLoading = true;
  setLoadingUI(true);
  startLoadingSteps();

  try {
    const res = await fetch('/api/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ repo_url: url, style: state.selectedStyle }),
    });

    const data = await res.json();

    if (!res.ok) {
      const msg = data.error || `Server error: ${res.status}`;
      throw new Error(msg);
    }

    state.generatedReadme = data.readme;
    state.repoData = data;

    renderRepoCard(data.repository);
    renderHealthCard(data.health_score);
    renderPreview(data.readme);
    showToast('README generated successfully! 🎉', 'success');

  } catch (err) {
    console.error(err);
    showErrorState(err.message);
    showToast(err.message || 'Generation failed. Please try again.', 'error');
  } finally {
    state.isLoading = false;
    setLoadingUI(false);
    stopLoadingSteps();
  }
}

function setLoadingUI(loading) {
  els.generateBtn.disabled = loading;
  els.generateBtn.classList.toggle('loading', loading);
  els.generateBtn.setAttribute('aria-busy', loading);

  if (loading) {
    showPanel('loading');
  }
}

/* ─── Repo Info Card ─────────────────────────────────────────────── */
function renderRepoCard(repo) {
  if (!repo) return;
  els.repoCardName.textContent = repo.full_name || repo.name;
  els.repoCardLang.textContent = repo.primary_language || 'Unknown';
  els.repoCardDesc.textContent = repo.description || 'No description provided';
  els.rcStars.textContent = formatNum(repo.stars);
  els.rcForks.textContent = formatNum(repo.forks);
  els.rcIssues.textContent = formatNum(repo.open_issues);

  // Avatar — use GitHub avatar CDN
  if (repo.owner) {
    els.repoOwnerAvatar.innerHTML = `<img src="https://github.com/${repo.owner}.png?size=64" alt="${repo.owner}" loading="lazy" onerror="this.style.display='none'" />`;
  }

  els.repoInfoCard.classList.remove('hidden');
}

/* ─── Health Card ────────────────────────────────────────────────── */
function renderHealthCard(health) {
  if (!health) return;
  const score = health.score || 0;
  els.healthScoreBadge.textContent = `${score}/100`;
  els.healthScoreBadge.className = 'health-badge' +
    (score >= 70 ? '' : score >= 45 ? ' score-mid' : ' score-low');

  // Animate bar
  requestAnimationFrame(() => {
    setTimeout(() => {
      els.healthBarFill.style.width = `${score}%`;
    }, 100);
  });

  // Suggestions
  els.suggestionsList.innerHTML = '';
  (health.suggestions || []).slice(0, 5).forEach(s => {
    const li = document.createElement('li');
    li.textContent = s;
    els.suggestionsList.appendChild(li);
  });

  els.healthCard.classList.remove('hidden');
}

/* ─── Preview Rendering ──────────────────────────────────────────── */
function renderPreview(markdown) {
  // Replace via.placeholder.com with placehold.co to avoid connection closed errors
  markdown = markdown.replace(/via\.placeholder\.com/g, 'placehold.co');

  // Fix literal '%' in shields.io badge URLs (replace % with %25 if not already encoded)
  markdown = markdown.replace(/(https?:\/\/img\.shields\.io\/[^\s\)]+)/g, (match) => {
    return match.replace(/%(?![0-9a-fA-F]{2})/g, '%25');
  });

  // Configure Marked
  marked.setOptions({
    gfm: true,
    breaks: true,
    highlight: (code, lang) => {
      if (lang && hljs.getLanguage(lang)) {
        try { return hljs.highlight(code, { language: lang }).value; } catch {}
      }
      return hljs.highlightAuto(code).value;
    },
  });

  els.markdownPreview.innerHTML = marked.parse(markdown);
  els.rawCode.textContent = markdown;

  // Show/hide action buttons
  els.copyBtn.classList.remove('hidden');
  els.downloadBtn.classList.remove('hidden');

  showPanel('preview');
}

function showErrorState(message) {
  els.markdownPreview.innerHTML = `
    <div style="text-align:center; padding:40px 20px; color:var(--accent-red);">
      <div style="font-size:2.5rem; margin-bottom:16px;">⚠️</div>
      <h3 style="font-size:1.1rem; margin-bottom:8px; color:var(--text-primary);">Generation Failed</h3>
      <p style="color:var(--text-secondary); max-width:400px; margin:0 auto; font-size:0.875rem;">${escapeHtml(message)}</p>
    </div>`;
  showPanel('preview');
}

/* ─── Panel Visibility ───────────────────────────────────────────── */
function showPanel(which) {
  els.emptyState.classList.add('hidden');
  els.loadingState.classList.add('hidden');
  els.previewContent.classList.add('hidden');
  els.rawContent.classList.add('hidden');

  if (which === 'empty') els.emptyState.classList.remove('hidden');
  else if (which === 'loading') els.loadingState.classList.remove('hidden');
  else if (which === 'preview') {
    if (state.activeTab === 'preview') els.previewContent.classList.remove('hidden');
    else els.rawContent.classList.remove('hidden');
  }
}

/* ─── Tab Switching ──────────────────────────────────────────────── */
function initTabs() {
  function switchTab(tab) {
    state.activeTab = tab;
    els.tabPreview.classList.toggle('active', tab === 'preview');
    els.tabRaw.classList.toggle('active', tab === 'raw');
    if (state.generatedReadme) {
      showPanel('preview');
    }
  }

  els.tabPreview.addEventListener('click', () => switchTab('preview'));
  els.tabRaw.addEventListener('click', () => switchTab('raw'));
}

/* ─── Copy & Download ────────────────────────────────────────────── */
function initActions() {
  els.copyBtn.addEventListener('click', async () => {
    if (!state.generatedReadme) return;
    try {
      await navigator.clipboard.writeText(state.generatedReadme);
      const original = els.copyBtn.innerHTML;
      els.copyBtn.innerHTML = `<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg> Copied!`;
      showToast('README copied to clipboard!', 'success');
      setTimeout(() => { els.copyBtn.innerHTML = original; }, 2000);
    } catch {
      showToast('Copy failed — use the Raw tab to select text manually.', 'error');
    }
  });

  els.downloadBtn.addEventListener('click', () => {
    if (!state.generatedReadme) return;
    const blob = new Blob([state.generatedReadme], { type: 'text/markdown;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'README.md';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast('README.md downloaded!', 'success');
  });
}

/* ─── Toast Notifications ────────────────────────────────────────── */
function showToast(message, type = 'info') {
  const container = $('toast-container');
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;

  const icons = {
    success: '✓',
    error: '✕',
    info: 'ℹ',
  };

  toast.innerHTML = `
    <span style="font-weight:700; flex-shrink:0;">${icons[type] || 'ℹ'}</span>
    <span>${escapeHtml(message)}</span>`;

  container.appendChild(toast);

  const dismiss = () => {
    toast.classList.add('toast-out');
    setTimeout(() => toast.remove(), 300);
  };

  toast.addEventListener('click', dismiss);
  setTimeout(dismiss, 5000);
}

/* ─── Helpers ────────────────────────────────────────────────────── */
function formatNum(n) {
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k';
  return String(n);
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.appendChild(document.createTextNode(String(str)));
  return div.innerHTML;
}

/* ─── Generate Button ────────────────────────────────────────────── */
function initGenerateButton() {
  els.generateBtn.addEventListener('click', generate);
}

/* ─── Init ───────────────────────────────────────────────────────── */
function init() {
  initTheme();
  initStyleSelector();
  initURLInput();
  initTabs();
  initActions();
  initGenerateButton();

  els.themeToggle.addEventListener('click', toggleTheme);
}

document.addEventListener('DOMContentLoaded', init);
