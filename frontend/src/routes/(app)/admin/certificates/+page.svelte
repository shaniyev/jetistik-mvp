<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import DataTable from '$lib/components/DataTable.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { t } from '$lib/i18n';

  let certs = $state<any[]>([]);
  let loading = $state(true);
  let error = $state('');
  let page = $state(1);
  let total = $state(0);
  const perPage = 20;

  let columns = $derived([
    { key: 'code', label: $t('common.code') },
    { key: 'name', label: $t('common.name') },
    { key: 'iin', label: 'IIN' },
    { key: 'status', label: $t('common.status') },
    { key: 'created_at', label: $t('common.created') },
    { key: 'actions', label: $t('common.actions'), class: 'text-right' },
  ]);

  let totalPages = $derived(Math.ceil(total / perPage));

  function maskIIN(iin: string) {
    if (!iin || iin.length < 6) return iin || '\u2014';
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

  async function revoke(cert: any) {
    const reason = prompt($t('admin.certs.revoke_reason'));
    if (reason === null) return;
    try {
      await api.post(`/api/v1/staff/certificates/${cert.id}/revoke`, { reason });
      load();
    } catch (e: any) {
      alert(e.message || 'Failed to revoke');
    }
  }

  onMount(() => { load(); });
</script>

<header class="flex justify-between items-end mb-10">
  <div class="space-y-1">
    <nav class="flex text-[10px] uppercase tracking-widest text-on-surface-variant/60 gap-2 mb-2">
      <a class="hover:text-primary transition-colors" href="/admin">{$t("admin.breadcrumb")}</a>
      <span>/</span>
      <span class="text-on-surface-variant">{$t("admin.certs.title")}</span>
    </nav>
    <h1 class="text-4xl font-extrabold tracking-tight text-on-surface font-display">{$t("admin.certs.title")}</h1>
    <p class="text-on-surface-variant max-w-2xl">{$t("admin.certs.subtitle")} Total: {total}</p>
  </div>
</header>

{#if error}
  <div class="bg-error-container text-on-error-container p-4 rounded-2xl text-sm mb-6 ring-1 ring-error/20">{error}</div>
{/if}

{#snippet row(cert: any, index: number)}
  <tr class="{index % 2 === 0 ? 'bg-surface-container-lowest' : 'bg-surface-container-low'} hover:bg-white transition-colors">
    <td class="px-6 py-5 text-xs font-mono text-on-surface-variant">{cert.code?.slice(0, 8)}...</td>
    <td class="px-6 py-5">
      <span class="font-semibold text-on-surface">{cert.name || '\u2014'}</span>
    </td>
    <td class="px-6 py-5 text-sm font-mono text-on-surface-variant">{maskIIN(cert.iin)}</td>
    <td class="px-6 py-5"><StatusBadge status={cert.status || 'valid'} /></td>
    <td class="px-6 py-5 text-sm text-on-surface-variant">{new Date(cert.created_at).toLocaleDateString()}</td>
    <td class="px-6 py-5 text-right">
      <a href={`/verify/${cert.code}`} target="_blank" class="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all inline-block" title="View">
        <span class="material-symbols-outlined">visibility</span>
      </a>
      <button
        onclick={() => revoke(cert)}
        disabled={cert.status === 'revoked'}
        class="p-2 text-outline hover:text-error hover:bg-error-container/30 rounded-lg transition-all disabled:opacity-30"
        title={cert.status === 'revoked' ? $t('admin.certs.alreadyRevoked') : $t('staff.certs.revoke')}
      >
        <span class="material-symbols-outlined">block</span>
      </button>
    </td>
  </tr>
{/snippet}

<DataTable
  {columns}
  data={certs}
  {loading}
  {row}
  empty={$t("admin.certs.empty")}
/>

{#if total > 0}
  <footer class="px-6 py-4 bg-surface-container-high/30 border-t border-outline-variant/10 flex items-center justify-between -mt-[1px] rounded-b-2xl">
    <p class="text-xs text-on-surface-variant">{$t("common.showing")} {(page - 1) * perPage + 1} to {Math.min(page * perPage, total)} {$t("common.of")} {total} {$t("admin.certs.entries")}</p>
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
