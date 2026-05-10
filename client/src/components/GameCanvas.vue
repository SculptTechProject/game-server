<template>
  <canvas ref="canvasRef" class="game-canvas"></canvas>
  <div class="hud">
    <span>🏠 Code: {{ roomId }}</span>
    <span>👥 {{ playerCount }}</span>
    <span>⭐ {{ score }}</span>
    <a class="back-link" href="#" @click.prevent="$emit('leave')">Leave</a>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref, computed } from "vue"
import { GameSocket } from "../game/websocket"
import type { MovePayload, CoinCollectedPayload, PlayerJoinedPayload, PlayerPositions, Coin, Particle } from "../game/types"

const WORLD = 2000
const WALL = 24
const PLAYER_R = 20
const COIN_R = 10
const SEND_INTERVAL = 50
const SPEED = 3.5
const RESPAWN_DELAY = 5000

const COLORS = [
  "#e94560", "#16c79a", "#f5a623", "#7c4dff",
  "#ff6b6b", "#48dbfb", "#ff9ff3", "#54a0ff",
]

const COINS: Coin[] = [
  { id: "c01", x: 300, y: 300 }, { id: "c02", x: 700, y: 200 }, { id: "c03", x: 1100, y: 400 },
  { id: "c04", x: 400, y: 700 }, { id: "c05", x: 900, y: 800 }, { id: "c06", x: 1500, y: 300 },
  { id: "c07", x: 200, y: 1100 }, { id: "c08", x: 600, y: 1300 }, { id: "c09", x: 1000, y: 1100 },
  { id: "c10", x: 1400, y: 900 }, { id: "c11", x: 1700, y: 600 }, { id: "c12", x: 300, y: 1500 },
  { id: "c13", x: 800, y: 1600 }, { id: "c14", x: 1300, y: 1500 }, { id: "c15", x: 1700, y: 1200 },
  { id: "c16", x: 500, y: 500 }, { id: "c17", x: 1200, y: 600 }, { id: "c18", x: 1600, y: 1600 },
  { id: "c19", x: 400, y: 1000 }, { id: "c20", x: 1000, y: 300 }, { id: "c21", x: 700, y: 900 },
  { id: "c22", x: 1400, y: 400 }, { id: "c23", x: 200, y: 600 }, { id: "c24", x: 1100, y: 1300 },
  { id: "c25", x: 600, y: 600 }, { id: "c26", x: 1500, y: 800 }, { id: "c27", x: 900, y: 1400 },
  { id: "c28", x: 300, y: 800 }, { id: "c29", x: 1300, y: 200 }, { id: "c30", x: 1600, y: 1000 },
]

const TREES = [
  { x: 100, y: 100 }, { x: 1900, y: 100 }, { x: 100, y: 1900 }, { x: 1900, y: 1900 },
  { x: 1000, y: 60 }, { x: 60, y: 1000 }, { x: 1940, y: 1000 }, { x: 1000, y: 1940 },
]

export default defineComponent({
  props: { host: String, playerId: String, roomId: String, nickname: String },
  emits: ["leave"],
  setup(props) {
    const canvasRef = ref<HTMLCanvasElement | null>(null)
    const ws = new GameSocket()
    const players = ref<PlayerPositions>({})
    const nicknames = ref<Record<string, string>>({})
    const score = ref(0)
    const playerCount = computed(() => Object.keys(players.value).length)

    const activeCoins = ref<Coin[]>(COINS.map((c) => ({ ...c })))
    const coinsCollected = new Set<string>()
    const particles: Particle[] = []
    const keys: Record<string, boolean> = {}
    const localPos = { x: 400, y: 300 }
    let lastSend = 0
    let animId = 0

    function getColor(id: string) {
      let h = 0
      for (let i = 0; i < id.length; i++) h = id.charCodeAt(i) + ((h << 5) - h)
      return COLORS[Math.abs(h) % COLORS.length]
    }

    function spawnParticles(x: number, y: number, count: number) {
      for (let i = 0; i < count; i++) {
        const angle = Math.random() * Math.PI * 2
        const speed = 0.5 + Math.random() * 1.5
        particles.push({
          x, y,
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          life: 30 + Math.random() * 20,
          maxLife: 50,
          color: "#aadd88",
          size: 2 + Math.random() * 3,
        })
      }
    }

    function drawMap(ctx: CanvasRenderingContext2D) {
      // Grass
      ctx.fillStyle = "#3a7d32"
      ctx.fillRect(0, 0, WORLD, WORLD)

      // Grid
      ctx.strokeStyle = "rgba(0,0,0,0.06)"
      ctx.lineWidth = 1
      for (let x = 0; x <= WORLD; x += 60) {
        ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, WORLD); ctx.stroke()
      }
      for (let y = 0; y <= WORLD; y += 60) {
        ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(WORLD, y); ctx.stroke()
      }

      // Border walls
      ctx.fillStyle = "#5c4033"
      ctx.fillRect(0, 0, WORLD, WALL)
      ctx.fillRect(0, WORLD - WALL, WORLD, WALL)
      ctx.fillRect(0, 0, WALL, WORLD)
      ctx.fillRect(WORLD - WALL, 0, WALL, WORLD)

      // Wall top highlight
      ctx.fillStyle = "#7a5a48"
      ctx.fillRect(0, 0, WORLD, 4)
      ctx.fillRect(0, 0, 4, WORLD)
      ctx.fillRect(0, WORLD - 4, WORLD, 4)
      ctx.fillRect(WORLD - 4, 0, 4, WORLD)

      // Trees
      for (const t of TREES) {
        ctx.beginPath()
        ctx.arc(t.x, t.y, 32, 0, Math.PI * 2)
        ctx.fillStyle = "#2a6e24"
        ctx.fill()
        ctx.beginPath()
        ctx.arc(t.x - 6, t.y - 6, 18, 0, Math.PI * 2)
        ctx.fillStyle = "#24681e"
        ctx.fill()
        ctx.beginPath()
        ctx.arc(t.x + 8, t.y - 4, 14, 0, Math.PI * 2)
        ctx.fillStyle = "#1e5c18"
        ctx.fill()
      }
    }

    function drawCoins(ctx: CanvasRenderingContext2D) {
      for (const coin of activeCoins.value) {
        // Glow
        ctx.beginPath()
        ctx.arc(coin.x, coin.y, COIN_R + 6, 0, Math.PI * 2)
        ctx.fillStyle = "rgba(255,215,0,0.2)"
        ctx.fill()

        // Coin
        ctx.beginPath()
        ctx.arc(coin.x, coin.y, COIN_R, 0, Math.PI * 2)
        ctx.fillStyle = "#ffd700"
        ctx.fill()
        ctx.strokeStyle = "#daa520"
        ctx.lineWidth = 2
        ctx.stroke()

        // Star
        ctx.fillStyle = "#b8860b"
        ctx.font = "12px sans-serif"
        ctx.textAlign = "center"
        ctx.textBaseline = "middle"
        ctx.fillText("★", coin.x, coin.y + 1)
      }
    }

    function drawPlayers(ctx: CanvasRenderingContext2D) {
      for (const [pid, pos] of Object.entries(players.value)) {
        const isLocal = pid === props.playerId
        const color = getColor(pid)
        const nick = nicknames.value[pid] || pid.slice(0, 8)

        // Shadow
        ctx.beginPath()
        ctx.arc(pos.x + 2, pos.y + 2, PLAYER_R + 2, 0, Math.PI * 2)
        ctx.fillStyle = "rgba(0,0,0,0.25)"
        ctx.fill()

        // Glow for local
        if (isLocal) {
          ctx.beginPath()
          ctx.arc(pos.x, pos.y, PLAYER_R + 8, 0, Math.PI * 2)
          ctx.fillStyle = "rgba(255,255,255,0.08)"
          ctx.fill()
        }

        // Main circle
        ctx.beginPath()
        ctx.arc(pos.x, pos.y, PLAYER_R, 0, Math.PI * 2)
        const grad = ctx.createRadialGradient(pos.x - 6, pos.y - 6, 2, pos.x, pos.y, PLAYER_R)
        grad.addColorStop(0, lighten(color, 40))
        grad.addColorStop(1, color)
        ctx.fillStyle = grad
        ctx.fill()
        ctx.strokeStyle = isLocal ? "#fff" : "rgba(255,255,255,0.5)"
        ctx.lineWidth = 3
        ctx.stroke()

        // Eyes
        ctx.fillStyle = "#fff"
        ctx.beginPath()
        ctx.arc(pos.x - 6, pos.y - 4, 4, 0, Math.PI * 2)
        ctx.arc(pos.x + 6, pos.y - 4, 4, 0, Math.PI * 2)
        ctx.fill()
        ctx.fillStyle = "#222"
        ctx.beginPath()
        ctx.arc(pos.x - 5, pos.y - 3, 2, 0, Math.PI * 2)
        ctx.arc(pos.x + 7, pos.y - 3, 2, 0, Math.PI * 2)
        ctx.fill()

        // Nickname
        ctx.fillStyle = "#fff"
        ctx.font = "bold 13px sans-serif"
        ctx.textAlign = "center"
        ctx.textBaseline = "bottom"
        ctx.strokeStyle = "rgba(0,0,0,0.6)"
        ctx.lineWidth = 3
        ctx.strokeText(nick, pos.x, pos.y - PLAYER_R - 6)
        ctx.fillText(nick, pos.x, pos.y - PLAYER_R - 6)
      }
    }

    function drawParticles(ctx: CanvasRenderingContext2D) {
      for (const p of particles) {
        const alpha = p.life / p.maxLife
        ctx.globalAlpha = alpha
        ctx.beginPath()
        ctx.arc(p.x, p.y, p.size * alpha, 0, Math.PI * 2)
        ctx.fillStyle = p.color
        ctx.fill()
      }
      ctx.globalAlpha = 1
    }

    function drawMinimap(ctx: CanvasRenderingContext2D, w: number, h: number) {
      const mmSize = 140
      const mmX = w - mmSize - 14
      const mmY = 56
      const scale = mmSize / WORLD

      ctx.fillStyle = "rgba(0,0,0,0.55)"
      ctx.strokeStyle = "rgba(255,255,255,0.2)"
      ctx.lineWidth = 1
      ctx.fillRect(mmX, mmY, mmSize, mmSize)
      ctx.strokeRect(mmX, mmY, mmSize, mmSize)

      for (const coin of activeCoins.value) {
        ctx.fillStyle = "#ffd700"
        ctx.fillRect(mmX + coin.x * scale - 1, mmY + coin.y * scale - 1, 3, 3)
      }

      for (const [pid, pos] of Object.entries(players.value)) {
        const isLocal = pid === props.playerId
        ctx.beginPath()
        ctx.arc(mmX + pos.x * scale, mmY + pos.y * scale, isLocal ? 4 : 3, 0, Math.PI * 2)
        ctx.fillStyle = isLocal ? "#fff" : getColor(pid)
        ctx.fill()
        if (isLocal) {
          ctx.strokeStyle = "rgba(255,255,255,0.5)"
          ctx.lineWidth = 1
          ctx.stroke()
        }
      }
    }

    function lighten(hex: string, amt: number) {
      let r = parseInt(hex.slice(1, 3), 16)
      let g = parseInt(hex.slice(3, 5), 16)
      let b = parseInt(hex.slice(5, 7), 16)
      r = Math.min(255, r + amt)
      g = Math.min(255, g + amt)
      b = Math.min(255, b + amt)
      return `rgb(${r},${g},${b})`
    }

    function gameLoop() {
      const canvas = canvasRef.value
      if (!canvas) { animId = requestAnimationFrame(gameLoop); return }
      const ctx = canvas.getContext("2d")
      if (!ctx) { animId = requestAnimationFrame(gameLoop); return }

      const w = canvas.width
      const h = canvas.height

      // Movement
      let dx = 0, dy = 0
      if (keys["w"] || keys["arrowup"]) dy = -SPEED
      if (keys["s"] || keys["arrowdown"]) dy = SPEED
      if (keys["a"] || keys["arrowleft"]) dx = -SPEED
      if (keys["d"] || keys["arrowright"]) dx = SPEED

      let moved = false
      if (dx || dy) {
        if (dx && dy) { dx *= 0.707; dy *= 0.707 }
        const nx = Math.max(WALL + PLAYER_R, Math.min(WORLD - WALL - PLAYER_R, localPos.x + dx))
        const ny = Math.max(WALL + PLAYER_R, Math.min(WORLD - WALL - PLAYER_R, localPos.y + dy))
        if (nx !== localPos.x || ny !== localPos.y) moved = true
        localPos.x = nx
        localPos.y = ny

        players.value = { ...players.value, [props.playerId!]: { x: localPos.x, y: localPos.y } }

        const now = Date.now()
        if (now - lastSend > SEND_INTERVAL) {
          lastSend = now
          ws.sendMove(localPos.x, localPos.y)
        }
      }

      if (moved) spawnParticles(localPos.x, localPos.y, 1)

      // Coin collection
      for (const coin of activeCoins.value) {
        const d = Math.hypot(localPos.x - coin.x, localPos.y - coin.y)
        if (d < PLAYER_R + COIN_R + 4) {
          if (!coinsCollected.has(coin.id)) {
            coinsCollected.add(coin.id)
            ws.sendCollect(coin.id)
            score.value += 10
            spawnParticles(coin.x, coin.y, 8)
          }
        }
      }

      // Update particles
      for (let i = particles.length - 1; i >= 0; i--) {
        const p = particles[i]
        p.x += p.vx
        p.y += p.vy
        p.life--
        if (p.life <= 0) particles.splice(i, 1)
      }

      // Camera
      const camX = localPos.x - w / 2
      const camY = localPos.y - h / 2

      ctx.clearRect(0, 0, w, h)
      ctx.save()
      ctx.translate(-camX, -camY)

      drawMap(ctx)
      drawCoins(ctx)
      drawPlayers(ctx)
      drawParticles(ctx)

      ctx.restore()
      drawMinimap(ctx, w, h)

      animId = requestAnimationFrame(gameLoop)
    }

    function onKey(e: KeyboardEvent, down: boolean) {
      keys[e.key.toLowerCase()] = down
    }

    function resize() {
      const c = canvasRef.value
      if (!c) return
      c.width = window.innerWidth
      c.height = window.innerHeight
    }

    onMounted(() => {
      resize()
      window.addEventListener("resize", resize)
      window.addEventListener("keydown", (e) => onKey(e, true))
      window.addEventListener("keyup", (e) => onKey(e, false))

      players.value = { [props.playerId!]: { x: localPos.x, y: localPos.y } }
      nicknames.value = { [props.playerId!]: props.nickname! }

      ws.on("player_moved", (ev) => {
        const pl = ev.payload as MovePayload
        if (ev.playerId !== props.playerId) {
          spawnParticles(pl.x, pl.y, 1)
        }
        players.value = { ...players.value, [ev.playerId!]: { x: pl.x, y: pl.y } }
      })

      ws.on("player_joined", (ev) => {
        const pl = ev.payload as PlayerJoinedPayload
        players.value = { ...players.value, [ev.playerId!]: { x: 400, y: 300 } }
        nicknames.value = { ...nicknames.value, [ev.playerId!]: pl?.nickname || ev.playerId!.slice(0, 8) }
      })

      ws.on("player_left", (ev) => {
        const copy = { ...players.value }
        delete copy[ev.playerId!]
        players.value = copy
      })

      ws.on("room_state", (ev) => {
        const pl = ev.payload as any
        if (pl?.positions) {
          for (const [pid, pos] of Object.entries(pl.positions)) {
            players.value = { ...players.value, [pid]: pos }
          }
        }
        if (pl?.nicknames) {
          nicknames.value = { ...nicknames.value, ...pl.nicknames }
        }
      })

      ws.on("coin_collected", (ev) => {
        const pl = ev.payload as CoinCollectedPayload
        activeCoins.value = activeCoins.value.filter((c) => c.id !== pl.coinId)
        setTimeout(() => {
          const orig = COINS.find((c) => c.id === pl.coinId)
          if (orig) {
            coinsCollected.delete(pl.coinId)
            activeCoins.value = [...activeCoins.value, { ...orig }]
          }
        }, RESPAWN_DELAY)
      })

      ws.connect(props.host!, props.roomId!, props.playerId!, props.nickname!)
      animId = requestAnimationFrame(gameLoop)
    })

    onUnmounted(() => {
      cancelAnimationFrame(animId)
      ws.disconnect()
      window.removeEventListener("resize", resize)
    })

    return { canvasRef, roomId: props.roomId, playerCount, score }
  },
})
</script>

<style scoped>
.game-canvas {
  display: block;
  width: 100%;
  height: 100%;
}
.hud {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  display: flex;
  gap: 20px;
  align-items: center;
  padding: 10px 16px;
  background: rgba(0, 0, 0, 0.55);
  font-size: 14px;
  z-index: 10;
  backdrop-filter: blur(6px);
}
.back-link {
  color: #e94560;
  text-decoration: none;
  margin-left: auto;
  font-weight: 600;
}
.back-link:hover {
  text-decoration: underline;
}
</style>
