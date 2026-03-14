// Module-level state
let topics = [];
let lastQuizResults = null;
let lastQuizTopicId = null;
let sessionId = null;

/**
 * Parse hash-based route.
 * Returns { page, params }
 */
function parseHash() {
  const hash = window.location.hash.replace(/^#/, '');
  const [page, queryString] = hash.split('?');
  const params = {};
  if (queryString) {
    queryString.split('&').forEach(pair => {
      const [key, val] = pair.split('=');
      if (key) params[decodeURIComponent(key)] = decodeURIComponent(val || '');
    });
  }
  return { page: page || 'home', params };
}

/**
 * Initialize or restore the user session.
 */
async function initSession() {
  const stored = getSessionId();

  if (stored) {
    // Validate existing session
    try {
      const data = await apiFetch(`/api/session/validate?sessionId=${encodeURIComponent(stored)}`);
      if (data.valid) {
        sessionId = stored;
        return;
      }
    } catch (_) {
      // Fall through to create a new session
    }
  }

  // Create a new session
  const data = await apiFetch('/api/session', { method: 'POST' });
  sessionId = data.sessionId;
  setSessionId(sessionId);
}

/**
 * Load topics from the API (cached).
 */
async function loadTopics(force = false) {
  if (topics.length > 0 && !force) return topics;
  const data = await apiFetch(`/api/topics?sessionId=${encodeURIComponent(sessionId || '')}`);
  topics = data || [];
  return topics;
}

/**
 * Route to the appropriate page based on the current hash.
 */
async function route() {
  hideError();
  const { page, params } = parseHash();

  try {
    if (page === 'home' || page === '') {
      showLoading();
      const t = await loadTopics();
      hideLoading();
      renderHome(t);

    } else if (page === 'quiz') {
      const topicId = params.topic;
      if (!topicId) {
        // No topic specified — redirect to home
        window.location.hash = '#home';
        return;
      }

      showLoading();
      const includeCustom = true;
      const quizData = await apiFetch(
        `/api/quiz/${encodeURIComponent(topicId)}?sessionId=${encodeURIComponent(sessionId)}&includeCustom=${includeCustom}`
      );
      hideLoading();

      if (!quizData.cards || quizData.cards.length === 0) {
        showError('No cards available for this topic.');
        window.location.hash = '#home';
        return;
      }

      renderQuiz(quizData, sessionId, (results, tid) => {
        lastQuizResults = results;
        lastQuizTopicId = tid || topicId;
        renderResults(results, lastQuizTopicId);
        // Update hash without triggering another route
        history.replaceState(null, '', '#results');
      });

    } else if (page === 'results') {
      if (!lastQuizResults) {
        // No results stored — redirect to home
        window.location.hash = '#home';
        return;
      }
      renderResults(lastQuizResults, lastQuizTopicId);

    } else if (page === 'custom-words') {
      showLoading();
      const filterTopicId = params.topic || '';
      const [t, wordsData] = await Promise.all([
        loadTopics(),
        apiFetch(`/api/custom-words?sessionId=${encodeURIComponent(sessionId)}${filterTopicId ? `&topicId=${filterTopicId}` : ''}`)
      ]);
      hideLoading();
      renderCustomWords(wordsData, t, sessionId, filterTopicId);

    } else {
      // Unknown route — go home
      window.location.hash = '#home';
    }
  } catch (err) {
    hideLoading();
    showError('Error: ' + err.message);
    console.error('Route error:', err);
  }
}

/**
 * Application entry point.
 */
document.addEventListener('DOMContentLoaded', async () => {
  showLoading();

  try {
    await initSession();
  } catch (err) {
    hideLoading();
    showError('Failed to initialize session. Please refresh the page.');
    console.error('Session init error:', err);
    return;
  }

  hideLoading();

  // Listen to hash changes for navigation
  window.addEventListener('hashchange', route);

  // Initial route
  await route();
});
