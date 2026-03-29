<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { auth, isAuthenticated, isLoading, currentUser } from "$lib/stores/auth";

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

  // Role-based redirect when landing on a generic path
  $effect(() => {
    if (!initialized || !$currentUser) return;
    const path = $page.url.pathname;
    // Only redirect from exact root-level app paths
    if (path === "/") {
      const role = $currentUser.role;
      if (role === "student") goto("/student");
      else if (role === "teacher") goto("/teacher");
      else if (role === "staff") goto("/staff/events");
      else if (role === "admin") goto("/admin/organizations");
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
