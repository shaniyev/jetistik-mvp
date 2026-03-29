<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import DataTable from '$lib/components/DataTable.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let logs = $state<any[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  let actionFilter = $state('');
  const perPage = 20;

  const actions = ['', 'create_event', 'upload_template', 'upload_data', 'generate', 'revoke', 'unrevoke', 'certificate_edit', 'certificate_delete', 'batch_delete', 'mapping_update', 'delete_event'];

  async function load() {
    loading = true;
    try {
      let url = `/api/v1/admin/audit-log?page=${page}&per_page=${perPage}`;
      if (actionFilter) url += `&action=${actionFilter}`;
      const res: any = await api.get(url);
      logs = res.data || [];
      total = res.pagination?.total || 0;
    } catch { logs = []; }
    loading = false;
  }

  onMount(() => { load(); });
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h1 class="text-2xl font-display font-bold text-on-surface">Audit Log</h1>
    <select bind:value={actionFilter} onchange={() => { page = 1; load(); }} class="px-3 py-2 rounded-lg bg-surface-lowest text-sm text-on-surface border-0 border-b-2 border-outline-variant focus:border-primary outline-none">
      <option value="">All Actions</option>
      {#each actions.filter(a => a) as action}
        <option value={action}>{action}</option>
      {/each}
    </select>
  </div>

  {#snippet row(log: any)}
    <td class="px-4 py-3 text-sm text-on-surface-variant">{new Date(log.created_at).toLocaleString()}</td>
    <td class="px-4 py-3 text-sm">{log.actor_id || '—'}</td>
    <td class="px-4 py-3"><StatusBadge status={log.action} /></td>
    <td class="px-4 py-3 text-sm text-on-surface-variant">{log.object_type} {log.object_id}</td>
  {/snippet}

  <DataTable
    columns={['Time', 'Actor', 'Action', 'Object']}
    items={logs}
    {loading}
    {row}
    emptyMessage="No audit logs found"
  />

  {#if total > perPage}
    <div class="flex justify-center gap-2 mt-4">
      <button onclick={() => { page--; }} disabled={page <= 1} class="px-3 py-1 rounded-md bg-surface-low text-sm disabled:opacity-50">Previous</button>
      <span class="px-3 py-1 text-sm text-on-surface-variant">Page {page}</span>
      <button onclick={() => { page++; }} disabled={page * perPage >= total} class="px-3 py-1 rounded-md bg-surface-low text-sm disabled:opacity-50">Next</button>
    </div>
  {/if}
</div>
