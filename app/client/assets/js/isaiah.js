/**
 * This file holds absolutely all the logic for the app
 * on the frontend
 *
 * It is responsible for handling :
 * - keyboard navigation
 * - mouse navigation
 * - websocket transactions
 * - ui rendering
 * - remote commands execution
 *
 * The app logic is operated mostly like in a video game :
 * - First loop :
 *   - 1. key press
 *   - 2. run command(s)
 *   - 3. update local state
 *   - 4. refresh render
 * - Second loop :
 *   - 1. message received from server
 *   - 2. run command(s)
 *   - 3. update local state
 *   - 4. refresh render
 * - Third loop :
 *   - 1. mouse click
 *   - 2. run command(s)
 *   - 3. update local state
 *   - 4. refresh render
 *
 * A command usually falls into one of two categories :
 * - Local  : Update local state in prevision for future render
 * - Remote : Send a command to the server via websocket
 * - Additionally, a command can be private or public, a.k.a
 *   directly mapped to a key press (public), or used
 *   internally to facilitate some operations (private).
 */
((window) => {
  // === Handy methods and aliases

  /**
   * @param {string} s
   * @returns {HTMLElement}
   */
  const q = (s) => document.querySelector(s);

  /**
   * @param {string} s
   * @returns {Array<HTMLElement>}
   */
  const qq = (s) => [...document.querySelectorAll(s)];

  /**
   * @param {string} s
   * @returns {boolean}
   */
  const e = (s) => (document.querySelector(s) ? true : false);

  /**
   * Prevent xss
   * @param {string} str
   * @returns {string}
   */
  const s = (str) => {
    return str
      .replace(/javascript:/gi, '')
      .replace(/[^\w-_. ]/gi, function (c) {
        return `&#${c.charCodeAt(0)};`;
      });
  };

  /**
   * Prevent artifacts from CLI color codes
   * @param {string} str
   * @returns {string}
   */
  const removeEscapeSequences = (str) =>
    str.replace(
      /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g,
      ''
    );

  // === Handy HTML-querying methods

  /**
   * @returns {HTMLElement}
   */
  const hgetApp = () => q(`.app-wrapper`);

  /**
   * @returns {HTMLElement}
   */
  const hgetPopupContainer = () => q(`.popup-layer`);

  /**
   * @param {string} key
   * @returns {HTMLElement}
   */
  const hgetScreen = (key) => q(`.screen.for-${key}`);

  /**
   * @param {string} key
   * @returns {HTMLElement}
   */
  const hgetTab = (key) => q(`.tab.for-${key}`);

  /**
   * @param {string} key
   * @returns {HTMLElement}
   */
  const hgetPopup = (key) => q(`.popup.for-${key}`);

  /**
   * @param {string} key
   * @returns {Array<HTMLElement>}
   */
  const hgetTabRows = (key) => qq(`.tab.for-${key} .rows`);

  /**
   * @param {string} key
   * @returns {Array<HTMLElement>}
   */
  const hgetPopupRows = (key) => qq(`.popup.for-${key} .rows`);

  /**
   * @param {string} key
   * @param {number} index (1-indexed)
   * @returns {HTMLElement}
   */
  const hgetTabRow = (key, index) =>
    q(`.tab.for-${key} .row:nth-of-type(${index})`);

  /**
   * @param {string} key
   * @param {number} index (1-indexed)
   * @returns {HTMLElement}
   */
  const hgetPopupRow = (key, index) =>
    q(`.popup.for-${key} .row:nth-of-type(${index})`);

  /**
   * @param {string} key
   * @returns {HTMLElement}
   */
  const hgetHelper = (key) => q(`.help.for-${key}`);

  /**
   * @param {string} key
   * @returns {HTMLElement}
   */
  const hgetConnectionIndicator = (key) => q(`.indicator.for-${key}`);

  /**
   * @returns {Array<HTMLElement>}
   */
  const hgetConnectionIndicators = () => qq(`.indicator`);

  /**
   * @returns {HTMLElement}
   */
  const hgetTtyInput = () => q('#tty-input');

  /**
   * @returns {HTMLElement}
   */
  const hgetPromptInput = () => q('#prompt-input');

  /**
   * @returns {HTMLElement}
   */
  const hgetMobileControl = (action) =>
    q(`.mobile-controls button[data-action="${action}"]`);

  const hgetMobileControls = (action) =>
    qq(`.mobile-controls button[data-action]`);

  // === Render-related methods

  /**
   * @param {object} action
   * @param {string} action.Prompt
   * @param {string} action.PromptInput
   * @param {string} action.Command
   * @param {string} action.Key
   * @param {string} action.Label
   * @param {boolean} action.RequiresResource
   * @param {boolean} action.RunLocally
   * @param {number} maxKeyWidth
   * @returns {string}
   */
  const renderMenuAction = (action, maxKeyWidth) => {
    let html = `<button type="button" class="row"`;

    if (action.Prompt) html += ` data-prompt="${action.Prompt}"`;
    if (action.PromptInput)
      html += ` data-prompt-input="${action.PromptInput}"`;
    if (action.RequiresResource) html += ` data-use-row="true"`;
    if (action.RunLocally) html += ` data-run-locally="true"`;
    html += ` data-command="${action.Command}">`;

    if (action.Key)
      html += `<span class="cell">${s(
        action.Key.padEnd(maxKeyWidth + 1, whitespace)
      )}</span>`;
    html += `<span class="cell">${action.Label}</span>`;
    html += `</button>`;

    return html;
  };

  /**
   * @returns {string}
   */
  const renderMenuActionCancel = () => {
    return `<button type="button" class="row" data-action="reject" data-cancel>
              <span class="cell" data-action="reject">cancel</span>
            </button>`;
  };

  /**
   * @typedef {object} Cell
   * @property {string} field
   * @property {string} representation
   * @property {string} value
   */

  /**
   * @param {object} cell
   * @param {number} cell.Width
   * @param {Cell|string} cell.Content
   * @returns {string}
   */
  const renderCell = (cell) => {
    let html = `<div class="cell" data-navigate="cell" `;

    // Cell object
    if (typeof cell.Content === 'object') {
      html += `data-value="${s(cell.Content.value)}" `;
      html += `data-field="${s(cell.Content.field)}">`;

      if (!cell.Content.representation)
        html += `${s(cell.Content.value.padEnd(cell.Width, whitespace))}`;
      else
        html += `${s(
          cell.Content.representation.padEnd(cell.Width, whitespace)
        )}`;
    }
    // Raw string
    else {
      html += `>`;
      html += `${s(cell.Content.padEnd(cell.Width, whitespace))}`;
    }

    html += `</div>`;

    return html;
  };

  /**
   * @param {object} content
   * @param {boolean} isSubObject
   * @returns {string}
   */
  const renderJSON = (content, isSubObject = false) => {
    let html = '';
    for (const entry of Object.entries(content)) {
      // prettier-ignore
      html += `<div class="row ${!isSubObject ? '' : 'sub-row'} is-not-interactive is-colored is-json">`;
      html += `<div class="cell">${entry[0]}:</div>`;

      // Case when Array
      if (Array.isArray(entry[1])) {
        if (entry[1].length > 0)
          for (const cell of entry[1]) {
            if (typeof cell === 'object') html += renderJSON(cell, true);
            else
              html += `<div class="row sub-row is-not-interactive is-colored is-json">
                         <div class="cell">-</div>
                         <div class="cell is-array-value">${renderJSONCell(
                           cell
                         )}</div>
                       </div>`;
          }
        else html += `<div class="cell">[]</div>`;
        // Case when Object
      } else if (entry[1] !== null && typeof entry[1] === 'object')
        html += renderJSON(entry[1], true);
      // Case when flat value
      else {
        html += `<div class="cell">${renderJSONCell(entry[1])}</div>`;
      }

      html += `</div>`;
    }
    return html;
  };

  /**
   * @param {} cell
   * @returns {string}
   */
  const renderJSONCell = (cell) => {
    if (cell === null) return `null`;
    if (Array.isArray(cell)) {
      if (cell.length > 0)
        for (const v of cell)
          html += `<div class="row sub-row is-not-interactive is-colored is-json">
                       <div class="cell">-</div>
                       <div class="cell is-array-value">${renderJSONCell(
                         v
                       )}</div>
                    </div>`;
      else html += `<div class="cell">[]</div>`;
    }
    if (typeof cell === 'object') return renderJSON(cell, true);
    if (typeof cell === 'string') return cell ? `"${cell}"` : '""';
    return `${cell}`;
  };

  /**
   * Render rows with cell padding according to the longest cell of each column
   * @param {Array<Row>} rows
   * @returns {string}
   */
  const renderRows = (rows) => {
    let html = '';

    let maxs = [];
    for (let i = 0; i < rows[0]._representation.length; i++) maxs[i] = -1;

    // Find the max length of each column
    for (const row of rows) {
      for (const [index, cell] of row._representation.entries()) {
        // Cell object
        if (typeof cell === 'object') {
          if (!cell.representation)
            maxs[index] = Math.max(maxs[index], cell.value.length);
          else maxs[index] = Math.max(maxs[index], cell.representation.length);
        }
        // Raw string
        else maxs[index] = Math.max(maxs[index], cell.length);
      }
    }

    // Rows creation
    for (const row of rows) {
      html += '<div class="row" data-navigate="row">';
      for (const [index, cell] of row._representation.entries())
        html += renderCell({ Width: maxs[index], Content: cell });
      html += '</div>';
    }

    return html;
  };

  /**
   * @param {object} tab
   * @param {string} tab.Key
   * @param {string} tab.Title
   * @param {Array<Row>} tab.Rows
   * @returns {string}
   */
  const renderTab = (tab) => {
    let html = `<div class="tab for-${tab.Key}">`;
    html += `<button class="tab-title" data-navigate="tab.${tab.Key}">${tab.Title}</button>`;

    html += `<div class="tab-content">`;
    if (tab.Rows.length > 0) {
      html += renderRows(tab.Rows);
    }
    html += `</div>`;

    /*
    html += `<div class="tab-scroller">
               <div class="up">▲</div>
               <div class="track">
                 <div class="thumb"></div>
               </div>
               <div class="down">▼</div>
              </div>`;
    */
    html += `</div>`;

    return html;
  };

  /**
   * @param {Inspector} inspector
   * @returns {string}
   */
  const renderInspector = (inspector) => {
    let html = `<div class="tab for-inspector`;
    if (inspector.isEnabled) html += ` is-active`;
    html += ` " data-tab="${inspector.currentTab}">`;

    html += `<div class="tab-title-group">`;
    for (const tabName of inspector.availableTabs) {
      html += `<button class="tab-sub-title`;
      if (tabName === inspector.currentTab) html += ' is-active';
      html += `" data-navigate="inspector.${tabName}">${tabName}</button>`;
    }
    html += `</div>`;

    html += `<div class="tab-content">`;
    if (inspector.content.length > 0) {
      // Render Inspector Content
      for (const inspectorPart of inspector.content) {
        switch (inspectorPart.Type) {
          // Render Rows
          case 'rows':
            html += renderRows(inspectorPart.Content);
            break;

          // Render a table
          case 'table':
            html += `<table>`;
            html += `<thead>`;
            html += `<tr>`;
            for (const header of inspectorPart.Content.Headers)
              html += `<th>${header}</th>`;
            html += `</tr>`;
            html += `</thead>`;
            html += '<tbody>';
            for (const row of inspectorPart.Content.Rows) {
              html += `<tr>`;
              for (const cell of row) html += `<td>${cell}</td>`;
              html += `</tr>`;
            }
            html += '</tbody>';
            html += '</table>';
            break;

          // Render a JSON structure
          case 'json':
            html += renderJSON(inspectorPart.Content);
            break;

          // Render raw lines
          case 'lines':
            if (Array.isArray(inspectorPart.Content))
              for (const line of inspectorPart.Content)
                html += `<div class="row is-textual is-not-interactive">${line}</div>`;
            else
              html += `<div class="row is-textual is-not-interactive">${inspectorPart.Content}</div>`;
            break;
        }

        // Empty row separator between every content part (except for raw lines)
        if (inspectorPart.Type !== 'lines')
          html += `<div class="row is-not-interactive"></div>`;
      }
    }
    html += `</div>`;

    /*
    html += `<div class="tab-scroller">
               <div class="up">▲</div>
               <div class="track">
                 <div class="thumb"></div>
               </div>
               <div class="down">▼</div>
              </div>`;
    */
    html += `</div>`;

    return html;
  };

  /**
   * @param {Prompt} prompt
   * @returns {string}
   */
  const renderPrompt = (prompt) => {
    const classname = prompt.isForAuthentication ? 'for-login' : '';
    const title = prompt.input.isEnabled ? 'Input' : 'Confirm';
    const body = prompt.input.isEnabled
      ? `<div class="cell">${prompt.input.name}:</div><input placeholder="${
          prompt.input.placeholder
        } "type="${
          prompt.isForAuthentication ? 'password' : 'text'
        }" id="prompt-input"/>`
      : `<p class="request">${prompt.text}</p>`;

    return `
      <div class="popup for-prompt ${classname}">
        <div class="tab is-active">
          <span class="tab-title">${title}</span>
          <div class="tab-content">
            <div class="row is-not-interactive is-textual">
              ${body}
            </div>
          </div>
        </div>
      </div>
      `;
  };

  /**
   * @param {Message} message
   * @returns {string}
   */
  const renderMessage = (message) => {
    return `
      <div class="popup for-message" data-category="${message.category}" data-type="${message.type}">
        <div class="tab is-active">
          <span class="tab-title">${message.title}</span>
          <div class="tab-content">
            <div class="row is-not-interactive is-textual">
              <p>${message.content}</p>
            </div>
          </div>
        </div>
      </div>
      `;
  };

  /**
   * @param {"menu"|"bulk"}
   * @param {Menu} menu
   * @returns {string}
   */
  const renderMenu = (key, menu) => {
    // Cell padding according to the longest key
    let maxKeyWidth = -1;
    for (const action of menu.actions)
      maxKeyWidth = Math.max(maxKeyWidth, action.Key.length);

    return `
      <div class="popup for-${key}">
        <div class="tab is-active">
          <span class="tab-title">
            ${key === 'menu' ? 'Menu' : 'Bulk actions'}
          </span>
          <div class="tab-content">
            ${menu.actions
              .map((a, i) => renderMenuAction(a, maxKeyWidth))
              .join('')}
            ${renderMenuActionCancel()}
          </div>
        </div>
      </div>
      `;
  };

  /**
   * @param {TTY} tty
   * @returns {string}
   */
  const renderTty = (tty) => {
    const { history, historyCursor } = tty;
    const lines = tty.lines.map(removeEscapeSequences);

    let html = `
      <div class="popup for-tty">
        <div class="tab is-active">
          <span class="tab-title">
            Shell (${tty.type})
          </span>
          <div class="tab-content">`;

    if (lines.length > 0) {
      html += lines
        .slice(0, -1)
        .map((l) => `<div class="row is-not-interactive is-textual">${l}</div>`)
        .join('');

      html += `
              <div class="row is-not-interactive is-textual">
                <div class="cell">${lines[lines.length - 1].trim()}</div>
                <input
                  type="text"
                  id="tty-input"
                  value="${
                    historyCursor >= 0 && historyCursor < history.length
                      ? s(history[historyCursor])
                      : tty._tmpCommand
                      ? s(tty._tmpCommand)
                      : ''
                  }"/>
              </div>
      `;
    }

    html += `</div>
        </div>`;

    return html;
  };

  /**
   * @returns {string}
   */
  const renderHelp = () => `
      <div class="popup for-help">
        <div class="tab is-active">
          <span class="tab-title">
            Help
          </span>
          <div class="tab-content">
             <div class="row is-not-interactive">
               <span class="cell">Tab      </span>
               <span class="cell">switch panel</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">← →      </span>
               <span class="cell">switch panel / scroll</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">↑ ↓      </span>
               <span class="cell">switch row / scroll</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">[ ]      </span>
               <span class="cell">switch inspector tab</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">- +      </span>
               <span class="cell">switch layout</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">1234     </span>
               <span class="cell">go to panel</span>
             </div>
             <div class="row is-not-interactive"></div>
             <div class="row is-not-interactive">
               <span class="cell">y n      </span>
               <span class="cell">confirm/reject</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">Enter    </span>
               <span class="cell">confirm/submit</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">Escape   </span>
               <span class="cell">reject/exit inspector</span>
             </div>
             <div class="row is-not-interactive"></div>
             <div class="row is-not-interactive">
               <span class="cell">x        </span>
               <span class="cell">open menu</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">b        </span>
               <span class="cell">view bulk commands</span>
             </div>
             <div class="row is-not-interactive"></div>
             <div class="row is-not-interactive">
               <span class="cell">S        </span>
               <span class="cell">open system shell</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">R        </span>
               <span class="cell">reload everything/inspector</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">G        </span>
               <span class="cell">open project on Github</span>
             </div>
             <div class="row is-not-interactive"></div>
             <div class="row is-not-interactive">
               <span class="cell">Ctrl+C   </span>
               <span class="cell">clear command (shell)</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">Ctrl+L   </span>
               <span class="cell">clear screen (shell)</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">Ctrl+D   </span>
               <span class="cell">quit (shell)</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">↑ ↓      </span>
               <span class="cell">cycle history (shell)</span>
             </div>
             <div class="row is-not-interactive"></div>
             <div class="row is-not-interactive">
               <span class="cell">q        </span>
               <span class="cell">cancel/close/quit</span>
             </div>
          </div>
        </div>
      </div>
  `;

  /**
   * Main rendering function, responsible for updating the DOM
   * from scratch, using the supplied _state argument
   *
   * @param {state} _state
   */
  const renderApp = (_state) => {
    let html;

    // 0. Determine screen to display

    if (!_state.hasEstablishedConnection)
      hgetScreen('loading').classList.add('is-active');
    else {
      hgetScreen('loading').classList.remove('is-active');

      if (_state.isAuthenticated)
        hgetScreen('dashboard').classList.add('is-active');
    }

    // 1. Reset DOM

    // 1.0. Set app layout
    hgetScreen('dashboard').setAttribute(
      'data-layout',
      _state.appearance.currentLayout
    );

    // 1.1. Erase overview tabs
    hgetScreen('dashboard').querySelector('.left').innerHTML = '';

    // 1.2. Erase inspector
    hgetScreen('dashboard').querySelector('.right').innerHTML = '';

    // 1.3. Erase popup
    hgetPopupContainer().innerHTML = '';

    // 1.4. Hide popup layer
    hgetPopupContainer().classList.remove('is-active');

    // 1.5. Hide helper
    if (e('.help.is-active'))
      q('.help.is-active').classList.remove('is-active');

    // 1.6. Hide connection indicators
    hgetConnectionIndicators().forEach((i) => i.classList.remove('is-active'));

    // 2. Build every tab
    html = _state.tabs.map(renderTab).join('');
    hgetScreen('dashboard').querySelector('.left').innerHTML = html;

    // 3. Build inspector
    html = renderInspector(_state.inspector);
    hgetScreen('dashboard').querySelector('.right').innerHTML = html;

    // 4. Build current popup
    if (state.popup) {
      hgetPopupContainer().classList.add('is-active');

      html = '';

      // 4.1. Popup - Prompt
      if (_state.prompt.isEnabled) html = renderPrompt(_state.prompt);
      // 4.2. Popup - Message
      else if (_state.message.isEnabled) html = renderMessage(_state.message);
      // 4.3. Popup - Menu
      else if (_state.menu.actions.length > 0)
        html = renderMenu(_state.popup, _state.menu);
      // 4.4. Popup - Tty
      else if (_state.tty.isEnabled) html = renderTty(_state.tty);
      // 4.5. Popup - Help
      else if (_state.popup === 'help') html = renderHelp();

      hgetPopupContainer().innerHTML = html;
    }

    // 5. Set focus on DOM elements

    // 5.1. Set focus on inspector
    if (_state.inspector.isEnabled)
      hgetTab('inspector').classList.add('is-active');

    // 5.1.1. Scroll horizontally on the inspector
    hgetTab('inspector').querySelector('.tab-content').scrollLeft =
      _state.inspector.horizontalScroll;

    // 5.1.2. Scroll vertically on the inspector
    hgetTab('inspector').querySelector('.tab-content').scrollTop =
      _state.inspector.verticalScroll;

    // 5.1.3. Scroll down the inspector if it's for logs
    if (_state.inspector.content.length > 0) {
      if (_state.inspector.content[0].Type === 'lines') {
        const _inspector = hgetTab('inspector');
        const _content = _inspector.querySelector('.tab-content');
        _content.scrollTo(
          _state.inspector.horizontalScroll,
          _content.scrollHeight
        );
      }
    }

    // 5.2. Set focus on tab
    if (_state.navigation.currentTab) {
      hgetTab(_state.navigation.currentTab).classList.add('is-active');

      // 5.2.1. Set focus on row - Tab
      const currentRow = hgetTabRow(
        _state.navigation.currentTab,
        _state.navigation.currentTabsRows[_state.navigation.currentTab]
      );
      currentRow.classList.add('is-active');
      currentRow.scrollIntoView({ block: 'nearest', inline: 'nearest' });

      /*
       * Disabled part - Works for panels, but not for inspector
       *                 Should be reworked for handling any type of
       *                 content
       *                 Native scroll is used in place
       */

      /*
        // 5.2.2. Update current tab scroll indicator
        const currentTab = hgetTab(_state.navigation.currentTab);
        const currentTabContent = currentTab.querySelector('.tab-content');

        // 5.2.2.1. Define whether the scrollbar should show
        if (currentTabContent.scrollHeight > currentTabContent.clientHeight) {
          currentTab.classList.add('is-scrollable');

          // 5.2.2.1.1. Define the height and position of the thumb
          const thumb = currentTab.querySelector('.thumb');
          const trackHeight = currentTab.querySelector('.track').clientHeight;

          const rowCount = currentTabContent.children.length;
          const rowHeight = currentRow.clientHeight;
          const rowIndex =
            Array.from(currentTabContent.children).indexOf(currentRow) + 1;

          const visibleRowCount = Math.floor(
            currentTabContent.clientHeight / rowHeight
          );
          const stepCount = Math.ceil(rowCount / visibleRowCount) + 1; // + 1 : account for last row mechanism
          const stepHeight = trackHeight / stepCount;

          thumb.style.top = '0';
          thumb.style.height = `${stepHeight}px`;

          if (rowIndex > visibleRowCount) {
            // This mechanism makes the thumb reach the end only when the last row is focused
            if (rowIndex === rowCount)
              thumb.style.top = `${stepHeight * (stepCount - 1)}px`;
            else
              thumb.style.top = `${
                stepHeight * Math.floor(rowIndex / visibleRowCount)
              }px`;
          }
        }
      */
    }

    // 5.3. Set focus on menu
    if (_state.isMenuIng) {
      // 5.3.1. Set focus on row - Menu
      if (_state.menu.actions.length > 0)
        hgetPopupRow(
          _state.popup,
          _state.navigation.currentMenuRow
        ).classList.add('is-active');
    }

    // 5.4. Set focus on tty
    if (_state.tty.isEnabled && _state.tty.lines.length > 0) {
      const ttyInput = hgetTtyInput();
      ttyInput.focus();
      ttyInput.scrollIntoView({ block: 'nearest', inline: 'nearest' });

      // 5.4.1. Focus end of input
      setTimeout(() => {
        ttyInput.selectionStart = ttyInput.selectionEnd = ttyInput.value.length;
      }, _state._delays.default / 2);
    }

    // 5.5. Set focus on prompt input
    if (_state.prompt.isEnabled && _state.prompt.input.isEnabled) {
      // Dev-only (case when the server spontaneously tells us we're authenticated)
      if (hgetPromptInput()) hgetPromptInput().focus();
    }

    // 5.6. Set flag on previous/current tab for the "focus" layout
    if (_state.navigation.currentTab)
      hgetTab(_state.navigation.currentTab).classList.add('is-current');
    else if (_state.navigation.previousTab)
      hgetTab(_state.navigation.previousTab).classList.add('is-current');

    // 6. Set helper
    hgetHelper(_state.helper).classList.add('is-active');

    // 7. Set connection indicator
    hgetConnectionIndicator(
      _state.isConnected ? 'connected' : 'disconnected'
    ).classList.add('is-active');
    if (_state.isLoading)
      hgetConnectionIndicator('loading').classList.add('is-active');

    // 8. Reset mobile controls' visibility
    hgetMobileControls().forEach((e) => {
      e.classList.remove('is-active');
    });

    // 9. Update the mobile controls' visibility

    // 9.1. Case when menuing
    if (_state.isMenuIng) {
      hgetMobileControl('reject').classList.add('is-active');
      hgetMobileControl('confirm').classList.add('is-active');
    }
    // 9.2. Case when tty-ing / showing a message
    else if (_state.popup === 'tty' || _state.popup === 'message') {
      hgetMobileControl('_ttyQuit').classList.add('is-active');
    }
    // 9.3. Case when prompting
    else if (_state.prompt.isEnabled || _state.prompt.input.isEnabled) {
      hgetMobileControl('reject').classList.add('is-active');
      hgetMobileControl('confirm').classList.add('is-active');
    }
    // 9.4. Every other case (default navigation)
    else {
      hgetMobileControl('previousTab').classList.add('is-active');
      hgetMobileControl('nextTab').classList.add('is-active');
      hgetMobileControl('menu').classList.add('is-active');
      hgetMobileControl('bulk').classList.add('is-active');
      hgetMobileControl('shellSystem').classList.add('is-active');
    }
  };

  // === Websocket-related methods

  /**
   * Initiate a Websocket connection with the remote server
   */
  const websocketConnect = () => {
    const socket = new WebSocket(
      `${!wsSSL ? 'ws' : 'wss'}://${wsHost}:${wsPort}/ws`
    );
    socket.onopen = listenerSocketOpen;
    socket.onmessage = listenerSocketMessage;
    socket.onerror = listenerSocketError;
    socket.onclose = listenerSocketClose;
    wsSocket = socket;
  };

  /**
   * Send an object as a JSON string over Websocket
   * @param {object} o
   */
  const websocketSend = (o) => {
    wsSocket.send(JSON.stringify(o));
  };

  // === State

  let state = {
    /**
     * @type {boolean}
     */
    hasEstablishedConnection: false,

    /**
     * @type {boolean}
     */
    isConnected: false,

    /**
     * @type {boolean}
     */
    isAuthenticated: false,

    /**
     * @type {boolean}
     */
    isLoading: false,

    /**
     * @returns {boolean}
     */
    get isMenuIng() {
      return state.popup && ['menu', 'bulk'].includes(state.popup);
    },

    /**
     * @typedef {object} Prompt
     * @property {string} text
     * @property {PromptInput} input
     * @property {function} callback
     * @property {Array} callbackArgs
     * @property {boolean} isEnabled
     * @property {boolean} isForAuthentication
     */

    /**
     * @typedef {object} PromptInput
     * @property {boolean} isEnabled
     * @property {string} placeholder
     * @property {string} name
     */

    /**
     * @type {Prompt}
     */
    prompt: {
      text: null,
      input: { isEnabled: false, name: null, placeholder: null },
      callback: null,
      callbackArgs: [],
      isEnabled: false,
    },

    /**
     * @typedef {object} Menu
     * @property {Array<MenuAction>} actions
     */

    /**
     * @typedef {object} MenuAction
     * @property {string} Prompt
     * @property {string} Command
     * @property {string} Key
     * @property {string} Label
     * @property {boolean} RequiresResource
     * @property {boolean} RunLocally
     */

    /**
     * @type {Menu}
     */
    menu: {
      /**
       * @type {Array<MenuAction>}
       */
      actions: [],
    },

    /**
     * @type {'default'|'menu'|'prompt'|'prompt-input'|'message'}
     */
    helper: 'default',

    /**
     * @type {"menu"|"bulk"|"prompt"|"message"|"tty"|"help"}
     */
    popup: null,

    appearance: {
      /**
       * @type {Array<string>}
       */
      availableLayouts: ['default', 'half', 'focus'],

      /**
       * @type {"default"|"half"|"focus"}
       */
      currentLayout: 'default',
    },

    /**
     * @typedef {object} Message
     * @property {boolean} isEnabled
     * @property {string} category
     * @property {string} type
     * @property {string} title
     * @property {string} content
     */

    /**
     * @type {Message}
     */
    message: {
      isEnabled: false,
      category: null,
      type: null,
      title: null,
      content: null,
    },

    /**
     * @typedef Navigation
     * @property {string} currentTab
     * @property {string} previousTab
     * @property {object.<string, number>} currentTabsRows - Used for tabs
     * @property {number} currentMenuRow
     * @property {number} previousMenuRow
     */

    /**
     * @type {Navigation}
     */
    navigation: {
      currentTab: null,
      previousTab: null,
      currentTabsRows: {},
      currentMenuRow: null,
      previousMenuRow: null,
    },

    /**
     * @typedef Inspector
     * @property {boolean} isEnabled
     * @property {boolean} wasEnabled
     * @property {string} currentTab
     * @property {string} previousTab
     * @property {Array<string>} availableTabs
     * @property {Array<Row|Table|object>} content
     * @property {number} horizontalScroll
     * @property {number} verticalScroll
     */

    /**
     * @type Inspector
     */
    inspector: {
      isEnabled: false,
      wasEnabled: false,
      currentTab: null,
      previousTab: null,
      availableTabs: [],
      content: [],
      horizontalScroll: 0,
      verticalScroll: 0,
    },

    /**
     * @typedef Row
     * @property {Array<string>} _representation
     */

    /**
     * @typedef Tab
     * @property {string} Key
     * @property {string} Title
     * @property {Array<Row>} Rows
     */

    /**
     * @type {Array<Tab>}
     */
    tabs: [],

    /**
     * @typedef TTY
     * @property {bool} isEnabled
     * @property {Array<string>} lines
     * @property {Array<string>} history
     * @property {number} historyCursor
     * @property {"system"|"container"|"volume"} type
     * @property {string} _buffer
     * @property {string} _tmpCommand
     */

    /**
     * @type TTY
     */
    tty: {
      isEnabled: false,
      lines: [],
      history: [],
      historyCursor: -1,
      type: null,
      _buffer: '',
      _tmpCommand: null,
    },

    _delays: {
      /**
       * @type {number}
       */
      forAuthentication: 2000,

      /**
       * @type {number}
       */
      forTTYBufferClear: 50,

      /**
       * @type {number}
       */
      forTTYInputFocus: 50,

      /**
       * @type {number}
       */
      default: 250,
    },
  };

  // === State-related handy methods

  /**
   * @returns {string}
   */
  const sgetCurrentTabKey = () => {
    return !state.navigation.currentTab && state.navigation.previousTab
      ? state.navigation.previousTab
      : state.navigation.currentTab;
  };

  /**
   * @returns {Tab}
   */
  const sgetCurrentTab = () => {
    const currentTabKey = sgetCurrentTabKey();
    return state.tabs.find((t) => t.Key === currentTabKey);
  };

  /**
   * @returns {Row}
   */
  const sgetCurrentRow = () => {
    const currentTab = sgetCurrentTab();
    return currentTab.Rows[
      state.navigation.currentTabsRows[currentTab.Key] - 1
    ];
  };

  // === Commands-related methods

  /**
   * @param {string} cmd
   * @returns {boolean}
   */
  const cmdAllowed = (cmd) => {
    // Prevent running anyrhing when not connected to the remote server
    if (!state.isConnected) return false;

    // Prevent running anything while loading
    if (state.isLoading) return false;

    // if (
    //   state.isLoading &&
    //   ![
    //     'scrollUp',
    //     'scrollDown',
    //     'scrollLeft',
    //     'scrollRight',
    //     'nextTab',
    //     'previousTab',
    //     'nextSubTab',
    //     'previousSubTab',
    //   ].includes(cmd)
    // )
    //   return false;

    // Prevent running anything other than submit while unauthenticated
    if (!state.isAuthenticated && !['confirm'].includes(cmd)) return false;

    // Force yes/no on prompts && messages
    if (
      (state.prompt.isEnabled ||
        state.message.isEnabled ||
        (state.popup && !state.isMenuIng)) &&
      !['confirm', 'reject', 'quit'].includes(cmd)
    )
      return false;

    // Force yes/no/arrows on menus,
    if (
      state.isMenuIng &&
      !['scrollUp', 'scrollDown', 'confirm', 'reject', 'quit'].includes(cmd)
    )
      return false;

    // Prevent multi-popup
    if (
      state.popup &&
      ['menu', 'bulk', 'prompt', 'message', 'shell'].includes(cmd)
    )
      return false;

    // Prevent multi-inspect
    if (state.inspector.isEnabled && ['confirm'].includes(cmd)) return false;

    return true;
  };

  /**
   * @param {function} cmd
   * @param {Array} args
   */
  const cmdRun = (cmd, ...args) => {
    const cmdString = cmd.name;
    const isKnown = cmdString in cmds;

    // console.log(cmdString);

    if (!isKnown) return;

    const isPrivate = cmdString[0] === '_';
    const isAllowed = isPrivate ? true : cmdAllowed(cmdString);

    if (!isAllowed) return;

    cmd(...args);

    renderApp(state);
  };

  // === Commands

  const cmds = {
    /**
     * Private - An empty function used to trigger the render loop
     */
    _render: function () {},

    /**
     * Private - Request the initial data from the server
     */
    _init: function () {
      websocketSend({ action: 'init' });
    },

    /**
     * Private - Log out
     */
    _exit: function () {
      state.isAuthenticated = false;
      websocketSend({ action: 'auth.logout' });

      setTimeout(() => {
        cmdRun(cmds._showAuthentication);
      }, state._delays.forTTYInputFocus);
    },

    /**
     * Private - Show a blocking prompt to the user
     * @param {Prompt} args
     */
    _showPrompt: function (args) {
      state.popup = 'prompt';
      state.helper = args.input ? 'prompt-input' : 'prompt';

      state.prompt.text = args.text;
      state.prompt.callback = args.callback;
      state.prompt.callbackArgs = args.callbackArgs || [];
      state.prompt.isEnabled = true;
      state.prompt.input = args.input || {
        isEnabled: false,
        placeholder: null,
        name: null,
      };
      state.prompt.isForAuthentication = args.isForAuthentication || false;

      if (state.navigation.currentTab) {
        state.navigation.previousTab = state.navigation.currentTab;
        state.navigation.currentTab = null;
      }

      if (state.navigation.currentMenuRow) {
        state.navigation.previousMenuRow = state.navigation.currentMenuRow;
        state.navigation.currentMenuRow = null;
      }

      if (state.inspector.isEnabled) {
        state.inspector.wasEnabled = true;
        state.inspector.isEnabled = false;
      }
    },

    /**
     * Private - Clear the last prompt and get back to previous display
     */
    _clearPrompt: function () {
      state.popup = null;
      state.helper = 'default';

      state.prompt.text = null;
      state.prompt.callback = null;
      state.prompt.callbackArgs = [];
      state.prompt.isEnabled = false;
      state.prompt.isForAuthentication = false;
      state.prompt.input = { isEnabled: false, name: null, placeholder: null };

      if (!state.inspector.wasEnabled) {
        state.navigation.currentTab = state.navigation.previousTab;
        state.navigation.currentMenuRow = state.navigation.previousMenuRow;
      } else {
        state.inspector.isEnabled = state.inspector.wasEnabled;
        state.inspector.wasEnabled = false;
      }
    },

    /**
     * Private - Show a popup
     * @param {string} key
     */
    _showPopup: function (key) {
      state.popup = key;

      if (state.navigation.currentTab) {
        state.navigation.previousTab = state.navigation.currentTab;
        state.navigation.currentTab = null;
      }

      if (state.navigation.currentMenuRow) {
        state.navigation.previousMenuRow = state.navigation.currentMenuRow;
        state.navigation.currentMenuRow = 1;
      }

      if (state.inspector.isEnabled) {
        state.inspector.wasEnabled = true;
        state.inspector.isEnabled = false;
      }
    },

    /**
     * Private - Clear the last popup and get back to previous display
     */
    _clearPopup: function () {
      state.popup = null;
      state.helper = 'default';
      state.menu.actions = [];

      if (!state.inspector.wasEnabled) {
        state.navigation.currentTab = state.navigation.previousTab;
        state.navigation.currentMenuRow = state.navigation.previousMenuRow;
      } else {
        state.inspector.isEnabled = state.inspector.wasEnabled;
        state.inspector.wasEnabled = false;
      }
    },

    /**
     * Private - Clear the message popup and get back to the previous display
     */
    _clearMessage: function () {
      state.popup = null;
      state.helper = 'default';

      state.message.isEnabled = false;
      state.message.type = null;
      state.message.title = null;
      state.message.content = null;

      state.navigation.currentTab = state.navigation.previousTab;
      state.navigation.currentMenuRow = state.navigation.previousMenuRow;
    },

    /**
     * Private - Enter inspection mode
     */
    _enterInspect: function () {
      state.inspector.isEnabled = true;

      // No need to check for currentTab as inspection
      // is only possible from a row contained inside a tab
      state.navigation.previousTab = state.navigation.currentTab;
      state.navigation.currentTab = null;
    },

    /**
     * Private - Leave inspection mode and get back to the previous display
     */
    _exitInspect: function () {
      state.inspector.isEnabled = false;
      state.navigation.currentTab = state.navigation.previousTab;
      state.navigation.currentMenuRow = state.navigation.previousMenuRow;
    },

    /**
     * Private - Send data through the current Websocket connection
     * @param {object} object
     */
    _wsSend: function (object) {
      websocketSend(object);
    },

    /**
     * Private - Enable tty mode
     */
    _ttyStart: function (type) {
      state.tty.isEnabled = true;
      cmdRun(cmds._showPopup, 'tty');
    },

    /**
     * Private - TTY-only - Clear tty screen
     */
    _ttyClear: function () {
      let lastLine = state.tty.lines[state.tty.lines.length - 1];
      lastLine = lastLine.split('<wbr />')[0];
      state.tty.lines = [lastLine];

      state.tty._tmpCommand = hgetTtyInput().value;
    },

    /**
     * Private - TTY-only - Erase the current command
     */
    _ttyErase: function () {
      hgetTtyInput().value = '';
    },

    /**
     * Private - TTY-only - Run a command through the TTY
     * @param {string} command
     */
    _ttyExec: function (command) {
      state.tty.history.push(command);
      state.tty.historyCursor = state.tty.history.length;
      state.tty._tmpCommand = null;
      websocketSend({
        action: 'shell.command',
        args: { Command: command },
      });
    },

    /**
     * Public - Mobile-only - TTY-only - Quit the current TTY
     */
    ttyQuit: function () {
      state.tty.isEnabled = false;
      state.tty.lines = [];
      state.tty.history = [];
      state.tty.historyCursor = [];
      cmdRun(cmds._clearPopup);
    },

    /**
     * Private - TTY-only - Quit and close the TTY session
     */
    _ttyQuit: function () {
      state.tty.isEnabled = false;
      state.tty.lines = [];
      state.tty.history = [];
      state.tty.historyCursor = [];
      cmdRun(cmds._clearPopup);
    },

    /**
     * Private - TTY-only - Set a command from the history
     */
    _ttySetHistoryPrevious: function () {
      if (state.tty.history.length === 0) return;
      if (state.tty.historyCursor > -1) state.tty.historyCursor--;
    },

    /**
     * Private - TTY-only - Set a command from the history
     */
    _ttySetHistoryNext: function () {
      if (state.tty.history.length === 0) return;
      if (state.tty.historyCursor < state.tty.history.length)
        state.tty.historyCursor++;
    },

    /**
     * Private - Image-only - Pull an image
     * @param {object} args
     * @param {string} args.Image
     */
    _imagePull: function (args) {
      if (!args.Image || args.Image.length === 0) return;
      websocketSend({
        action: 'image.pull',
        args: { Image: args.Image },
      });
    },

    /**
     * Private - Image-only - Run an image
     * @param {object} args
     * @param {string} args.Name (container name)
     */
    _imageRun: function (args) {
      if (!args.Name || args.Name.length === 0) return;
      websocketSend({
        action: 'image.run',
        args: { Resource: sgetCurrentRow(), Name: args.Name },
      });
    },

    /**
     * Private - Container-only - Rename a container
     * @param {object} args
     * @param {string} args.Name (new container's name)
     */
    _containerRename: function (args) {
      if (!args.Name || args.Name.length === 0) return;
      websocketSend({
        action: 'container.rename',
        args: { Resource: sgetCurrentRow(), Name: args.Name },
      });
    },

    /**
     * Private - Get available inspector (sub) tabs for the current tab (containers, images, etc.)
     */
    _inspectorTabs: function () {
      const currentTabKey = sgetCurrentTabKey();
      websocketSend({ action: `${currentTabKey.slice(0, -1)}.inspect.tabs` });
    },

    /**
     * Private - Refresh the inspector data (request new data from the server)
     */
    _refreshInspector: function () {
      state.inspector.content = [];
      state.inspector.horizontalScroll = 0;
      state.inspector.verticalScroll = 0;

      const currentRow = sgetCurrentRow();
      const currentTabKey = sgetCurrentTabKey();
      const currentInspectorTab = state.inspector.currentTab;

      // Produces something like : <tab>.inspect.<sub-tab>
      // Such as : container.inspect.logs
      websocketSend({
        // prettier-ignore
        action: `${currentTabKey.slice(0,-1)}.inspect.${currentInspectorTab.toLowerCase()}`,
        args: { Resource: currentRow },
      });
    },

    /**
     * Private - Show the authentication popup
     */
    _showAuthentication: function () {
      cmdRun(cmds._showPrompt, {
        input: {
          isEnabled: true,
          name: 'Password',
          placeholder: 'Please fill in your server secret',
        },
        isForAuthentication: true,
        callback: cmds._authenticate,
      });
    },

    /**
     * Private - Send password to the server as a mean for authenticating
     */
    _authenticate: function () {
      const password = hgetPromptInput().value;

      if (!password) return;

      websocketSend({ action: 'auth.login', args: { Password: password } });
    },

    /**
     * Public - Quit the app / Quit the current popup
     * Requires prompt
     */
    quit: function () {
      if (!state.popup) {
        // prettier-ignore
        cmdRun(cmds.prompt, {
          text: 'Are you sure you want to quit?',
          callback: cmds._exit
        });
        return;
      }

      if (state.prompt.isEnabled) {
        cmdRun(cmds._clearPrompt);
        return;
      }

      if (state.message.isEnabled) {
        cmdRun(cmds._clearMessage);
        return;
      }

      cmdRun(cmds._clearPopup);
    },

    /**
     * Public - Show the general help
     */
    help: function () {
      state.helper = 'message';
      cmdRun(cmds._showPopup, 'help');
    },

    /**
     * Public - Show the menu associated with the current tab
     */
    menu: function () {
      const currentTab = state.inspector.isEnabled
        ? state.navigation.previousTab
        : state.navigation.currentTab;

      websocketSend({
        action: `${currentTab.slice(0, -1)}.menu`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public - Show the bulk menu associated with the current tab
     */
    bulk: function () {
      const currentTab = state.inspector.isEnabled
        ? state.navigation.previousTab
        : state.navigation.currentTab;

      websocketSend({ action: `${currentTab}.bulk` });

      cmdRun(cmds._showPopup, 'bulk');
      state.helper = 'menu';
    },

    /**
     * Public - Show a confirm prompt
     * @param {Prompt} args
     */
    prompt: function (args) {
      cmdRun(cmds._showPrompt, args);
    },

    /**
     * Public - Confirm the current context
     * When prompt      : confirm the requested action, and run the callback associated
     * When menu/popup  : run the action associated with the current row
     * When tab         : inspect the resource associated with the current row
     * When message     : close the popup
     */
    confirm: function () {
      // Prompt confirm (run callback)
      if (state.prompt.isEnabled) {
        if (!state.prompt.input.isEnabled)
          cmdRun(state.prompt.callback, ...state.prompt.callbackArgs);
        else
          cmdRun(state.prompt.callback, {
            [state.prompt.input.name]: hgetPromptInput().value,
          });

        cmdRun(cmds._clearPrompt);
        return;
      }

      // Message confirm (close popup)
      if (state.message.isEnabled) {
        cmdRun(cmds._clearMessage);
        return;
      }

      // Menu / Bulk confirm (run action)
      if (state.isMenuIng && state.navigation.currentMenuRow > 0) {
        // prettier-ignore
        const row = hgetPopupRow(state.popup, state.navigation.currentMenuRow);
        const attributes = row.dataset;

        if ('cancel' in attributes) {
          cmdRun(cmds._clearPopup);
          return;
        }

        if (attributes.prompt) {
          cmdRun(cmds._showPrompt, {
            text: attributes.prompt,
            callback: cmds._wsSend,
            callbackArgs: [
              {
                action: attributes.command,
                args: { Resource: sgetCurrentRow() },
              },
            ],
          });
          return;
        }

        if (!attributes.runLocally) {
          if (!attributes.useRow)
            cmdRun(cmds._wsSend, { action: attributes.command });
          else
            cmdRun(cmds._wsSend, {
              action: attributes.command,
              args: { Resource: sgetCurrentRow() },
            });

          // Can clear anytime as the command is private ("_wsSend")
          cmdRun(cmds._clearPopup);
        } else {
          // Must clear first to enable running a command out of menu context
          cmdRun(cmds._clearPopup);

          if (!attributes.useRow) cmdRun(cmds[attributes.command]);
          else cmdRun(cmds[attributes.command], sgetCurrentRow());
        }

        return;
      }

      // Help / Any other popup (close popup)
      if (state.popup) {
        cmdRun(cmds._clearPopup);
        return;
      }

      // Tab confirm (inspect)
      cmdRun(cmds._enterInspect);
    },

    /**
     * Public - Reject the current context
     * When prompt     : close the prompt, and ignore the callback associated
     * When menu/popup : close the popup
     * When tab        : do nothing
     * When message    : close the popup
     * When inspect    : exit inspect
     */
    reject: function () {
      if (!state.isAuthenticated) return;

      if (state.prompt.isEnabled) {
        cmdRun(cmds._clearPrompt);
        return;
      }

      if (state.message.isEnabled) {
        cmdRun(cmds._clearMessage);
        return;
      }

      if (state.popup) {
        cmdRun(cmds._clearPopup);
        return;
      }

      if (state.inspector.isEnabled) {
        cmdRun(cmds._exitInspect);
        return;
      }
    },

    /**
     * Public - Navigate to the previous tab
     */
    previousTab: function () {
      if (state.inspector.isEnabled) cmdRun(cmds._exitInspect);

      const currentIndex = state.tabs.findIndex(
        (t) => t.Key === state.navigation.currentTab
      );
      let prevIndex = currentIndex - 1;

      if (prevIndex === -1) prevIndex = state.tabs.length - 1;

      state.navigation.currentTab = state.tabs[prevIndex].Key;

      cmdRun(cmds._inspectorTabs);
    },

    /**
     * Public - Navigate to the next tab
     */
    nextTab: function () {
      if (state.inspector.isEnabled) cmdRun(cmds._exitInspect);

      const currentIndex = state.tabs.findIndex(
        (t) => t.Key === state.navigation.currentTab
      );
      let nextIndex = currentIndex + 1;

      if (nextIndex > state.tabs.length - 1) nextIndex = 0;

      state.navigation.currentTab = state.tabs[nextIndex].Key;

      cmdRun(cmds._inspectorTabs);
    },

    /**
     * Public - Navigate to the previous inspector tab
     */
    previousSubTab: function () {
      const currentIndex = state.inspector.availableTabs.indexOf(
        state.inspector.currentTab
      );
      let prevIndex = currentIndex - 1;

      if (prevIndex == -1) prevIndex = state.inspector.availableTabs.length - 1;

      state.inspector.currentTab =
        state.inspector.availableTabs[
          prevIndex % state.inspector.availableTabs.length
        ];

      cmdRun(cmds._refreshInspector);
    },

    /**
     * Public - Navigate to the next inspector tab
     */
    nextSubTab: function () {
      const currentIndex = state.inspector.availableTabs.indexOf(
        state.inspector.currentTab
      );
      const nextIndex = currentIndex + 1;
      state.inspector.currentTab =
        state.inspector.availableTabs[
          nextIndex % state.inspector.availableTabs.length
        ];

      cmdRun(cmds._refreshInspector);
    },

    /**
     * Public - Activate the next layout for the app
     */
    nextLayout: function () {
      const currentIndex = state.appearance.availableLayouts.indexOf(
        state.appearance.currentLayout
      );

      if (currentIndex === state.appearance.availableLayouts.length - 1)
        state.appearance.currentLayout = state.appearance.availableLayouts[0];
      else
        state.appearance.currentLayout =
          state.appearance.availableLayouts[currentIndex + 1];
    },

    /**
     * Public - Activate the previous layout for the app
     */
    previousLayout: function () {
      const currentIndex = state.appearance.availableLayouts.indexOf(
        state.appearance.currentLayout
      );

      if (currentIndex === 0)
        state.appearance.currentLayout =
          state.appearance.availableLayouts[
            state.appearance.availableLayouts.length - 1
          ];
      else
        state.appearance.currentLayout =
          state.appearance.availableLayouts[currentIndex - 1];
    },

    /**
     * Public - Scroll left / Navigate to previous tab
     */
    scrollLeft: function () {
      if (state.inspector.isEnabled) {
        if (state.inspector.horizontalScroll > 0)
          state.inspector.horizontalScroll -= 20;
        return;
      }

      cmdRun(cmds.previousTab);
    },

    /**
     * Public - Scroll right / Navigate to next tab
     */
    scrollRight: function () {
      if (state.inspector.isEnabled) {
        const _inspector = hgetTab('inspector');
        const _content = _inspector.querySelector('.tab-content');
        const maxScroll = _content.scrollWidth - _content.clientWidth;
        if (state.inspector.horizontalScroll < maxScroll)
          state.inspector.horizontalScroll += 20;
        return;
      }

      cmdRun(cmds.nextTab);
    },

    /**
     * Public - Scroll down / Navigate to next row
     */
    scrollDown: function () {
      // Menu - Next row
      if (state.isMenuIng) {
        const availableRows = state.menu.actions;
        state.navigation.currentMenuRow += 1;

        // +1 is added to account for the extra "cancel" option
        if (state.navigation.currentMenuRow > availableRows.length + 1)
          state.navigation.currentMenuRow = 1;

        return;
      }

      // Tab - Next row
      if (state.navigation.currentTab) {
        const availableRows = state.tabs.find(
          (t) => t.Key === state.navigation.currentTab
        ).Rows;
        state.navigation.currentTabsRows[state.navigation.currentTab] += 1;

        if (
          state.navigation.currentTabsRows[state.navigation.currentTab] >
          availableRows.length
        )
          state.navigation.currentTabsRows[state.navigation.currentTab] = 1;

        cmdRun(cmds._refreshInspector);
        return;
      }

      // Inspector - Scroll down
      if (state.inspector.isEnabled) {
        const _inspector = hgetTab('inspector');
        const _content = _inspector.querySelector('.tab-content');
        const maxScroll = _content.scrollHeight - _content.clientHeight;
        if (state.inspector.verticalScroll < maxScroll)
          state.inspector.verticalScroll += 20;
      }
    },

    /**
     * Public - Scroll up / Navigate to previous row
     */
    scrollUp: function () {
      // Menu - Previous row
      if (state.isMenuIng) {
        const availableRows = state.menu.actions;
        state.navigation.currentMenuRow -= 1;

        // +1 is added to account for the extra "cancel" option
        if (state.navigation.currentMenuRow < 1)
          state.navigation.currentMenuRow = availableRows.length + 1;

        return;
      }

      // Tab - Previous row
      if (state.navigation.currentTab) {
        const availableRows = state.tabs.find(
          (t) => t.Key === state.navigation.currentTab
        ).Rows;
        state.navigation.currentTabsRows[state.navigation.currentTab] -= 1;

        if (state.navigation.currentTabsRows[state.navigation.currentTab] < 1)
          state.navigation.currentTabsRows[state.navigation.currentTab] =
            availableRows.length;

        cmdRun(cmds._refreshInspector);
        return;
      }

      // Inspector - Scroll up
      if (state.inspector.isEnabled) {
        if (state.inspector.verticalScroll > 0)
          state.inspector.verticalScroll -= 20;
      }
    },

    /**
     * Public - Display the remove menu of the highlighted resource
     */
    remove: function () {
      websocketSend({
        action: `${sgetCurrentTabKey().slice(0, -1)}.menu.remove`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public - Container-only - Pause/Unpause
     */
    pause: function () {
      if (sgetCurrentTabKey() !== 'containers') return;
      websocketSend({
        action: `container.pause`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public - Container-only - Stop
     */
    stop: function () {
      if (sgetCurrentTabKey() !== 'containers') return;

      websocketSend({
        action: `container.stop`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public
     * - Container-only - Restart
     * - Image-only - Run
     */
    run_restart: function () {
      const currentTabKey = sgetCurrentTabKey();

      if (currentTabKey === 'containers')
        websocketSend({
          action: `container.restart`,
          args: { Resource: sgetCurrentRow() },
        });
      else if (currentTabKey === 'images')
        cmdRun(cmds.prompt, {
          input: {
            isEnabled: true,
            name: 'Name',
            placeholder: 'Please fill in a name for the new container',
          },
          callback: cmds._imageRun,
        });
    },

    /**
     * Public - Container-only - Rename
     */
    rename: function () {
      if (sgetCurrentTabKey() !== 'containers') return;

      cmdRun(cmds.prompt, {
        input: {
          isEnabled: true,
          name: 'Name',
          placeholder: 'Please fill in a new name for the container',
        },
        callback: cmds._containerRename,
      });
    },

    /**
     * Public - Container-only - Exec shell
     */
    shellContainer: function () {
      if (sgetCurrentTabKey() !== 'containers') return;
      websocketSend({
        action: `container.shell`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public - System-only - Exec shell
     */
    shellSystem: function () {
      websocketSend({ action: `shell` });
    },

    /**
     * Public - Container-only - Open in browser
     */
    browser: function () {
      if (sgetCurrentTabKey() !== 'containers') return;
      websocketSend({
        action: `container.browser`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public - Image-only - Open in Docker Hub
     */
    hub: function () {
      if (sgetCurrentTabKey() !== 'images') return;
      window.open(`https://hub.docker.com/r/${sgetCurrentRow().Name}`);
    },

    /**
     * Public - Image-only - Pull
     */
    pull: function () {
      if (sgetCurrentTabKey() !== 'images') return;

      cmdRun(cmds.prompt, {
        input: {
          isEnabled: true,
          name: 'Image',
          placeholder: '[repository/]name[:tag]',
        },
        callback: cmds._imagePull,
      });
    },

    /**
     * Public - Volume-only - Browse (Open a terminal on the server, and cd to the volume's mountpoint)
     */
    browse: function () {
      if (sgetCurrentTabKey() !== 'volumes') return;
      websocketSend({
        action: `volume.browse`,
        args: { Resource: sgetCurrentRow() },
      });
    },

    /**
     * Public - Reload the current inspector tab / Reload everything
     */
    reload: function () {
      if (state.inspector.isEnabled) cmdRun(cmds._refreshInspector);
      else cmdRun(cmds._init);
    },

    /**
     * Public - Open Isaiah repository on Github
     */
    github: function () {
      window.open(`https://github.com/will-moss/isaiah/?from=instance`);
    },

    /**
     * Public - Navigate to the nth tab
     */
    firstTab: function () {
      if (!state.tabs[0]) return;
      state.navigation.currentTab = state.tabs[0].Key;
    },
    secondTab: function () {
      if (!state.tabs[1]) return;
      state.navigation.currentTab = state.tabs[1].Key;
    },
    thirdTab: function () {
      if (!state.tabs[2]) return;
      state.navigation.currentTab = state.tabs[2].Key;
    },
    fourthTab: function () {
      if (!state.tabs[3]) return;
      state.navigation.currentTab = state.tabs[3].Key;
    },
  };

  // === Variables

  /**
   * @type {string}
   */
  const wsHost = window.location.hostname;

  /**
   * @type {number}
   */
  const wsPort = window.location.port;

  /**
   * @type {boolean}
   */
  const wsSSL = window.location.protocol === 'https:';

  /**
   * @type {number} - Milliseconds
   */
  const wsRetryInterval = 1000;

  /**
   * @type {WebSocket}
   */
  let wsSocket = null;

  /**
   * @type {object.<string, string>}
   */
  const kbMap = {
    // Navigation
    ArrowUp: 'scrollUp',
    ArrowDown: 'scrollDown',
    ArrowLeft: 'scrollLeft',
    ArrowRight: 'scrollRight',

    Tab: 'nextTab',
    ShiftTab: 'previousTab',

    ']': 'nextSubTab',
    '[': 'previousSubTab',

    1: 'firstTab',
    2: 'secondTab',
    3: 'thirdTab',
    4: 'fourthTab',

    // Interaction
    Enter: 'confirm',
    y: 'confirm',

    Escape: 'reject',
    n: 'reject',

    // Menu
    x: 'menu',
    b: 'bulk',

    // Sub commands
    q: 'quit',
    d: 'remove',
    p: 'pause',
    s: 'stop',
    r: 'run_restart',
    m: 'rename',
    E: 'shellContainer',
    S: 'shellSystem',
    R: 'reload',
    P: 'pull',
    B: 'browse',
    w: 'browser',
    h: 'hub',
    G: 'github',
    '?': 'help',

    // Misc
    '+': 'nextLayout', // Next layout
    '-': 'previousLayout', // Previous layout
  };

  /**
   * @type {string}
   */
  const whitespace = '\xa0';

  // === Listeners

  /**
   * Called every time the user presses a key down
   * on their keyboard. Will call the internal command
   * associated with the key pressed, or take the appropriate
   * behavior if tty is enabled
   */
  const listenerKeyDown = (evt) => {
    if (state.tty.isEnabled) {
      listenerTtyKeyDown(evt);
      return;
    }

    if (state.prompt.input.isEnabled) {
      listenerPromptInputKeyDown(evt);
      return;
    }

    if (evt.metaKey) return;

    let { key } = evt;

    if (!key || !(key in kbMap)) return;

    if (evt.shiftKey && `Shift${key}` in kbMap) key = `Shift${key}`;

    evt.stopPropagation();
    evt.preventDefault();

    cmdRun(cmds[kbMap[key]]);
  };

  /**
   * Called every time the user presses a key down
   * on their keyboard, while in TTY mode. Will
   * appropriately perform TTY actions and update  the state
   */
  const listenerTtyKeyDown = (evt) => {
    let { key } = evt;

    switch (key) {
      // Run command
      case 'Enter':
        cmdRun(cmds._ttyExec, hgetTtyInput().value);
        break;

      // Clear screen
      case 'l':
        if (evt.ctrlKey) cmdRun(cmds._ttyClear);
        break;

      // Erase command
      case 'c':
        if (evt.ctrlKey) cmdRun(cmds._ttyErase);
        break;

      // Quit
      case 'd':
        if (evt.ctrlKey) cmdRun(cmds._ttyExec, 'exit');
        break;

      // Previous command
      case 'ArrowUp':
        cmdRun(cmds._ttySetHistoryPrevious);
        break;

      case 'ArrowDown':
        cmdRun(cmds._ttySetHistoryNext);
        break;

      case 'Tab':
        evt.stopPropagation();
        evt.preventDefault();
        hgetTtyInput().focus();
        break;
    }
  };

  /**
   * Called every time the user presses a key down
   * on their keyboard, while in Prompt Input  mode. Will
   * allow the user to input data using any key, while
   * Escape and Enter will be used to Leave/Confirm the
   * Prompt
   */
  const listenerPromptInputKeyDown = (evt) => {
    if (evt.metaKey) return;

    const { key } = evt;

    if (!key || !(key in kbMap)) return;

    if (evt.shiftKey && `Shift${key}` in kbMap) key = `Shift${key}`;

    // Only allow Escape and Enter
    if (!['Escape', 'Enter'].includes(key)) return;

    evt.stopPropagation();
    evt.preventDefault();

    cmdRun(cmds[kbMap[key]]);
  };

  /**
   * Called every time the user clicks with their
   * mouse. Will call the internal command
   * associated with the DOM element that's been
   * targeted
   */
  const listenerMouseClick = (evt) => {
    evt.preventDefault();

    // Prevent doing anything while the server is loading
    if (state.isLoading) return;

    // Prevent doing anything while the connection is lost
    if (!state.isConnected) return;

    const { target } = evt;

    // 1. Explicit navigation via data-navigate attribute (e.g. data-navigate="tab.containers"/"inspector.Logs"/"row")
    if (target.hasAttribute('data-navigate')) {
      const [part, key] = target.getAttribute('data-navigate').split('.');

      // 1.1. Tab Header
      if (part === 'tab') {
        if (state.inspector.isEnabled) cmdRun(cmds._exitInspect);
        state.navigation.currentTab = key;
        cmdRun(cmds._inspectorTabs);
      }
      // 1.2. Inspector Tab Header
      else if (part === 'inspector') {
        if (!state.inspector.isEnabled) cmdRun(cmds._enterInspect);
        state.inspector.currentTab = key;
        cmdRun(cmds._refreshInspector);
      }
      // 1.3. Tab Row
      else if (part === 'row') {
        const tab = target.parentNode.parentNode;

        if (tab.classList.contains('for-inspector')) return;

        const tabHeader = tab.querySelector('.tab-title[data-navigate]');
        const tabContent = tab.querySelector('.tab-content');

        // 1.3.1. Navigate to the clicked row's parent tab
        // prettier-ignore
        const [_part, _key] = tabHeader.getAttribute('data-navigate').split('.');
        if (_key !== state.navigation.currentTab) {
          if (state.inspector.isEnabled) cmdRun(cmds._exitInspect);
          state.navigation.currentTab = _key;
        }

        // 1.3.2. Focus the clicked row and refresh the inspector
        const rowIndex = Array.from(tabContent.children).indexOf(target);
        state.navigation.currentTabsRows[_key] = rowIndex + 1;

        cmdRun(cmds._inspectorTabs);
      }
      // 1.4. Tab Row's Cell
      else if (part === 'cell') {
        const tab = target.parentNode.parentNode.parentNode;

        // 1.4.1. Clicked in the inspector
        if (tab.classList.contains('for-inspector')) {
          // 1.4.1.1. Focus the inspector
          if (!state.inspector.isEnabled) cmdRun(cmds._enterInspect);
        }
        // 1.4.2. Clicked in a regular tab
        else {
          const tabHeader = tab.querySelector('.tab-title[data-navigate]');
          const tabContent = tab.querySelector('.tab-content');
          const tabRow = target.parentNode;

          // 1.4.2.1. Navigate to the clicked cell's row's parent tab
          // prettier-ignore
          const [_part, _key] = tabHeader.getAttribute('data-navigate').split('.');
          if (_key !== state.navigation.currentTab) {
            if (state.inspector.isEnabled) cmdRun(cmds._exitInspect);
            state.navigation.currentTab = _key;
          }

          // 1.4.2.2. Focus the clicked row and refresh the inspector
          const rowIndex = Array.from(tabContent.children).indexOf(tabRow);
          state.navigation.currentTabsRows[_key] = rowIndex + 1;

          cmdRun(cmds._inspectorTabs);
        }
      }
    }

    // 2. Explicit action via data-action attribute (e.g. data-action="help")
    else if (target.hasAttribute('data-action')) {
      const action = target.getAttribute('data-action');
      cmdRun(cmds[action]);
    }

    // 3. Explicit command (local or remote) via data-command attribute (menu actions only)
    else if (target.hasAttribute('data-command')) {
      // 3.1. Focus the clicked row and trigger it
      const tabContent = target.parentNode;
      const rowIndex = Array.from(tabContent.children).indexOf(target) + 1;

      state.navigation.currentMenuRow = rowIndex;
      cmdRun(cmds.confirm);
    }

    // 4. Clicked anywhere else
    else {
      // 4.0. If tty-ing, focus the tty input
      if (state.tty.isEnabled) {
        hgetTtyInput().focus();
        return;
      }

      // 4.1. If any popup is activated, while not menuing and not tty-ing, dismiss it
      if (!state.isMenuIng && state.popup) cmdRun(cmds.reject);
      // 4.2. If menuing, and clicked outside the menu, dismiss it
      else if (state.isMenuIng && target.classList.contains('popup-layer'))
        cmdRun(cmds.reject);
      // 4.3. If menuing, and clicked inside the menu, run the 3. scenario (assumption: we clicked a span inside a row)
      else {
        const tabContent = target.parentNode.parentNode;
        const tabRow = target.parentNode;
        const rowIndex = Array.from(tabContent.children).indexOf(tabRow) + 1;

        state.navigation.currentMenuRow = rowIndex;
        cmdRun(cmds.confirm);
      }
    }

    cmdRun(cmds._render);
  };

  /**
   * Called when the browser has succesfully established a
   * Websocket connection with the server
   */
  const listenerSocketOpen = (evt) => {
    state.isConnected = true;
    state.hasEstablishedConnection = true;
    cmdRun(cmds._showAuthentication);
  };

  /**
   * Called when the Websocket connection was closed
   * or encountered an error while connecting to the server
   */
  const listenerSocketError = (evt) => {
    if (wsSocket) wsSocket.close();
  };

  /**
   * Called when the browser receives a message from the server
   * through the Websocket connection established prior
   */
  const listenerSocketMessage = (evt) => {
    const { data } = evt;

    /**
     * @typedef Notification
     * @property {"init"|"refresh"|"loading"|"report"|"prompt"|"tty"|"auth"} Category
     * @property {string} Type
     * @property {string} Title
     * @property {object} Content
     * @property {string} Follow
     * @property {boolean} Display
     */

    /**
     * @type {Notification}
     */
    const notification = JSON.parse(data);

    switch (notification.Category) {
      case 'init':
        state.tabs = notification.Content.Tabs;
        state.navigation.currentTab = state.tabs[0].Key;
        state.navigation.currentTabsRows = state.tabs.reduce(
          (a, b) => ({ ...a, [b.Key]: 1 }),
          {}
        );
        state.isLoading = false;
        cmdRun(cmds._inspectorTabs);
        break;

      case 'auth':
        if ('Authentication' in notification.Content) {
          state.message.category = 'authentication';
          state.message.type = notification.Type;
          state.message.title = notification.Title;
          state.message.content = notification.Content.Authentication.Message;
          state.message.isEnabled = true;
          state.helper = 'message';

          // Authentication error
          if ('error' === notification.Type) {
            cmdRun(cmds._showPopup, 'message');
            setTimeout(() => {
              cmdRun(cmds._clearMessage);
              cmdRun(cmds._showAuthentication);
            }, state._delays.forAuthentication);
          }
          // Authentication success
          else if ('success' === notification.Type) {
            // Normal case
            if (!notification.Content.Authentication.Spontaneous) {
              cmdRun(cmds._showPopup, 'message');
              setTimeout(() => {
                cmdRun(cmds._clearMessage);
                state.isAuthenticated = true;
                cmdRun(cmds._init);
              }, state._delays.forAuthentication);
            }
            // Dev-only
            else {
              cmdRun(cmds._clearMessage);
              cmdRun(cmds._clearPrompt);
              state.isAuthenticated = true;
              cmdRun(cmds._init);
            }
          }
        }
        break;

      case 'refresh':
        if ('Tab' in notification.Content)
          if (notification.Content.Tab.Rows.length > 0) {
            state.tabs = state.tabs.map((t) =>
              t.Key === notification.Content.Tab.Key
                ? notification.Content.Tab
                : t
            );
            state.navigation.currentTabsRows[notification.Content.Tab.Key] = 1;
          } else {
            state.tabs = state.tabs.filter(
              (t) => t.Key !== notification.Content.Tab.Key
            );
            state.navigation.currentTab = state.tabs[0].Key;
            state.navigation.previousTab = state.tabs[0].Key;
            state.navigation.currentTabsRows[state.navigation.currentTab] = 1;
          }

        if ('Actions' in notification.Content) {
          state.menu.actions = notification.Content.Actions;
          state.navigation.currentMenuRow = 1;

          cmdRun(cmds._showPopup, 'menu');
          state.helper = 'menu';
        }

        if ('Address' in notification.Content) {
          window.open(notification.Content.Address, '_blank');
        }

        if ('Inspector' in notification.Content) {
          if ('Tabs' in notification.Content.Inspector) {
            state.inspector.availableTabs = notification.Content.Inspector.Tabs;
            state.inspector.currentTab = notification.Content.Inspector.Tabs[0];
            cmdRun(cmds._refreshInspector);
          }
          if ('Content' in notification.Content.Inspector) {
            // When raw lines are received, append them
            if (notification.Content.Inspector.Content[0].Type === 'lines')
              state.inspector.content.push(
                ...notification.Content.Inspector.Content
              );
            // Else, replace the current content with the one received
            else
              state.inspector.content = notification.Content.Inspector.Content;
          }
        }

        state.isLoading = false;
        break;

      case 'loading':
        state.isLoading = true;
        break;

      case 'report':
        state.message.category = notification.Category;
        state.message.type = notification.Type;
        state.message.title = notification.Title;
        state.message.content = notification.Content.Message;
        state.isLoading = false;

        if (notification.Display) {
          state.message.isEnabled = true;
          state.helper = 'message';
          cmdRun(cmds._showPopup, 'message');
        }

        break;

      case 'prompt':
        cmdRun(cmds._showPrompt, {
          text: notification.Content.Message,
          callback: cmds._wsSend,
          callbackArgs: [
            {
              action: notification.Content.Command,
              args: { Resource: sgetCurrentRow() },
            },
          ],
        });
        state.isLoading = false;
        break;

      case 'tty':
        if ('Status' in notification.Content) {
          switch (notification.Content.Status) {
            case 'started':
              state.tty.type = notification.Content.Type;
              cmdRun(cmds._ttyStart);
              break;

            case 'exited':
              cmdRun(cmds._ttyQuit);
              break;
          }
        }

        if ('Output' in notification.Content) {
          const { Output } = notification.Content;
          state.tty._buffer += Output;

          // Fill a buffer with stdout content until we meet a newline character
          if (!['\r', '\n'].some((n) => Output.includes(n))) {
            // If no new content was received after X ms, consider it's the end of output, and print it
            setTimeout(() => {
              if (state.tty._buffer.length > 0) {
                state.tty.lines.push(state.tty._buffer);
                state.tty._buffer = '';
                cmdRun(cmds._render);
              }
            }, state._delays.forTTYBufferClear);
            break;
          }

          // If the command was run by us, apply a string transform to print it on the same line as last stdout
          if (state.tty._buffer.trim().endsWith('#ISAIAH')) {
            const command = state.tty._buffer.split('#ISAIAH')[0];
            state.tty.lines[state.tty.lines.length - 1] += `<wbr />${command}`;
            state.tty._buffer = '';
            break;
          }

          // Regular stdout lines
          state.tty.lines.push(
            ...(state.tty._buffer.includes('\r')
              ? state.tty._buffer.split('\r\n').filter((l) => l)
              : state.tty._buffer.split('\n').filter((l) => l))
          );
          state.tty._buffer = '';
        }

        state.isLoading = false;
    }

    if (notification.Follow)
      // Delay added to prevent caching / refreshing data issues with the Docker client
      setTimeout(() => {
        websocketSend({ action: notification.Follow });
      }, state._delays.default);

    renderApp(state);
  };

  /**
   * Called when the Websocket connection between the browser
   * and the server is closed. Attempts to reconnect and establish
   * the Websocket connection again
   */
  const listenerSocketClose = (evt) => {
    state.isConnected = false;
    renderApp(state);

    if (!wsSocket || wsSocket.readyState === WebSocket.CLOSED)
      setTimeout(websocketConnect, wsRetryInterval);
  };

  // ===
  // === Entry Point
  // ===
  window.addEventListener('load', () => {
    // 1. Connect to server (first execution loop)
    websocketConnect();

    // 2. Set keyboard listener (second execution loop)
    window.addEventListener('keydown', listenerKeyDown);

    // 3. Set mouse listener (third execution loop)
    window.addEventListener('click', listenerMouseClick);
  });
})(window);
