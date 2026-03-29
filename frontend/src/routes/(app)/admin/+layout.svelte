<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { auth, currentUser, userRole } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import { onMount } from "svelte";

  let { children } = $props();

  let currentPath = $derived($page.url.pathname);

  const navItems = [
    { href: "/admin/organizations", labelKey: "admin.organizations" as const, icon: "corporate_fare" },
    { href: "/admin/users", labelKey: "admin.users" as const, icon: "group" },
    { href: "/admin/events", labelKey: "admin.events" as const, icon: "event" },
    { href: "/admin/certificates", labelKey: "admin.certificates" as const, icon: "workspace_premium" },
    { href: "/admin/audit", labelKey: "admin.auditLog" as const, icon: "history_edu" },
  ];

  function getPageTitle(path: string): string {
    const segments = path.split('/').filter(Boolean).slice(1);
    return segments.map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' / ');
  }

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
  <aside class="h-screen w-64 fixed left-0 top-0 bg-slate-50 flex flex-col p-4 gap-2 border-r border-slate-100 z-40">
    <div class="mb-8 px-2 py-4">
      <a href="/" class="text-2xl font-bold tracking-tighter text-primary font-display hover:opacity-80 transition-opacity">Jetistik</a>
      <p class="text-[10px] text-on-surface-variant uppercase tracking-[0.2em] mt-1">{$t("admin.title")}</p>
    </div>

    <nav class="flex-1 space-y-1">
      {#each navItems as item}
        {@const isActive = currentPath.startsWith(item.href)}
        <a
          href={item.href}
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 ease-in-out
            {isActive
              ? 'bg-blue-50 text-blue-700'
              : 'text-slate-500 hover:text-slate-900 hover:bg-slate-100'}"
        >
          <span class="material-symbols-outlined text-[20px]">{item.icon}</span>
          <span>{$t(item.labelKey)}</span>
        </a>
      {/each}
    </nav>

    <div class="mt-auto pt-6 flex items-center gap-3 px-3 border-t border-slate-200/50 py-4">
      <div class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary text-sm font-bold">
        {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
      </div>
      <div class="flex flex-col flex-1 min-w-0">
        <span class="text-xs font-bold text-on-surface truncate">{$currentUser?.username ?? ""}</span>
        <span class="text-[10px] text-on-surface-variant capitalize">{$currentUser?.role ?? ""}</span>
      </div>
      <button
        onclick={() => auth.logout()}
        class="ml-auto text-slate-400 hover:text-error transition-colors"
        title={$t("admin.signout")}
      >
        <span class="material-symbols-outlined text-[20px]">logout</span>
      </button>
    </div>
  </aside>

  <!-- Main content -->
  <main class="ml-64 flex-1 min-h-screen p-8 bg-surface">
    {@render children()}
  </main>
</div>
