<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { api, ApiError, type ApiResponse } from "$lib/api/client";
  import StatusBadge from "$lib/components/StatusBadge.svelte";

  interface Event {
    id: number;
    title: string;
    date: string;
    city: string;
    description: string;
    status: string;
  }

  interface Template {
    id: number;
    file_path: string;
    tokens: string[];
  }

  interface Batch {
    id: number;
    file_path: string;
    status: string;
    rows_total: number;
    rows_ok: number;
    rows_failed: number;
    created_at: string;
  }

  let eventId = $derived($page.params.id);
  let event = $state<Event | null>(null);
  let template = $state<Template | null>(null);
  let batches = $state<Batch[]>([]);
  let loading = $state(true);
  let uploading = $state(false);
  let uploadingBatch = $state(false);
  let error = $state("");

  async function loadEvent() {
    loading = true;
    try {
      const res = await api.get<Event>(`/api/v1/staff/events/${eventId}`);
      event = res.data;

      // Load template
      try {
        const tmplRes = await api.get<Template>(`/api/v1/staff/events/${eventId}/template`);
        template = tmplRes.data;
      } catch {
        template = null;
      }

      // Load batches
      try {
        const batchRes = await api.get<Batch[]>(`/api/v1/staff/events/${eventId}/batches`) as unknown as { data: Batch[] };
        batches = batchRes.data ?? [];
      } catch {
        batches = [];
      }
    } catch {
      error = "Failed to load event";
    } finally {
      loading = false;
    }
  }

  async function uploadTemplate(e: globalThis.Event) {
    const input = e.target as HTMLInputElement;
    if (!input.files?.length) return;

    uploading = true;
    error = "";
    const formData = new FormData();
    formData.append("file", input.files[0]);

    try {
      const res = await api.upload<Template>(`/api/v1/staff/events/${eventId}/template`, formData);
      template = res.data;
    } catch (err) {
      error = err instanceof ApiError ? err.message : "Failed to upload template";
    } finally {
      uploading = false;
      input.value = "";
    }
  }

  async function deleteTemplate() {
    if (!confirm("Delete template? This cannot be undone.")) return;
    try {
      await api.delete(`/api/v1/staff/events/${eventId}/template`);
      template = null;
    } catch (err) {
      error = err instanceof ApiError ? err.message : "Failed to delete template";
    }
  }

  async function uploadBatch(e: globalThis.Event) {
    const input = e.target as HTMLInputElement;
    if (!input.files?.length) return;

    uploadingBatch = true;
    error = "";
    const formData = new FormData();
    formData.append("file", input.files[0]);

    try {
      const res = await api.upload<{ batch: Batch }>(`/api/v1/staff/events/${eventId}/batches`, formData);
      const batch = res.data.batch;
      batches = [batch, ...batches];
      // Redirect to mapping page
      window.location.href = `/staff/events/${eventId}/batches/${batch.id}`;
    } catch (err) {
      error = err instanceof ApiError ? err.message : "Failed to upload batch";
    } finally {
      uploadingBatch = false;
      input.value = "";
    }
  }

  onMount(loadEvent);
</script>

{#if loading}
  <div class="text-center py-12 text-on-surface-variant">Loading event...</div>
{:else if event}
  <div class="space-y-8">
    <!-- Header -->
    <div>
      <a href="/staff/events" class="text-sm text-on-surface-variant hover:text-primary transition-colors">
        &larr; Back to events
      </a>
      <div class="flex items-start justify-between mt-2">
        <div>
          <h1 class="font-display text-2xl font-bold text-on-surface">{event.title}</h1>
          <div class="flex items-center gap-3 mt-1 text-sm text-on-surface-variant">
            {#if event.date}<span>{event.date}</span>{/if}
            {#if event.city}<span>{event.city}</span>{/if}
            <StatusBadge status={event.status} />
          </div>
        </div>
        <a
          href="/staff/events/{eventId}/certificates"
          class="px-4 py-2 rounded-lg text-sm font-medium bg-surface-low text-on-surface hover:bg-surface-high transition-colors"
        >
          View Certificates
        </a>
      </div>
      {#if event.description}
        <p class="text-sm text-on-surface-variant mt-2">{event.description}</p>
      {/if}
    </div>

    {#if error}
      <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm">{error}</div>
    {/if}

    <!-- Template Section -->
    <section class="bg-surface-lowest rounded-lg p-6 space-y-4">
      <h2 class="font-display text-lg font-semibold text-on-surface">Template</h2>

      {#if template}
        <div class="flex items-center justify-between p-4 rounded-lg bg-surface">
          <div>
            <p class="text-sm font-medium text-on-surface">
              {template.file_path.split("/").pop()}
            </p>
            <div class="flex flex-wrap gap-1.5 mt-2">
              {#each template.tokens as token}
                <span class="px-2 py-0.5 rounded bg-primary-fixed text-on-primary-container text-xs font-mono">
                  {token}
                </span>
              {/each}
            </div>
          </div>
          <button
            onclick={deleteTemplate}
            class="text-xs text-error hover:underline shrink-0 ml-4"
          >
            Delete
          </button>
        </div>
      {:else}
        <div class="text-center py-6">
          <p class="text-sm text-on-surface-variant mb-3">Upload a PPTX template for this event</p>
          <label class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium cursor-pointer
                        bg-gradient-to-br from-primary to-primary-container text-on-primary
                        hover:shadow-lg transition-shadow {uploading ? 'opacity-50 pointer-events-none' : ''}">
            {uploading ? "Uploading..." : "Upload .pptx"}
            <input type="file" accept=".pptx" onchange={uploadTemplate} class="sr-only" />
          </label>
        </div>
      {/if}
    </section>

    <!-- Batch Upload Section -->
    <section class="bg-surface-lowest rounded-lg p-6 space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="font-display text-lg font-semibold text-on-surface">Import Batches</h2>
        {#if template}
          <label class="inline-flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium cursor-pointer
                        bg-surface-low text-on-surface hover:bg-surface-high transition-colors
                        {uploadingBatch ? 'opacity-50 pointer-events-none' : ''}">
            {uploadingBatch ? "Uploading..." : "Upload CSV/XLSX"}
            <input type="file" accept=".csv,.xlsx" onchange={uploadBatch} class="sr-only" />
          </label>
        {/if}
      </div>

      {#if !template}
        <p class="text-sm text-on-surface-variant">Upload a template first before importing participant data.</p>
      {:else if batches.length === 0}
        <p class="text-sm text-on-surface-variant">No batches uploaded yet.</p>
      {:else}
        <div class="space-y-2">
          {#each batches as batch}
            <a
              href="/staff/events/{eventId}/batches/{batch.id}"
              class="flex items-center justify-between p-3 rounded-lg bg-surface hover:bg-surface-low transition-colors"
            >
              <div class="flex items-center gap-3">
                <StatusBadge status={batch.status} />
                <span class="text-sm text-on-surface">
                  {batch.rows_total} rows
                </span>
              </div>
              <span class="text-xs text-on-surface-variant">
                {new Date(batch.created_at).toLocaleString()}
              </span>
            </a>
          {/each}
        </div>
      {/if}
    </section>
  </div>
{:else}
  <div class="text-center py-12 text-error">Event not found</div>
{/if}
