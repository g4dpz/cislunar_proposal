// main.js — Client-side progressive enhancement

(function () {
  "use strict";

  // ─── Copy-to-Clipboard ────────────────────────────────────────────────────

  /**
   * Attach click handlers to all .btn-copy buttons.
   * Each button copies the text content of the adjacent <pre><code> block.
   */
  function initCopyButtons() {
    var buttons = document.querySelectorAll(".btn-copy");

    buttons.forEach(function (button) {
      button.addEventListener("click", function () {
        var wrapper = button.closest(".code-block-wrapper");
        if (!wrapper) return;

        var codeEl = wrapper.querySelector("code");
        if (!codeEl) return;

        var text = codeEl.textContent || "";

        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard.writeText(text).then(function () {
            showCopied(button);
          }).catch(function () {
            fallbackCopy(text, button);
          });
        } else {
          fallbackCopy(text, button);
        }
      });
    });
  }

  /**
   * Fallback copy method using a temporary textarea element.
   */
  function fallbackCopy(text, button) {
    var textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();

    try {
      document.execCommand("copy");
      showCopied(button);
    } catch (_e) {
      // Silently fail — progressive enhancement
    }

    document.body.removeChild(textarea);
  }

  /**
   * Briefly change button text to indicate successful copy.
   */
  function showCopied(button) {
    var originalText = button.textContent;
    button.textContent = "Copied!";
    button.classList.add("copied");

    setTimeout(function () {
      button.textContent = originalText;
      button.classList.remove("copied");
    }, 2000);
  }

  // ─── Mobile Nav Toggle (Progressive Enhancement) ──────────────────────────

  /**
   * Ensure mobile nav toggle works even without Bootstrap JS loaded.
   * This is a fallback — Bootstrap's collapse plugin handles this normally.
   */
  function initMobileNav() {
    var toggler = document.querySelector(".navbar-toggler");
    var navCollapse = document.querySelector("#mainNav");

    if (!toggler || !navCollapse) return;

    // Only add fallback if Bootstrap's Collapse is not available
    if (typeof window.bootstrap === "undefined" || !window.bootstrap.Collapse) {
      toggler.addEventListener("click", function () {
        var isExpanded = toggler.getAttribute("aria-expanded") === "true";
        toggler.setAttribute("aria-expanded", String(!isExpanded));

        if (isExpanded) {
          navCollapse.classList.remove("show");
        } else {
          navCollapse.classList.add("show");
        }
      });
    }
  }

  // ─── Initialization ───────────────────────────────────────────────────────

  document.addEventListener("DOMContentLoaded", function () {
    initCopyButtons();
    initMobileNav();
  });
})();
