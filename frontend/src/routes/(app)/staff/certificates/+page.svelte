<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { t } from '$lib/i18n';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let certs = $state<any[]>([]);
  let loading = $state(true);
  let total = $state(0);

  function maskIIN(iin: string) {
    if (!iin || iin.length < 6) return iin || '—';
    return iin.slice(0, 4) + '****' + iin.slice(-2);
  }

  async function load() {
    loading = true;
    try {
      // Load all events first, then certs from each
      const evRes: any = await api.get('/api/v1/staff/events?page=1&per_page=100');
      const events = evRes.data ?? [];
      const allCerts: any[] = [];
      for (const ev of events) {
        try {
          const certRes: any = await api.get(`/api/v1/staff/events/${ev.id}/certificates?page=1&per_page=200`);
          for (const c of (certRes.data ?? [])) {
            allCerts.push({ ...c, event_title: ev.title });
          }
        } catch {}
      }
      certs = allCerts.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
      total = certs.length;
    } catch {
      certs = [];
    } finally {
      loading = false;
    }
  }

  async function revoke(cert: any) {
    const reason = prompt($t("staff.certs.revokeReason"));
    if (reason === null) return;
    try {
      await api.post(`/api/v1/staff/certificates/${cert.id}/revoke`, { reason });
      load();
    } catch (e: any) { alert(e.message || 'Failed'); }
  }

  async function unrevoke(cert: any) {
    try {
      await api.post(`/api/v1/staff/certificates/${cert.id}/unrevoke`);
      load();
    } catch (e: any) { alert(e.message || 'Failed'); }
  }

  onMount(load);
</script>

<header class="mb-8">
  <h1 class="font-display text-2xl font-bold text-on-surface">{$t("staff.certs.title")}</h1>
  <p class="text-on-surface-variant text-sm mt-1">{total} {$t("staff.certs.total")}</p>
</header>

{#if loading}
  <div class="text-center py-16 text-on-surface-variant">{$t("common.loading")}</div>
{:else if certs.length === 0}
  <div class="text-center py-16 text-on-surface-variant">{$t("staff.certs.empty")}</div>
{:else}
  <div class="bg-surface-container-lowest rounded-2xl border border-outline-variant/10 overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-sm text-left">
        <thead class="bg-surface-container-low">
          <tr>
            <th class="px-5 py-3.5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.name")}</th>
            <th class="px-5 py-3.5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.event")}</th>
            <th class="px-5 py-3.5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.status")}</th>
            <th class="px-5 py-3.5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.date")}</th>
            <th class="px-5 py-3.5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest text-right">{$t("common.actions")}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant/10">
          {#each certs as cert}
            <tr class="hover:bg-surface-container-low/30 transition-colors">
              <td class="px-5 py-4">
                <p class="font-medium text-on-surface text-sm">{cert.name || '—'}</p>
                <p class="text-xs text-on-surface-variant font-mono mt-0.5">{maskIIN(cert.iin)}</p>
              </td>
              <td class="px-5 py-4 text-sm text-on-surface-variant max-w-[200px] truncate">{cert.event_title || '—'}</td>
              <td class="px-5 py-4"><StatusBadge status={cert.status || 'valid'} /></td>
              <td class="px-5 py-4 text-sm text-on-surface-variant">{new Date(cert.created_at).toLocaleDateString()}</td>
              <td class="px-5 py-4 text-right">
                <div class="flex gap-2 justify-end">
                  <a href="/verify/{cert.code}" class="text-primary text-xs font-medium hover:underline">{$t("verify.verify")}</a>
                  {#if cert.status === 'valid'}
                    <button onclick={() => revoke(cert)} class="text-error text-xs font-medium hover:underline">{$t("staff.certs.revoke")}</button>
                  {:else if cert.status === 'revoked'}
                    <button onclick={() => unrevoke(cert)} class="text-emerald-600 text-xs font-medium hover:underline">{$t("staff.certs.restore")}</button>
                  {/if}
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>
{/if}
