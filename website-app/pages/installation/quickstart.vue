<script setup lang="ts">
definePageMeta({ layout: "docs" });

useHead({
  title: "Quickstart | env-sync",
  meta: [
    {
      name: "description",
      content:
        "Get env-sync running in under a minute. One command install, initialize, and start syncing secrets across your machines.",
    },
    {
      name: "keywords",
      content:
        "env-sync quickstart, dotenv sync setup, fast install, peer-to-peer sync",
    },
    { property: "og:title", content: "Quickstart | env-sync" },
    {
      property: "og:description",
      content:
        "Get env-sync running in under a minute. One command install, initialize, and start syncing secrets across your machines.",
    },
    { property: "og:type", content: "article" },
    {
      property: "og:url",
      content: "https://envsync.arnav.tech/installation/quickstart",
    },
    {
      property: "og:image",
      content: "https://envsync.arnav.tech/assets/cover.png",
    },
    { name: "twitter:card", content: "summary_large_image" },
    { name: "twitter:title", content: "Quickstart | env-sync" },
    {
      name: "twitter:description",
      content:
        "Get env-sync running in under a minute. One command, no accounts, no server.",
    },
    {
      name: "twitter:image",
      content: "https://envsync.arnav.tech/assets/cover.png",
    },
  ],
  link: [
    {
      rel: "canonical",
      href: "https://envsync.arnav.tech/installation/quickstart",
    },
  ],
});
</script>

<template>
  <NuxtLink class="back-link" to="/installation"
    >← Installation guides</NuxtLink
  >

  <div class="subpage-hero">
    <h1>Quickstart</h1>
    <p>
      Get peer-to-peer secret synchronization running in under a minute. One
      command, no accounts, no server to provision.
    </p>
  </div>

  <section class="panel">
    <div class="section-header">
      <h2>
        <div class="number">1</div>
        Install
      </h2>
    </div>
    <pre><code># system-wide install
curl -fsSL https://envsync.arnav.tech/install.sh | sudo bash

# or user-only install (no sudo)
curl -fsSL https://envsync.arnav.tech/install.sh | bash -s -- --user</code></pre>
    <p>
      The installer detects your platform, downloads the binary, and places it
      in your <code>PATH</code>. Works on Linux, macOS, and WSL2.
    </p>
  </section>

  <section class="panel">
    <div class="section-header">
      <h2>
        <div class="number">2</div>
        Initialize
      </h2>
    </div>
    <pre><code># verify installation
env-sync --version

# initialize (uses trusted-owner-ssh mode by default)
env-sync init</code></pre>
    <p>
      This creates the configuration directory at
      <code>~/.config/env-sync/</code> and generates any keys needed for your
      security mode.
    </p>
  </section>

  <section class="panel">
    <div class="section-header">
      <h2>
        <div class="number">3</div>
        Add secrets &amp; sync
      </h2>
    </div>
    <pre><code># add your first secret
env-sync add OPENAI_API_KEY="sk-abc123xyz"

# discover other machines on your network
env-sync discover

# sync with all discovered peers
env-sync sync</code></pre>
  </section>

  <section class="panel">
    <div class="section-header">
      <h2>
        <div class="number">4</div>
        Automate (optional)
      </h2>
    </div>
    <pre><code># install a cron job to sync every 30 minutes
env-sync cron --install

# or load secrets automatically on shell startup
# add this to ~/.bashrc or ~/.zshrc:
eval "$(env-sync load 2>/dev/null)"</code></pre>
  </section>

  <section class="panel">
    <div class="section-header">
      <h2>Prerequisites</h2>
    </div>
    <div class="prereq-list">
      <div class="prereq-item">
        <strong>Go 1.24+</strong>
        <span>Only needed if building from source</span>
      </div>
      <div class="prereq-item">
        <strong>SSH client</strong>
        <span
          >For <code>trusted-owner-ssh</code> mode (installed by default)</span
        >
      </div>
      <div class="prereq-item">
        <strong>mDNS support</strong>
        <span>Avahi on Linux, Bonjour on macOS (built-in)</span>
      </div>
    </div>
  </section>

  <section class="panel">
    <div class="section-header">
      <h2>Choosing the right mode</h2>
    </div>
    <div class="mode-options">
      <div class="mode-option">
        <h3>Personal fleet</h3>
        <p>Use <code>trusted-owner-ssh</code>. Your laptop, desktop, server, NUC — all behind SSH trust you already manage.</p>
      </div>
      <div class="mode-option">
        <h3>Team collaboration</h3>
        <p>Use <code>secure-peer</code>. Team members get secrets without SSH access to each other's machines.</p>
      </div>
      <div class="mode-option">
        <h3>Quick debugging</h3>
        <p>Use <code>dev-plaintext-http</code>. Fast setup for throwaway testing — never store real credentials here.</p>
      </div>
    </div>
  </section>

  <section class="panel">
    <h2>Build from source</h2>
    <pre><code>git clone https://github.com/championswimmer/env.sync.local.git
cd env.sync.local
make build
make test
sudo make install        # system-wide
# or: make install-user  # user-only</code></pre>
  </section>

  <section class="panel">
    <h2>Upgrade &amp; uninstall</h2>
    <pre><code># upgrade — re-run the installer
curl -fsSL https://envsync.arnav.tech/install.sh | sudo bash

# uninstall
env-sync service uninstall
rm -rf ~/.config/env-sync
sudo rm -f /usr/local/bin/env-sync</code></pre>
  </section>

  <section class="cta-banner">
    <h2>What's next?</h2>
    <p>
      Choose the setup guide that matches your scenario — personal devices or
      cross-team collaboration.
    </p>
    <div class="cta-row" style="justify-content: center">
      <NuxtLink class="btn btn-primary" to="/installation/trusted-peers"
        >Trusted Peers guide →</NuxtLink
      >
      <NuxtLink class="btn btn-secondary" to="/installation/secure-peers"
        >Secure Peers guide →</NuxtLink
      >
    </div>
  </section>
</template>
