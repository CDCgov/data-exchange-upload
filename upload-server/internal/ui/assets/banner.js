(function webBanner() {
  const banner = document.querySelector('#nav-banner');

  if (banner) {
    banner.innerHTML = `
      <div role="navigation" id="nav" class="nav-container">
        <a href="/">
          <img src="/assets/dex_logo.svg" /> Upload
        </a>
      </div>
    `;
  }
})();