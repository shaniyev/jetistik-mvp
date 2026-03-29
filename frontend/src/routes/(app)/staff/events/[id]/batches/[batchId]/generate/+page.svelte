<script lang="ts">
  import { page } from "$app/stores";
  import { onMount, onDestroy } from "svelte";
  import { api, ApiError } from "$lib/api/client";

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

  function logStatusClasses(status: string): string {
    if (status === "ok" || status === "valid") {
      return "bg-primary-fixed text-on-primary-fixed-variant";
    }
    if (status === "failed" || status === "error") {
      return "bg-error-container text-on-error-container";
    }
    return "bg-surface-container-highest text-on-surface-variant";
  }

  onMount(loadBatch);

  onDestroy(() => {
    eventSource?.close();
  });
</script>

{#if loading}
  <div class="flex items-center justify-center py-24 text-on-surface-variant">
    <span class="material-symbols-outlined animate-spin mr-2">progress_activity</span>
    Loading...
  </div>
{:else if batch}
  <!-- Header Section -->
  <header class="sticky top-0 z-30 bg-white/80 backdrop-blur-xl border-b border-outline-variant/10 px-8 py-6">
    <div class="max-w-6xl mx-auto flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div>
        <a
          href="/staff/events/{eventId}/batches/{batchId}"
          class="inline-flex items-center text-primary hover:text-surface-tint transition-colors mb-2 text-sm font-medium group"
        >
          <span class="material-symbols-outlined text-sm mr-1 group-hover:-translate-x-1 transition-transform">arrow_back</span>
          Вернуться к сопоставлению / Карталауга оралу
        </a>
        <h1 class="font-display text-3xl font-extrabold tracking-tight text-on-surface">
          Генерация: Батч #{batch.id}
        </h1>
        <p class="text-sm text-on-surface-variant mt-1">
          Генерациялау: №{batch.id} топтама
        </p>
      </div>
      <div class="flex items-center gap-3">
        <button
          onclick={() => loadBatch()}
          class="px-5 py-2.5 rounded-xl border border-outline-variant text-on-surface font-semibold text-sm flex items-center gap-2 hover:bg-surface-container-low transition-all active:scale-95"
        >
          <span class="material-symbols-outlined text-[20px]">refresh</span>
          Обновить / Жанарту
        </button>
        {#if !generating && !completed}
          <button
            onclick={startGeneration}
            class="px-6 py-2.5 rounded-xl bg-gradient-to-br from-primary to-primary-container text-white font-semibold text-sm flex items-center gap-2 shadow-lg shadow-primary/20 active:scale-95 transition-transform"
          >
            <span class="material-symbols-outlined text-[20px]">play_circle</span>
            Запустить генерацию / Жіберу
          </button>
        {:else if completed}
          <a
            href="/staff/events/{eventId}/certificates"
            class="px-6 py-2.5 rounded-xl bg-gradient-to-br from-primary to-primary-container text-white font-semibold text-sm flex items-center gap-2 shadow-lg shadow-primary/20 active:scale-95 transition-transform"
          >
            <span class="material-symbols-outlined text-[20px]">verified</span>
            Просмотреть сертификаты
          </a>
        {:else}
          <button
            class="px-6 py-2.5 rounded-xl bg-slate-200 text-slate-400 font-semibold text-sm flex items-center gap-2 cursor-not-allowed"
            disabled
          >
            <span class="material-symbols-outlined text-[20px]">play_circle</span>
            Запустить генерацию / Жіберу
          </button>
        {/if}
      </div>
    </div>
  </header>

  <!-- Content Area -->
  <div class="p-8 max-w-6xl mx-auto w-full flex-1">
    {#if error}
      <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm mb-6">{error}</div>
    {/if}

    <!-- Progress Section: Bento Grid Layout -->
    <div class="grid grid-cols-1 md:grid-cols-12 gap-6 mb-8">
      <!-- Large Progress Card -->
      <div class="md:col-span-8 bg-surface-container-lowest p-8 rounded-3xl shadow-sm border border-outline-variant/5">
        <div class="flex justify-between items-start mb-10">
          <div>
            <h3 class="font-display text-lg font-bold text-on-surface mb-1">Текущий прогресс</h3>
            <p class="text-on-surface-variant text-xs uppercase tracking-widest font-medium">Агымдагы барысы</p>
          </div>
          <div class="text-right">
            <span class="text-4xl font-display font-extrabold text-primary">{percentComplete()}%</span>
          </div>
        </div>
        <div class="relative w-full h-4 bg-surface-container-high rounded-full overflow-hidden mb-4">
          <div
            class="absolute top-0 left-0 h-full rounded-full transition-all duration-300 ease-out
              {completed
                ? rowsFailed > 0
                  ? 'bg-gradient-to-r from-amber-400 to-amber-500'
                  : 'bg-gradient-to-r from-emerald-400 to-emerald-500'
                : 'bg-gradient-to-r from-primary to-primary-container shadow-[0_0_12px_rgba(0,74,198,0.3)]'}"
            style="width: {percentComplete()}%"
          ></div>
        </div>
        <div class="flex justify-between text-sm">
          <span class="text-on-surface-variant font-medium">
            {#if generating && !completed}
              Обработка сертификатов...
            {:else if completed}
              Генерация завершена
            {:else}
              Готово к генерации
            {/if}
          </span>
          <span class="text-on-surface font-bold">{progress} из {total.toLocaleString()}</span>
        </div>
      </div>

      <!-- Stats Counters Stack -->
      <div class="md:col-span-4 flex flex-col gap-4">
        <div class="flex-1 bg-surface-container-lowest p-5 rounded-2xl border border-outline-variant/5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-primary-fixed/30 flex items-center justify-center text-primary">
            <span class="material-symbols-outlined">all_inbox</span>
          </div>
          <div>
            <p class="text-[10px] text-on-surface-variant uppercase tracking-widest font-bold">Всего / Барлыгы</p>
            <p class="text-2xl font-display font-extrabold">{total.toLocaleString()}</p>
          </div>
        </div>
        <div class="flex-1 bg-surface-container-lowest p-5 rounded-2xl border border-outline-variant/5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-emerald-100 flex items-center justify-center text-emerald-600">
            <span class="material-symbols-outlined active-nav-icon">check_circle</span>
          </div>
          <div>
            <p class="text-[10px] text-emerald-600 uppercase tracking-widest font-bold">Успешно / Сатті</p>
            <p class="text-2xl font-display font-extrabold">{rowsOk.toLocaleString()}</p>
          </div>
        </div>
        <div class="flex-1 bg-surface-container-lowest p-5 rounded-2xl border border-outline-variant/5 flex items-center gap-4">
          <div class="w-12 h-12 rounded-xl bg-error-container/40 flex items-center justify-center text-error">
            <span class="material-symbols-outlined active-nav-icon">error</span>
          </div>
          <div>
            <p class="text-[10px] text-error uppercase tracking-widest font-bold">Ошибки / Кателер</p>
            <p class="text-2xl font-display font-extrabold">{rowsFailed.toLocaleString()}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Detailed Log Table -->
    {#if logs.length > 0 || generating || completed}
      <div class="bg-surface-container-low rounded-3xl p-1 overflow-hidden">
        <div class="bg-surface-container-lowest rounded-[1.4rem] overflow-hidden">
          <div class="px-6 py-5 border-b border-outline-variant/10 flex justify-between items-center bg-white">
            <h2 class="font-display font-bold text-on-surface">Журнал выполнения / Орындалу журналы</h2>
            {#if generating && !completed}
              <span class="px-3 py-1 bg-surface-container-high rounded-full text-[10px] font-bold text-on-surface-variant uppercase tracking-tighter">Live Updates</span>
            {:else}
              <span class="text-sm text-on-surface-variant">{logs.length} entries</span>
            {/if}
          </div>
          <div class="overflow-x-auto max-h-[450px] overflow-y-auto">
            <table class="w-full text-left border-collapse">
              <thead class="sticky top-0 bg-surface-container-lowest">
                <tr class="bg-surface-container-low/30">
                  <th class="px-6 py-4 text-xs font-bold text-on-surface-variant uppercase tracking-widest">ID</th>
                  <th class="px-6 py-4 text-xs font-bold text-on-surface-variant uppercase tracking-widest">Получатель / Алушы</th>
                  <th class="px-6 py-4 text-xs font-bold text-on-surface-variant uppercase tracking-widest">Статус / Куйі</th>
                  <th class="px-6 py-4 text-xs font-bold text-on-surface-variant uppercase tracking-widest text-right">Время / Уакыты</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-outline-variant/5">
                {#each logs as entry (entry.rowId)}
                  <tr class="hover:bg-surface-container-low transition-colors group">
                    <td class="px-6 py-4 font-mono text-sm text-on-surface-variant">#CERT-{entry.rowId}</td>
                    <td class="px-6 py-4">
                      <p class="text-sm font-semibold text-on-surface">{entry.name}</p>
                      {#if entry.iin}
                        <p class="text-[11px] text-on-surface-variant">{maskIIN(entry.iin)}</p>
                      {/if}
                    </td>
                    <td class="px-6 py-4">
                      <span class="inline-flex items-center px-3 py-1 rounded-full text-[11px] font-bold uppercase tracking-tight {logStatusClasses(entry.status)}">
                        {#if entry.error}
                          {entry.status}: {entry.error}
                        {:else}
                          {entry.status === "ok" ? "Valid" : entry.status}
                        {/if}
                      </span>
                    </td>
                    <td class="px-6 py-4 text-right text-sm text-on-surface-variant font-medium">{entry.time}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
          {#if logs.length > 10}
            <div class="px-6 py-4 bg-surface-container-low/20 border-t border-outline-variant/10 flex justify-center">
              <button class="text-primary font-bold text-xs uppercase tracking-widest hover:underline flex items-center gap-2">
                Показать все записи / Барлык жазбаларды корсету
                <span class="material-symbols-outlined text-sm">keyboard_arrow_down</span>
              </button>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
{:else}
  <div class="text-center py-12 text-error">Batch not found</div>
{/if}
