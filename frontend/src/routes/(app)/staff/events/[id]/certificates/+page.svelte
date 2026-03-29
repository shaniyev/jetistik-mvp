<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { api, ApiError, type PaginatedResponse } from "$lib/api/client";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";

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
  <div>
    <a href="/staff/events/{eventId}" class="text-sm text-on-surface-variant hover:text-primary transition-colors">
      &larr; Back to event
    </a>
    <h1 class="font-display text-2xl font-bold text-on-surface mt-2">Certificates</h1>
    <p class="text-sm text-on-surface-variant mt-1">{total} total certificates</p>
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
