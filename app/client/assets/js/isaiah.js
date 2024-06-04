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
   * Retrieve a value from localStorage, with support for booleans and default value
   * @param {string} key
   * @param {*} d
   * @returns {*}
   */
  const lsGet = (key, d) => {
    const v = localStorage.getItem(key);
    if (v === null) return d;
    if (v === 'true') return true;
    if (v === 'false') return false;
    return v;
  };

  /**
   * Determine the general type of a variable
   * @param {*} v
   * @returns {'string'|'numeric'|'date'}
   */
  const getGeneralType = (v) => {
    if (typeof v !== 'string') v = v.toString();

    const isNumeric = !isNaN(v) && !isNaN(parseFloat(v));

    return isNumeric ? 'numeric' : 'string';
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

  /**
   * Turn a given method into a debounced version of it
   * @param {Function} fn
   * @param {int} delay
   */
  const debounce = (fn, delay) => {
    let timeoutId = null;
    return (...args) => {
      window.clearTimeout(timeoutId);
      timeoutId = window.setTimeout(() => {
        fn(...args);
      }, delay);
    };
  };

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
  const hgetJumpInput = () => q('#jump-input');

  /**
   * @returns {HTMLElement}
   */
  const hgetJumpResults = () => q('.jump-results');

  /**
   * @returns {HTMLElement}
   */
  const hgetSearchControl = () => q('.search-control');

  /**
   * @returns {HTMLElement}
   */
  const hgetSearchInput = () => q('#search-input');

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
   * @param {boolean} action.RequiresMenuAction
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
    if (action.RequiresMenuAction) html += ` data-use-menu-action="true"`;
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
   * @param {string} tab.SortBy
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

    if (inspector.currentTab === 'Logs') {
      html += `<div class="tab-title-group for-controls">`;
      html += `<button class="tab-sub-title" data-action="inspectorCopyLogs">Copy</button>`;
      html += `</div>`;
    }

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
   * @param {Menu} menu
   * @param {Row} row
   * @returns {string}
   */
  const renderMenu = (menu, row = null) => {
    // Cell padding according to the longest key
    let maxKeyWidth = -1;

    if (menu.actions.length > 0 && menu.actions[0].Key)
      for (const action of menu.actions)
        maxKeyWidth = Math.max(maxKeyWidth, action.Key.length);
    else maxKeyWidth = 0;

    let title = {
      menu: 'Menu',
      bulk: 'Bulk actions',
      theme: 'Theme',
      agent: 'Agent',
      host: 'Host',
      parameters: 'Parameters',
    }[menu.key];
    if (menu.key === 'menu' && row) title += ` (${row.Name})`;

    return `
      <div class="popup for-menu">
        <div class="tab is-active">
          <span class="tab-title">
            ${title}
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
             <div class="row is-not-interactive">
               <span class="cell">< >      </span>
               <span class="cell">switch node</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">l k      </span>
               <span class="cell">switch host</span>
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
               <span class="cell">O        </span>
               <span class="cell">show overview</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">J        </span>
               <span class="cell">jump to any resource</span>
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
             <div class="row is-not-interactive">
               <span class="cell">V        </span>
               <span class="cell">show version</span>
             </div>
             <div class="row is-not-interactive"></div>
             <div class="row is-not-interactive">
               <span class="cell">T        </span>
               <span class="cell">open theme picker</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">A        </span>
               <span class="cell">open agent picker</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">H        </span>
               <span class="cell">open host picker</span>
             </div>
             <div class="row is-not-interactive">
               <span class="cell">X        </span>
               <span class="cell">open parameters manager</span>
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
   * @param {Overview} overview
   * @returns {string}
   */
  const renderOverview = (overview) => `
      <div class="popup for-overview">
        <div class="tab is-active">
          <span class="tab-title">
            Overview
          </span>
          <div class="tab-content">
          ${overview.Instances.map(
            (i) => `
             <div
               data-role="${i.Server.Role}"
               data-name="${i.Server.Name}"
               class="row"
             >
               <div class="row-logo">
                 <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="M21.75 17.25v-.228a4.5 4.5 0 0 0-.12-1.03l-2.268-9.64a3.375 3.375 0 0 0-3.285-2.602H7.923a3.375 3.375 0 0 0-3.285 2.602l-2.268 9.64a4.5 4.5 0 0 0-.12 1.03v.228m19.5 0a3 3 0 0 1-3 3H5.25a3 3 0 0 1-3-3m19.5 0a3 3 0 0 0-3-3H5.25a3 3 0 0 0-3 3m16.5 0h.008v.008h-.008v-.008Zm-3 0h.008v.008h-.008v-.008Z" /> </svg>
               </div>
               <div class="row-content">
                 <div class="row-summary">
                   <p>
                     ${i.Server.Role === 'Agent' ? ' <i>Agent</i>' : ''}
                     ${i.Server.Name} (Docker ${i.Docker.Version})
                     <em>
                       ${
                         (i.Docker.Host.includes('://')
                           ? i.Docker.Host.split('://')[1]
                           : i.Docker.Host
                         ).split(':')[0]
                       }
                     </em>
                   </p>
                 </div>
                 <div class="row-information">
                   <div class="row-information-box for-containers">
                     <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /> </svg>
                     <span>${i.Resources.Containers.Count} Containers</span>
                   </div>
                   <div class="row-information-box for-volumes">
                   <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" /> </svg>
                   <span>${i.Resources.Volumes.Count} Volumes</span>
                   </div>
                   <div class="row-information-box for-images">
                     <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /> </svg>
                     <span>${i.Resources.Images.Count} Images</span>
                   </div>
                   <div class="row-information-box for-networks">
                     <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="M8.288 15.038a5.25 5.25 0 0 1 7.424 0M5.106 11.856c3.807-3.808 9.98-3.808 13.788 0M1.924 8.674c5.565-5.565 14.587-5.565 20.152 0M12.53 18.22l-.53.53-.53-.53a.75.75 0 0 1 1.06 0Z" /> </svg>
                     <span>${i.Resources.Networks.Count} Networks</span>
                   </div>

                   ${
                     i.Server.Host.startsWith('unix://') ||
                     (i.Server.Host === '' &&
                       i.Docker.Host.startsWith('unix://'))
                       ? `
                         <div class="row-information-specs">
                           <div class="row-information-box">
                             <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 0 0 2.25-2.25V6.75a2.25 2.25 0 0 0-2.25-2.25H6.75A2.25 2.25 0 0 0 4.5 6.75v10.5a2.25 2.25 0 0 0 2.25 2.25Zm.75-12h9v9h-9v-9Z" /> </svg>
                             <span>${i.Server.CountCPU} CPU</span>
                           </div>
                           <div class="row-information-box">
                             <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" d="m3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75Z" /> </svg>
                             <span>
                               ${(
                                 i.Server.AmountRAM /
                                 (1024 * 1024 * 1024)
                               ).toFixed(0)} GB RAM
                             </span>
                           </div>
                         </div>
                        `
                       : '<span class="row-filler">Remote host</span>'
                   }
                 </div>
               </div>
             </div>
          `
          ).join('')}
          </div>
        </div>
      </div>
  `;

  /**
   * @param {Jump} jump
   * @returns {string}
   */
  const renderJump = (jump) => `
      <div class="popup for-jump">
        <div class="tab is-active">
          <span class="tab-title">
            Jump
          </span>
          <div class="tab-content">
            <div class="jump-input-wrapper">
              <input
                id="jump-input"
                type="text"
                placeholder="Type the name of a resource"
                value="${jump.search ? jump.search : ''}"
              />
            </div>
            <div class="jump-results">
              ${renderJumpResults(jump)}
            </div>
          </div>
        </div>
      </div>
  `;

  /**
   * @param {Jump} jump
   * @returns {string}
   */
  const renderJumpResults = (jump) => `
      ${
        !jump.search
          ? `<p><i>Start typing, and results will appear</i></p>`
          : jump.results.length === 0
          ? `<p class="no-result-message"><i>No resource found</i></p>`
          : jump.results
              .map(
                (r) =>
                  `
                  <div
                    class="jump-result row"
                    data-jump="${r.Host ? `${r.Host}.` : ''}${r.ParentKey}.${
                    r.ID || r.Name
                  }"
                  >
                    ${r.Host ? `<span class="for-host">(${r.Host})</span>` : ''}
                    <span class="for-tab">${r.Parent}</span>
                    &gt;
                    <span class="for-resource">${r.Name}</span>
                  </div>
                `
              )
              .join('')
      }
  `;

  /**
   * Main rendering function, responsible for updating the DOM
   * from scratch, using the supplied _state argument
   *
   * @param {state} _state
   */
  const renderApp = (_state) => {
    if (_state.shouldStopRendering) return;

    let html;

    // -1. Set app's theme
    document.body.setAttribute('data-theme', _state.appearance.currentTheme);

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
    // 1.3.1. Re-use the popup when in Jump mode, as opposed to erasing it
    if (!_state.jump.isEnabled) hgetPopupContainer().innerHTML = '';

    // 1.4. Hide popup layer
    // 1.4.1. Re-use the popup when in Jump mode, as opposed to erasing it
    if (!_state.jump.isEnabled)
      hgetPopupContainer().classList.remove('is-active');

    // 1.5. Hide helper
    if (e('.help.is-active'))
      q('.help.is-active').classList.remove('is-active');

    // 1.6. Hide connection indicators
    hgetConnectionIndicators().forEach((i) => i.classList.remove('is-active'));

    // 1.7. Hide search control
    hgetSearchControl().classList.remove('is-active');

    // 2. Build every tab
    if (!_state.isFullyEmpty) {
      html = _state.tabs.map(renderTab).join('');
      hgetScreen('dashboard').querySelector('.left').innerHTML = html;

      // 3. Build inspector
      html = renderInspector(_state.inspector);
      hgetScreen('dashboard').querySelector('.right').innerHTML = html;
    }

    // 4. Build current popup
    if (state.popup) {
      hgetPopupContainer().classList.add('is-active');

      html = '';

      // 4.1. Popup - Prompt
      if (_state.prompt.isEnabled) html = renderPrompt(_state.prompt);
      // 4.2. Popup - Message
      else if (_state.message.isEnabled) html = renderMessage(_state.message);
      // 4.3. Popup - Menu / Theme / Agent / Host / Parameters
      else if (_state.menu.actions.length > 0) {
        if (sgetCurrentTab().Rows)
          html = renderMenu(_state.menu, sgetCurrentRow());
        else html = renderMenu(_state.menu);
      }
      // 4.4. Popup - Tty
      else if (_state.tty.isEnabled) html = renderTty(_state.tty);
      // 4.5. Popup - Help
      else if (_state.popup === 'help') html = renderHelp();
      // 4.6. Popup - Overview
      else if (_state.overview.isEnabled)
        html = renderOverview(_state.overview);
      // 4.7. Popup - Jump
      else if (_state.jump.isEnabled) {
        // 4.7.1. Perform partial re-render on subsequent render, to avoid regenerating the input while typing
        if (_state.jump.search) {
          html = renderJumpResults(_state.jump);
          hgetJumpResults().innerHTML = html;
        }
        // 4.7.2. Perform full re-render on first opening ( = no search performed yet )
        else {
          html = renderJump(_state.jump);
          hgetPopupContainer().innerHTML = html;
        }
      }

      if (!_state.jump.isEnabled) hgetPopupContainer().innerHTML = html;
    }

    // 5. Show search control
    if (!_state.isFullyEmpty) {
      if (_state.search.isEnabled)
        hgetSearchControl().classList.add('is-active');
    }

    // 6. Set focus on DOM elements

    // 6.1. Set focus on inspector
    if (!_state.isFullyEmpty) {
      if (
        _state.inspector.isEnabled &&
        (!_state.search.isEnabled || _state.search.isPending)
      )
        hgetTab('inspector').classList.add('is-active');

      // 6.1.1. Scroll horizontally on the inspector
      hgetTab('inspector').querySelector('.tab-content').scrollLeft =
        _state.inspector.horizontalScroll;

      // 6.1.2. Scroll vertically on the inspector
      hgetTab('inspector').querySelector('.tab-content').scrollTop =
        _state.inspector.verticalScroll;

      // 6.1.3. Scroll down the inspector if it's for logs
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

      // 6.2. Set focus on tab
      if (
        _state.navigation.currentTab &&
        (!_state.search.isEnabled || _state.search.isPending)
      ) {
        hgetTab(_state.navigation.currentTab).classList.add('is-active');

        // 6.2.1. Set focus on row - Tab
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
    }

    // 6.3. Set focus on menu
    if (_state.isMenuIng) {
      // 6.3.1. Set focus on row - Menu
      if (_state.menu.actions.length > 0)
        hgetPopupRow(
          _state.popup,
          _state.navigation.currentMenuRow
        ).classList.add('is-active');

      // 6.3.2. Set focus on row - Overview
      if (_state.overview.isEnabled)
        hgetPopupRow(
          _state.popup,
          _state.navigation.currentMenuRow
        ).classList.add('is-active');

      // 6.3.3. Set focus on row - Jump
      if (_state.jump.isEnabled && _state.jump.results.length > 0) {
        hgetPopupRow(
          _state.popup,
          _state.navigation.currentMenuRow
        ).classList.add('is-active');
        hgetPopupRow(
          _state.popup,
          _state.navigation.currentMenuRow
        ).scrollIntoView({ block: 'nearest', inline: 'nearest' });
      }
    }

    if (!_state.isFullyEmpty) {
      // 6.4. Set focus on tty
      if (_state.tty.isEnabled && _state.tty.lines.length > 0) {
        const ttyInput = hgetTtyInput();
        if (ttyInput) {
          ttyInput.focus();
          ttyInput.scrollIntoView({ block: 'nearest', inline: 'nearest' });

          // 6.4.1. Focus end of input
          setTimeout(() => {
            ttyInput.selectionStart = ttyInput.selectionEnd =
              ttyInput.value.length;
          }, _state._delays.default / 2);
        }
      }
    }

    // 6.5. Set focus on prompt input
    if (_state.prompt.isEnabled && _state.prompt.input.isEnabled) {
      // Dev-only (lack of input happens when the server spontaneously tells us we're authenticated)
      // (hence the need for checking it)
      if (hgetPromptInput()) hgetPromptInput().focus();
    }

    if (!_state.isFullyEmpty) {
      // 6.6. Set focus on search input
      if (_state.search.isEnabled && !_state.search.isPending) {
        hgetSearchInput().focus();
      }
      // 6.6.1. Unfocus search input when pending
      if (_state.search.isEnabled && _state.search.isPending) {
        hgetSearchInput().blur();
      }

      // 6.7. Set flag on previous/current tab for the "focus" layout
      if (_state.navigation.currentTab)
        hgetTab(_state.navigation.currentTab).classList.add('is-current');
      else if (_state.navigation.previousTab)
        hgetTab(_state.navigation.previousTab).classList.add('is-current');
    }

    // 6.8. Set focus on jump input
    if (_state.jump.isEnabled) {
      if (!_state.jump.search) hgetJumpInput().focus();
    }

    // 7. Set helper
    if (!_state.search.isEnabled)
      hgetHelper(_state.helper).classList.add('is-active');

    // 8. Set connection indicator
    hgetConnectionIndicator(
      _state.isConnected ? 'connected' : 'disconnected'
    ).classList.add('is-active');
    if (_state.isLoading)
      hgetConnectionIndicator('loading').classList.add('is-active');

    // 9. Set communication (master / agent / host) indicator
    if (
      _state.communication.availableAgents.length > 0 ||
      _state.communication.availableHosts.length > 0
    ) {
      hgetConnectionIndicator('communication-target').classList.add(
        'is-active'
      );

      let fullIndicator =
        _state.communication.availableAgents.length > 0
          ? `<button data-action="agent">${
              _state.communication.currentAgent || 'Master'
            }</button>`
          : 'Master';
      if (_state.communication.currentHost)
        fullIndicator = `${fullIndicator} (<button data-action="host">${_state.communication.currentHost}</button>)`;

      hgetConnectionIndicator('communication-target').innerHTML = fullIndicator;
    }

    // 10. Reset mobile controls' visibility
    hgetMobileControls().forEach((e) => {
      e.classList.remove('is-active');
    });

    // 11. Update the mobile controls' visibility

    // 11.1. Case when menuing
    if (_state.isMenuIng) {
      hgetMobileControl('reject').classList.add('is-active');
      hgetMobileControl('confirm').classList.add('is-active');
    }
    // 11.2. Case when tty-ing / showing a message
    else if (_state.popup === 'tty' || _state.popup === 'message') {
      hgetMobileControl('ttyQuit').classList.add('is-active');
    }
    // 11.3. Case when prompting
    else if (_state.prompt.isEnabled || _state.prompt.input.isEnabled) {
      hgetMobileControl('reject').classList.add('is-active');
      hgetMobileControl('confirm').classList.add('is-active');
    }
    // 11.4. Every other case (default navigation)
    else {
      hgetMobileControl('previousTab').classList.add('is-active');
      hgetMobileControl('nextTab').classList.add('is-active');
      hgetMobileControl('menu').classList.add('is-active');
      hgetMobileControl('bulk').classList.add('is-active');
      hgetMobileControl('shellSystem').classList.add('is-active');
      hgetMobileControl('theme').classList.add('is-active');
      hgetMobileControl('parameters').classList.add('is-active');
      if (_state.communication.availableAgents.length > 0)
        hgetMobileControl('agent').classList.add('is-active');
      if (_state.communication.availableHosts.length > 0)
        hgetMobileControl('host').classList.add('is-active');
      hgetMobileControl('overview').classList.add('is-active');
      hgetMobileControl('jump').classList.add('is-active');
    }

    // 12. Apply extra user settings if any

    // 12.1. Log lines wrap
    if (_state.settings.enableLogLinesWrap)
      if (_state.inspector.currentTab === 'Logs')
        hgetTab('inspector').classList.add('no-wrap');

    // 12.2. Log lines stripped background
    if (_state.settings.enableLogLinesStrippedBackground)
      if (_state.inspector.currentTab === 'Logs')
        hgetTab('inspector').classList.add('stripped-background');
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
   * @param {boolean} skipAutoForward
   */
  const websocketSend = (o, skipAutoForward = false) => {
    const copy = { ...o };

    if (!skipAutoForward) {
      if (state.communication.currentAgent)
        copy.Agent = state.communication.currentAgent;

      if (state.communication.currentHost)
        copy.Host = state.communication.currentHost;
    }

    wsSocket.send(JSON.stringify(copy));
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
     * @type {boolean}
     */
    isFullyEmpty: true,

    /**
     * @type {boolean}
     */
    shouldStopRendering: false,

    /**
     * @returns {boolean}
     */
    get isMenuIng() {
      return (
        state.popup &&
        [
          'menu',
          'bulk',
          'theme',
          'agent',
          'host',
          'parameters',
          'overview',
          'jump',
        ].includes(state.popup)
      );
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
     * @property {'menu'|'bulk'|'theme'|'agent'|'host'|'parameters'} key
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

      /**
       * @type {'menu'|'bulk'|'theme'|'agent'|'host'|'parameters'}
       */
      key: null,
    },

    /**
     * @type {'default'|'menu'|'prompt'|'prompt-input'|'message'|'picker'|'parameters'}
     */
    helper: 'default',

    /**
     * @type {"menu"|"bulk"|"prompt"|"message"|"tty"|"help"|"overview"|"theme"|"agent"|"host"|"parameters"|"jump"}
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

      /**
       * @type {Array<string>}
       */
      availableThemes: ['default', 'moon', 'dawn'],

      /**
       * @type {"default"|"moon"|"dawn"}
       */
      currentTheme: 'default',
    },

    settings: {
      /**
       * @type {boolean}
       */
      enableLogLinesWrap: true,

      /**
       * @type {boolean}
       */
      enableTimestampDisplay: false,

      /**
       * @type {boolean}
       */
      enableMenuPrompt: true,

      /**
       * @type {boolean}
       */
      enableOverviewOnLaunch: true,

      /**
       * @type {boolean}
       */
      enableLogLinesStrippedBackground: true,

      /**
       * @type {boolean}
       */
      enableJumpFuzzySearch: true,
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
     * @property {string} SortBy
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

    communication: {
      /**
       * @type {string}
       */
      currentAgent: null,

      /**
       * @type {Array<string>}
       */
      availableAgents: [],

      /**
       * @type {string}
       */
      currentHost: null,

      /**
       * @type {Array<string>}
       */
      availableHosts: [],
    },

    /**
     * @typedef Search
     * @property {string} query
     * @property {boolean} isEnabled
     * @property {boolean} isPending
     * @property {'logs'|'resource'} startedOn
     * @property {Array<Row>|Array<*>} previousRows
     */

    /**
     * @type Search
     */
    search: {
      query: null,
      isEnabled: false,
      isPending: false,
      startedOn: null,
      previousRows: [],
    },

    /**
     * @typedef {object} Jump
     * @property {boolean} isEnabled
     * @property {string} search
     * @property {Array<Row>} results
     * @property {Array<Row>} remoteResources
     * @property {Row} backlog
     */

    /**
     * @type {Jump}
     */
    jump: {
      isEnabled: false,
      search: null,
      results: [],
      remoteResources: [],
      backlog: null,
    },

    _delays: {
      /**
       * @type {number}
       */
      forAuthentication: 2000,

      /**
       * @type {number}
       */
      forConfirmations: 2000,

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

    /**
     * @typedef Overview
     * @property {boolean} isEnabled
     * @property {Array<Instance>} Instances
     */

    /**
     * @typedef Instance
     * @property {object} Docker
     * @property {string} Docker.Version
     * @property {string} Docker.Host
     * @property {object} Resources
     * @property {object} Resources.Containers
     * @property {number} Resources.Containers.Count
     * @property {object} Resources.Images
     * @property {number} Resources.Images.Count
     * @property {object} Resources.Volumes
     * @property {number} Resources.Volumes.Count
     * @property {object} Resources.Networks
     * @property {number} Resources.Networks.Count
     * @property {object} Server
     * @property {string} Server.Host
     * @property {string} Server.Name
     * @property {string} Server.Role
     * @property {number} Server.CountCPU
     * @property {number} Server.AmountRAM
     * @property {Array<string>} Server.Agents
     */

    /**
     * @type Overview
     */
    overview: {
      isEnabled: false,
      Instances: [],
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

  /**
   * @returns {MenuAction}
   */
  const sgetCurrentMenuAction = () => {
    return state.menu.actions[state.navigation.currentMenuRow - 1];
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
    if (state.isLoading) {
      // Except in jump mode
      if (!state.jump.isEnabled) return false;
    }

    // Prevent running anything other than submit while unauthenticated
    if (!state.isAuthenticated && !['confirm'].includes(cmd)) return false;

    // Force yes/no on prompts & messages & search
    if (
      (state.prompt.isEnabled ||
        state.message.isEnabled ||
        (state.search.isEnabled && !state.search.isPending) ||
        (state.popup && !state.isMenuIng)) &&
      !['confirm', 'reject', 'quit', 'ttyQuit'].includes(cmd)
    )
      return false;

    // Force yes/no/arrows on menus
    if (
      state.isMenuIng &&
      !['scrollUp', 'scrollDown', 'confirm', 'reject', 'quit'].includes(cmd)
    )
      return false;

    // Prevent multi-popup
    if (
      state.popup &&
      [
        'menu',
        'bulk',
        'prompt',
        'message',
        'shell',
        'theme',
        'parameters',
        'agent',
        'host',
        'overview',
      ].includes(cmd)
    )
      return false;

    // Prevent multi-inspect
    if (
      state.inspector.isEnabled &&
      !state.search.isEnabled &&
      ['confirm'].includes(cmd)
    )
      return false;

    // Prevent anything other than popups and confirmation when fully empty
    if (
      state.isFullyEmpty &&
      ![
        'scrollUp',
        'scrollDown',
        'confirm',
        'reject',
        'quit',
        'agent',
        'host',
        'parameters',
        'shellSystem',
        'message',
        'prompt',
        'overview',
      ].includes(cmd)
    )
      return false;

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

    // console.log(cmdString, isAllowed);

    if (!isAllowed) return;

    // Any allowed command will deactivate the render-blocker
    if (!['_render', '_enterInspect', 'confirm'].includes(cmdString))
      state.shouldStopRendering = false;

    cmd(...args);

    renderApp(state);
  };

  const debouncedCmdRun = debounce((cmd, ...args) => {
    cmdRun(cmd, ...args);
  }, state._delays.default);

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
     * Private - Clear menu actions
     */
    _clearMenu: function () {
      state.menu.actions = [];
      state.menu.key = null;
    },

    /**
     * Private - Clear overview
     */
    _clearOverview: function () {
      state.overview.isEnabled = false;
      state.overview.Instances = [];
      cmdRun(cmds._clearPopup);
    },

    /**
     * Private - Clear jump
     */
    _clearJump: function () {
      state.jump.isEnabled = false;
      state.jump.search = null;
      state.jump.results = [];
      state.jump.remoteResources = [];
      cmdRun(cmds._clearPopup);
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
     * Private - Desktop-only - TTY-only - Quit and close the TTY session
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
      // No text typed -> Use the current image's name
      if (!args.Image || args.Image.length === 0) {
        const currentImage = sgetCurrentRow();
        websocketSend({
          action: 'image.pull',
          args: { Image: `${currentImage.Name}:${currentImage.Version}` },
        });
      }
      // Else -> Use the user-supplied name
      else
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

      let payload = { Resource: currentRow };
      if (currentInspectorTab === 'Logs')
        payload['showTimestamps'] = state.settings.enableTimestampDisplay;

      // Produces something like : <tab>.inspect.<sub-tab>
      // Such as : container.inspect.logs
      websocketSend({
        // prettier-ignore
        action: `${currentTabKey.slice(0,-1)}.inspect.${currentInspectorTab.toLowerCase()}`,
        args: payload,
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
     * Private - Pick a theme and store it in LocalStorage
     * @param {MenuAction} action
     */
    _pickTheme: function (action) {
      state.appearance.currentTheme = action.Label.toLowerCase();
      localStorage.setItem('theme', state.appearance.currentTheme);
    },

    /**
     * Private - Pick an agent for further communication
     * @param {MenuAction} action
     */
    _pickAgent: function (action) {
      cmdRun(cmds._clear);

      if (action.Label === 'Master') state.communication.currentAgent = null;
      else state.communication.currentAgent = action.Label;

      cmdRun(cmds._init);
    },

    /**
     * Private - Agent-only - Clear any open stream / tty
     */
    _clear: function () {
      if (!state.communication.currentAgent) return;
      websocketSend({ action: 'clear' });
    },

    /**
     * Private - Pick a new host for further communication
     * @param {MenuAction} action
     */
    _pickHost: function (action) {
      state.communication.currentHost = action.Label;
      cmdRun(cmds._init);
    },

    /**
     * Private - Toggle a parameter
     * @param {MenuAction} action
     */
    _toggleParameter: function (action) {
      state.settings[action.Metadata['Key']] = !state.settings[
        action.Metadata['Key']
      ];
      localStorage.setItem(
        action.Metadata['Key'],
        state.settings[action.Metadata['Key']]
      );
      cmdRun(cmds._init);
    },

    /**
     * Private - Clear search and reset display
     */
    _clearSearch: function () {
      const isForLogs =
        state.inspector.isEnabled && state.inspector.currentTab === 'Logs';

      if (!isForLogs) {
        const currentTabKey = sgetCurrentTabKey();

        state.tabs = state.tabs.map((t) =>
          t.Key === currentTabKey
            ? { ...t, Rows: [...state.search.previousRows] }
            : t
        );
        state.navigation.currentTabsRows[currentTabKey] = 1;
      } else {
        state.inspector.content = [...state.search.previousRows];
      }

      state.search.isEnabled = false;
      state.search.isPending = false;
      state.search.query = null;
      state.search.startedOn = null;
      state.search.previousRows = [];
      hgetSearchInput().value = '';
    },

    /**
     * Private - Perform search (local one)
     * @param {boolean} resetOriginalRows
     */
    _performSearch: function (resetOriginalRows = false) {
      const { query } = state.search;
      const currentTab = sgetCurrentTab();
      const currentTabKey = currentTab.Key;
      const isForLogs =
        state.inspector.isEnabled && state.inspector.currentTab === 'Logs';

      // Reset when empty
      if (!query && state.search.previousRows.length > 0) {
        if (!isForLogs) {
          state.tabs = state.tabs.map((t) =>
            t.Key === currentTabKey
              ? { ...t, Rows: [...state.search.previousRows] }
              : t
          );
          state.navigation.currentTabsRows[currentTabKey] = 1;
        } else {
          state.inspector.content = [...state.search.previousRows];
        }
        return;
      }

      // Save initial rows
      if (state.search.previousRows.length === 0 || resetOriginalRows) {
        if (!isForLogs) state.search.previousRows = [...currentTab.Rows];
        else state.search.previousRows = [...state.inspector.content];
      }

      // else state.search.previousRows = [];

      // Perform search
      const currentRows = state.search.previousRows;
      const filteredRows = !isForLogs
        ? currentRows.filter((r) =>
            r._representation
              .map((e) => e.value)
              .filter((v) => v)
              .join('~')
              .toLowerCase()
              .includes(query.toLowerCase())
          )
        : currentRows.filter((r) =>
            r.Content.some((l) => l.toLowerCase().includes(query.toLowerCase()))
          );

      // Update tab's rows
      if (!isForLogs) {
        state.tabs = state.tabs.map((t) =>
          t.Key === currentTabKey ? { ...t, Rows: filteredRows } : t
        );
        state.navigation.currentTabsRows[currentTabKey] = 1;

        if (filteredRows.length > 0) cmdRun(cmds._refreshInspector);
      } else {
        state.inspector.content = [...filteredRows];
      }
    },

    _performJumpSearch: function () {
      const { search } = state.jump;

      if (!search) {
        state.jump.results = [];
        return;
      }

      const resources = state.tabs
        .map((t) =>
          t.Rows.map((r) => ({ ...r, Parent: t.Title, ParentKey: t.Key }))
        )
        .flat();

      let identifiedResources = [];

      // Default search
      if (!state.settings.enableJumpFuzzySearch)
        identifiedResources = resources.filter((r) =>
          r._representation
            .map((f) => f.value)
            .join('|')
            .toLowerCase()
            .includes(search.toLowerCase())
        );
      // Fuzzy search
      else {
        const options = {
          keys: ['Name'],
          isCaseSensitive: false,
          ignoreLocation: true,
          threshold: 0.3,
        };
        const fuse = new Fuse(resources, options);
        const results = fuse.search(search.toLowerCase());
        identifiedResources = results.map((r) => r.item);
      }

      // Remote search
      if (state.jump.remoteResources.length > 0) {
        if (!state.settings.enableJumpFuzzySearch)
          // Default
          identifiedResources.push(
            ...state.jump.remoteResources.filter((r) =>
              r._representation
                .map((f) => f.value)
                .join('|')
                .toLowerCase()
                .includes(search.toLowerCase())
            )
          );
        // Fuzzy
        else {
          const options = {
            keys: ['Name'],
            isCaseSensitive: false,
            ignoreLocation: true,
            threshold: 0.3,
          };
          const fuse = new Fuse(state.jump.remoteResources, options);
          const results = fuse.search(search.toLowerCase());
          identifiedResources.push(...results.map((r) => r.item));
        }
      }

      state.jump.results = [...identifiedResources];
      state.navigation.currentMenuRow = 1;
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

      cmdRun(cmds.reject);
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

      state.menu.key = 'menu';

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

      state.menu.key = 'bulk';

      websocketSend({ action: `${currentTab}.bulk` });
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
      if (
        state.isMenuIng &&
        state.navigation.currentMenuRow > 0 &&
        !state.overview.isEnabled &&
        !state.jump.isEnabled
      ) {
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

          cmdRun(cmds._clearMenu);
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
          // Save the focused menu row in case we need it later
          const _menuAction = sgetCurrentMenuAction();

          // Must clear first to enable running a command out of menu context
          cmdRun(cmds._clearPopup);

          if (attributes.useRow) {
            cmdRun(cmds[attributes.command], sgetCurrentRow());
            cmdRun(cmds._clearMenu);
            return;
          }

          if (attributes.useMenuAction) {
            cmdRun(cmds[attributes.command], _menuAction);
            cmdRun(cmds._clearMenu);
            return;
          }

          cmdRun(cmds[attributes.command]);
          cmdRun(cmds._clearMenu);
        }

        return;
      }

      // Overview confirm
      if (
        state.isMenuIng &&
        state.navigation.currentMenuRow > 0 &&
        state.overview.isEnabled
      ) {
        // prettier-ignore
        const availableRows = state.overview.Instances;

        if (availableRows.length === 1) {
          cmdRun(cmds._clearOverview);
          cmdRun(cmds._init);
          return;
        }

        const row = hgetPopupRow(state.popup, state.navigation.currentMenuRow);
        const attributes = row.dataset;

        // Case when : Multi-agents
        if (state.communication.availableAgents.length > 0) {
          if (attributes.role === 'Agent')
            cmdRun(cmds._pickAgent, { Label: attributes.name });
          else cmdRun(cmds._pickAgent, { Label: 'Master' });
        }
        // Case when : Multi-hosts
        else if (state.overview.Instances.length > 0) {
          cmdRun(cmds._pickHost, { Label: attributes.name });
        }

        cmdRun(cmds._clearOverview);
        return;
      }

      // Jump confirm
      if (
        state.isMenuIng &&
        state.navigation.currentMenuRow > 0 &&
        state.jump.isEnabled
      ) {
        if (state.jump.results.length === 0) return;

        // prettier-ignore
        const currentResult = { ...state.jump.results[state.navigation.currentMenuRow - 1] };

        cmdRun(cmds._clearJump);

        if (currentResult.Host) {
          state.jump.backlog = currentResult;
          cmdRun(cmds._pickHost, { Label: currentResult.Host });
          return;
        }

        state.navigation.currentTab = currentResult.ParentKey;
        state.navigation.currentTabsRows[currentResult.ParentKey] =
          sgetCurrentTab().Rows.findIndex((r) =>
            r.ID ? r.ID === currentResult.ID : r.Name === currentResult.Name
          ) + 1;
        state.isLoading = false;

        cmdRun(cmds._inspectorTabs);

        return;
      }

      // Search confirm
      if (state.search.isEnabled) {
        const isForLogs =
          state.inspector.isEnabled && state.inspector.currentTab === 'Logs';

        if (!isForLogs) {
          // No row found
          if (sgetCurrentTab().Rows.length === 0) cmdRun(cmds._clearSearch);
          // Rows found
          else {
            state.search.isPending = true;
            cmdRun(cmds._refreshInspector);
          }
        } else {
          if (state.inspector.content.length === 0) cmdRun(cmds._clearSearch);
          else state.search.isPending = true;
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
      if (state.overview.isEnabled && state.isFullyEmpty) return;

      if (state.prompt.isEnabled) {
        cmdRun(cmds._clearPrompt);
        return;
      }

      if (state.message.isEnabled) {
        cmdRun(cmds._clearMessage);
        return;
      }

      if (state.overview.isEnabled) {
        cmdRun(cmds._clearOverview);
        return;
      }

      if (state.jump.isEnabled) {
        cmdRun(cmds._clearJump);
        state.isLoading = false;
        return;
      }

      if (state.popup) {
        cmdRun(cmds._clearPopup);
        return;
      }

      if (state.search.isEnabled) {
        cmdRun(cmds._clearSearch);
        cmdRun(cmds._refreshInspector);
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
      if (state.search.isEnabled && state.search.isPending)
        cmdRun(cmds._clearSearch);

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
      if (state.search.isEnabled && state.search.isPending)
        cmdRun(cmds._clearSearch);

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

      if (
        state.search.isEnabled &&
        state.search.isPending &&
        !state.navigation.currentTab &&
        state.inspector.currentTab === 'Logs'
      )
        cmdRun(cmds._clearSearch);

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

      if (
        state.search.isEnabled &&
        state.search.isPending &&
        !state.navigation.currentTab &&
        state.inspector.currentTab === 'Logs'
      )
        cmdRun(cmds._clearSearch);

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
      if (
        state.isMenuIng &&
        !state.overview.isEnabled &&
        !state.jump.isEnabled
      ) {
        const availableRows = state.menu.actions;
        state.navigation.currentMenuRow += 1;

        // +1 is added to account for the extra "cancel" option
        if (state.navigation.currentMenuRow > availableRows.length + 1)
          state.navigation.currentMenuRow = 1;

        return;
      }

      // Overview - Next row
      if (state.isMenuIng && state.overview.isEnabled) {
        const availableRows = state.overview.Instances;
        state.navigation.currentMenuRow += 1;

        if (state.navigation.currentMenuRow > availableRows.length)
          state.navigation.currentMenuRow = 1;

        return;
      }

      // Jump - Next row
      if (state.isMenuIng && state.jump.isEnabled) {
        const availableRows = state.jump.results;
        state.navigation.currentMenuRow += 1;

        if (state.navigation.currentMenuRow > availableRows.length)
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
      if (
        state.isMenuIng &&
        !state.overview.isEnabled &&
        !state.jump.isEnabled
      ) {
        const availableRows = state.menu.actions;
        state.navigation.currentMenuRow -= 1;

        // +1 is added to account for the extra "cancel" option
        if (state.navigation.currentMenuRow < 1)
          state.navigation.currentMenuRow = availableRows.length + 1;

        return;
      }

      // Overview - Previous row
      if (state.isMenuIng && state.overview.isEnabled) {
        const availableRows = state.overview.Instances;
        state.navigation.currentMenuRow -= 1;

        if (state.navigation.currentMenuRow < 1)
          state.navigation.currentMenuRow = availableRows.length;

        return;
      }

      // Jump - Previous row
      if (state.isMenuIng && state.jump.isEnabled) {
        const availableRows = state.jump.results;
        state.navigation.currentMenuRow -= 1;

        if (state.navigation.currentMenuRow < 1)
          state.navigation.currentMenuRow = availableRows.length;

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
      state.menu.key = 'menu';

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

      if (state.settings.enableMenuPrompt)
        cmdRun(cmds._showPrompt, {
          text: 'Are you sure you want to pause/unpause this container?',
          callback: cmds._wsSend,
          callbackArgs: [
            {
              action: `container.pause`,
              args: { Resource: sgetCurrentRow() },
            },
          ],
        });
      else
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

      if (state.settings.enableMenuPrompt)
        cmdRun(cmds._showPrompt, {
          text: 'Are you sure you want to stop this container?',
          callback: cmds._wsSend,
          callbackArgs: [
            {
              action: `container.stop`,
              args: { Resource: sgetCurrentRow() },
            },
          ],
        });
      else
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

      if (currentTabKey === 'containers') {
        if (state.settings.enableMenuPrompt)
          cmdRun(cmds._showPrompt, {
            text: 'Are you sure you want to restart this container?',
            callback: cmds._wsSend,
            callbackArgs: [
              {
                action: `container.restart`,
                args: { Resource: sgetCurrentRow() },
              },
            ],
          });
        else
          websocketSend({
            action: `container.restart`,
            args: { Resource: sgetCurrentRow() },
          });
      } else if (currentTabKey === 'images')
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
          placeholder:
            '[repository/]name[:tag] -- Leave empty to pull the current image',
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

    /**
     * Public - Show theme picker
     */
    theme: function () {
      state.helper = 'picker';
      state.menu.key = 'theme';
      state.menu.actions = state.appearance.availableThemes.map((t) => ({
        RunLocally: true,
        RequiresResource: false,
        RequiresMenuAction: true,
        Label: t,
        Command: '_pickTheme',
      }));
      state.navigation.currentMenuRow = 1;
      cmdRun(cmds._showPopup, 'menu');
    },

    /**
     * Public - Show agent picker
     */
    agent: function () {
      if (state.communication.availableAgents.length === 0) return;

      state.helper = 'picker';
      state.menu.key = 'agent';

      state.menu.actions = state.communication.availableAgents.map((t) => ({
        RunLocally: true,
        RequiresResource: false,
        RequiresMenuAction: true,
        Label: t,
        Command: '_pickAgent',
      }));

      state.menu.actions.unshift({
        RunLocally: true,
        RequiresResource: false,
        RequiresMenuAction: true,
        Label: 'Master',
        Command: '_pickAgent',
      });

      state.navigation.currentMenuRow = 1;
      cmdRun(cmds._showPopup, 'menu');
    },

    /**
     * Public - Show host picker
     */
    host: function () {
      if (state.communication.availableHosts.length === 0) return;

      state.helper = 'picker';
      state.menu.key = 'host';

      state.menu.actions = state.communication.availableHosts.map((t) => ({
        RunLocally: true,
        RequiresResource: false,
        RequiresMenuAction: true,
        Label: t,
        Command: '_pickHost',
      }));

      state.navigation.currentMenuRow = 1;
      cmdRun(cmds._showPopup, 'menu');
    },

    /**
     * Public - Show parameters manager
     */
    parameters: function () {
      state.helper = 'parameters';
      state.menu.key = 'parameters';
      state.menu.actions = Object.keys(state.settings).map((k) => ({
        RunLocally: true,
        RequiresResource: false,
        RequiresMenuAction: true,
        Label: `[${state.settings[k] ? 'x' : ' '}] ${k}`,
        Command: '_toggleParameter',
        Metadata: { Key: k },
      }));
      state.navigation.currentMenuRow = 1;
      cmdRun(cmds._showPopup, 'menu');
    },

    /**
     * Public - Switch to the previous agent for further communication
     */
    previousAgent: function () {
      if (state.communication.availableAgents.length === 0) return;

      cmdRun(cmds._clear);

      const currentIndex = state.communication.availableAgents.indexOf(
        state.communication.currentAgent
      );

      // Case when currently communicating with master
      if (currentIndex === -1) {
        state.communication.currentAgent =
          state.communication.availableAgents[
            state.communication.availableAgents.length - 1
          ];
      }

      // Case when currently communicating with the first agent
      else if (currentIndex === 0) {
        state.communication.currentAgent = null;
      }

      // Regular case, switching to the previous agent
      else
        state.communication.currentAgent =
          state.communication.availableAgents[currentIndex - 1];

      cmdRun(cmds._init);
    },

    /**
     * Public - Switch to the next agent for further communication
     */
    nextAgent: function () {
      if (state.communication.availableAgents.length === 0) return;

      cmdRun(cmds._clear);

      const currentIndex = state.communication.availableAgents.indexOf(
        state.communication.currentAgent
      );

      // Case when currently communicating with master
      if (currentIndex === -1) {
        state.communication.currentAgent =
          state.communication.availableAgents[0];
      }

      // Case when currently communicating with the last agent
      else if (
        currentIndex ===
        state.communication.availableAgents.length - 1
      ) {
        state.communication.currentAgent = null;
      }

      // Regular case, switching to the next agent
      else
        state.communication.currentAgent =
          state.communication.availableAgents[currentIndex + 1];

      cmdRun(cmds._init);
    },

    /**
     * Public - Switch to the previous host for further communication
     */
    previousHost: function () {
      if (state.communication.availableHosts.length === 0) return;

      const currentIndex = state.communication.availableHosts.indexOf(
        state.communication.currentHost
      );

      if (currentIndex === 0) {
        state.communication.currentHost =
          state.communication.availableHosts[
            state.communication.availableHosts.length - 1
          ];
      }
      // Regular case, switching to the previous agent
      else
        state.communication.currentHost =
          state.communication.availableHosts[currentIndex - 1];

      cmdRun(cmds._init);
    },

    /**
     * Public - Switch to the next host for further communication
     */
    nextHost: function () {
      if (state.communication.availableHosts.length === 0) return;

      const currentIndex = state.communication.availableHosts.indexOf(
        state.communication.currentHost
      );

      if (currentIndex === state.communication.availableHosts.length - 1) {
        state.communication.currentHost = state.communication.availableHosts[0];
      } else
        state.communication.currentHost =
          state.communication.availableHosts[currentIndex + 1];

      cmdRun(cmds._init);
    },

    /**
     * Public - Copy the inspector's logs to the clipboard
     */
    inspectorCopyLogs: function () {
      const button = q('button[data-action="inspectorCopyLogs"]');
      const inspectorContent = q('.tab.for-inspector .tab-content');
      const toCopy = inspectorContent.textContent;

      if (!toCopy) return;

      const _showConfirmation = () => {
        state.message.category = 'report';
        state.message.type = 'success';
        state.message.title = 'Confirmation';
        state.message.content =
          'The logs of this container were copied to your clipboard';
        state.message.isEnabled = true;
        state.helper = 'message';

        cmdRun(cmds._showPopup, 'message');
        setTimeout(() => {
          if (state.message.isEnabled) cmdRun(cmds._clearMessage);
        }, state._delays.forConfirmations);
      };

      // Modern
      if (navigator.clipboard) {
        navigator.clipboard.writeText(toCopy).then(_showConfirmation);
      }
      // Compatible
      else {
        const textArea = document.createElement('textarea');
        textArea.value = toCopy;
        textArea.style.top = '0';
        textArea.style.left = '0';
        textArea.style.position = 'fixed';

        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        try {
          const successful = document.execCommand('copy');
          if (successful) _showConfirmation();
        } catch (err) {}

        document.body.removeChild(textArea);
      }
    },

    /**
     * Public - Toggle search mode
     */
    search: function () {
      state.search.isEnabled = true;
      state.search.isPending = false;
      state.search.startedOn = 'resource';

      if (state.inspector.isEnabled && state.inspector.currentTab !== 'Logs')
        cmdRun(cmds._exitInspect);
      else if (
        state.inspector.isEnabled &&
        state.inspector.currentTab === 'Logs'
      )
        state.search.startedOn = 'logs';
    },

    /**
     * Public - On click - Focus the search input
     */
    focusSearch: function () {
      hgetSearchInput().focus();
    },

    /**
     * Public - Request a global overview from the server
     */
    overview: function () {
      // When no other Agent exist
      if (state.communication.availableAgents.length === 0)
        websocketSend({ action: `overview` });
      // When current Agent is Master, and other Agents exist
      else if (
        !state.communication.currentAgent &&
        state.communication.availableAgents.length > 0
      ) {
        websocketSend({ action: `overview` });
        for (const agent of state.communication.availableAgents)
          websocketSend({ action: `overview`, Agent: agent }, true);
      }
      // When current Agent isn't Master, and other Agents exist
      else if (
        state.communication.currentAgent &&
        state.communication.availableAgents.length > 0
      ) {
        websocketSend({ action: `overview` }, true);
        for (const agent of state.communication.availableAgents)
          websocketSend({ action: `overview`, Agent: agent }, true);
      }
    },

    /**
     * Public - Jump to any resource
     */
    jump: function () {
      state.jump.isEnabled = true;
      state.helper = 'jump';
      cmdRun(cmds._showPopup, 'jump');

      // When other hosts exist
      if (state.communication.availableHosts.length > 0) {
        for (const host of state.communication.availableHosts)
          if (host !== state.communication.currentHost)
            websocketSend({ action: `enumerate`, Host: host }, true);
      }
    },

    /**
     * Public - Display current version
     */
    version: function () {
      state.message.category = 'report';
      state.message.type = 'success';
      state.message.title = 'Version';
      state.message.content = `You are currently running the version <em class="has-accent">${VERSION}</em> of Isaiah.`;
      state.message.isEnabled = true;
      state.helper = 'message';
      cmdRun(cmds._showPopup, 'message');

      fetch(`https://api.github.com/repos/will-moss/isaiah/releases/latest`)
        .then((r) => r.json())
        .then((d) => {
          const latestVersion = d.tag_name;

          state.message.content +=
            d.tag_name === VERSION
              ? `<span class="line-break"></span>It is the most recent version available.`
              : `<span class="line-break"></span>The more recent version <em class="has-accent">${latestVersion}</em> is available.`;

          cmdRun(cmds._render);
        });
    },
  };

  // === Variables

  /**
   * @type {string}
   */
  const VERSION = '-VERSION-';

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
    X: 'parameters',
    T: 'theme',

    // Agents
    '<': 'previousAgent',
    '>': 'nextAgent',

    // Hosts & Agents
    k: 'previousHost',
    l: 'nextHost',
    A: 'agent',
    H: 'host',

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
    O: 'overview',

    // Misc
    '?': 'help',
    '/': 'search',
    J: 'jump',
    V: 'version',

    // Appearance
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

    if (state.search.isEnabled && !state.search.isPending) {
      return;
    }

    if (state.jump.isEnabled) {
      // Menu navigation
      if (['ArrowUp', 'ArrowDown'].includes(evt.key)) evt.preventDefault();
      else return;
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
   * Called every time the user releases a key up
   * on their keyboard. Will exclusively be handled during
   * search and jump, as it requires live input and
   * off-by-one characters wouldn't work
   */
  const listenerKeyUp = (evt) => {
    if (state.jump.isEnabled) {
      listenerJumpKeyUp(evt);
      return;
    }

    if (!state.search.isEnabled) return;
    if (state.search.isPending) return;

    listenerSearchInputKeyUp(evt);
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
   * Called every time the user presses a key down
   * on their keyboard, while in Search Input mode. Will
   * allow the user to input data using any key, while
   * Escape and Enter will be used to Leave/Confirm the
   * Search
   */
  const listenerSearchInputKeyUp = (evt) => {
    const { key } = evt;

    if (!key) return;

    if (['Escape', 'Enter'].includes(key)) {
      cmdRun(cmds[kbMap[key]]);
      return;
    }

    state.search.query = hgetSearchInput().value;
    cmdRun(cmds._performSearch);
  };

  /**
   * Called every time the user presses a key up
   * on their keyboard, while in jump mode. Will
   * allow the user to input data using any key, while
   * Escape and Enter will be used to Leave/Confirm the
   * mode, along with arrows keys for navigation
   */
  const listenerJumpKeyUp = (evt) => {
    if (evt.metaKey) return;

    const { key } = evt;

    if (!key) return;

    // Cancel / Confirm
    if (['Escape', 'Enter'].includes(key)) {
      cmdRun(cmds[kbMap[key]]);
      return;
    }

    // Normal input
    if (key.length === 1 && /[a-zA-Z0-9-_ ]/.test(key)) {
      // No need to update if search hasn't changed
      if (state.jump.search === hgetJumpInput().value) return;

      state.jump.search = hgetJumpInput().value;
      debouncedCmdRun(cmds._performJumpSearch);
    }

    // Erase
    if (['Backspace', 'Delete'].includes(key)) {
      state.jump.search = hgetJumpInput().value;
      debouncedCmdRun(cmds._performJumpSearch);
      return;
    }

    return;
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
        if (state.search.isEnabled) cmdRun(cmds._clearSearch);
        state.navigation.currentTab = key;
        cmdRun(cmds._inspectorTabs);
      }
      // 1.2. Inspector Tab Header
      else if (part === 'inspector') {
        if (state.search.isEnabled && state.search.startedOn === 'logs')
          cmdRun(cmds._clearSearch);

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
          if (_key !== state.navigation.previousTab) {
            if (state.search.isEnabled) cmdRun(cmds._clearSearch);
          }

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
            if (_key !== state.navigation.previousTab) {
              if (state.search.isEnabled) cmdRun(cmds._clearSearch);
            } else {
              if (state.search.isEnabled && state.search.startedOn === 'logs')
                cmdRun(cmds._clearSearch);
            }

            if (state.inspector.isEnabled) cmdRun(cmds._exitInspect);
            state.navigation.currentTab = _key;
          }

          // 1.4.2.2. Focus the clicked row and refresh the inspector
          const rowIndex = Array.from(tabContent.children).indexOf(tabRow);
          state.navigation.currentTabsRows[_key] = rowIndex + 1;

          if (state.search.isEnabled) state.search.isPending = true;

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

    // 4. Explicit jump via data-jump attribute (e.g. data-jump="[host.]images.great-author/wonderful-image")
    else if (target.hasAttribute('data-jump')) {
      const parts = target.getAttribute('data-jump').split('.');
      cmdRun(cmds._clearJump);

      let host, tabKey, resourceID;
      if (parts.length === 2) {
        tabKey = parts[0];
        resourceID = parts[1];
      } else if (parts.length === 3) {
        host = parts[0];
        tabKey = parts[1];
        resourceID = parts[2];
      }

      if (host) {
        state.jump.backlog = { ParentKey: tabKey, ID: resourceID, Host: host };
        cmdRun(cmds._pickHost, { Label: host });
        return;
      }

      state.navigation.currentTab = tabKey;
      state.navigation.currentTabsRows[tabKey] =
        sgetCurrentTab().Rows.findIndex((r) =>
          r.ID ? r.ID === resourceID : r.Name === resourceID
        ) + 1;

      cmdRun(cmds._inspectorTabs);

      return;
    }

    // 4. Clicked anywhere else
    else {
      // 4.0. If tty-ing, focus the tty input
      if (state.tty.isEnabled) {
        hgetTtyInput().focus();
        return;
      }

      // 4.1. If jump-ing, focus the jump input
      if (state.jump.isEnabled) {
        hgetJumpInput().focus();
        return;
      }

      // 4.2. If any popup is activated, while not menuing and not tty-ing, dismiss it
      if (!state.isMenuIng && state.popup) cmdRun(cmds.reject);
      // 4.3. If menuing, and clicked outside the menu, dismiss it
      else if (state.isMenuIng && target.classList.contains('popup-layer'))
        cmdRun(cmds.reject);
      // 4.4. If menuing, and clicked inside the menu, run the 3. scenario (assumption: we clicked a span inside a row)
      else {
        // Clicked inside Overview's menu
        let tabContent, tabRow, rowIndex;
        if (target.classList.contains('row')) {
          tabContent = target.parentNode;
          tabRow = target;
          rowIndex = Array.from(tabContent.children).indexOf(tabRow) + 1;
        }
        // Clicked inside regular menu
        else {
          tabContent = target.parentNode.parentNode;
          tabRow = target.parentNode;
          rowIndex = Array.from(tabContent.children).indexOf(tabRow) + 1;
        }

        state.navigation.currentMenuRow = rowIndex;
        cmdRun(cmds.confirm);
      }
    }

    cmdRun(cmds._render);
  };

  /**
   * Called every time the user moves their mouse.
   * This listener is responsible only for allowing
   * the user to interrupt the rendering loop when
   * hovering the Inspector, so they can select text
   * and not lose it on every render loop
   */
  const listenerMouseMove = (evt) => {
    // if (state.inspector.currentTab !== 'Logs') return;

    const inspector = hgetTab('inspector');

    state.shouldStopRendering = inspector.matches(':hover') ? true : false;
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

    const searchIsEnabled = state.search.isEnabled;
    const searchIsPending = state.search.isPending;
    const searchIsForLogs =
      searchIsEnabled && state.search.startedOn === 'logs';
    let reapplySearch = false;

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
        if (notification.Content.Tabs) {
          state.tabs = notification.Content.Tabs;
          state.navigation.currentTab = state.tabs[0].Key;
          state.navigation.currentTabsRows = state.tabs.reduce(
            (a, b) => ({ ...a, [b.Key]: 1 }),
            {}
          );
          state.isFullyEmpty = false;

          // Perform sort if applicable
          state.tabs = state.tabs.map((tab) => ({
            ...tab,
            Rows: !tab.SortBy
              ? tab.Rows
              : tab.Rows.toSorted((a, b) => {
                  const inReverse = tab.SortBy.startsWith('-');
                  const key = inReverse ? tab.SortBy.slice(1) : tab.SortBy;

                  let val1 = a[key];
                  let val2 = b[key];
                  const comparisonType = getGeneralType(!val1 ? val2 : val1);

                  if (comparisonType === 'string')
                    return !inReverse
                      ? val1.localeCompare(val2)
                      : val2.localeCompare(val1);
                  else if (comparisonType === 'numeric')
                    return !inReverse ? val1 - val2 : val2 - val1;
                }),
          }));
        } else {
          state.isFullyEmpty = true;
        }

        // Update agents list only on the very first init
        if (state.communication.availableAgents.length === 0)
          state.communication.availableAgents =
            notification.Content.Agents || [];

        // Update hosts list only on the very first init
        if (state.communication.availableHosts.length === 0) {
          state.communication.availableHosts = notification.Content.Hosts || [];

          // When the host was picked using the "Overview" pane, don't re-run an init sequence
          if (
            state.communication.availableHosts.length > 0 &&
            !state.communication.currentHost
          ) {
            state.communication.currentHost =
              state.communication.availableHosts[0];

            cmdRun(cmds._init);
          }
        }

        // Jump to the picked resource if previously Jumped to a new host
        if (state.jump.backlog) {
          state.navigation.currentTab = state.jump.backlog.ParentKey;
          state.navigation.currentTabsRows[state.jump.backlog.ParentKey] =
            sgetCurrentTab().Rows.findIndex((r) =>
              r.ID
                ? r.ID === state.jump.backlog.ID
                : r.Name === state.jump.backlog.Name
            ) + 1;
          state.jump.backlog = null;
        }

        state.isLoading = false;
        if (!state.isFullyEmpty) cmdRun(cmds._inspectorTabs);
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
                if (state.settings.enableOverviewOnLaunch)
                  cmdRun(cmds.overview);
                else cmdRun(cmds._init);
              }, state._delays.forAuthentication);
            }
            // Dev-only
            else {
              cmdRun(cmds._clearMessage);
              cmdRun(cmds._clearPrompt);
              state.isAuthenticated = true;
              if (state.settings.enableOverviewOnLaunch) cmdRun(cmds.overview);
              else cmdRun(cmds._init);
            }
          }
        }

        state.isLoading = false;
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

            if (searchIsEnabled) reapplySearch = true;
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
          state.helper = 'menu';
          state.isLoading = false;

          cmdRun(cmds._showPopup, 'menu');
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
            if (notification.Content.Inspector.Content[0].Type === 'lines') {
              if (!searchIsEnabled || !searchIsForLogs)
                state.inspector.content.push(
                  ...notification.Content.Inspector.Content
                );
              else if (searchIsEnabled && searchIsForLogs)
                state.search.previousRows.push(
                  ...notification.Content.Inspector.Content
                );
            }
            // Else, replace the current content with the one received
            else {
              if (!searchIsEnabled || !searchIsForLogs)
                state.inspector.content =
                  notification.Content.Inspector.Content;
              else if (searchIsEnabled && searchIsForLogs)
                state.search.previousRows =
                  notification.Content.Inspector.Content;
            }
          }
        }

        if ('Agents' in notification.Content) {
          cmdRun(cmds._clear);

          // prettier-ignore
          state.communication.availableAgents = notification.Content.Agents || [];
          state.communication.currentAgent = null;
          cmdRun(cmds._init);
        }

        if ('Overview' in notification.Content) {
          state.isLoading = false;

          // Case when : First launch with multiple Hosts
          if (
            state.isFullyEmpty &&
            notification.Content.Overview.Instances.length === 1 &&
            notification.Content.Overview.Instances[0].Server.Hosts &&
            state.communication.availableHosts.length === 0
          ) {
            state.communication.availableHosts = [
              ...notification.Content.Overview.Instances[0].Server.Hosts,
            ];
            cmdRun(cmds.overview);
            return;
          }

          // prettier-ignore
          state.overview.Instances = [
            ...state.overview.Instances, ...notification.Content.Overview.Instances,
          ];
          state.overview.isEnabled = true;
          state.navigation.currentMenuRow = 1;
          state.helper = state.isFullyEmpty ? 'overview' : 'picker';
          cmdRun(cmds._showPopup, 'overview');
          break;
        }

        if ('Enumeration' in notification.Content) {
          // Jump can be disabled if we chose a local resource before enumeration finished
          if (!state.jump.isEnabled) return;

          state.jump.remoteResources.push(
            ...notification.Content.Enumeration.map((t) =>
              t.Rows.map((r) => ({
                ...r,
                Host: notification.Content.Host,
                Parent: t.Title,
                ParentKey: t.Key,
              }))
            ).flat()
          );
          state.isLoading = false;
          cmdRun(cmds._performJumpSearch);
          break;
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

    // Prevent re-rendering when user is in prompt or menu
    if (!state.isMenuIng && !state.prompt.input.isEnabled) {
      renderApp(state);

      // Reapply search when search was enabled prior and new data was received
      if (searchIsEnabled && reapplySearch) {
        if (!searchIsForLogs && searchIsPending)
          cmdRun(cmds._performSearch, true);
      }
    }
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
    // 0. Append custom.css stylesheet
    document
      .getElementsByTagName('head')[0]
      .insertAdjacentHTML(
        'beforeend',
        `<link type="text/css" rel="stylesheet" href="/assets/css/custom.css?v=${new Date().getTime()}" />`
      );

    // 1. Load user settings if any
    // prettier-ignore
    {
      state.appearance.currentTheme = lsGet('theme', 'default');
      state.settings.enableMenuPrompt = lsGet('enableMenuPrompt', true);
      state.settings.enableLogLinesWrap = lsGet('enableLogLinesWrap', false);
      state.settings.enableLogLinesStrippedBackground = lsGet('enableLogLinesStrippedBackground', true);
      state.settings.enableTimestampDisplay = lsGet('enableTimestampDisplay', false);
      state.settings.enableOverviewOnLaunch = lsGet('enableOverviewOnLaunch', true);
      state.settings.enableJumpFuzzySearch = lsGet('enableJumpFuzzySearch', true);
    }

    // 1.1. Load fuzzy-search library if enabled by the user
    if (state.settings.enableJumpFuzzySearch) {
      const script = document.createElement('script');
      script.type = `text/javascript`;
      script.src = `/assets/js/lib.fuse.min.js`;
      document.head.appendChild(script);
    }

    // 2. Connect to server (first execution loop)
    websocketConnect();

    // 3. Set keyboard listener (second execution loop)
    window.addEventListener('keydown', listenerKeyDown);
    window.addEventListener('keyup', listenerKeyUp);

    // 4. Set mouse listener (third execution loop)
    window.addEventListener('click', listenerMouseClick);
    window.addEventListener('mousemove', listenerMouseMove);
  });
})(window);
