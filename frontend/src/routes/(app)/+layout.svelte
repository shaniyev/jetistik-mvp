<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth, isAuthenticated, isLoading } from "$lib/stores/auth";

  let { children } = $props();
  let initialized = $state(false);

  onMount(async () => {
    // Try to refresh token on initial load
    const ok = await auth.refresh();
    if (!ok) {
      goto("/login");
      return;
    }
    initialized = true;
  });

  // Redirect if auth state changes to unauthenticated after init
  $effect(() => {
    if (initialized && !$isLoading && !$isAuthenticated) {
      goto("/login");
    }
  });
</script>

{#if !initialized}
  <div class="min-h-screen bg-surface flex items-center justify-center">
    <div class="text-on-surface-variant">Loading...</div>
  </div>
{:else}
  {@render children()}
{/if}
