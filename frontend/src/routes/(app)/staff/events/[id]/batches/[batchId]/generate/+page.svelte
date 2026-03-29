<script lang="ts">
  import { page } from "$app/stores";
  import { onMount, onDestroy } from "svelte";
  import { api, ApiError } from "$lib/api/client";
  import StatusBadge from "$lib/components/StatusBadge.svelte";

  const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

  interface Batch {
    id: number;
    event_id: number;
    status: string;
    rows_total: number;
    rows_ok: number;
    rows_failed: number;
    mapping: Record<string, string>;
    tokens: string[];
  }

  interface ProgressEvent {
    row_id: number;
    row_name: string;
    row_iin: string;
    status: string;
    error?: string;
    progress: number;
    total: number;
    rows_ok: number;
    rows_failed: number;
  }

  interface LogEntry {
    rowId: number;
    name: string;
    iin: string;
    status: string;
    error?: string;
    time: string;
  }

  let eventId = $derived($page.params.id);
  let batchId = $derived($page.params.batchId);

  let batch = $state<Batch | null>(null);
  let loading = $state(true);
  let generating = $state(false);
  let error = $state("");
  let progress = $state(0);
  let total = $state(0);
  let rowsOk = $state(0);
  let rowsFailed = $state(0);
  let logs = $state<LogEntry[]>([]);
  let completed = $state(false);
  let eventSource: EventSource | null = null;

  function maskIIN(iin: string): string {
    if (!iin || iin.length < 8) return iin || "";
    return iin.slice(0, 4) + "****" + iin.slice(8);
  }

  function percentComplete(): number {
    if (total === 0) return 0;
    return Math.round((progress / total) * 100);
  }

  async function loadBatch() {
    loading = true;
    try {
      const res = await api.get<Batch>(`/api/v1/staff/batches/${batchId}`);
      batch = res.data;
      total = batch.rows_total;
      rowsOk = batch.rows_ok;
      rowsFailed = batch.rows_failed;
      progress = rowsOk + rowsFailed;

      if (batch.status === "generating") {
        generating = true;
        connectSSE();
      } else if (batch.status === "done" || batch.status === "done_with_errors" || batch.status === "failed") {
        completed = true;
        progress = total;
      }
    } catch {
      error = "Failed to load batch";
    } finally {
      loading = false;
    }
  }

  async function startGeneration() {
    generating = true;
    error = "";
    logs = [];
    progress = 0;
    rowsOk = 0;
    rowsFailed = 0;
    completed = false;

    try {
      await api.post(`/api/v1/staff/batches/${batchId}/generate`);
      connectSSE();
    } catch (err) {
      error = err instanceof ApiError ? err.message : "Failed to start generation";
      generating = false;
    }
  }

  function connectSSE() {
    if (eventSource) {
      eventSource.close();
    }

    const url = `${API_BASE}/api/v1/staff/batches/${batchId}/progress`;
    eventSource = new EventSource(url, { withCredentials: true });

    eventSource.onmessage = (event) => {
      try {
        const data: ProgressEvent = JSON.parse(event.data);
        progress = data.progress;
        total = data.total;
        rowsOk = data.rows_ok;
        rowsFailed = data.rows_failed;

        if (data.status === "complete") {
          completed = true;
          generating = false;
          eventSource?.close();
          eventSource = null;
          loadBatch();
          return;
        }

        if (data.row_id) {
          const entry: LogEntry = {
            rowId: data.row_id,
            name: data.row_name || "",
            iin: data.row_iin || "",
            status: data.status,
            error: data.error,
            time: new Date().toLocaleTimeString(),
          };
          logs = [...logs, entry];
        }
      } catch {
        // ignore parse errors
      }
    };

    eventSource.onerror = () => {
      if (!completed) {
        // Connection lost, try to reload batch status
        setTimeout(() => loadBatch(), 2000);
      }
      eventSource?.close();
      eventSource = null;
    };
  }

  onMount(loadBatch);

  onDestroy(() => {
    eventSource?.close();
  });
</script>

{#if loading}
  <div class="text-center py-12 text-on-surface-variant">Loading...</div>
{:else if batch}
  <div class="space-y-6 max-w-4xl">
    <!-- Header -->
    <div class="flex items-start justify-between">
      <div>
        <a
          href="/staff/events/{eventId}/batches/{batchId}"
          class="text-sm text-on-surface-variant hover:text-primary transition-colors"
        >
          &larr; Back to mapping / batch
        </a>
        <h1 class="font-display text-2xl font-bold text-on-surface mt-2">
          Generation: Batch #{batch.id}
        </h1>
        <p class="text-sm text-on-surface-variant mt-1">
          {batch.rows_total} participants
        </p>
      </div>
      <div class="flex gap-2">
        {#if !generating && !completed}
          <button
            onclick={startGeneration}
            class="px-5 py-2.5 rounded-lg text-sm font-medium
                   bg-gradient-to-br from-primary to-primary-container text-on-primary
                   hover:shadow-lg transition-all"
          >
            Generate
          </button>
        {/if}
        {#if completed}
          <a
            href="/staff/events/{eventId}/certificates"
            class="px-5 py-2.5 rounded-lg text-sm font-medium
                   bg-gradient-to-br from-primary to-primary-container text-on-primary
                   hover:shadow-lg transition-all"
          >
            View Certificates
          </a>
        {/if}
      </div>
    </div>

    {#if error}
      <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm">{error}</div>
    {/if}

    <!-- Stats Cards -->
    <div class="grid grid-cols-3 gap-4">
      <div class="bg-surface-lowest rounded-lg p-5">
        <p class="text-sm text-on-surface-variant">Total</p>
        <p class="text-3xl font-bold text-on-surface mt-1 font-display">{total.toLocaleString()}</p>
      </div>
      <div class="bg-surface-lowest rounded-lg p-5">
        <p class="text-sm text-emerald-600">OK</p>
        <p class="text-3xl font-bold text-emerald-700 mt-1 font-display">{rowsOk.toLocaleString()}</p>
      </div>
      <div class="bg-surface-lowest rounded-lg p-5">
        <p class="text-sm text-red-600">Failed</p>
        <p class="text-3xl font-bold text-red-700 mt-1 font-display">{rowsFailed.toLocaleString()}</p>
      </div>
    </div>

    <!-- Progress Section -->
    <div class="bg-surface-lowest rounded-lg p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="font-display text-lg font-semibold text-on-surface">Progress</h2>
        <span class="text-2xl font-bold text-primary font-display">{percentComplete()}%</span>
      </div>

      <div class="w-full bg-surface-high rounded-full h-3 overflow-hidden">
        <div
          class="h-full rounded-full transition-all duration-300 ease-out
                 {completed
                    ? rowsFailed > 0
                      ? 'bg-gradient-to-r from-amber-400 to-amber-500'
                      : 'bg-gradient-to-r from-emerald-400 to-emerald-500'
                    : 'bg-gradient-to-r from-primary to-primary-container'}"
          style="width: {percentComplete()}%"
        ></div>
      </div>

      <p class="text-sm text-on-surface-variant">
        {#if generating && !completed}
          Processing certificates... {progress} of {total}
        {:else if completed}
          Generation complete. {rowsOk} succeeded, {rowsFailed} failed.
        {:else}
          Ready to generate {total} certificates.
        {/if}
      </p>
    </div>

    <!-- Log Table -->
    {#if logs.length > 0 || completed}
      <div class="bg-surface-lowest rounded-lg p-6 space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="font-display text-lg font-semibold text-on-surface">Generation Log</h2>
          <span class="text-sm text-on-surface-variant">{logs.length} entries</span>
        </div>

        <div class="overflow-auto max-h-[400px]">
          <table class="w-full text-sm">
            <thead class="sticky top-0 bg-surface-lowest">
              <tr class="text-left text-on-surface-variant">
                <th class="pb-2 pr-4 font-medium">ID</th>
                <th class="pb-2 pr-4 font-medium">Participant / IIN</th>
                <th class="pb-2 pr-4 font-medium">Status</th>
                <th class="pb-2 font-medium">Time</th>
              </tr>
            </thead>
            <tbody>
              {#each logs as entry (entry.rowId)}
                <tr class="border-t border-surface-high/50">
                  <td class="py-2.5 pr-4 text-on-surface-variant font-mono text-xs">
                    CERT-{entry.rowId}
                  </td>
                  <td class="py-2.5 pr-4">
                    <div class="text-on-surface font-medium">{entry.name}</div>
                    {#if entry.iin}
                      <div class="text-xs text-on-surface-variant">{maskIIN(entry.iin)}</div>
                    {/if}
                  </td>
                  <td class="py-2.5 pr-4">
                    <StatusBadge status={entry.status} />
                    {#if entry.error}
                      <p class="text-xs text-red-600 mt-0.5 max-w-[200px] truncate" title={entry.error}>
                        {entry.error}
                      </p>
                    {/if}
                  </td>
                  <td class="py-2.5 text-on-surface-variant text-xs">{entry.time}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div class="text-center py-12 text-error">Batch not found</div>
{/if}
