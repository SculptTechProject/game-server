<template>
  <div class="lobby">
    <div class="card">
      <h1>🎮 Game Server</h1>

      <div class="field">
        <label>Server</label>
        <input v-model="host" placeholder="localhost:8080" />
      </div>

      <div class="field">
        <label>Nickname</label>
        <input v-model="nickname" placeholder="Your nickname" maxlength="20" />
      </div>

      <div class="field">
        <label>Room Code <span class="hint">(leave empty to create new)</span></label>
        <input v-model="roomId" placeholder="e.g. 7429" maxlength="4" />
      </div>

      <div v-if="createdCode" class="code-badge">
        Your room code: <strong>{{ createdCode }}</strong>
        <span class="share-hint">Share it with friends!</span>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <button :disabled="loading || !nickname" @click="play">
        {{ loading ? "Connecting..." : "Play" }}
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue"

export default defineComponent({
  emits: ["start"],
  setup(_props, { emit }) {
    const host = ref("localhost:8080")
    const nickname = ref("")
    const roomId = ref("")
    const loading = ref(false)
    const error = ref("")
    const createdCode = ref("")

    async function apiPost(path: string, body: object) {
      const res = await fetch(`http://${host.value}${path}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      })
      const data = await res.json()
      if (!data.success) throw new Error(data.error?.message || "API error")
      return data.data
    }

    async function play() {
      loading.value = true
      error.value = ""

      try {
        const player = await apiPost("/create-player", { nickname: nickname.value.trim() })
        let rid = roomId.value.trim()

        if (!rid) {
          const room = await apiPost("/create-room", { name: nickname.value + "'s room", "max-players": 10 })
          rid = room.id
          createdCode.value = room.id
        }

        await apiPost("/join-room", { roomId: rid, playerId: player.id })

        emit("start", {
          host: host.value,
          playerId: player.id,
          roomId: rid,
          nickname: nickname.value.trim(),
        })
      } catch (err: any) {
        error.value = err.message || "Failed to connect"
      } finally {
        loading.value = false
      }
    }

    return { host, nickname, roomId, loading, error, createdCode, play }
  },
})
</script>

<style scoped>
.lobby {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card {
  background: #16213e;
  padding: 40px;
  border-radius: 16px;
  width: 380px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

h1 {
  text-align: center;
  margin-bottom: 8px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

label {
  font-size: 13px;
  color: #888;
}

.hint {
  color: #555;
  font-weight: 400;
}

input {
  background: #0f3460;
  border: 1px solid #1a4a8a;
  color: #eee;
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 15px;
  outline: none;
  transition: border-color 0.2s;
}

input:focus {
  border-color: #e94560;
}

button {
  background: #e94560;
  color: #fff;
  border: none;
  padding: 12px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  margin-top: 8px;
}

button:hover:not(:disabled) {
  background: #d63851;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error {
  color: #e94560;
  font-size: 13px;
  text-align: center;
}

.code-badge {
  background: #1a4a8a;
  padding: 10px 14px;
  border-radius: 8px;
  text-align: center;
  font-size: 18px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.code-badge strong {
  font-size: 28px;
  letter-spacing: 4px;
  color: #ffd700;
}

.share-hint {
  font-size: 11px;
  color: #888;
}
</style>
