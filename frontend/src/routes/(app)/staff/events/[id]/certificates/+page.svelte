<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { api, ApiError, getAccessToken, type PaginatedResponse } from "$lib/api/client";
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
      error = "Failed to load certificates";
    } finally {
      loading = false;
    }
  }

  async function revoke(id: number) {
    const reason = prompt("Revoke reason:");
    if (!reason) return;
    try {
      await api.post(`/api/v1/staff/certificates/${id}/revoke`, { reason });
      loadCerts();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to revoke");
    }
  }

  async function unrevoke(id: number) {
    try {
      await api.post(`/api/v1/staff/certificates/${id}/unrevoke`);
      loadCerts();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to unrevoke");
    }
  }

  async function downloadAll() {
    downloading = true;
    downloadProgress = "Fetching certificates...";
    try {
      const apiBase = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
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
        downloadProgress = `Downloading ${done + 1} / ${allCerts.length}...`;
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

      downloadProgress = "Creating ZIP...";
      const zipBlob = await zip.generateAsync({ type: "blob" });
      const a = document.createElement("a");
      a.href = URL.createObjectURL(zipBlob);
      a.download = `certificates_event_${eventId}.zip`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(a.href);
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to download certificates");
    } finally {
      downloading = false;
      downloadProgress = "";
    }
  }

  onMount(loadCerts);

  const columns = [
    { key: "name", label: "Name" },
    { key: "iin", label: "IIN" },
    { key: "code", label: "Code" },
    { key: "status", label: "Status" },
    { key: "created_at", label: "Created" },
    { key: "actions", label: "", class: "w-32" },
  ];
</script>

<div class="space-y-6">
  <div class="flex items-start justify-between">
    <div>
      <a href="/staff/events/{eventId}" class="text-sm text-on-surface-variant hover:text-primary transition-colors">
        &larr; Back to event
      </a>
      <h1 class="font-display text-2xl font-bold text-on-surface mt-2">Certificates</h1>
      <p class="text-sm text-on-surface-variant mt-1">{total} total certificates</p>
    </div>
    {#if total > 0}
      <button
        onclick={downloadAll}
        disabled={downloading}
        class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium
               bg-gradient-to-br from-primary to-primary-container text-on-primary
               hover:shadow-lg transition-shadow disabled:opacity-50"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
        </svg>
        {downloading ? downloadProgress : "Download All"}
      </button>
    {/if}
  </div>

  {#if error}
    <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm">{error}</div>
  {/if}

  <DataTable {columns} data={certs} {loading} empty="No certificates generated yet.">
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
            <a
              href="/api/v1/staff/certificates/{cert.id}/download"
              target="_blank"
              class="text-xs text-primary hover:underline"
            >
              PDF
            </a>
            {#if cert.status === "valid"}
              <button onclick={() => revoke(cert.id)} class="text-xs text-error hover:underline">
                Revoke
              </button>
            {:else if cert.status === "revoked"}
              <button onclick={() => unrevoke(cert.id)} class="text-xs text-emerald-600 hover:underline">
                Restore
              </button>
            {/if}
          </div>
        </td>
      </tr>
    {/snippet}
  </DataTable>

  {#if total > perPage}
    <div class="flex items-center justify-between text-sm text-on-surface-variant">
      <span>Page {currentPage} of {Math.ceil(total / perPage)}</span>
      <div class="flex gap-2">
        <button
          disabled={currentPage <= 1}
          onclick={() => { currentPage--; loadCerts(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          Previous
        </button>
        <button
          disabled={currentPage * perPage >= total}
          onclick={() => { currentPage++; loadCerts(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  {/if}
</div>
