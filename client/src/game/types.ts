export interface Player {
  id: string
  nickname: string
}

export interface GameEvent {
  type: string
  roomId: string
  playerId?: string
  payload?: unknown
  timestamp: string
}

export interface MovePayload {
  x: number
  y: number
}

export interface CoinCollectedPayload {
  coinId: string
}

export interface PlayerJoinedPayload {
  nickname: string
}

export type PlayerPositions = Record<string, { x: number; y: number }>

export interface Coin {
  id: string
  x: number
  y: number
}

export interface Particle {
  x: number
  y: number
  life: number
  maxLife: number
  vx: number
  vy: number
  color: string
  size: number
}
