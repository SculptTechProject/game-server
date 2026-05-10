import type { GameEvent } from "./types"

export type EventHandler = (event: GameEvent) => void

export class GameSocket {
  private ws: WebSocket | null = null
  private handlers = new Map<string, EventHandler>()

  connect(host: string, roomId: string, playerId: string, nickname: string) {
    const url = `ws://${host}/ws?roomId=${roomId}&playerId=${playerId}&nickname=${encodeURIComponent(nickname)}`
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log("WS connected")
    }

    this.ws.onmessage = (e) => {
      try {
        const event: GameEvent = JSON.parse(e.data)
        const handler = this.handlers.get(event.type)
        if (handler) handler(event)
      } catch (err) {
        console.error("WS parse error:", err)
      }
    }

    this.ws.onclose = () => {
      console.log("WS disconnected")
    }

    this.ws.onerror = (err) => {
      console.error("WS error:", err)
    }
  }

  disconnect() {
    this.ws?.close()
    this.ws = null
  }

  sendMove(x: number, y: number) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: "move", x, y }))
    }
  }

  sendCollect(coinId: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: "collect", coinId }))
    }
  }

  on(type: string, handler: EventHandler) {
    this.handlers.set(type, handler)
  }
}
