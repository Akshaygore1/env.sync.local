<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '@/stores/settings'
import { useToast } from '@/composables/useToast'

const router = useRouter()
const settings = useSettingsStore()
const toast = useToast()

const step = ref(1)
const totalSteps = 4
const selectedMode = ref('trusted-owner-ssh')
const enableEncryption = ref(true)
const initializing = ref(false)
const localKey = ref('')

const modeOptions = [
  {
    value: 'trusted-owner-ssh',
    label: 'Trusted Owner (SSH)',
    desc: 'All devices belong to one user. Uses SSH for transport. Recommended for personal use.',
    icon: '🏠',
  },
  {
    value: 'secure-peer',
    label: 'Secure Peer (mTLS)',
    desc: 'Different owners sharing secrets. Uses mTLS and mandatory AGE encryption.',
    icon: '🔐',
  },
  {
    value: 'dev-plaintext-http',
    label: 'Dev Plaintext (HTTP)',
    desc: 'Debug only. No encryption, no security. Not for production use.',
    icon: '⚠️',
  },
]

function next() {
  if (step.value < totalSteps) {
    step.value++
  }
}

function prev() {
  if (step.value > 1) {
    step.value--
  }
}

async function initialize() {
  initializing.value = true
  try {
    // Set mode first
    await settings.setMode(selectedMode.value, false)

    // Initialize secrets
    await settings.initialize(enableEncryption.value)

    // Try to get the local key
    try {
      // @ts-expect-error Wails bindings
      const key = await window.go.main.KeysService.GetLocalKey()
      localKey.value = key?.publicKey || ''
    } catch {
      // Key might not exist in all modes
    }

    step.value = totalSteps
    toast.success('env-sync initialized successfully!')
  } catch (e) {
    toast.error('Initialization failed: ' + e)
  } finally {
    initializing.value = false
  }
}

function finish() {
  router.push('/')
}

onMounted(async () => {
  await settings.fetchAvailableModes()
})
</script>

<template>
  <div class="wizard-overlay">
    <div class="wizard">
      <!-- Progress -->
      <div class="wizard-progress">
        <div
          v-for="s in totalSteps"
          :key="s"
          class="progress-dot"
          :class="{ active: s === step, done: s < step }"
        ></div>
      </div>

      <!-- Step 1: Welcome -->
      <div v-if="step === 1" class="wizard-step">
        <div class="wizard-icon">🔐</div>
        <h1 class="wizard-title">Welcome to env-sync</h1>
        <p class="wizard-desc">
          Sync your secrets across machines on your local network.
          No cloud, no servers — just peer-to-peer.
        </p>
        <div class="wizard-features">
          <div class="feature">
            <span class="feature-icon">🔄</span>
            <span>Auto-sync across devices</span>
          </div>
          <div class="feature">
            <span class="feature-icon">🔒</span>
            <span>End-to-end encryption</span>
          </div>
          <div class="feature">
            <span class="feature-icon">🌐</span>
            <span>Local network only</span>
          </div>
        </div>
        <button class="btn btn-primary btn-lg" @click="next">Get Started →</button>
      </div>

      <!-- Step 2: Mode Selection -->
      <div v-if="step === 2" class="wizard-step">
        <h2 class="wizard-title">Choose Security Mode</h2>
        <p class="wizard-desc">How will you use env-sync?</p>
        <div class="mode-options">
          <label
            v-for="opt in modeOptions"
            :key="opt.value"
            class="mode-option"
            :class="{ selected: selectedMode === opt.value }"
          >
            <input type="radio" v-model="selectedMode" :value="opt.value" />
            <div class="mode-content">
              <span class="mode-icon">{{ opt.icon }}</span>
              <div>
                <div class="mode-label">{{ opt.label }}</div>
                <div class="mode-desc">{{ opt.desc }}</div>
              </div>
            </div>
          </label>
        </div>
        <div class="wizard-actions">
          <button class="btn btn-secondary" @click="prev">← Back</button>
          <button class="btn btn-primary" @click="next">Next →</button>
        </div>
      </div>

      <!-- Step 3: Initialize -->
      <div v-if="step === 3" class="wizard-step">
        <h2 class="wizard-title">Initialize</h2>
        <p class="wizard-desc">Create your secrets file and encryption keys.</p>

        <div class="init-options">
          <label class="checkbox-label" v-if="selectedMode !== 'dev-plaintext-http'">
            <input type="checkbox" v-model="enableEncryption" />
            <span>Enable AGE encryption {{ selectedMode === 'secure-peer' ? '(required)' : '(recommended)' }}</span>
          </label>

          <div class="init-summary card">
            <div class="info-row">
              <span class="text-muted">Mode</span>
              <span class="badge badge-info">{{ selectedMode }}</span>
            </div>
            <div class="info-row">
              <span class="text-muted">Encryption</span>
              <span>{{ enableEncryption ? 'AGE encrypted' : 'Plaintext' }}</span>
            </div>
          </div>
        </div>

        <div class="wizard-actions">
          <button class="btn btn-secondary" @click="prev">← Back</button>
          <button class="btn btn-primary" @click="initialize" :disabled="initializing">
            <span v-if="initializing" class="spinner"></span>
            Initialize
          </button>
        </div>
      </div>

      <!-- Step 4: Done -->
      <div v-if="step === 4" class="wizard-step">
        <div class="wizard-icon">✅</div>
        <h2 class="wizard-title">All Set!</h2>
        <p class="wizard-desc">env-sync is configured and ready to use.</p>

        <div class="done-summary card" v-if="localKey">
          <div class="info-row">
            <span class="text-muted">Your Public Key</span>
          </div>
          <code class="mono key-display">{{ localKey }}</code>
          <p class="text-muted" style="font-size: 12px; margin-top: 8px">
            Share this key with peers so they can encrypt secrets for you.
          </p>
        </div>

        <button class="btn btn-primary btn-lg" @click="finish">Go to Dashboard →</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wizard-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.wizard {
  max-width: 560px;
  width: 100%;
  padding: 40px;
}

.wizard-progress {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-bottom: 32px;
}

.progress-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--bg-hover);
  transition: all 0.2s ease;
}

.progress-dot.active {
  background: var(--accent-color);
  transform: scale(1.2);
}

.progress-dot.done {
  background: var(--success-color);
}

.wizard-step {
  text-align: center;
}

.wizard-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.wizard-title {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.wizard-desc {
  color: var(--text-secondary);
  margin-bottom: 24px;
  line-height: 1.5;
}

.wizard-features {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 32px;
  text-align: left;
}

.feature {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.feature-icon {
  font-size: 20px;
}

.mode-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 24px;
  text-align: left;
}

.mode-option {
  display: block;
  padding: 16px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.15s ease;
}

.mode-option:hover {
  border-color: var(--accent-color);
}

.mode-option.selected {
  border-color: var(--accent-color);
  background: rgba(99, 102, 241, 0.1);
}

.mode-option input[type="radio"] {
  display: none;
}

.mode-content {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.mode-icon {
  font-size: 24px;
  min-width: 32px;
  text-align: center;
}

.mode-label {
  font-weight: 600;
  margin-bottom: 4px;
}

.mode-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.init-options {
  text-align: left;
  margin-bottom: 24px;
}

.init-summary {
  margin-top: 16px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.wizard-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.done-summary {
  text-align: left;
  margin-bottom: 24px;
}

.key-display {
  display: block;
  font-size: 12px;
  word-break: break-all;
  margin-top: 8px;
  padding: 8px;
  background: var(--bg-primary);
  border-radius: var(--radius-sm);
}

.btn-lg {
  padding: 12px 32px;
  font-size: 16px;
}
</style>
