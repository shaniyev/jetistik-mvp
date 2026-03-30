<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { api, ApiError, getAccessToken, type PaginatedResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import JSZip from "jszip";

  interface Certificate {
    id: number;
    name: string;
    iin: string;
    code: string;
    status: string;
    created_at: string;
  }

  let eventId = $derived($page.params.id);
  let certs = $state<Certificate[]>([]);
  let loading = $state(true);
  let currentPage = $state(1);
  let total = $state(0);
  let error = $state("");
  let downloading = $state(false);
  let downloadProgress = $state("");
  const perPage = 20;

  async function loadCerts() {
    loading = true;
    try {
      const res = await api.get<Certificate[]>(
        `/api/v1/staff/events/${eventId}/certificates?page=${currentPage}&per_page=${perPage}`
      ) as PaginatedResponse<Certificate>;
      certs = res.data;
      total = res.pagination.total;
    } catch (e) {
      error = $t("common.unexpectedError");
    } finally {
      loading = false;
    }
  }

  async function revoke(id: number) {
    const reason = prompt($t("staff.certs.revokeReason"));
    if (!reason) return;
    try {
      await api.post(`/api/v1/staff/certificates/${id}/revoke`, { reason });
      loadCerts();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : $t("common.unexpectedError"));
    }
  }

  async function unrevoke(id: number) {
    try {
      await api.post(`/api/v1/staff/certificates/${id}/unrevoke`);
      loadCerts();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : $t("common.unexpectedError"));
    }
  }

  async function downloadPdf(cert: Certificate) {
    try {
      const token = getAccessToken();
      const headers: Record<string, string> = {};
      if (token) headers["Authorization"] = `Bearer ${token}`;
      const res = await fetch(`/api/v1/staff/certificates/${cert.id}/download`, { headers, credentials: "include" });
      if (!res.ok) throw new Error("Download failed");
      const blob = await res.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = `${cert.name}_${cert.code.slice(0, 8)}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(a.href);
    } catch (e) {
      alert($t("common.unexpectedError"));
    }
  }

  // Edit
  let editingCert = $state<Certificate | null>(null);
  let editName = $state("");
  let editIIN = $state("");
  let editSaving = $state(false);

  function startEdit(cert: Certificate) {
    editingCert = cert;
    editName = cert.name;
    editIIN = cert.iin;
  }

  async function saveEdit() {
    if (!editingCert) return;
    editSaving = true;
    try {
      await api.patch(`/api/v1/staff/certificates/${editingCert.id}`, { name: editName, iin: editIIN });
      editingCert = null;
      loadCerts();
    } catch (e: any) {
      alert(e.message || $t("common.unexpectedError"));
    } finally {
      editSaving = false;
    }
  }

  async function downloadAll() {
    downloading = true;
    downloadProgress = $t("staff.certs.fetching");
    try {
      const apiBase = import.meta.env.VITE_API_URL || "";
      const token = getAccessToken();
      const headers: Record<string, string> = {};
      if (token) headers["Authorization"] = `Bearer ${token}`;

      // Fetch all pages of certificates
      let allCerts: Certificate[] = [];
      let p = 1;
      while (true) {
        const res = await api.get<Certificate[]>(
          `/api/v1/staff/events/${eventId}/certificates?page=${p}&per_page=100`
        ) as PaginatedResponse<Certificate>;
        allCerts = allCerts.concat(res.data);
        if (allCerts.length >= res.pagination.total) break;
        p++;
      }

      const zip = new JSZip();
      let done = 0;
      for (const cert of allCerts) {
        downloadProgress = `${$t("staff.certs.downloading")} ${done + 1} / ${allCerts.length}...`;
        try {
          const res = await fetch(`${apiBase}/api/v1/staff/certificates/${cert.id}/download`, {
            headers,
            credentials: "include",
          });
          if (res.ok) {
            const blob = await res.blob();
            const name = `${cert.name}_${cert.code.slice(0, 8)}.pdf`;
            zip.file(name, blob);
          }
        } catch {
          // Skip failed downloads
        }
        done++;
      }

      downloadProgress = $t("staff.certs.creating_zip");
      const zipBlob = await zip.generateAsync({ type: "blob" });
      const a = document.createElement("a");
      a.href = URL.createObjectURL(zipBlob);
      a.download = `certificates_event_${eventId}.zip`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(a.href);
    } catch (err) {
      alert(err instanceof ApiError ? err.message : $t("common.unexpectedError"));
    } finally {
      downloading = false;
      downloadProgress = "";
    }
  }

  onMount(loadCerts);

  let totalPages = $derived(Math.ceil(total / perPage));
</script>

<div class="p-6 lg:p-10 pb-32">
  <!-- Header -->
  <header class="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-12">
    <div class="space-y-1">
      <a href="/staff/events/{eventId}" class="flex items-center gap-2 text-primary font-semibold text-sm mb-2 hover:underline">
        <span class="material-symbols-outlined text-sm">arrow_back</span>
        <span>{$t("staff.event.backToEvents")}</span>
      </a>
      <h1 class="font-display text-4xl font-extrabold tracking-tight text-on-surface">{$t("staff.certs.title")}</h1>
      <p class="text-on-surface-variant">{total} {$t("staff.certs.total")}</p>
    </div>
    {#if total > 0}
      <button
        onclick={downloadAll}
        disabled={downloading}
        class="px-5 py-2.5 rounded-lg bg-gradient-to-br from-primary to-primary-container text-white font-semibold text-sm shadow-lg shadow-primary/20 flex items-center gap-2 active:scale-95 transition-transform disabled:opacity-50"
      >
        <span class="material-symbols-outlined text-lg">download</span>
        {downloading ? downloadProgress : $t("staff.certs.downloadAll")}
      </button>
    {/if}
  </header>

  {#if error}
    <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm mb-6">{error}</div>
  {/if}

  <!-- Table Card -->
  <section class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-left">
        <thead class="bg-surface-container-low">
          <tr>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.name")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.code")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.status")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.created")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant text-right">{$t("common.actions")}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant/10">
          {#if loading}
            <tr>
              <td colspan="5" class="px-6 py-16 text-center text-on-surface-variant">
                <div class="flex items-center justify-center gap-2">
                  <span class="material-symbols-outlined animate-spin text-primary">progress_activity</span>
                  <span>{$t("common.loading")}</span>
                </div>
              </td>
            </tr>
          {:else if certs.length === 0}
            <tr>
              <td colspan="5" class="px-6 py-16 text-center text-on-surface-variant">{$t("staff.certs.empty")}</td>
            </tr>
          {:else}
            {#each certs as cert (cert.id)}
              <tr class="hover:bg-surface-container-low/30 transition-colors">
                <td class="px-6 py-5">
                  <p class="font-medium text-on-surface text-sm">{cert.name}</p>
                  <p class="text-xs text-on-surface-variant font-mono mt-0.5">{cert.iin}</p>
                </td>
                <td class="px-6 py-5 font-mono text-sm text-on-surface-variant">{cert.code.slice(0, 8)}</td>
                <td class="px-6 py-5">
                  <StatusBadge status={cert.status} />
                </td>
                <td class="px-6 py-5 text-sm text-on-surface-variant">
                  {new Date(cert.created_at).toLocaleDateString()}
                </td>
                <td class="px-6 py-5 text-right">
                  <div class="flex items-center justify-end gap-3">
                    <button onclick={() => downloadPdf(cert)} class="text-primary text-xs font-semibold hover:underline">
                      PDF
                    </button>
                    <button onclick={() => startEdit(cert)} class="text-on-surface-variant text-xs font-semibold hover:text-primary hover:underline">
                      {$t("common.edit")}
                    </button>
                    {#if cert.status === "valid"}
                      <button onclick={() => revoke(cert.id)} class="text-error text-xs font-semibold hover:underline">
                        {$t("staff.certs.revoke")}
                      </button>
                    {:else if cert.status === "revoked"}
                      <button onclick={() => unrevoke(cert.id)} class="text-emerald-600 text-xs font-semibold hover:underline">
                        {$t("staff.certs.restore")}
                      </button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>

  <!-- Pagination -->
  {#if total > perPage}
    <div class="flex items-center justify-between mt-6">
      <p class="text-sm text-on-surface-variant">
        {$t("staff.certs.page_of")} {currentPage} / {totalPages}
      </p>
      <div class="flex items-center gap-2">
        <button
          disabled={currentPage <= 1}
          onclick={() => { currentPage--; loadCerts(); }}
          class="px-4 py-2 rounded-lg border border-outline-variant/20 text-sm font-semibold text-on-surface hover:bg-surface-container transition-colors disabled:opacity-40"
        >
          {$t("common.previous")}
        </button>
        <button
          disabled={currentPage * perPage >= total}
          onclick={() => { currentPage++; loadCerts(); }}
          class="px-4 py-2 rounded-lg border border-outline-variant/20 text-sm font-semibold text-on-surface hover:bg-surface-container transition-colors disabled:opacity-40"
        >
          {$t("common.next")}
        </button>
      </div>
    </div>
  {/if}
</div>

<!-- Edit Modal -->
{#if editingCert}
  <div class="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4" onclick={() => { editingCert = null; }}>
    <div class="bg-surface-container-lowest rounded-2xl p-8 w-full max-w-md shadow-2xl border border-outline-variant/10" onclick={(e) => e.stopPropagation()}>
      <h3 class="font-display text-xl font-bold text-on-surface mb-6">{$t("common.edit")} — {editingCert.code.slice(0, 8)}</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-1.5">{$t("common.name")}</label>
          <input
            type="text"
            bind:value={editName}
            class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-shadow"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-1.5">IIN</label>
          <input
            type="text"
            bind:value={editIIN}
            maxlength="12"
            inputmode="numeric"
            class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm font-mono focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-shadow"
          />
        </div>
      </div>
      <div class="flex gap-3 mt-6">
        <button
          onclick={saveEdit}
          disabled={editSaving}
          class="flex-1 py-3 rounded-xl bg-gradient-to-br from-primary to-primary-container text-white font-semibold text-sm shadow-lg shadow-primary/20 hover:shadow-xl transition-all disabled:opacity-50 active:scale-95"
        >
          {editSaving ? '...' : $t("common.save")}
        </button>
        <button
          onclick={() => { editingCert = null; }}
          class="px-6 py-3 rounded-xl border border-outline-variant/20 text-on-surface-variant font-semibold text-sm hover:bg-surface-container transition-colors"
        >
          {$t("common.cancel")}
        </button>
      </div>
    </div>
  </div>
{/if}
