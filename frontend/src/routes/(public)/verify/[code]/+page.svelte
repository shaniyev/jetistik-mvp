<script lang="ts">
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { page } from '$app/stores';
  import { t } from '$lib/i18n';

  let loading = $state(true);
  let result = $state<any>(null);
  let certificates = $state<any[]>([]);
  let isIIN = $state(false);
  let currentCode = $state('');

  const API_BASE = import.meta.env.VITE_API_URL || '';

  function maskIIN(iin: string) {
    if (!iin || iin.length < 6) return iin || '—';
    return iin.slice(0, 4) + '****' + iin.slice(-2);
  }

  async function loadData(code: string) {
    loading = true;
    result = null;
    certificates = [];
    currentCode = code;
    isIIN = /^\d{12}$/.test(code);

    try {
      const res = await fetch(`${API_BASE}/api/v1/verify/${code}`);
      if (!res.ok) {
        result = null;
        certificates = [];
      } else {
        const body = await res.json();
        if (isIIN) {
          certificates = Array.isArray(body.data) ? body.data : [];
        } else {
          result = body.data;
        }
      }
    } catch {
      result = null;
      certificates = [];
    }
    loading = false;
  }

  // React to URL param changes (same route, different code)
  $effect(() => {
    const code = $page.params.code;
    if (code && code !== currentCode) {
      loadData(code);
    }
  });

  // Initial load
  $effect(() => {
    if (!currentCode && $page.params.code) {
      loadData($page.params.code);
    }
  });
</script>

<svelte:head>
  <title>Verify Certificate — Jetistik</title>
</svelte:head>

<div class="min-h-screen bg-surface flex items-center justify-center px-4 py-12">
  <div class="w-full max-w-lg">
    <div class="text-center mb-8">
      <a href="/" class="inline-block">
        <h1 class="font-display text-3xl font-bold text-on-surface">Jetistik</h1>
      </a>
      <p class="text-sm text-on-surface-variant mt-1">{$t("verify.title")}</p>
    </div>

    {#if loading}
      <div class="bg-surface-lowest rounded-lg p-12 shadow-[0_4px_40px_rgba(0,74,198,0.04)] text-center">
        <p class="text-on-surface-variant">{$t("verify.loading")}</p>
      </div>

    {:else if isIIN}
      <!-- IIN search results -->
      <div class="bg-surface-lowest rounded-lg p-6 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
        <div class="flex items-center justify-between mb-6">
          <div>
            <p class="text-sm text-on-surface-variant">{$t("verify.certificatesForIIN")}</p>
            <p class="font-mono text-lg font-semibold text-on-surface mt-1">{maskIIN(currentCode)}</p>
          </div>
          <div class="flex gap-2">
            {#if certificates.length > 0}
              <a
                href="{API_BASE}/api/v1/certificates/download-zip?iin={currentCode}"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-gradient-to-br from-primary to-primary-container text-on-primary text-xs font-medium hover:opacity-90 transition-opacity"
              >
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
                </svg>
                {$t("verify.downloadAll")}
              </a>
            {/if}
            <a
              href="/"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-surface-low text-on-surface text-xs font-medium hover:bg-surface-high transition-colors"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182" />
              </svg>
              {$t("verify.changeIIN")}
            </a>
          </div>
        </div>

        {#if certificates.length === 0}
          <div class="text-center py-8">
            <div class="w-16 h-16 mx-auto rounded-full bg-surface-low flex items-center justify-center mb-3">
              <svg class="w-8 h-8 text-on-surface-variant" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m6.75 12H9m1.5-12H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
              </svg>
            </div>
            <p class="text-on-surface-variant">{$t("verify.noCertsFound")}</p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each certificates as cert}
              <div class="p-4 rounded-lg bg-surface">
                <div class="flex items-start justify-between">
                  <div class="flex-1 min-w-0">
                    <p class="font-medium text-on-surface">{cert.name || '—'}</p>
                    <p class="text-xs text-on-surface-variant mt-1">
                      {cert.org_name || ''}{cert.event_title ? ` • ${cert.event_title}` : ''}
                    </p>
                    <p class="text-xs text-on-surface-variant mt-0.5">
                      {new Date(cert.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <StatusBadge status={cert.status || 'valid'} />
                </div>
                <div class="flex items-center gap-3 mt-3 pt-3 border-t border-surface-high/50">
                  <a
                    href="{API_BASE}/api/v1/certificates/{cert.code}/download"
                    target="_blank"
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-gradient-to-br from-primary to-primary-container text-on-primary text-xs font-medium hover:opacity-90 transition-opacity"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
                    </svg>
                    PDF
                  </a>
                  <a
                    href="/verify/{cert.code}"
                    class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-surface-low text-on-surface text-xs font-medium hover:bg-surface-high transition-colors"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
                    </svg>
                    {$t("verify.verify")}
                  </a>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

    {:else if !result}
      <!-- Not found -->
      <div class="bg-surface-lowest rounded-lg p-8 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
        <div class="text-center space-y-3">
          <div class="w-16 h-16 mx-auto rounded-full bg-error-container flex items-center justify-center">
            <svg class="w-8 h-8 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="font-display text-xl font-semibold text-on-surface">{$t("verify.notFoundTitle")}</h2>
          <p class="text-sm text-on-surface-variant">
            Certificate with code <code class="font-mono bg-surface-low px-1.5 py-0.5 rounded">{currentCode}</code> was not found.
          </p>
        </div>
      </div>

    {:else if result.valid}
      <!-- Valid certificate -->
      <div class="bg-surface-lowest rounded-lg shadow-[0_4px_40px_rgba(0,74,198,0.04)] overflow-hidden">
        <!-- Status header -->
        <div class="bg-emerald-50 px-8 py-6 text-center">
          <div class="w-14 h-14 mx-auto rounded-full bg-emerald-100 flex items-center justify-center mb-3">
            <svg class="w-7 h-7 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
            </svg>
          </div>
          <span class="inline-block bg-emerald-600 text-white text-xs font-semibold px-3 py-1 rounded-full uppercase tracking-wider">{$t("verify.valid")}</span>
        </div>

        <div class="p-8 space-y-0">
          <!-- Name -->
          <div class="text-center pb-4 mb-4" style="border-bottom: 1px solid var(--color-surface-high);">
            <p class="text-xs text-on-surface-variant uppercase tracking-wider mb-1">{$t("verify.recipient")}</p>
            <p class="font-display text-xl font-bold text-on-surface">{result.name}</p>
          </div>

          <!-- IIN -->
          {#if result.iin}
            <div class="flex justify-between py-2.5">
              <span class="text-sm text-on-surface-variant">{$t("verify.iinLabel")}</span>
              <span class="text-sm font-mono font-medium text-on-surface">{maskIIN(result.iin)}</span>
            </div>
          {/if}

          <!-- Organization -->
          {#if result.org_name}
            <div class="flex justify-between py-2.5">
              <span class="text-sm text-on-surface-variant">{$t("verify.orgLabel")}</span>
              <span class="text-sm font-medium text-on-surface text-right max-w-[60%]">{result.org_name}</span>
            </div>
          {/if}

          <!-- Event -->
          {#if result.event_title}
            <div class="flex justify-between py-2.5">
              <span class="text-sm text-on-surface-variant">{$t("verify.eventLabel")}</span>
              <span class="text-sm font-medium text-on-surface text-right max-w-[60%]">{result.event_title}</span>
            </div>
          {/if}

          <!-- Payload details -->
          {#if result.payload}
            {#if result.payload.school}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.schoolLabel")}</span>
                <span class="text-sm font-medium text-on-surface text-right max-w-[60%]">{result.payload.school}</span>
              </div>
            {/if}
            {#if result.payload.class}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.classLabel")}</span>
                <span class="text-sm font-medium text-on-surface">{result.payload.class}</span>
              </div>
            {/if}
            {#if result.payload.text}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.descriptionLabel")}</span>
                <span class="text-sm font-medium text-on-surface text-right max-w-[60%]">{result.payload.text}</span>
              </div>
            {/if}
            {#if result.payload.diplom}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.typeLabel")}</span>
                <span class="text-sm font-medium text-on-surface">{result.payload.diplom}</span>
              </div>
            {/if}
            {#if result.payload.id}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.numberLabel")}</span>
                <span class="text-sm font-mono font-medium text-on-surface">{result.payload.id}</span>
              </div>
            {/if}
            {#if result.payload.data || result.payload.event_date}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.dateLabel")}</span>
                <span class="text-sm font-medium text-on-surface">{result.payload.data || result.payload.event_date}</span>
              </div>
            {/if}
            {#if result.payload.event_city}
              <div class="flex justify-between py-2.5">
                <span class="text-sm text-on-surface-variant">{$t("verify.cityLabel")}</span>
                <span class="text-sm font-medium text-on-surface">{result.payload.event_city}</span>
              </div>
            {/if}
          {/if}

          <!-- Dates & code -->
          <div class="mt-4 pt-4" style="border-top: 1px solid var(--color-surface-high);">
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">{$t("verify.issuedLabel")}</span>
              <span class="text-sm font-medium text-on-surface">
                {new Date(result.created_at).toLocaleDateString()}
              </span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">{$t("verify.codeLabel")}</span>
              <span class="text-xs font-mono text-on-surface-variant">{result.code}</span>
            </div>
          </div>

          <!-- Download button -->
          <div class="mt-4 pt-4 text-center" style="border-top: 1px solid var(--color-surface-high);">
            <a
              href="{API_BASE}/api/v1/certificates/{result.code}/download"
              class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-gradient-to-br from-primary to-primary-container text-on-primary text-sm font-medium hover:opacity-90 transition-opacity"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
              </svg>
              {$t("verify.downloadPdf")}
            </a>
          </div>
        </div>
      </div>

    {:else}
      <!-- Revoked certificate -->
      <div class="bg-surface-lowest rounded-lg p-8 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
        <div class="text-center space-y-4">
          <div class="w-16 h-16 mx-auto rounded-full bg-error-container flex items-center justify-center">
            <svg class="w-8 h-8 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
            </svg>
          </div>
          <h2 class="font-display text-xl font-semibold text-error">{$t("verify.revokedTitle")}</h2>
          <p class="text-sm text-on-surface-variant">{$t("verify.revokedDesc")}</p>
          {#if result.revoked_reason}
            <p class="text-sm text-on-surface-variant">Reason: {result.revoked_reason}</p>
          {/if}
          <div class="space-y-2 text-left mt-4">
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">{$t("verify.recipientLabel")}</span>
              <span class="text-sm font-medium text-on-surface">{result.name}</span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">{$t("verify.codeSmall")}</span>
              <span class="text-sm font-mono text-on-surface-variant">{result.code}</span>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <p class="text-center text-xs text-on-surface-variant mt-6">
      {$t("verify.poweredBy")} <a href="/" class="text-primary hover:underline">Jetistik</a>
    </p>
  </div>
</div>
