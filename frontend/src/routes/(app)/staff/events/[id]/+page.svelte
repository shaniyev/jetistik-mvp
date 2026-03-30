<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { api, ApiError, type ApiResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";

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
  let editingEvent = $state(false);
  let editForm = $state({ title: "", date: "", city: "", description: "" });
  let savingEvent = $state(false);

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
      error = $t("common.unexpectedError");
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
      error = err instanceof ApiError ? err.message : $t("common.unexpectedError");
    } finally {
      uploading = false;
      input.value = "";
    }
  }

  async function deleteTemplate() {
    if (!confirm($t("staff.event.deleteTemplateConfirm"))) return;
    try {
      await api.delete(`/api/v1/staff/events/${eventId}/template`);
      template = null;
    } catch (err) {
      error = err instanceof ApiError ? err.message : $t("common.unexpectedError");
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
      error = err instanceof ApiError ? err.message : $t("common.unexpectedError");
    } finally {
      uploadingBatch = false;
      input.value = "";
    }
  }

  async function deleteBatch(batchId: number) {
    if (!confirm($t("staff.event.deleteBatchConfirm"))) return;
    try {
      await api.delete(`/api/v1/staff/batches/${batchId}`);
      batches = batches.filter((b) => b.id !== batchId);
    } catch (err) {
      error = err instanceof ApiError ? err.message : $t("common.unexpectedError");
    }
  }

  function startEditEvent() {
    if (!event) return;
    editForm = {
      title: event.title,
      date: event.date,
      city: event.city,
      description: event.description,
    };
    editingEvent = true;
  }

  async function saveEvent() {
    savingEvent = true;
    error = "";
    try {
      const res = await api.patch<Event>(`/api/v1/staff/events/${eventId}`, editForm);
      event = res.data;
      editingEvent = false;
    } catch (err) {
      error = err instanceof ApiError ? err.message : $t("common.unexpectedError");
    } finally {
      savingEvent = false;
    }
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "—";
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString("en-US", { month: "long", day: "numeric", year: "numeric" });
    } catch {
      return dateStr;
    }
  }

  function batchStatusClasses(status: string): string {
    switch (status) {
      case "generating":
        return "bg-primary-fixed text-on-primary-fixed-variant";
      case "done":
      case "completed":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-error-container text-on-error-container";
      default:
        return "bg-surface-container-highest text-on-surface-variant";
    }
  }

  function batchStatusIcon(status: string): string {
    switch (status) {
      case "generating": return "";
      case "done":
      case "completed": return "check_circle";
      case "failed": return "error";
      default: return "";
    }
  }

  onMount(loadEvent);
</script>

{#if loading}
  <div class="flex items-center justify-center py-24 text-on-surface-variant">
    <span class="material-symbols-outlined animate-spin mr-2">progress_activity</span>
    {$t("staff.event.loading")}
  </div>
{:else if event}
  <div class="p-6 lg:p-10 pb-32">
    <!-- Header Actions Bar -->
    <header class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-12">
      <div class="space-y-1">
        <a href="/staff/events" class="flex items-center gap-2 text-primary font-semibold text-sm mb-2 hover:underline">
          <span class="material-symbols-outlined text-sm">arrow_back</span>
          <span>{$t("staff.event.backToEvents")}</span>
        </a>
        <h1 class="font-display text-4xl font-extrabold tracking-tight text-on-surface">{event.title}</h1>
        {#if event.description}
          <p class="text-on-surface-variant max-w-2xl">{event.description}</p>
        {/if}
      </div>
      <div class="flex items-center gap-3">
        <button
          onclick={startEditEvent}
          class="px-5 py-2.5 rounded-lg border border-outline-variant font-semibold text-sm hover:bg-surface-container transition-colors flex items-center gap-2"
        >
          <span class="material-symbols-outlined text-lg">edit</span>
          {$t("staff.event.edit")}
        </button>
        <a
          href="/staff/events/{eventId}/certificates"
          class="px-5 py-2.5 rounded-lg bg-gradient-to-br from-primary to-primary-container text-white font-semibold text-sm shadow-lg shadow-primary/20 flex items-center gap-2 active:scale-95 transition-transform"
        >
          <span class="material-symbols-outlined text-lg">download_for_offline</span>
          {$t("staff.event.downloadAll")}
        </a>
      </div>
    </header>

    {#if error}
      <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm mb-6">{error}</div>
    {/if}

    <!-- Edit Event Modal -->
    {#if editingEvent}
      <div class="mb-8 bg-surface-container-lowest rounded-xl p-6 shadow-sm border border-outline-variant/10">
        <h3 class="font-display font-bold text-lg mb-4">{$t("staff.event.editEvent")}</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label class="block text-xs text-on-surface-variant font-medium uppercase tracking-wider mb-1">{$t("common.title")}</label>
            <input
              type="text"
              bind:value={editForm.title}
              class="w-full px-3 py-2 rounded-lg bg-surface text-sm text-on-surface border border-outline-variant/30 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div>
            <label class="block text-xs text-on-surface-variant font-medium uppercase tracking-wider mb-1">{$t("common.date")}</label>
            <input
              type="text"
              bind:value={editForm.date}
              placeholder="2026-03-29"
              class="w-full px-3 py-2 rounded-lg bg-surface text-sm text-on-surface border border-outline-variant/30 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div>
            <label class="block text-xs text-on-surface-variant font-medium uppercase tracking-wider mb-1">{$t("common.city")}</label>
            <input
              type="text"
              bind:value={editForm.city}
              class="w-full px-3 py-2 rounded-lg bg-surface text-sm text-on-surface border border-outline-variant/30 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
            />
          </div>
          <div>
            <label class="block text-xs text-on-surface-variant font-medium uppercase tracking-wider mb-1">{$t("common.description")}</label>
            <input
              type="text"
              bind:value={editForm.description}
              class="w-full px-3 py-2 rounded-lg bg-surface text-sm text-on-surface border border-outline-variant/30 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
            />
          </div>
        </div>
        <div class="flex items-center gap-3 mt-4">
          <button
            onclick={saveEvent}
            disabled={savingEvent || !editForm.title}
            class="px-5 py-2.5 rounded-lg text-sm font-semibold bg-gradient-to-br from-primary to-primary-container text-white shadow-lg shadow-primary/20 disabled:opacity-50 transition-all active:scale-95"
          >
            {savingEvent ? $t("staff.event.saving") : $t("staff.event.saveChanges")}
          </button>
          <button
            onclick={() => { editingEvent = false; }}
            class="px-5 py-2.5 rounded-lg text-sm font-semibold text-on-surface-variant hover:bg-surface-container transition-colors"
          >
            {$t("common.cancel")}
          </button>
        </div>
      </div>
    {/if}

    <!-- Bento Grid Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Left Column: Event Specs -->
      <div class="lg:col-span-4 space-y-6">
        <!-- Meta Card -->
        <section class="bg-surface-container-lowest rounded-xl p-6 shadow-sm border border-outline-variant/10">
          <h3 class="font-display font-bold text-lg mb-6">{$t("staff.events.event_logistics")}</h3>
          <div class="space-y-4">
            <div class="flex items-start gap-4">
              <div class="p-2 bg-surface-container rounded-lg">
                <span class="material-symbols-outlined text-primary">calendar_today</span>
              </div>
              <div>
                <p class="text-xs text-on-surface-variant font-medium uppercase tracking-wider">{$t("common.date")}</p>
                <p class="font-semibold">{formatDate(event.date)}</p>
              </div>
            </div>
            <div class="flex items-start gap-4">
              <div class="p-2 bg-surface-container rounded-lg">
                <span class="material-symbols-outlined text-primary">location_on</span>
              </div>
              <div>
                <p class="text-xs text-on-surface-variant font-medium uppercase tracking-wider">{$t("common.city")}</p>
                <p class="font-semibold">{event.city || "—"}</p>
              </div>
            </div>
            <div class="flex items-start gap-4">
              <div class="p-2 bg-surface-container rounded-lg">
                <span class="material-symbols-outlined text-primary">fingerprint</span>
              </div>
              <div>
                <p class="text-xs text-on-surface-variant font-medium uppercase tracking-wider">{$t("staff.events.event_id")}</p>
                <p class="font-mono text-sm">EVT-{event.id}</p>
              </div>
            </div>
          </div>
        </section>

        <!-- Detected Tokens -->
        {#if template}
          <section class="bg-surface-container-lowest rounded-xl p-6 shadow-sm border border-outline-variant/10">
            <div class="flex items-center justify-between mb-6">
              <h3 class="font-display font-bold text-lg">{$t("staff.events.detected_tokens")}</h3>
              <span class="px-2 py-0.5 bg-secondary-container text-on-surface-variant text-[10px] font-bold rounded uppercase">{$t("staff.events.auto_parsed")}</span>
            </div>
            <div class="flex flex-wrap gap-2">
              {#each template.tokens as token}
                <div class="px-3 py-1.5 bg-surface-container rounded flex items-center gap-2 border border-outline-variant/20">
                  <span class="text-xs font-mono text-primary">{token}</span>
                </div>
              {/each}
            </div>
            <p class="mt-4 text-xs text-on-surface-variant leading-relaxed">
              {$t("staff.events.tokens_hint")}
            </p>
            <button
              onclick={deleteTemplate}
              class="mt-3 text-xs text-error hover:underline flex items-center gap-1"
            >
              <span class="material-symbols-outlined text-sm">delete</span>
              {$t("staff.events.remove_template")}
            </button>
          </section>
        {/if}
      </div>

      <!-- Right Column: Main Upload Zones -->
      <div class="lg:col-span-8 space-y-6">
        <!-- Template Upload Area -->
        {#if !template}
          <section class="bg-surface-container-lowest rounded-xl p-8 border-2 border-dashed border-outline-variant flex flex-col items-center justify-center text-center group hover:border-primary/50 transition-colors">
            <div class="w-16 h-16 bg-primary-fixed rounded-2xl flex items-center justify-center text-primary mb-4 group-hover:scale-110 transition-transform">
              <span class="material-symbols-outlined text-3xl">upload_file</span>
            </div>
            <h3 class="font-display font-bold text-xl mb-2">{$t("staff.events.certificate_template")}</h3>
            <p class="text-on-surface-variant text-sm mb-6 max-w-sm">
              {$t("staff.events.template_drag_hint")}
            </p>
            <div class="flex items-center gap-4">
              <label class="px-6 py-2 bg-surface-container text-on-surface font-semibold text-sm rounded-lg hover:bg-surface-container-high transition-colors cursor-pointer {uploading ? 'opacity-50 pointer-events-none' : ''}">
                {uploading ? $t("staff.batch.uploading") : $t("staff.events.browse_files")}
                <input type="file" accept=".pptx" onchange={uploadTemplate} class="sr-only" />
              </label>
              <span class="text-xs text-on-surface-variant font-medium">Max 25MB</span>
            </div>
          </section>
        {:else}
          <section class="bg-surface-container-lowest rounded-xl p-8 border border-outline-variant/10 shadow-sm">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 bg-primary-fixed rounded-xl flex items-center justify-center text-primary">
                <span class="material-symbols-outlined text-2xl">description</span>
              </div>
              <div>
                <h3 class="font-display font-bold text-lg">{$t("staff.events.certificate_template")}</h3>
                <p class="text-sm text-on-surface-variant">{template.file_path.split("/").pop()}</p>
              </div>
            </div>
            <label class="inline-flex items-center gap-2 px-4 py-2 bg-surface-container text-on-surface font-semibold text-sm rounded-lg hover:bg-surface-container-high transition-colors cursor-pointer {uploading ? 'opacity-50 pointer-events-none' : ''}">
              <span class="material-symbols-outlined text-lg">swap_horiz</span>
              {uploading ? $t("staff.batch.uploading") : $t("staff.events.replace_template")}
              <input type="file" accept=".pptx" onchange={uploadTemplate} class="sr-only" />
            </label>
          </section>
        {/if}

        <!-- Batch History -->
        <section class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 overflow-hidden">
          <div class="p-6 border-b border-surface-container flex items-center justify-between">
            <div>
              <h3 class="font-display font-bold text-lg">{$t("staff.events.batch_history")}</h3>
              <p class="text-sm text-on-surface-variant">{$t("staff.events.batch_track_hint")}</p>
            </div>
            {#if template}
              <label class="px-4 py-2 bg-primary/10 text-primary font-bold text-sm rounded-lg flex items-center gap-2 hover:bg-primary/20 transition-colors cursor-pointer {uploadingBatch ? 'opacity-50 pointer-events-none' : ''}">
                <span class="material-symbols-outlined text-lg">add</span>
                {uploadingBatch ? $t("staff.batch.uploading") : $t("staff.events.import_data")}
                <input type="file" accept=".csv,.xlsx" onchange={uploadBatch} class="sr-only" />
              </label>
            {/if}
          </div>

          {#if !template}
            <div class="p-8 text-center text-on-surface-variant text-sm">
              {$t("staff.events.no_template_hint")}
            </div>
          {:else if batches.length === 0}
            <div class="p-8 text-center text-on-surface-variant text-sm">
              {$t("staff.events.no_batches")}
            </div>
          {:else}
            <div class="overflow-x-auto">
              <table class="w-full text-left">
                <thead class="bg-surface-container-low">
                  <tr>
                    <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-on-surface-variant">{$t("staff.events.col.batch_id")}</th>
                    <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-on-surface-variant">{$t("staff.events.col.file_source")}</th>
                    <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-on-surface-variant">{$t("staff.events.col.records")}</th>
                    <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-on-surface-variant">{$t("common.status")}</th>
                    <th class="px-6 py-4 text-xs font-bold uppercase tracking-wider text-on-surface-variant text-right">{$t("staff.events.col.action")}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-surface-container">
                  {#each batches as batch (batch.id)}
                    <tr class="hover:bg-surface-container-low transition-colors">
                      <td class="px-6 py-5 font-mono text-sm text-on-surface">B-{batch.id}</td>
                      <td class="px-6 py-5">
                        <div class="flex items-center gap-2">
                          <span class="material-symbols-outlined text-green-600">table_view</span>
                          <span class="text-sm font-medium">{batch.file_path?.split("/").pop() || "—"}</span>
                        </div>
                      </td>
                      <td class="px-6 py-5 text-sm">{batch.rows_total}</td>
                      <td class="px-6 py-5">
                        <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold uppercase {batchStatusClasses(batch.status)}">
                          {#if batch.status === "generating"}
                            <span class="w-1.5 h-1.5 rounded-full bg-primary animate-pulse"></span>
                          {:else if batchStatusIcon(batch.status)}
                            <span class="material-symbols-outlined text-[12px]">{batchStatusIcon(batch.status)}</span>
                          {/if}
                          {batch.status}
                        </span>
                      </td>
                      <td class="px-6 py-5 text-right">
                        <div class="flex items-center justify-end gap-2">
                          <a
                            href="/staff/events/{eventId}/batches/{batch.id}"
                            class="text-on-surface-variant hover:text-primary transition-colors"
                          >
                            <span class="material-symbols-outlined">visibility</span>
                          </a>
                          <button
                            onclick={() => deleteBatch(batch.id)}
                            class="text-on-surface-variant hover:text-error transition-colors"
                          >
                            <span class="material-symbols-outlined">delete</span>
                          </button>
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </section>
      </div>
    </div>
  </div>
{:else}
  <div class="text-center py-12 text-error">{$t("staff.events.event_not_found")}</div>
{/if}
