import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { DiscoveredPeer, PeerInfo, InviteInfo, TrustInfo } from '@/types'

export const usePeersStore = defineStore('peers', () => {
  const discovered = ref<DiscoveredPeer[]>([])
  const registered = ref<PeerInfo[]>([])
  const pending = ref<PeerInfo[]>([])
  const loading = ref(false)
  const discovering = ref(false)

  async function discover(timeout = 5) {
    discovering.value = true
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.DiscoveryService.Discover(timeout)
      discovered.value = result || []
    } catch (e) {
      console.error('Discovery failed:', e)
    } finally {
      discovering.value = false
    }
  }

  async function discoverSSH(timeout = 5) {
    discovering.value = true
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.DiscoveryService.DiscoverSSH(timeout)
      discovered.value = result || []
    } catch (e) {
      console.error('SSH discovery failed:', e)
    } finally {
      discovering.value = false
    }
  }

  async function collectKeys(timeout = 5): Promise<number> {
    // @ts-expect-error Wails bindings
    return await window.go.main.DiscoveryService.CollectKeys(timeout)
  }

  async function fetchPeers() {
    loading.value = true
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.PeerService.ListPeers()
      registered.value = result || []
    } catch (e) {
      console.error('Failed to fetch peers:', e)
    } finally {
      loading.value = false
    }
  }

  async function fetchPending() {
    try {
      // @ts-expect-error Wails bindings
      const result = await window.go.main.PeerService.ListPending()
      pending.value = result || []
    } catch (e) {
      console.error('Failed to fetch pending peers:', e)
    }
  }

  async function approvePeer(peerID: string) {
    // @ts-expect-error Wails bindings
    await window.go.main.PeerService.ApprovePeer(peerID)
    await fetchPeers()
    await fetchPending()
  }

  async function revokePeer(peerID: string) {
    // @ts-expect-error Wails bindings
    await window.go.main.PeerService.RevokePeer(peerID)
    await fetchPeers()
  }

  async function createInvite(expiryHours = 24): Promise<InviteInfo> {
    // @ts-expect-error Wails bindings
    return await window.go.main.PeerService.CreateInvite(expiryHours)
  }

  async function getTrustInfo(): Promise<TrustInfo> {
    // @ts-expect-error Wails bindings
    return await window.go.main.PeerService.GetTrustInfo()
  }

  async function isSecurePeerMode(): Promise<boolean> {
    // @ts-expect-error Wails bindings
    return await window.go.main.PeerService.IsSecurePeerMode()
  }

  return {
    discovered,
    registered,
    pending,
    loading,
    discovering,
    discover,
    discoverSSH,
    collectKeys,
    fetchPeers,
    fetchPending,
    approvePeer,
    revokePeer,
    createInvite,
    getTrustInfo,
    isSecurePeerMode,
  }
})
