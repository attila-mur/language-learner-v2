// Base URL for API - empty string means same origin
const API_BASE = '';

/**
 * Fetch JSON from the API.
 * @param {string} path - API path (e.g. '/api/session')
 * @param {object} options - fetch options
 * @returns {Promise<any>} parsed JSON response
 */
async function apiFetch(path, options = {}) {
  const url = API_BASE + path;

  const fetchOptions = { ...options };

  // Set Content-Type for requests with a body
  if (fetchOptions.body && !fetchOptions.headers) {
    fetchOptions.headers = { 'Content-Type': 'application/json' };
  } else if (fetchOptions.body && fetchOptions.headers && !fetchOptions.headers['Content-Type']) {
    fetchOptions.headers['Content-Type'] = 'application/json';
  }

  const response = await fetch(url, fetchOptions);

  let data;
  const contentType = response.headers.get('content-type') || '';
  if (contentType.includes('application/json')) {
    data = await response.json();
  } else {
    data = await response.text();
  }

  if (!response.ok) {
    const message = (data && data.error) ? data.error : `HTTP ${response.status}: ${response.statusText}`;
    throw new Error(message);
  }

  return data;
}

/**
 * Get the stored session ID.
 * @returns {string|null}
 */
function getSessionId() {
  return localStorage.getItem('sessionId');
}

/**
 * Save a session ID to localStorage.
 * @param {string} id
 */
function setSessionId(id) {
  localStorage.setItem('sessionId', id);
}

/**
 * Show the full-page loading spinner.
 */
function showLoading() {
  const el = document.getElementById('loading');
  if (el) el.style.display = 'flex';
}

/**
 * Hide the full-page loading spinner.
 */
function hideLoading() {
  const el = document.getElementById('loading');
  if (el) el.style.display = 'none';
}

/**
 * Show an error banner at the top of the page.
 * @param {string} msg
 */
function showError(msg) {
  const el = document.getElementById('error-banner');
  if (!el) return;
  el.textContent = msg;
  el.style.display = 'block';
  // Auto-hide after 5 seconds
  setTimeout(() => {
    el.style.display = 'none';
  }, 5000);
}

/**
 * Hide the error banner.
 */
function hideError() {
  const el = document.getElementById('error-banner');
  if (el) el.style.display = 'none';
}

/**
 * Shorthand for document.querySelector
 * @param {string} selector
 * @returns {Element|null}
 */
function $(selector) {
  return document.querySelector(selector);
}

/**
 * Shorthand for document.querySelectorAll
 * @param {string} selector
 * @returns {NodeList}
 */
function $$(selector) {
  return document.querySelectorAll(selector);
}

/**
 * Escape HTML to prevent XSS.
 * @param {string} str
 * @returns {string}
 */
function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
