<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import DataTable from '$lib/components/DataTable.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let certs = $state<any[]>([]);
  let loading = $state(true);
  let error = $state('');
  let page = $state(1);
  let total = $state(0);
  const perPage = 20;

  function maskIIN(iin: string) {
    if (!iin || iin.length < 6) return iin || '—';
    return iin.slice(0, 4) + '****' + iin.slice(-2);
  }

  async function load() {
    loading = true;
    error = '';
    try {
      const res: any = await api.get(`/api/v1/admin/certificates?page=${page}&per_page=${perPage}`);
      certs = res.data || [];
      total = res.pagination?.total || 0;
    } catch (e: any) {
      error = e.message || 'Failed to load certificates';
      certs = [];
    } finally {
      loading = false;
    }
  }

  onMount(() => { load(); });
</script>

<div class="space-y-6">
  <h1 class="text-2xl font-display font-bold text-on-surface">Certificates</h1>
  <p class="text-on-surface-variant">Total: {total}</p>

  {#if error}
    <div class="bg-error-container text-on-error-container p-4 rounded-lg text-sm">{error}</div>
  {/if}

  {#snippet row(cert: any)}
    <td class="px-4 py-3 text-sm font-mono text-on-surface-variant">{cert.code?.slice(0, 8)}...</td>
    <td class="px-4 py-3">
      <div class="font-medium text-on-surface">{cert.name || '—'}</div>
    </td>
    <td class="px-4 py-3 text-sm">{maskIIN(cert.iin)}</td>
    <td class="px-4 py-3"><StatusBadge status={cert.status || 'valid'} /></td>
    <td class="px-4 py-3 text-sm text-on-surface-variant">{new Date(cert.created_at).toLocaleDateString()}</td>
    <td class="px-4 py-3">
      {#if cert.pdf_path}
        <a href={`/api/v1/staff/certificates/${cert.id}/download`} target="_blank" class="text-primary text-sm hover:underline">PDF</a>
      {/if}
    </td>
  {/snippet}

  <DataTable
    columns={['Code', 'Name', 'IIN', 'Status', 'Created', 'Actions']}
    items={certs}
    {loading}
    {row}
    emptyMessage="No certificates found"
  />

  {#if total > perPage}
    <div class="flex justify-center gap-2 mt-4">
      <button onclick={() => { page--; load(); }} disabled={page <= 1} class="px-3 py-1 rounded-md bg-surface-low text-sm disabled:opacity-50">Previous</button>
      <span class="px-3 py-1 text-sm text-on-surface-variant">Page {page} of {Math.ceil(total / perPage)}</span>
      <button onclick={() => { page++; load(); }} disabled={page * perPage >= total} class="px-3 py-1 rounded-md bg-surface-low text-sm disabled:opacity-50">Next</button>
    </div>
  {/if}
</div>
