<script lang="ts">
  import { onMount } from 'svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';

  let { data }: { data: { code: string } } = $props();

  let loading = $state(true);
  let result = $state<any>(null);
  let certificates = $state<any[]>([]);
  let isIIN = $state(false);

  const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

  function maskIIN(iin: string) {
    if (!iin || iin.length < 6) return iin || '—';
    return iin.slice(0, 4) + '****' + iin.slice(-2);
  }

  onMount(async () => {
    const code = data.code;
    isIIN = /^\d{12}$/.test(code);

    try {
      const res = await fetch(`${API_BASE}/api/v1/verify/${code}`);
      if (!res.ok) {
        result = null;
        certificates = [];
      } else {
        const body = await res.json();
        if (isIIN) {
          // IIN search returns array of certificates
          certificates = Array.isArray(body.data) ? body.data : [];
        } else {
          // Code verification returns single result
          result = body.data;
        }
      }
    } catch {
      result = null;
      certificates = [];
    }
    loading = false;
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
      <p class="text-sm text-on-surface-variant mt-1">Certificate Verification</p>
    </div>

    {#if loading}
      <div class="bg-surface-lowest rounded-lg p-12 shadow-[0_4px_40px_rgba(0,74,198,0.04)] text-center">
        <p class="text-on-surface-variant">Loading...</p>
      </div>

    {:else if isIIN}
      <!-- IIN search results -->
      <div class="bg-surface-lowest rounded-lg p-6 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
        <div class="text-center mb-6">
          <p class="text-sm text-on-surface-variant">Certificates for IIN</p>
          <p class="font-mono text-lg font-semibold text-on-surface mt-1">{maskIIN(data.code)}</p>
        </div>

        {#if certificates.length === 0}
          <div class="text-center py-8">
            <div class="w-16 h-16 mx-auto rounded-full bg-surface-low flex items-center justify-center mb-3">
              <svg class="w-8 h-8 text-on-surface-variant" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m6.75 12H9m1.5-12H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
              </svg>
            </div>
            <p class="text-on-surface-variant">No certificates found for this IIN</p>
          </div>
        {:else}
          <div class="space-y-3">
            {#each certificates as cert}
              <a
                href="/verify/{cert.code}"
                class="block p-4 rounded-lg bg-surface hover:bg-surface-low transition-colors"
              >
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-medium text-on-surface">{cert.name || '—'}</p>
                    <p class="text-xs text-on-surface-variant mt-1">
                      {cert.event_title || ''} {cert.org_name ? `• ${cert.org_name}` : ''}
                    </p>
                  </div>
                  <StatusBadge status={cert.status || 'valid'} />
                </div>
                <div class="flex items-center justify-between mt-2">
                  <span class="text-xs text-on-surface-variant font-mono">{cert.code?.slice(0, 12)}...</span>
                  <span class="text-xs text-on-surface-variant">{new Date(cert.created_at).toLocaleDateString()}</span>
                </div>
              </a>
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
          <h2 class="font-display text-xl font-semibold text-on-surface">Not Found</h2>
          <p class="text-sm text-on-surface-variant">
            Certificate with code <code class="font-mono bg-surface-low px-1.5 py-0.5 rounded">{data.code}</code> was not found.
          </p>
        </div>
      </div>

    {:else if result.valid}
      <!-- Valid certificate -->
      <div class="bg-surface-lowest rounded-lg p-8 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
        <div class="text-center space-y-4">
          <div class="w-16 h-16 mx-auto rounded-full bg-emerald-50 flex items-center justify-center">
            <svg class="w-8 h-8 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
            </svg>
          </div>
          <h2 class="font-display text-xl font-semibold text-emerald-700">Valid Certificate</h2>

          <div class="space-y-3 text-left mt-6">
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">Recipient</span>
              <span class="text-sm font-medium text-on-surface">{result.name}</span>
            </div>
            {#if result.iin}
              <div class="flex justify-between py-2">
                <span class="text-sm text-on-surface-variant">IIN</span>
                <span class="text-sm font-mono text-on-surface">{maskIIN(result.iin)}</span>
              </div>
            {/if}
            {#if result.event_title}
              <div class="flex justify-between py-2">
                <span class="text-sm text-on-surface-variant">Event</span>
                <span class="text-sm font-medium text-on-surface">{result.event_title}</span>
              </div>
            {/if}
            {#if result.org_name}
              <div class="flex justify-between py-2">
                <span class="text-sm text-on-surface-variant">Organization</span>
                <span class="text-sm font-medium text-on-surface">{result.org_name}</span>
              </div>
            {/if}
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">Issued</span>
              <span class="text-sm font-medium text-on-surface">
                {new Date(result.created_at).toLocaleDateString()}
              </span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">Code</span>
              <span class="text-sm font-mono text-on-surface-variant">{result.code}</span>
            </div>
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
          <h2 class="font-display text-xl font-semibold text-error">Revoked Certificate</h2>
          <p class="text-sm text-on-surface-variant">This certificate has been revoked.</p>
          {#if result.revoked_reason}
            <p class="text-sm text-on-surface-variant">Reason: {result.revoked_reason}</p>
          {/if}
          <div class="space-y-2 text-left mt-4">
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">Recipient</span>
              <span class="text-sm font-medium text-on-surface">{result.name}</span>
            </div>
            <div class="flex justify-between py-2">
              <span class="text-sm text-on-surface-variant">Code</span>
              <span class="text-sm font-mono text-on-surface-variant">{result.code}</span>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <p class="text-center text-xs text-on-surface-variant mt-6">
      Powered by <a href="/" class="text-primary hover:underline">Jetistik</a>
    </p>
  </div>
</div>
