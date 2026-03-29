<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import DataTable from '$lib/components/DataTable.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let events = $state<any[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  const perPage = 20;

  const columns = [
    { key: 'id', label: 'ID', class: 'w-20' },
    { key: 'title', label: 'Event' },
    { key: 'status', label: 'Status' },
    { key: 'date', label: 'Date' },
    { key: 'created_at', label: 'Created' },
  ];

  let totalPages = $derived(Math.ceil(total / perPage));

  async function load() {
    loading = true;
    try {
      const res: any = await api.get(`/api/v1/admin/events?page=${page}&per_page=${perPage}`);
      events = res.data || [];
      total = res.pagination?.total || 0;
    } catch { events = []; }
    finally { loading = false; }
  }

  onMount(() => { load(); });
</script>

<header class="flex justify-between items-end mb-10">
  <div class="space-y-1">
    <nav class="flex text-[10px] uppercase tracking-widest text-on-surface-variant/60 gap-2 mb-2">
      <a class="hover:text-primary transition-colors" href="/admin">Admin</a>
      <span>/</span>
      <span class="text-on-surface-variant">Events</span>
    </nav>
    <h1 class="text-4xl font-extrabold tracking-tight text-on-surface font-display">Events</h1>
    <p class="text-on-surface-variant max-w-2xl">View and manage all events across organizations.</p>
  </div>
</header>

{#snippet row(event: any, index: number)}
  <tr class="{index % 2 === 0 ? 'bg-surface-container-lowest' : 'bg-surface-container-low'} hover:bg-white transition-colors">
    <td class="px-6 py-5 text-xs font-mono text-on-surface-variant">{event.id}</td>
    <td class="px-6 py-5">
      <div class="flex flex-col">
        <span class="font-semibold text-on-surface">{event.title}</span>
        <span class="text-xs text-on-surface-variant">{event.city || ''}</span>
      </div>
    </td>
    <td class="px-6 py-5"><StatusBadge status={event.status || 'active'} /></td>
    <td class="px-6 py-5 text-sm text-on-surface-variant">{event.date || '\u2014'}</td>
    <td class="px-6 py-5 text-sm text-on-surface-variant">{new Date(event.created_at).toLocaleDateString()}</td>
  </tr>
{/snippet}

<DataTable {columns} data={events} {loading} {row} empty="No events found" />

{#if total > 0}
  <footer class="px-6 py-4 bg-surface-container-high/30 border-t border-outline-variant/10 flex items-center justify-between -mt-[1px] rounded-b-2xl">
    <p class="text-xs text-on-surface-variant">Showing {(page - 1) * perPage + 1} to {Math.min(page * perPage, total)} of {total} entries</p>
    <div class="flex items-center gap-1">
      <button disabled={page <= 1} onclick={() => { page--; load(); }} class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-surface-container transition-colors text-outline disabled:opacity-30">
        <span class="material-symbols-outlined text-[18px]">chevron_left</span>
      </button>
      {#each Array.from({ length: Math.min(3, totalPages) }, (_, i) => i + 1) as p}
        <button onclick={() => { page = p; load(); }} class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium transition-colors {p === page ? 'bg-primary text-white font-bold' : 'hover:bg-surface-container text-on-surface'}">
          {p}
        </button>
      {/each}
      {#if totalPages > 4}
        <span class="px-2 text-outline text-xs">...</span>
        <button onclick={() => { page = totalPages; load(); }} class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium transition-colors {totalPages === page ? 'bg-primary text-white font-bold' : 'hover:bg-surface-container text-on-surface'}">
          {totalPages}
        </button>
      {/if}
      <button disabled={page * perPage >= total} onclick={() => { page++; load(); }} class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-surface-container transition-colors text-outline disabled:opacity-30">
        <span class="material-symbols-outlined text-[18px]">chevron_right</span>
      </button>
    </div>
  </footer>
{/if}
