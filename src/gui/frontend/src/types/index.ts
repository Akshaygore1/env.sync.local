export interface SecretEntry {
  key: string
  value: string
  updatedAt: string
}

export interface SyncResult {
  success: boolean
  message: string
  source: string
}

export interface DiscoveredPeer {
  hostname: string
  address: string
  port: number
  sshAccess: boolean
  hasPubKey: boolean
  reachable: boolean
}

export interface KeyInfo {
  hostname: string
  publicKey: string
  fingerprint: string
  isLocal: boolean
}

export interface AccessRequest {
  hostname: string
  publicKey: string
  timestamp: string
}

export interface PeerInfo {
  id: string
  hostname: string
  state: string
  tlsFingerprint: string
  agePubKey: string
  addedAt: string
  updatedAt: string
}

export interface InviteInfo {
  token: string
  createdBy: string
  fingerprint: string
  expiresAt: string
  command: string
}

export interface TrustInfo {
  localHostname: string
  tlsFingerprint: string
  certValidUntil: string
  agePublicKey: string
  trustedPeers: PeerInfo[]
}

export interface ModeInfo {
  current: string
  description: string
  features: string[]
  transport: string
  encryption: string
}

export interface FileStatus {
  exists: boolean
  path: string
  version: string
  timestamp: string
  host: string
  encrypted: boolean
  modTime: string
}

export interface ServerStatus {
  running: boolean
  port: string
  pid: number
}

export interface PeerStatus {
  hostname: string
  sshAccess: boolean
  hasPubKey: boolean
}

export interface BackupEntry {
  number: number
  timestamp: string
  size: number
  path: string
}

export interface StatusInfo {
  secretsFile: FileStatus
  server: ServerStatus
  peers: PeerStatus[]
  backups: BackupEntry[]
  mode: ModeInfo
}

export interface CronInfo {
  installed: boolean
  interval: number
}

export interface ConfigPaths {
  configDir: string
  secretsFile: string
  backupDir: string
  logDir: string
  keysDir: string
}

export interface LogEntry {
  timestamp: string
  level: string
  message: string
}

export type ToastType = 'success' | 'error' | 'info'

export interface Toast {
  id: number
  type: ToastType
  message: string
}
