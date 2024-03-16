<script setup lang="ts">
import { Service } from "@/api/generated";
import api from "@/api";
import { onBeforeUnmount, ref, watch } from "vue";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";

const props = defineProps<{
  service: Service;
}>();

const loaded = ref(false);

const replica = ref<number | null>(null);

let ws: WebSocket | null = null;

let unmount = false;

let onResize: (() => void) | null = null;

const pingHandler = setInterval(() => ws?.send("ping"), 10_000);

onBeforeUnmount(() => {
  clearInterval(pingHandler);

  term?.dispose();

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

  if (replica.value != null && replica.value >= props.service.replicas) {
    replica.value = null;
  }

  const token = await api.TokenApi.createTempToken()
    .then((r) => r.data)
    .then((data) => data.token);

  ws = api.ServiceLogsApi.connectToLogStream(
    props.service.id,
    token,
    replica.value != null ? replica.value : undefined,
  );

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
      foreground: "rgb(33,37,41)",
      selectionBackground: "rgb(150, 150, 170)",
    },
    allowTransparency: true,
    scrollSensitivity: 8,
  });

  const fitAddon = new FitAddon();
  term.loadAddon(fitAddon);

  term.open(termElement);

  hideWidthCacheDiv();

  if (term.textarea) term.textarea.readOnly = true;

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

// workaround of issues that widthCache causes by creating very big element
// which causes browser to render page with incorrect size, adding unnecessary scroll
function hideWidthCacheDiv() {
  const core = (term as any)._core;
  const renderService = core?._renderService;
  const renderer = renderService?._renderer?._value;
  const widthCache = renderer?._widthCache;
  const container = widthCache?._container as HTMLElement | null;
  if (container) {
    // @ts-ignore
    container.style.contentVisibility = "hidden";
  }
}

function reconnect() {
  ws?.close();
}

function replicaSelectOptions() {
  return [...Array(props.service.replicas).keys()].map((idx) => ({
    value: idx,
    text: `Replica ${idx}`,
  }));
}

load().catch(() =>
  setTimeout(function () {
    load();
  }, 1000),
);

watch(replica, () => reconnect());
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

    <b-row class="my-2">
      <b-col style="max-width: 10rem">
        <b-form-select
          v-model="replica"
          :options="replicaSelectOptions()"
          size="sm"
        />
      </b-col>
      <b-col>
        <b-button @click="reconnect" variant="outline-danger" size="sm">
          Reconnect
        </b-button>
      </b-col>
    </b-row>
  </b-container>
</template>

<style>
@import "xterm/css/xterm.css";
</style>
