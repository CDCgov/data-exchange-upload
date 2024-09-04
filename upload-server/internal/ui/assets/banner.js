(function webBanner() {
  const banner = document.querySelector('#nav-banner');

  if (banner) {
    banner.innerHTML = `
      <a class="skip-main" href="#main" tabindex="0">Skip to main content</a>
      <div role="navigation" id="nav" class="nav-container">
        <a href="/">
          <img src="/assets/dex_logo.svg" alt="DEX logo" /> Upload
        </a>
      </div>
    `;
  }
})();