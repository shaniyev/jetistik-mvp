<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { api, ApiError } from "$lib/api/client";
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
      error = "Failed to load batch";
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
      error = err instanceof ApiError ? err.message : "Failed to save mapping";
    } finally {
      saving = false;
    }
  }

  onMount(loadData);
</script>

{#if loading}
  <div class="text-center py-12 text-on-surface-variant">Loading batch...</div>
{:else if batch}
  <div class="max-w-2xl space-y-6">
    <div>
      <a href="/staff/events/{eventId}" class="text-sm text-on-surface-variant hover:text-primary transition-colors">
        &larr; Back to event
      </a>
      <h1 class="font-display text-2xl font-bold text-on-surface mt-2">Column Mapping</h1>
      <p class="text-sm text-on-surface-variant mt-1">
        Map CSV/XLSX columns to template tokens. {batch.rows_total} rows found.
        <StatusBadge status={batch.status} />
      </p>
    </div>

    {#if error}
      <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm">{error}</div>
    {/if}

    <div class="bg-surface-lowest rounded-lg p-6 space-y-4">
      <div class="grid grid-cols-[1fr_auto_1fr] gap-3 items-center text-sm font-medium text-on-surface-variant">
        <span>Template Token</span>
        <span></span>
        <span>CSV/XLSX Column</span>
      </div>

      {#each templateTokens as token}
        <div class="grid grid-cols-[1fr_auto_1fr] gap-3 items-center">
          <div class="px-3 py-2.5 rounded-md bg-primary-fixed text-on-primary-container text-sm font-mono">
            {token}
          </div>
          <svg class="w-5 h-5 text-on-surface-variant" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
          </svg>
          <select
            bind:value={mapping[token]}
            class="w-full px-3 py-2.5 rounded-md bg-surface text-on-surface text-sm
                   focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow"
          >
            <option value="">— not mapped —</option>
            {#each batch.tokens ?? [] as col}
              <option value={col}>{col}</option>
            {/each}
          </select>
        </div>
      {/each}
    </div>

    <div class="flex gap-3">
      <button
        onclick={saveMapping}
        disabled={saving}
        class="flex-1 py-2.5 rounded-lg text-sm font-medium
               bg-gradient-to-br from-primary to-primary-container text-on-primary
               hover:shadow-lg disabled:opacity-50 transition-all"
      >
        {saving ? "Saving..." : "Save Mapping"}
      </button>
      {#if batch.mapping && Object.keys(batch.mapping).length > 0}
        <a
          href="/staff/events/{eventId}/batches/{batchId}/generate"
          class="px-6 py-2.5 rounded-lg text-sm font-medium bg-emerald-50 text-emerald-700 hover:bg-emerald-100 transition-colors"
        >
          Generate
        </a>
      {/if}
      <a
        href="/staff/events/{eventId}"
        class="px-6 py-2.5 rounded-lg text-sm font-medium bg-surface-low text-on-surface hover:bg-surface-high transition-colors"
      >
        Cancel
      </a>
    </div>
  </div>
{:else}
  <div class="text-center py-12 text-error">Batch not found</div>
{/if}
