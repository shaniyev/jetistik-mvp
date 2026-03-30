<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { api, ApiError } from "$lib/api/client";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";

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

  interface Template {
    tokens: string[];
  }

  let eventId = $derived($page.params.id);
  let batchId = $derived($page.params.batchId);

  let batch = $state<Batch | null>(null);
  let templateTokens = $state<string[]>([]);
  let mapping = $state<Record<string, string>>({});
  let loading = $state(true);
  let saving = $state(false);
  let error = $state("");

  async function loadData() {
    loading = true;
    try {
      const batchRes = await api.get<Batch>(`/api/v1/staff/batches/${batchId}`);
      batch = batchRes.data;

      // Load template tokens
      try {
        const tmplRes = await api.get<Template>(`/api/v1/staff/events/${eventId}/template`);
        templateTokens = tmplRes.data.tokens;
      } catch {
        templateTokens = [];
      }

      // Initialize mapping from batch or default
      if (batch.mapping && Object.keys(batch.mapping).length > 0) {
        mapping = { ...batch.mapping };
      } else {
        // Create empty mapping for each template token
        for (const token of templateTokens) {
          mapping[token] = "";
        }
      }
    } catch (e) {
      error = $t("common.unexpectedError");
    } finally {
      loading = false;
    }
  }

  async function saveMapping() {
    saving = true;
    error = "";
    try {
      await api.patch(`/api/v1/staff/batches/${batchId}/mapping`, { mapping });
      goto(`/staff/events/${eventId}`);
    } catch (err) {
      error = err instanceof ApiError ? err.message : $t("common.unexpectedError");
    } finally {
      saving = false;
    }
  }

  onMount(loadData);
</script>

{#if loading}
  <div class="flex items-center justify-center py-24 text-on-surface-variant">
    <span class="material-symbols-outlined animate-spin mr-2">progress_activity</span>
    {$t("staff.batch.loading")}
  </div>
{:else if batch}
  <div class="p-6 lg:p-10 pb-32 max-w-3xl">
    <!-- Header -->
    <header class="mb-12">
      <a href="/staff/events/{eventId}" class="flex items-center gap-2 text-primary font-semibold text-sm mb-2 hover:underline">
        <span class="material-symbols-outlined text-sm">arrow_back</span>
        <span>{$t("staff.batch.back_to_event")}</span>
      </a>
      <h1 class="font-display text-4xl font-extrabold tracking-tight text-on-surface">{$t("staff.mapping.title")}</h1>
      <div class="flex items-center gap-3 mt-2">
        <p class="text-on-surface-variant">
          {batch.rows_total} {$t("staff.batch.rows_found")}
        </p>
        <StatusBadge status={batch.status} />
      </div>
    </header>

    {#if error}
      <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm mb-6">{error}</div>
    {/if}

    <!-- Mapping Card -->
    <section class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 overflow-hidden">
      <!-- Card Header -->
      <div class="px-6 py-5 border-b border-outline-variant/10">
        <div class="grid grid-cols-[1fr_auto_1fr] gap-4 items-center">
          <span class="text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("staff.mapping.templateToken")}</span>
          <span></span>
          <span class="text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("staff.mapping.csvColumn")}</span>
        </div>
      </div>

      <!-- Mapping Rows -->
      <div class="divide-y divide-outline-variant/10">
        {#each templateTokens as token}
          <div class="grid grid-cols-[1fr_auto_1fr] gap-4 items-center px-6 py-4">
            <div class="px-4 py-2.5 rounded-lg bg-primary-fixed text-on-primary-container text-sm font-mono font-medium">
              {token}
            </div>
            <span class="material-symbols-outlined text-on-surface-variant/50">arrow_forward</span>
            <select
              bind:value={mapping[token]}
              class="w-full px-4 py-2.5 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm font-medium
                     focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-shadow"
            >
              <option value="">{$t("staff.mapping.notMapped")}</option>
              {#each batch.tokens ?? [] as col}
                <option value={col}>{col}</option>
              {/each}
            </select>
          </div>
        {/each}
      </div>
    </section>

    <!-- Actions -->
    <div class="flex items-center gap-3 mt-8">
      <button
        onclick={saveMapping}
        disabled={saving}
        class="px-6 py-3 rounded-xl bg-gradient-to-br from-primary to-primary-container text-white font-semibold text-sm shadow-lg shadow-primary/20 hover:shadow-xl transition-all disabled:opacity-50 active:scale-95"
      >
        {saving ? $t("staff.batch.saving") : $t("staff.mapping.save")}
      </button>
      {#if batch.mapping && Object.keys(batch.mapping).length > 0}
        <a
          href="/staff/events/{eventId}/batches/{batchId}/generate"
          class="px-6 py-3 rounded-xl bg-emerald-50 text-emerald-700 font-semibold text-sm hover:bg-emerald-100 transition-colors flex items-center gap-2"
        >
          <span class="material-symbols-outlined text-lg">play_arrow</span>
          {$t("staff.mapping.generate")}
        </a>
      {/if}
      <a
        href="/staff/events/{eventId}"
        class="px-6 py-3 rounded-xl border border-outline-variant/20 text-on-surface-variant font-semibold text-sm hover:bg-surface-container transition-colors"
      >
        {$t("common.cancel")}
      </a>
    </div>
  </div>
{:else}
  <div class="flex items-center justify-center py-24 text-error">{$t("staff.mapping.batchNotFound")}</div>
{/if}
