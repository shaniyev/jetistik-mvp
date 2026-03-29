<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { t } from '$lib/i18n';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let certs = $state<any[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  const perPage = 20;

  function maskIIN(iin: string) {
    if (!iin || iin.length < 6) return iin || '—';
    return iin.slice(0, 4) + '****' + iin.slice(-2);
  }

  async function load() {
    loading = true;
    try {
      const res: any = await api.get(`/api/v1/staff/events`);
      const events = res.data ?? [];
      // Load certificates from all events
      const allCerts: any[] = [];
      for (const ev of events) {
        try {
          const certRes: any = await api.get(`/api/v1/staff/events/${ev.id}/certificates?page=1&per_page=100`);
          const eventCerts = (certRes.data ?? []).map((c: any) => ({ ...c, event_title: ev.title }));
          allCerts.push(...eventCerts);
        } catch {}
      }
      certs = allCerts;
      total = allCerts.length;
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
    } catch (e: any) {
      alert(e.message || 'Failed');
    }
  }

  async function unrevoke(cert: any) {
    try {
      await api.post(`/api/v1/staff/certificates/${cert.id}/unrevoke`);
      load();
    } catch (e: any) {
      alert(e.message || 'Failed');
    }
  }

  onMount(load);
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="font-display text-2xl font-bold text-on-surface">{$t("staff.certs.title")}</h1>
      <p class="text-on-surface-variant text-sm mt-1">{total} {$t("staff.certs.total")}</p>
    </div>
  </div>

  {#if loading}
    <div class="text-center py-12 text-on-surface-variant">{$t("common.loading")}</div>
  {:else if certs.length === 0}
    <div class="text-center py-12 text-on-surface-variant">{$t("staff.certs.empty")}</div>
  {:else}
    <div class="bg-surface-container-lowest rounded-2xl border border-outline-variant/10 overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-surface-container-low">
          <tr>
            <th class="px-6 py-4 text-left text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.name")}</th>
            <th class="px-6 py-4 text-left text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.event")}</th>
            <th class="px-6 py-4 text-left text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.status")}</th>
            <th class="px-6 py-4 text-left text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.date")}</th>
            <th class="px-6 py-4 text-right text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("common.actions")}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant/10">
          {#each certs as cert}
            <tr class="hover:bg-surface-container-low/30 transition-colors">
              <td class="px-6 py-4">
                <div class="font-medium text-on-surface">{cert.name || '—'}</div>
                <div class="text-xs text-on-surface-variant font-mono">{maskIIN(cert.iin)}</div>
              </td>
              <td class="px-6 py-4 text-on-surface-variant">{cert.event_title || '—'}</td>
              <td class="px-6 py-4"><StatusBadge status={cert.status || 'valid'} /></td>
              <td class="px-6 py-4 text-on-surface-variant">{new Date(cert.created_at).toLocaleDateString()}</td>
              <td class="px-6 py-4 text-right">
                <div class="flex gap-2 justify-end">
                  <a href="/verify/{cert.code}" class="text-primary text-xs hover:underline">{$t("verify.verify")}</a>
                  {#if cert.status === 'valid'}
                    <button onclick={() => revoke(cert)} class="text-error text-xs hover:underline">{$t("staff.certs.revoke")}</button>
                  {:else if cert.status === 'revoked'}
                    <button onclick={() => unrevoke(cert)} class="text-emerald-600 text-xs hover:underline">{$t("staff.certs.restore")}</button>
                  {/if}
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>
