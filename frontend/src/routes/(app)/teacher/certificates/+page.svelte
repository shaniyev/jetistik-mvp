<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { t } from '$lib/i18n';

  interface Certificate {
    code: string;
    name: string;
    event_title: string;
    org_name: string;
    status: string;
    created_at: string;
  }

  let certificates = $state<Certificate[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const res = await api.get<Certificate[]>('/api/v1/teacher/certificates');
      certificates = res.data ?? [];
    } catch (e) {
      console.error('Failed to load', e);
    } finally {
      loading = false;
    }
  });
</script>

<div class="space-y-6">
  <h1 class="font-display text-2xl font-bold text-on-surface">{$t("staff.nav.certificates")}</h1>

  {#if loading}
    <div class="text-center py-12 text-on-surface-variant">{$t("common.loading")}</div>
  {:else if certificates.length === 0}
    <div class="text-center py-12 text-on-surface-variant">{$t("student.noCertificates")}</div>
  {:else}
    <div class="grid gap-3">
      {#each certificates as cert}
        <a href="/verify/{cert.code}" class="bg-surface-container-lowest p-5 rounded-xl flex flex-col sm:flex-row sm:items-center gap-4 hover:shadow-md transition-all border border-outline-variant/5">
          <div class="w-12 h-12 rounded-lg {cert.status === 'valid' ? 'bg-primary-fixed' : 'bg-surface-container-high'} flex items-center justify-center shrink-0">
            <span class="material-symbols-outlined {cert.status === 'valid' ? 'text-primary' : 'text-outline'} text-2xl">workspace_premium</span>
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-display font-bold text-on-surface">{cert.event_title}</h3>
            <p class="text-xs text-on-surface-variant mt-0.5">{cert.org_name} &middot; {cert.name}</p>
          </div>
          <div class="flex items-center gap-3 shrink-0">
            {#if cert.status === 'valid'}
              <span class="px-3 py-1 rounded-full bg-emerald-50 text-emerald-700 text-[10px] font-bold uppercase">{$t("student.filter.valid")}</span>
            {:else}
              <span class="px-3 py-1 rounded-full bg-error-container text-on-error-container text-[10px] font-bold uppercase">{$t("student.filter.revoked")}</span>
            {/if}
            <span class="text-xs text-on-surface-variant">{new Date(cert.created_at).toLocaleDateString()}</span>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</div>
