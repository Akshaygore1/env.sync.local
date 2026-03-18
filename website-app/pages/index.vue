<script setup lang="ts">
useHead({
  title: "env-sync | Sync .env secrets across machines",
  meta: [
    {
      name: "description",
      content:
        "env-sync is a peer-to-peer .env secrets synchronization tool for local networks. Auto-discover peers with mDNS, sync over SSH or mTLS, and keep every machine in sync — no central server required.",
    },
    {
      property: "og:title",
      content: "env-sync | Sync .env secrets across machines",
    },
    {
      property: "og:description",
      content:
        "env-sync is a peer-to-peer .env secrets synchronization tool for local networks. Auto-discover peers with mDNS, sync over SSH or mTLS, and keep every machine in sync.",
    },
    { property: "og:url", content: "https://envsync.arnav.tech/" },
    { property: "og:type", content: "website" },
    {
      name: "twitter:title",
      content: "env-sync | Sync .env secrets across machines",
    },
    {
      name: "twitter:description",
      content:
        "env-sync is a peer-to-peer .env secrets synchronization tool for local networks. Auto-discover peers with mDNS, sync over SSH or mTLS, and keep every machine in sync.",
    },
  ],
  link: [{ rel: "canonical", href: "https://envsync.arnav.tech/" }],
});
</script>

<template>
  <div class="hero">
    <div class="hero-content">
      <h1 class="font-serif">Sync secrets across<br />your local machines</h1>
      <p>
        Synchronize your .env files without changing your workflow. Keep secrets
        in sync across your devices via mDNS without any centralized server.
      </p>
      <div class="hero-cta">
        <NuxtLink class="btn btn-primary" to="/installation"
          >Get Started</NuxtLink
        >
        <NuxtLink class="btn btn-secondary" to="/usage">Documentation</NuxtLink>
      </div>
    </div>

    <div class="terminal-block">
      <div class="terminal-header">
        <div class="terminal-dot"></div>
        <div class="terminal-dot"></div>
        <div class="terminal-dot"></div>
      </div>
      <pre class="terminal-body">
<span class="term-comment"># 📡 Discovered peer via mDNS</span>
<span class="term-key">env-sync</span> <span class="term-string">sync</span>

<span class="term-comment"># 🔐 Fetching and merging...</span>
<span class="term-key">✔</span> <span class="term-string">OPENAI_API_KEY</span> updated from mbp16.local
<span class="term-key">✔</span> <span class="term-string">DATABASE_URL</span> remains unchanged

<span class="term-comment"># ✅ Successfully synchronized</span>
<span class="term-key">env-sync</span> <span class="term-string">status</span>
Mode: secure-peer (mTLS + AGE encryption)
Last sync: just now
      </pre>
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <h2>A secure peer-to-peer env sync</h2>
    </div>
    <p>
      We all copy <code>.env</code> files around. But maintaining
      synchronization across multiple local laptops, servers, and desktops is a
      pain. <strong>env-sync</strong> uses mDNS (Bonjour/Avahi) to auto-discover
      peers on your local network and syncs <code>.env</code> files
      automatically without touching any external cloud server.
    </p>
  </div>

  <div class="section">
    <div class="section-header">
      <h2>Three security modes for your team.</h2>
    </div>
    <div class="comparisons">
      <NuxtLink to="/modes#trusted-owner-ssh" class="flex-col">
        <div style="width: 100%">
          <h3
            class="font-sans font-bold"
            style="font-size: 1.125rem; margin-bottom: 0.25rem"
          >
            Trusted Owner (Default)
          </h3>
          <p class="text-sm m-0">
            Zero-config SSH transport for your own devices.
          </p>
        </div>
      </NuxtLink>
      <NuxtLink to="/modes#secure-peer" class="flex-col">
        <div style="width: 100%">
          <h3
            class="font-sans font-bold"
            style="font-size: 1.125rem; margin-bottom: 0.25rem"
          >
            Secure Peer
          </h3>
          <p class="text-sm m-0">
            mTLS + AGE encryption for teams and cross-owner collaboration.
          </p>
        </div>
      </NuxtLink>
      <NuxtLink to="/modes#dev-plaintext-http" class="flex-col">
        <div style="width: 100%">
          <h3
            class="font-sans font-bold"
            style="font-size: 1.125rem; margin-bottom: 0.25rem"
          >
            Dev Plaintext
          </h3>
          <p class="text-sm m-0">
            Plain HTTP for immediate local debugging (not recommended for prod).
          </p>
        </div>
      </NuxtLink>
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <h2>Growing rapidly.</h2>
    </div>
    <p>
      env-sync is built for developers who care about operational security and
      offline-ability.
    </p>
    <div class="features-grid">
      <div class="panel">
        <h3 class="font-sans font-bold text-xl mb-2">mDNS Auto-Discovery</h3>
        <p class="text-sm m-0">
          Machines find each other automatically using Avahi or Bonjour. No
          manual IP configuration needed.
        </p>
      </div>
      <div class="panel">
        <h3 class="font-sans font-bold text-xl mb-2">
          Smart merge &amp; backup
        </h3>
        <p class="text-sm m-0">
          Per-key timestamps prevent data loss. Automatic backups before every
          write — keeps last 5 versions.
        </p>
      </div>
      <div class="panel">
        <h3 class="font-sans font-bold text-xl mb-2">Lightweight &amp; fast</h3>
        <p class="text-sm m-0">
          Single Go binary. No runtime dependencies, no background database, no
          Docker required.
        </p>
      </div>
      <div class="panel">
        <h3 class="font-sans font-bold text-xl mb-2">100% Open Source</h3>
        <p class="text-sm m-0">
          The entire stack is completely open-source and ready for auditing.
        </p>
      </div>
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <h2>How does env-sync compare?</h2>
    </div>
    <p>
      env-sync fills the gap that centralized tools leave open — secrets on
      local machines that drift apart.
    </p>
    <div class="comparisons mt-4">
      <NuxtLink to="/comparisons/vault-vs-envsync">
        <div class="flex-row">
          <div class="vs-badge">VS</div>
          <h3 class="font-sans font-bold m-0" style="font-size: 1.125rem">
            HashiCorp Vault
          </h3>
        </div>
        <p class="text-sm m-0 hidden md-block">
          Enterprise secrets platform vs lightweight peer sync.
        </p>
      </NuxtLink>
      <NuxtLink to="/comparisons/infisical-vs-envsync">
        <div class="flex-row">
          <div class="vs-badge">VS</div>
          <h3 class="font-sans font-bold m-0" style="font-size: 1.125rem">
            Infisical
          </h3>
        </div>
        <p class="text-sm m-0 hidden md-block">
          Open-source secrets platform vs decentralized LAN mesh.
        </p>
      </NuxtLink>
      <NuxtLink to="/comparisons/doppler-vs-envsync">
        <div class="flex-row">
          <div class="vs-badge">VS</div>
          <h3 class="font-sans font-bold m-0" style="font-size: 1.125rem">
            Doppler
          </h3>
        </div>
        <p class="text-sm m-0 hidden md-block">
          Cloud secrets operations vs offline-first sync.
        </p>
      </NuxtLink>
      <NuxtLink to="/comparisons/dotenvx-vs-envsync">
        <div class="flex-row">
          <div class="vs-badge">VS</div>
          <h3 class="font-sans font-bold m-0" style="font-size: 1.125rem">
            dotenvx
          </h3>
        </div>
        <p class="text-sm m-0 hidden md-block">
          Encrypted dotenv runtime vs peer synchronization.
        </p>
      </NuxtLink>
      <NuxtLink to="/comparisons/sops-vs-envsync">
        <div class="flex-row">
          <div class="vs-badge">VS</div>
          <h3 class="font-sans font-bold m-0" style="font-size: 1.125rem">
            SOPS
          </h3>
        </div>
        <p class="text-sm m-0 hidden md-block">
          Git-native encrypted files vs always-on machine sync.
        </p>
      </NuxtLink>
    </div>
  </div>

  <div class="section">
    <div class="section-header">
      <h2>Easy to install.</h2>
    </div>
    <div
      class="panel mb-4"
      style="background-color: var(--bg); border-color: var(--border-accent)"
    >
      <code class="text-muted block mb-2"># Install via script</code>
      <code class="text-main font-bold"
        >curl -fsSL https://envsync.arnav.tech/install.sh | bash -s --
        --user</code
      >
    </div>
    <p><NuxtLink to="/installation">See all install options →</NuxtLink></p>
  </div>
</template>

<style scoped>
.text-sm {
  font-size: 0.875rem;
}
.m-0 {
  margin: 0;
}
.mt-4 {
  margin-top: 1rem;
}
.mb-1 {
  margin-bottom: 0.25rem;
}
.mb-2 {
  margin-bottom: 0.5rem;
}
.mb-4 {
  margin-bottom: 1rem;
}
.text-xl {
  font-size: 1.25rem;
}
.block {
  display: block;
}
.flex-col {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}
.flex-row {
  display: flex;
  gap: 1rem;
  align-items: center;
}
.hidden {
  display: none;
}
@media (min-width: 768px) {
  .md-block {
    display: block;
  }
}
</style>
