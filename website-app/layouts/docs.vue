<script setup lang="ts"></script>

<template>
  <div class="docs-layout">
    <!-- Header -->
    <LayoutSiteHeader />

    <!-- Main docs layout -->
    <div class="docs-container">
      <!-- Sidebar -->
      <DocsSidebar />

      <!-- Content -->
      <main class="docs-content">
        <slot />
      </main>
    </div>

    <!-- Footer -->
    <LayoutSiteFooter />
  </div>
</template>

<style>
.docs-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.docs-container {
  flex-grow: 1;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

@media (min-width: 768px) {
  .docs-container {
    flex-direction: row;
    gap: 3rem;
  }
}

.docs-content {
  flex: 1;
  min-width: 0;
  padding-bottom: 4rem;
}

/* 
  Styles to make markdown-like content look good inside the docs template.
*/
.docs-content h1 {
  font-family: var(--font-serif);
  font-size: 2.5rem;
  color: var(--text-main);
  margin-top: 0;
  margin-bottom: 1.5rem;
}

.docs-content h2 {
  font-family: var(--font-serif);
  font-size: 2rem;
  color: var(--text-main);
  /* margin-top: 2.5rem;
  margin-bottom: 1.25rem; */
  border-bottom: 1px solid var(--border);
  padding-bottom: 1.25rem;
}

.docs-content h3 {
  font-family: var(--font-serif);
  font-size: 1.5rem;
  color: var(--text-main);
  margin-top: 2rem;
  margin-bottom: 1rem;
}

.docs-content p {
  margin-bottom: 1.25rem;
  line-height: 1.7;
}

.docs-content ul,
.docs-content ol {
  margin-bottom: 1.25rem;
  padding-left: 1.5rem;
  color: var(--text-muted);
}

.docs-content li {
  margin-bottom: 0.5rem;
  line-height: 1.6;
}

.docs-content pre {
  background: #000;
  padding: 1.25rem;
  border-radius: 12px;
  border: 1px solid var(--border-accent);
  overflow-x: auto;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  line-height: 1.5;
}

.docs-content code {
  background: var(--accent-light);
  color: #60a5fa;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  border: 1px solid rgba(59, 130, 246, 0.2);
  font-size: 0.85em;
  font-family: var(--font-mono);
}

.docs-content pre code {
  background: transparent;
  color: inherit;
  padding: 0;
  border: none;
}

.docs-content a {
  color: var(--accent);
  text-decoration: none;
}

.docs-content a:hover {
  text-decoration: underline;
  color: var(--accent-hover);
}

.docs-content blockquote {
  border-left: 4px solid var(--accent);
  padding-left: 1rem;
  color: var(--text-muted);
  background: var(--accent-light);
  padding: 0.75rem 1rem;
  border-radius: 0 8px 8px 0;
  margin-bottom: 1.5rem;
}
</style>
