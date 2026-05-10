<template>
  <Lobby v-if="screen === 'lobby'" @start="onStart" />
  <GameCanvas
    v-else
    :host="gameInfo.host"
    :player-id="gameInfo.playerId"
    :room-id="gameInfo.roomId"
    :nickname="gameInfo.nickname"
    @leave="screen = 'lobby'"
  />
</template>

<script lang="ts">
import { defineComponent, ref } from "vue"
import Lobby from "./components/Lobby.vue"
import GameCanvas from "./components/GameCanvas.vue"
import "./style.css"

interface GameInfo {
  host: string
  playerId: string
  roomId: string
  nickname: string
}

export default defineComponent({
  components: { Lobby, GameCanvas },
  setup() {
    const screen = ref<"lobby" | "game">("lobby")
    const gameInfo = ref<GameInfo>({ host: "", playerId: "", roomId: "", nickname: "" })

    function onStart(info: GameInfo) {
      gameInfo.value = info
      screen.value = "game"
    }

    return { screen, gameInfo, onStart }
  },
})
</script>
