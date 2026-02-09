(function () {
  function applyColumnVisibility(table, visibleByIndex) {
    visibleByIndex.forEach(function (isVisible, index) {
      var nth = index + 1;
      var cells = table.querySelectorAll('tr > *:nth-child(' + nth + ')');
      cells.forEach(function (cell) {
        cell.style.display = isVisible ? '' : 'none';
      });
    });
  }

  function initWindowColumnPicker(windowEl) {
    var table = windowEl.querySelector('table.table');
    var tableWrap = windowEl.querySelector('.table-wrap');

    if (!table || !tableWrap) {
      return;
    }

    var headers = Array.prototype.slice.call(table.querySelectorAll('thead th'));
    if (headers.length === 0) {
      return;
    }

    var configEl = document.createElement('section');
    configEl.className = 'column-config';
    configEl.innerHTML =
      '<div class="column-config-head">' +
      '  <span class="column-config-title">Fields</span>' +
      '  <span class="column-config-count"></span>' +
      '  <button class="button ghost column-config-toggle" type="button" aria-expanded="false">Choose Columns</button>' +
      '</div>' +
      '<div class="column-options" hidden></div>';

    tableWrap.parentNode.insertBefore(configEl, tableWrap);

    var countEl = configEl.querySelector('.column-config-count');
    var toggleButton = configEl.querySelector('.column-config-toggle');
    var optionsEl = configEl.querySelector('.column-options');

    headers.forEach(function (header, index) {
      var optionEl = document.createElement('label');
      optionEl.className = 'column-option';
      optionEl.innerHTML =
        '<input type="checkbox" data-column-index="' + index + '" checked />' +
        '<span>' + header.textContent.trim() + '</span>';
      optionsEl.appendChild(optionEl);
    });

    function sync() {
      var inputs = Array.prototype.slice.call(optionsEl.querySelectorAll('input[type="checkbox"]'));
      var checkedCount = inputs.filter(function (input) {
        return input.checked;
      }).length;

      // At least one column must remain visible for predictable table output.
      if (checkedCount === 0 && inputs[0]) {
        inputs[0].checked = true;
        checkedCount = 1;
      }

      var visibility = inputs.map(function (input) {
        return input.checked;
      });

      countEl.textContent = checkedCount + ' selected';
      applyColumnVisibility(table, visibility);
    }

    toggleButton.addEventListener('click', function () {
      var willOpen = optionsEl.hidden;
      optionsEl.hidden = !willOpen;
      toggleButton.setAttribute('aria-expanded', String(willOpen));
      toggleButton.textContent = willOpen ? 'Close Columns' : 'Choose Columns';
    });

    optionsEl.addEventListener('change', function (event) {
      if (!(event.target instanceof HTMLInputElement)) {
        return;
      }
      sync();
    });

    sync();
  }

  function init() {
    var windows = document.querySelectorAll('.data-window');
    windows.forEach(function (windowEl) {
      initWindowColumnPicker(windowEl);
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
