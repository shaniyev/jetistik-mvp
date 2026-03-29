<script lang="ts">
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  interface VerifyResult {
    valid: boolean;
    code: string;
    name: string;
    event_title: string;
    org_name: string;
    status: string;
    revoked_reason: string;
    created_at: string;
  }

  let result = $derived(data.result as VerifyResult | null);
  let code = $derived(data.code);
</script>

<svelte:head>
  <title>Verify Certificate — Jetistik</title>
</svelte:head>

<div class="min-h-screen bg-surface flex items-center justify-center px-4 py-12">
  <div class="w-full max-w-md">
    <div class="text-center mb-8">
      <h1 class="font-display text-3xl font-bold text-on-surface">Jetistik</h1>
      <p class="text-sm text-on-surface-variant mt-1">Certificate Verification</p>
    </div>

    <div class="bg-surface-lowest rounded-lg p-8 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
      {#if !result}
        <div class="text-center space-y-3">
          <div class="w-16 h-16 mx-auto rounded-full bg-error-container flex items-center justify-center">
            <svg class="w-8 h-8 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 class="font-display text-xl font-semibold text-on-surface">Not Found</h2>
          <p class="text-sm text-on-surface-variant">
            Certificate with code <code class="font-mono bg-surface-low px-1.5 py-0.5 rounded">{code}</code> was not found.
          </p>
        </div>
      {:else if result.valid}
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
      {:else}
        <div class="text-center space-y-4">
          <div class="w-16 h-16 mx-auto rounded-full bg-error-container flex items-center justify-center">
            <svg class="w-8 h-8 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
            </svg>
          </div>
          <h2 class="font-display text-xl font-semibold text-error">Revoked Certificate</h2>
          <p class="text-sm text-on-surface-variant">
            This certificate has been revoked.
          </p>
          {#if result.revoked_reason}
            <p class="text-sm text-on-surface-variant">
              Reason: {result.revoked_reason}
            </p>
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
      {/if}
    </div>

    <p class="text-center text-xs text-on-surface-variant mt-6">
      Powered by <a href="/" class="text-primary hover:underline">Jetistik</a>
    </p>
  </div>
</div>
