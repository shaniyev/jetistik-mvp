<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { auth, currentUser } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  let { children } = $props();

  $effect(() => {
    if ($currentUser && $currentUser.role !== "teacher") {
      goto("/");
    }
  });

  const navItems = [
    { href: "/teacher", label: "teacher.nav.students" as const, icon: "users" },
    { href: "/teacher/certificates", label: "teacher.nav.certificates" as const, icon: "certificate" },
  ];

  let currentPath = $derived($page.url.pathname);
</script>

<div class="min-h-screen bg-surface flex">
  <!-- Sidebar -->
  <aside class="w-64 bg-surface-lowest flex-col shrink-0 hidden md:flex">
    <div class="p-6">
      <div class="flex items-center gap-2">
        <div class="w-7 h-7 rounded-md bg-gradient-to-br from-primary to-primary-container flex items-center justify-center">
          <svg class="w-4 h-4 text-on-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 7.74-3.342" />
          </svg>
        </div>
        <div>
          <h1 class="font-display text-base font-bold text-on-surface">Jetistik</h1>
          <p class="text-[10px] text-on-surface-variant uppercase tracking-wider">{$t("teacher.title")}</p>
        </div>
      </div>
    </div>

    <nav class="flex-1 px-3 space-y-1">
      {#each navItems as item}
        {@const isActive = currentPath === item.href || (item.href !== "/teacher" && currentPath.startsWith(item.href))}
        <a
          href={item.href}
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors
            {isActive
              ? 'bg-primary/10 text-primary'
              : 'text-on-surface-variant hover:bg-surface-low hover:text-on-surface'}"
        >
          {#if item.icon === "users"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
            </svg>
          {:else if item.icon === "certificate"}
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
            </svg>
          {/if}
          {$t(item.label)}
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
          <p class="text-xs text-on-surface-variant">{$t("teacher.role")}</p>
        </div>
      </div>
      <button
        onclick={() => auth.logout()}
        class="mt-3 w-full text-xs text-on-surface-variant hover:text-error transition-colors text-left"
      >
        {$t("teacher.signout")}
      </button>
    </div>
  </aside>

  <!-- Mobile header -->
  <div class="md:hidden fixed top-0 left-0 right-0 z-30 bg-surface-lowest/80 backdrop-blur-xl">
    <div class="px-4 h-14 flex items-center justify-between">
      <span class="font-display text-lg font-bold text-on-surface">Jetistik</span>
      <div class="flex items-center gap-3">
        {#each navItems as item}
          {@const isActive = currentPath === item.href || (item.href !== "/teacher" && currentPath.startsWith(item.href))}
          <a
            href={item.href}
            class="text-xs font-medium px-2 py-1 rounded transition-colors {isActive ? 'text-primary' : 'text-on-surface-variant'}"
          >
            {$t(item.label)}
          </a>
        {/each}
        <button
          onclick={() => auth.logout()}
          class="text-xs text-on-surface-variant hover:text-error transition-colors"
        >
          {$t("nav.logout")}
        </button>
      </div>
    </div>
  </div>

  <!-- Main content -->
  <main class="flex-1 p-4 sm:p-8 md:pt-8 pt-18">
    {@render children()}
  </main>
</div>
