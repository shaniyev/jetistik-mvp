<script lang="ts">
  import { page } from "$app/stores";
  import { auth, currentUser } from "$lib/stores/auth";

  let { children } = $props();

  const navItems = [
    { href: "/staff/events", label: "Events", icon: "calendar" },
    { href: "/staff/audit", label: "Audit Log", icon: "shield" },
  ];

  let currentPath = $derived($page.url.pathname);
</script>

<div class="min-h-screen bg-surface flex">
  <!-- Sidebar -->
  <aside class="w-64 bg-surface-lowest flex flex-col shrink-0">
    <div class="p-6">
      <h1 class="font-display text-xl font-bold text-on-surface">Jetistik</h1>
      <p class="text-xs text-on-surface-variant mt-1">Staff Panel</p>
    </div>

    <nav class="flex-1 px-3 space-y-1">
      {#each navItems as item}
        {@const isActive = currentPath.startsWith(item.href)}
        <a
          href={item.href}
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors
            {isActive
              ? 'bg-primary/10 text-primary'
              : 'text-on-surface-variant hover:bg-surface-low hover:text-on-surface'}"
        >
          {#if item.icon === "calendar"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
            </svg>
          {:else if item.icon === "shield"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
            </svg>
          {/if}
          {item.label}
        </a>
      {/each}
    </nav>

    <div class="p-4">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary text-sm font-bold">
          {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-on-surface truncate">{$currentUser?.username ?? ""}</p>
          <p class="text-xs text-on-surface-variant capitalize">{$currentUser?.role ?? ""}</p>
        </div>
      </div>
      <button
        onclick={() => auth.logout()}
        class="mt-3 w-full text-xs text-on-surface-variant hover:text-error transition-colors text-left"
      >
        Sign out
      </button>
    </div>
  </aside>

  <!-- Main content -->
  <main class="flex-1 p-8">
    {@render children()}
  </main>
</div>
