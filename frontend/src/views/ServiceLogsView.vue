<script setup lang="ts">
import { Service } from "@/api/generated";
import api from "@/api";
import { onBeforeUnmount, ref } from "vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";

const props = defineProps<{
  service: Service;
}>();

const loaded = ref(false);

let ws: WebSocket | null = null;

let unmount = false;

let onResize: (() => void) | null = null;

const pingHandler = setInterval(() => ws?.send("ping"), 10_000);

onBeforeUnmount(() => {
  clearInterval(pingHandler);

  if (ws == null) return;
  unmount = true;
  try {
    ws.close();
  } catch (e) {
    console.log("Failed to close WebSocket");
  }
});

async function load() {
  loaded.value = false;

  const token = await api.TokenApi.createTempToken()
    .then((r) => r.data)
    .then((data) => data.token);

  ws = api.ServiceLogsApi.connectToLogStream(props.service.id, token);

  ws.onopen = () => {
    createTerm();
  };
  ws.onmessage = (message) => {
    const line = message.data as string;
    if (line.endsWith("\n")) term?.writeln(line.substring(0, line.length - 1));
    else term?.write(line);
  };
  ws.onclose = () => {
    if (unmount) return;
    setTimeout(function () {
      load();
    }, 1000);
  };

  loaded.value = true;
}

let term: Terminal | null;

function createTerm() {
  if (ws == null) {
    console.log("WebSocket is null"); // TODO: throw exception
    return;
  }

  const termElement: HTMLElement | null = document.getElementById("terminal");
  if (termElement == null) {
    console.log("Element with id 'terminal' does not exist");
    return;
  }

  if (term != null) {
    term.dispose();
  }
  term = new Terminal({
    theme: {
      background: "rgba(240, 240, 245, 0.8)",
      foreground: "#212529",
    },
    allowTransparency: true,
  });

  const fitAddon = new FitAddon();
  term.loadAddon(fitAddon);

  term.open(termElement);
  term.focus();

  fitAddon.fit();

  if (onResize != null) {
    window.removeEventListener("resize", onResize);
  }
  onResize = function () {
    if (term == null) {
      return;
    }
    fitAddon.fit();
  };
  window.addEventListener("resize", onResize);
  onResize();
}

function reconnect() {
  ws?.close();
}

await load().catch(() =>
  setTimeout(function () {
    load();
  }, 1000)
);
</script>

<template>
  <b-container>
    <b-overlay :show="!loaded">
      <template #overlay>
        <div class="text-center">
          <p>Please wait for the session to be created</p>
          <b-spinner />
        </div>
      </template>

      <div id="terminal" />
    </b-overlay>

    <b-button @click="reconnect" class="mt-1" variant="outline-danger">
      Reconnect
    </b-button>
  </b-container>
</template>

<style>
@import "xterm/css/xterm.css";
</style>
