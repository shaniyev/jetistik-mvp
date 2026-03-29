<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import DataTable from '$lib/components/DataTable.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { t } from '$lib/i18n';

  let logs = $state<any[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  let actionFilter = $state('');
  const perPage = 20;

  let columns = $derived([
    { key: 'time', label: $t('common.time') },
    { key: 'actor', label: $t('common.actor') },
    { key: 'action', label: $t('common.action') },
    { key: 'object', label: $t('common.object') },
  ]);

  const actions = ['', 'create_event', 'upload_template', 'upload_data', 'generate', 'revoke', 'unrevoke', 'certificate_edit', 'certificate_delete', 'batch_delete', 'mapping_update', 'delete_event'];

  let totalPages = $derived(Math.ceil(total / perPage));

  async function load() {
    loading = true;
    try {
      let url = `/api/v1/admin/audit-log?page=${page}&per_page=${perPage}`;
      if (actionFilter) url += `&action=${actionFilter}`;
      const res: any = await api.get(url);
      logs = res.data || [];
      total = res.pagination?.total || 0;
    } catch { logs = []; }
    finally { loading = false; }
  }

  onMount(() => { load(); });

  function applyFilter() {
    page = 1;
    load();
  }
</script>

<header class="flex justify-between items-end mb-10">
  <div class="space-y-1">
    <nav class="flex text-[10px] uppercase tracking-widest text-on-surface-variant/60 gap-2 mb-2">
      <a class="hover:text-primary transition-colors" href="/admin">{$t("admin.breadcrumb")}</a>
      <span>/</span>
      <span class="text-on-surface-variant">{$t("admin.audit.title")}</span>
    </nav>
    <h1 class="text-4xl font-extrabold tracking-tight text-on-surface font-display">{$t("admin.audit.title")}</h1>
    <p class="text-on-surface-variant max-w-2xl">{$t("admin.audit.subtitle")}</p>
  </div>
  <select
    bind:value={actionFilter}
    onchange={applyFilter}
    class="px-4 py-2.5 rounded-lg bg-surface-container-lowest text-sm text-on-surface ring-1 ring-outline-variant/30 focus:ring-2 focus:ring-primary/50 outline-none transition-all"
  >
    <option value="">{$t("admin.audit.allActions")}</option>
    {#each actions.filter(a => a) as action}
      <option value={action}>{action}</option>
    {/each}
  </select>
</header>

{#snippet row(log: any, index: number)}
  <tr class="{index % 2 === 0 ? 'bg-surface-container-lowest' : 'bg-surface-container-low'} hover:bg-white transition-colors">
    <td class="px-6 py-5 text-sm text-on-surface-variant">{new Date(log.created_at).toLocaleString()}</td>
    <td class="px-6 py-5 text-sm font-medium text-on-surface">{log.actor_id || '\u2014'}</td>
    <td class="px-6 py-5"><StatusBadge status={log.action} /></td>
    <td class="px-6 py-5 text-sm text-on-surface-variant">{log.object_type} {log.object_id}</td>
  </tr>
{/snippet}

<DataTable {columns} data={logs} {loading} {row} empty={$t("admin.audit.empty")} />

{#if total > 0}
  <footer class="px-6 py-4 bg-surface-container-high/30 border-t border-outline-variant/10 flex items-center justify-between -mt-[1px] rounded-b-2xl">
    <p class="text-xs text-on-surface-variant">{$t("common.showing")} {(page - 1) * perPage + 1} to {Math.min(page * perPage, total)} {$t("common.of")} {total} {$t("admin.audit.entries")}</p>
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
