/**
 * Render the home page with topic cards.
 * @param {Array} topics - array of topic objects
 */
function renderHome(topics) {
  const app = $('#app');
  if (!app) return;

  const seeded = (topics || []).filter(t => !t.isCustom);
  const custom = (topics || []).filter(t => t.isCustom);

  function topicCardHTML(topic) {
    return `
      <div class="topic-card${topic.isCustom ? ' topic-card-custom' : ''}" data-topic-id="${topic.id}">
        ${topic.isCustom ? '<span class="topic-custom-badge">My Topic</span>' : ''}
        <div class="topic-card-name">${escapeHtml(topic.name)}</div>
        <div class="topic-card-desc">${escapeHtml(topic.description || '')}</div>
        <div class="topic-card-actions">
          <button class="btn btn-primary btn-sm start-quiz-btn" data-topic-id="${topic.id}">
            Start Quiz
          </button>
          <button class="btn btn-outline btn-sm add-words-btn" data-topic-id="${topic.id}">
            + Add Words
          </button>
        </div>
      </div>`;
  }

  app.innerHTML = `
    <h2 class="page-title">Choose a Topic</h2>
    <p class="page-subtitle">Select a vocabulary topic to start a quiz. Test yourself in Hungarian!</p>

    ${seeded.length > 0 ? `<div class="topics-grid">${seeded.map(topicCardHTML).join('')}</div>` : ''}

    ${custom.length > 0 ? `
      <div class="section-heading" style="margin-top:2rem;">My Topics</div>
      <div class="topics-grid">${custom.map(topicCardHTML).join('')}</div>
    ` : ''}

    <div style="text-align:center; margin-top:2rem;">
      <a href="#custom-words" class="btn btn-outline">Manage My Words & Topics</a>
    </div>`;

  app.querySelectorAll('.start-quiz-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      window.location.hash = `#quiz?topic=${btn.dataset.topicId}`;
    });
  });

  app.querySelectorAll('.add-words-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      window.location.hash = `#custom-words?topic=${btn.dataset.topicId}`;
    });
  });

  app.querySelectorAll('.topic-card').forEach(card => {
    card.addEventListener('click', () => {
      window.location.hash = `#quiz?topic=${card.dataset.topicId}`;
    });
  });
}

/**
 * Render the quiz page.
 * Shows cards one at a time; collects answers; submits at the end.
 * @param {object} quizData - quiz response from API
 * @param {string} sessionId
 * @param {Function} onResults - callback with submit response
 */
function renderQuiz(quizData, sessionId, onResults) {
  const app = $('#app');
  if (!app) return;

  const { cards, topicName, topicId } = quizData;
  const totalCards = cards.length;
  let currentIndex = 0;
  const answers = [];

  function renderCard(index) {
    const card = cards[index];
    const cardNum = index + 1;
    const progressPct = Math.round((index / totalCards) * 100);
    const isLast = index === totalCards - 1;

    app.innerHTML = `
      <div class="quiz-wrapper">
        <div class="quiz-header">
          <div class="quiz-topic">${escapeHtml(topicName)}</div>
          <div class="quiz-progress-label">Card ${cardNum} of ${totalCards}</div>
          <div class="progress-bar-track">
            <div class="progress-bar-fill" style="width: ${progressPct}%"></div>
          </div>
        </div>

        <div class="quiz-card">
          <div class="quiz-prompt-label">Translate to English</div>
          ${card.isCustom ? '<span class="quiz-custom-badge">My Word</span>' : ''}
          <div class="quiz-hungarian-word">${escapeHtml(card.hungarian)}</div>
          <div class="quiz-input-group">
            <input
              type="text"
              class="quiz-input"
              id="answer-input"
              placeholder="Type English translation..."
              autocomplete="off"
              autocorrect="off"
              autocapitalize="off"
              spellcheck="false"
            />
            <button class="btn btn-primary btn-lg" id="next-btn">
              ${isLast ? 'Submit Quiz' : 'Next Card →'}
            </button>
          </div>
        </div>

        <div style="text-align:center;">
          <button class="btn btn-outline btn-sm" id="quit-btn">Quit Quiz</button>
        </div>
      </div>`;

    const input = $('#answer-input');
    const nextBtn = $('#next-btn');
    const quitBtn = $('#quit-btn');

    // Focus input
    if (input) input.focus();

    // Allow Enter to advance
    if (input) {
      input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          e.preventDefault();
          advance();
        }
      });
    }

    if (nextBtn) {
      nextBtn.addEventListener('click', advance);
    }

    if (quitBtn) {
      quitBtn.addEventListener('click', () => {
        window.location.hash = '#home';
      });
    }
  }

  async function advance() {
    const input = $('#answer-input');
    const userAnswer = input ? input.value : '';

    answers.push({
      cardId: cards[currentIndex].id,
      userAnswer: userAnswer
    });

    currentIndex++;

    if (currentIndex < totalCards) {
      renderCard(currentIndex);
    } else {
      // Submit answers
      showLoading();
      try {
        const result = await apiFetch('/api/submit', {
          method: 'POST',
          body: JSON.stringify({
            sessionId: sessionId,
            topicId: topicId,
            answers: answers
          })
        });
        hideLoading();
        onResults(result, topicId);
      } catch (err) {
        hideLoading();
        showError('Failed to submit quiz: ' + err.message);
        // Still show results with what we have if possible
      }
    }
  }

  renderCard(0);
}

/**
 * Render the results page.
 * @param {object} results - submit response from API
 * @param {number} topicId - topic ID for retake
 */
function renderResults(results, topicId) {
  const app = $('#app');
  if (!app) return;

  const pct = results.percentage;
  let pctClass = 'poor';
  let message = 'Keep practicing! You can do it!';
  if (pct >= 90) { pctClass = 'excellent'; message = 'Excellent work! You nailed it!'; }
  else if (pct >= 70) { pctClass = 'good'; message = 'Good job! Keep it up!'; }
  else if (pct >= 50) { pctClass = 'fair'; message = 'Not bad! A bit more practice will help.'; }

  app.innerHTML = `
    <div class="results-wrapper">
      <div class="results-score-card">
        <div class="results-percentage ${pctClass}">${pct}%</div>
        <div class="results-score-text">${results.score} / ${results.totalCards} correct</div>
        <div class="results-message">${message}</div>
        <div class="results-actions">
          <a href="#home" class="btn btn-outline">← Back to Topics</a>
          ${topicId ? `<button class="btn btn-primary" id="retake-btn">Retake Quiz</button>` : ''}
        </div>
      </div>

      <div class="section-heading">Results Breakdown</div>
      <div class="results-list">
        ${(results.results || []).map(r => `
          <div class="result-item ${r.isCorrect ? 'correct' : 'incorrect'}">
            <div class="result-icon">${r.isCorrect ? '✓' : '✗'}</div>
            <div class="result-details">
              <div class="result-hungarian">${escapeHtml(r.hungarian)}</div>
              <div class="result-answers">
                ${r.isCorrect
                  ? `<span class="result-answer-correct">${escapeHtml(r.userAnswer || '(blank)')}</span>`
                  : `<span class="result-answer-wrong">${escapeHtml(r.userAnswer || '(blank)')}</span>
                     → <span class="result-answer-correct">${escapeHtml(r.correct)}</span>`
                }
              </div>
            </div>
          </div>
        `).join('')}
      </div>
    </div>`;

  const retakeBtn = $('#retake-btn');
  if (retakeBtn && topicId) {
    retakeBtn.addEventListener('click', () => {
      window.location.hash = `#quiz?topic=${topicId}`;
    });
  }
}

/**
 * Render the custom words page.
 * @param {object} data - {count, words}
 * @param {Array} topics - array of topic objects
 * @param {string} sessionId
 * @param {number|null} filterTopicId - pre-selected topic filter
 */
function renderCustomWords(data, topics, sessionId, filterTopicId) {
  const app = $('#app');
  if (!app) return;

  const words = data.words || [];
  const topicsMap = {};
  (topics || []).forEach(t => { topicsMap[t.id] = t.name; });

  // Filter words by topic if needed
  let displayedWords = words;
  let activeTopicFilter = filterTopicId || '';

  function getTopicName(topicId) {
    if (!topicId) return null;
    return topicsMap[topicId] || null;
  }

  function render(wordsToShow, topicFilter) {
    app.innerHTML = `
      <div class="custom-words-wrapper">
        <h2 class="page-title">My Custom Words</h2>
        <p class="page-subtitle">Add your own vocabulary to practice alongside the built-in words.</p>

        <div class="add-word-form">
          <h3>Add New Word</h3>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label" for="hw-input">Hungarian</label>
              <input type="text" id="hw-input" class="form-input" placeholder="e.g. kutya" />
            </div>
            <div class="form-group">
              <label class="form-label" for="en-input">English</label>
              <input type="text" id="en-input" class="form-input" placeholder="e.g. dog" />
            </div>
          </div>
          <div class="form-group" style="margin-bottom:12px;">
            <label class="form-label" for="topic-select">Topic (optional)</label>
            <select id="topic-select" class="form-select">
              <option value="">-- No topic --</option>
              ${(topics || []).map(t => `<option value="${t.id}" ${String(topicFilter) === String(t.id) ? 'selected' : ''}>${escapeHtml(t.name)}${t.isCustom ? ' (mine)' : ''}</option>`).join('')}
              <option value="__new__">+ Create new topic...</option>
            </select>
          </div>
          <div class="form-group" id="new-topic-group" style="display:none; margin-bottom:12px;">
            <label class="form-label" for="new-topic-input">New topic name</label>
            <input type="text" id="new-topic-input" class="form-input" placeholder="e.g. Travel, Work, Sports..." />
          </div>
          <div class="form-actions">
            <button class="btn btn-primary" id="add-word-btn">Add Word</button>
          </div>
        </div>

        <div class="words-filter-bar">
          <label for="filter-topic">Filter by topic:</label>
          <select id="filter-topic" class="form-select" style="width:auto;">
            <option value="">All topics</option>
            ${(topics || []).map(t => `<option value="${t.id}" ${String(topicFilter) === String(t.id) ? 'selected' : ''}>${escapeHtml(t.name)}</option>`).join('')}
          </select>
          <span class="words-count-badge">${wordsToShow.length} word${wordsToShow.length !== 1 ? 's' : ''}</span>
        </div>

        <div class="words-list" id="words-list">
          ${wordsToShow.length === 0 ? `
            <div class="empty-state">
              <div class="empty-state-icon">📝</div>
              <div class="empty-state-text">No custom words yet.</div>
              <div class="empty-state-sub">Add your first word using the form above.</div>
            </div>` :
            wordsToShow.map(w => {
              const topicName = getTopicName(w.topicId);
              return `
              <div class="word-item" data-word-id="${w.id}">
                <div class="word-pair">
                  <span class="word-hungarian">${escapeHtml(w.hungarian)}</span>
                  <span class="word-arrow">→</span>
                  <span class="word-english">${escapeHtml(w.english)}</span>
                </div>
                ${topicName ? `<span class="word-topic-badge">${escapeHtml(topicName)}</span>` : ''}
                <button class="btn btn-danger delete-word-btn" data-word-id="${w.id}">Delete</button>
              </div>`;
            }).join('')
          }
        </div>
      </div>`;

    // Show/hide new topic input when "Create new topic..." is selected
    const topicSelectEl = $('#topic-select');
    if (topicSelectEl) {
      topicSelectEl.addEventListener('change', () => {
        const newTopicGroup = $('#new-topic-group');
        if (newTopicGroup) {
          newTopicGroup.style.display = topicSelectEl.value === '__new__' ? 'block' : 'none';
        }
      });
    }

    // Add word handler
    const addBtn = $('#add-word-btn');
    if (addBtn) {
      addBtn.addEventListener('click', async () => {
        const hungarian = $('#hw-input').value.trim();
        const english = $('#en-input').value.trim();
        const topicSelect = $('#topic-select');

        if (!hungarian || !english) {
          showError('Please enter both Hungarian and English words.');
          return;
        }

        try {
          addBtn.disabled = true;
          addBtn.textContent = 'Adding...';

          let topicId = null;

          if (topicSelect && topicSelect.value === '__new__') {
            // Create the new topic first
            const newTopicName = ($('#new-topic-input') || {}).value?.trim();
            if (!newTopicName) {
              showError('Please enter a name for the new topic.');
              addBtn.disabled = false;
              addBtn.textContent = 'Add Word';
              return;
            }
            const newTopic = await apiFetch('/api/topics', {
              method: 'POST',
              body: JSON.stringify({ sessionId, name: newTopicName })
            });
            topicId = newTopic.id;
            // Bust the topics cache so the new topic appears on home
            topics = [];
          } else if (topicSelect && topicSelect.value) {
            topicId = parseInt(topicSelect.value);
          }

          await apiFetch('/api/custom-words', {
            method: 'POST',
            body: JSON.stringify({ sessionId, hungarian, english, topicId })
          });

          // Reload the page content
          const newHash = '#custom-words' + (activeTopicFilter ? `?topic=${activeTopicFilter}` : '');
          if (window.location.hash === newHash) {
            window.dispatchEvent(new Event('hashchange'));
          } else {
            window.location.hash = newHash;
          }
        } catch (err) {
          showError(err.message);
          addBtn.disabled = false;
          addBtn.textContent = 'Add Word';
        }
      });
    }

    // Filter handler
    const filterSelect = $('#filter-topic');
    if (filterSelect) {
      filterSelect.addEventListener('change', async () => {
        const val = filterSelect.value;
        activeTopicFilter = val;
        showLoading();
        try {
          const path = `/api/custom-words?sessionId=${encodeURIComponent(sessionId)}${val ? `&topicId=${val}` : ''}`;
          const newData = await apiFetch(path);
          hideLoading();
          render(newData.words || [], val);
        } catch (err) {
          hideLoading();
          showError('Failed to filter words: ' + err.message);
        }
      });
    }

    // Delete handlers
    app.querySelectorAll('.delete-word-btn').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        const wordId = btn.dataset.wordId;
        if (!confirm('Delete this word?')) return;
        try {
          btn.disabled = true;
          await apiFetch(`/api/custom-words/${wordId}?sessionId=${encodeURIComponent(sessionId)}`, {
            method: 'DELETE'
          });
          // Remove the word item from DOM
          const item = app.querySelector(`.word-item[data-word-id="${wordId}"]`);
          if (item) item.remove();
          // Update count badge
          const remaining = app.querySelectorAll('.word-item').length;
          const badge = app.querySelector('.words-count-badge');
          if (badge) badge.textContent = `${remaining} word${remaining !== 1 ? 's' : ''}`;
          if (remaining === 0) {
            const list = $('#words-list');
            if (list) {
              list.innerHTML = `
                <div class="empty-state">
                  <div class="empty-state-icon">📝</div>
                  <div class="empty-state-text">No custom words yet.</div>
                  <div class="empty-state-sub">Add your first word using the form above.</div>
                </div>`;
            }
          }
        } catch (err) {
          showError('Failed to delete word: ' + err.message);
          btn.disabled = false;
        }
      });
    });
  }

  render(displayedWords, activeTopicFilter);
}
