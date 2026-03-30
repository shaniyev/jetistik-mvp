<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { api, ApiError, getAccessToken, type PaginatedResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";
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

  let columns = $derived([
    { key: "name", label: $t("common.name") },
    { key: "iin", label: $t("auth.iin") },
    { key: "code", label: $t("common.code") },
    { key: "status", label: $t("common.status") },
    { key: "created_at", label: $t("common.created") },
    { key: "actions", label: "", class: "w-32" },
  ]);
</script>

<div class="space-y-8">
  <!-- Header -->
  <div>
    <a href="/staff/events/{eventId}" class="inline-flex items-center gap-1 text-primary text-sm font-medium hover:underline mb-3">
      <span class="material-symbols-outlined text-sm">arrow_back</span>
      {$t("staff.event.backToEvents")}
    </a>
    <div class="flex items-end justify-between">
      <div>
        <h1 class="font-display text-3xl font-extrabold tracking-tight text-on-surface">{$t("staff.certs.title")}</h1>
        <p class="text-on-surface-variant mt-1">{total} {$t("staff.certs.total")}</p>
      </div>
      {#if total > 0}
        <button
          onclick={downloadAll}
          disabled={downloading}
          class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold
                 bg-gradient-to-br from-primary to-primary-container text-on-primary
                 hover:shadow-lg hover:shadow-primary/20 transition-all active:scale-95 disabled:opacity-50"
        >
          <span class="material-symbols-outlined text-lg">download</span>
          {downloading ? downloadProgress : $t("staff.certs.downloadAll")}
        </button>
      {/if}
    </div>
  </div>

  {#if error}
    <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm">{error}</div>
  {/if}

  <DataTable {columns} data={certs} {loading} empty={$t("staff.certs.empty")}>
    {#snippet row(cert: Certificate)}
      <tr class="hover:bg-surface-low/50 transition-colors">
        <td class="px-4 py-3 text-sm text-on-surface">{cert.name}</td>
        <td class="px-4 py-3 text-sm text-on-surface-variant font-mono">{cert.iin}</td>
        <td class="px-4 py-3 text-sm text-on-surface-variant font-mono">{cert.code}</td>
        <td class="px-4 py-3">
          <StatusBadge status={cert.status} />
        </td>
        <td class="px-4 py-3 text-xs text-on-surface-variant">
          {new Date(cert.created_at).toLocaleDateString()}
        </td>
        <td class="px-4 py-3">
          <div class="flex items-center gap-2">
            <button onclick={() => downloadPdf(cert)} class="text-xs text-primary hover:underline font-medium">
              PDF
            </button>
            <button onclick={() => startEdit(cert)} class="text-xs text-on-surface-variant hover:text-primary hover:underline font-medium">
              {$t("common.edit")}
            </button>
            {#if cert.status === "valid"}
              <button onclick={() => revoke(cert.id)} class="text-xs text-error hover:underline font-medium">
                {$t("staff.certs.revoke")}
              </button>
            {:else if cert.status === "revoked"}
              <button onclick={() => unrevoke(cert.id)} class="text-xs text-emerald-600 hover:underline font-medium">
                {$t("staff.certs.restore")}
              </button>
            {/if}
          </div>
        </td>
      </tr>
    {/snippet}
  </DataTable>

  {#if total > perPage}
    <div class="flex items-center justify-between text-sm text-on-surface-variant">
      <span>{$t("staff.certs.page_of")} {currentPage} / {Math.ceil(total / perPage)}</span>
      <div class="flex gap-2">
        <button
          disabled={currentPage <= 1}
          onclick={() => { currentPage--; loadCerts(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          {$t("common.previous")}
        </button>
        <button
          disabled={currentPage * perPage >= total}
          onclick={() => { currentPage++; loadCerts(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
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
    <div class="bg-surface-container-lowest rounded-2xl p-8 w-full max-w-md shadow-2xl" onclick={(e) => e.stopPropagation()}>
      <h3 class="font-display text-xl font-bold text-on-surface mb-6">{$t("common.edit")} — {editingCert.code.slice(0, 8)}</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-1.5">{$t("common.name")}</label>
          <input
            type="text"
            bind:value={editName}
            class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm focus:border-primary focus:ring-1 focus:ring-primary/20 outline-none"
          />
        </div>
        <div>
          <label class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-1.5">IIN</label>
          <input
            type="text"
            bind:value={editIIN}
            maxlength="12"
            inputmode="numeric"
            class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm font-mono focus:border-primary focus:ring-1 focus:ring-primary/20 outline-none"
          />
        </div>
      </div>
      <div class="flex gap-3 mt-6">
        <button
          onclick={saveEdit}
          disabled={editSaving}
          class="flex-1 py-3 rounded-xl bg-gradient-to-br from-primary to-primary-container text-on-primary font-semibold text-sm hover:shadow-lg transition-all disabled:opacity-50"
        >
          {editSaving ? '...' : $t("common.save")}
        </button>
        <button
          onclick={() => { editingCert = null; }}
          class="px-6 py-3 rounded-xl bg-surface-container-low text-on-surface-variant font-semibold text-sm hover:bg-surface-container-high transition-colors"
        >
          {$t("common.cancel")}
        </button>
      </div>
    </div>
  </div>
{/if}
