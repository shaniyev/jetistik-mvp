<script lang="ts">
  import { api } from '$lib/api/client';
  import DataTable from '$lib/components/DataTable.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let events = $state<any[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  const perPage = 20;

  async function load() {
    loading = true;
    try {
      const res: any = await api.get(`/api/v1/admin/events?page=${page}&per_page=${perPage}`);
      events = res.data || [];
      total = res.pagination?.total || 0;
    } catch { events = []; }
    loading = false;
  }

  $effect(() => { load(); });
</script>

<div class="space-y-6">
  <h1 class="text-2xl font-display font-bold text-on-surface">Events</h1>

  {#snippet row(event: any)}
    <td class="px-4 py-3 text-sm font-medium">{event.id}</td>
    <td class="px-4 py-3">
      <div class="font-medium text-on-surface">{event.title}</div>
      <div class="text-xs text-on-surface-variant">{event.city || ''}</div>
    </td>
    <td class="px-4 py-3"><StatusBadge status={event.status || 'active'} /></td>
    <td class="px-4 py-3 text-sm text-on-surface-variant">{event.date || '—'}</td>
    <td class="px-4 py-3 text-sm text-on-surface-variant">{new Date(event.created_at).toLocaleDateString()}</td>
  {/snippet}

  <DataTable
    columns={['ID', 'Event', 'Status', 'Date', 'Created']}
    items={events}
    {loading}
    {row}
    emptyMessage="No events found"
  />

  {#if total > perPage}
    <div class="flex justify-center gap-2 mt-4">
      <button onclick={() => { page--; }} disabled={page <= 1} class="px-3 py-1 rounded-md bg-surface-low text-sm disabled:opacity-50">Previous</button>
      <span class="px-3 py-1 text-sm text-on-surface-variant">Page {page}</span>
      <button onclick={() => { page++; }} disabled={page * perPage >= total} class="px-3 py-1 rounded-md bg-surface-low text-sm disabled:opacity-50">Next</button>
    </div>
  {/if}
</div>
