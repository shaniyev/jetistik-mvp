<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { auth, currentUser, userRole } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import { onMount } from "svelte";

  let { children } = $props();

  let currentPath = $derived($page.url.pathname);

  const navItems = [
    { href: "/admin/organizations", labelKey: "admin.organizations" as const, icon: "building" },
    { href: "/admin/users", labelKey: "admin.users" as const, icon: "users" },
    { href: "/admin/events", labelKey: "admin.events" as const, icon: "calendar" },
    { href: "/admin/certificates", labelKey: "admin.certificates" as const, icon: "certificate" },
    { href: "/admin/audit", labelKey: "admin.auditLog" as const, icon: "shield" },
  ];

  onMount(() => {
    const unsub = userRole.subscribe((role) => {
      if (role && role !== "admin") {
        goto("/");
      }
    });
    return unsub;
  });
</script>

<div class="min-h-screen bg-surface flex">
  <!-- Sidebar -->
  <aside class="w-64 bg-surface-lowest flex flex-col shrink-0">
    <div class="p-6">
      <h1 class="font-display text-xl font-bold text-primary">Jetistik</h1>
      <p class="text-[0.65rem] uppercase tracking-widest text-on-surface-variant mt-1 font-medium">{$t("admin.title")}</p>
    </div>

    <nav class="flex-1 px-3 space-y-0.5">
      {#each navItems as item}
        {@const isActive = currentPath.startsWith(item.href)}
        <a
          href={item.href}
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors
            {isActive
              ? 'bg-primary/10 text-primary'
              : 'text-on-surface-variant hover:bg-surface-low hover:text-on-surface'}"
        >
          {#if item.icon === "building"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" />
            </svg>
          {:else if item.icon === "users"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
            </svg>
          {:else if item.icon === "calendar"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
            </svg>
          {:else if item.icon === "certificate"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
            </svg>
          {:else if item.icon === "shield"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
            </svg>
          {/if}
          {$t(item.labelKey)}
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
        {$t("admin.signout")}
      </button>
    </div>
  </aside>

  <!-- Main content -->
  <main class="flex-1 p-8 overflow-y-auto">
    <div class="text-xs text-on-surface-variant mb-6 uppercase tracking-wide">
      Admin / {#each $page.url.pathname.split('/').filter(Boolean).slice(1) as segment, i}
        {#if i > 0} / {/if}
        <span class="capitalize">{segment}</span>
      {/each}
    </div>
    {@render children()}
  </main>
</div>
