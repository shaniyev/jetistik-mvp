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
    { href: "/teacher", label: "teacher.nav.students" as const, icon: "group" },
    { href: "/teacher/certificates", label: "teacher.nav.certificates" as const, icon: "verified" },
  ];

  let currentPath = $derived($page.url.pathname);
  let sidebarOpen = $state(false);

  function closeSidebar() {
    sidebarOpen = false;
  }
</script>

<div class="min-h-screen bg-surface font-body text-on-surface">
  <!-- Sidebar (Desktop always visible, Mobile toggleable) -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  {#if sidebarOpen}
    <div class="lg:hidden fixed inset-0 bg-black/30 z-40" onclick={closeSidebar}></div>
  {/if}

  <aside class="h-screen w-64 fixed left-0 top-0 bg-slate-50 flex flex-col py-6 z-50 transition-transform duration-300
    {sidebarOpen ? 'translate-x-0' : '-translate-x-full'} lg:translate-x-0">
    <div class="px-6 mb-10 flex justify-between items-center">
      <div>
        <h1 class="text-xl font-bold text-blue-700 font-display tracking-tight">Sovereign Ledger</h1>
        <p class="text-xs font-semibold text-slate-500 uppercase tracking-widest mt-1">{$t("teacher.title")}</p>
      </div>
      <button class="lg:hidden text-slate-500 p-1" onclick={closeSidebar}>
        <span class="material-symbols-outlined">close</span>
      </button>
    </div>

    <nav class="flex-1 px-3 space-y-1">
      {#each navItems as item}
        {@const isActive = currentPath === item.href || (item.href !== "/teacher" && currentPath.startsWith(item.href))}
        <a
          href={item.href}
          onclick={closeSidebar}
          class="flex items-center px-3 py-3 transition-colors duration-200
            {isActive
              ? 'text-blue-700 font-bold border-r-4 border-blue-600 bg-blue-50/50'
              : 'text-slate-500 hover:text-blue-600 hover:bg-slate-100'}"
        >
          <span class="material-symbols-outlined mr-3">{item.icon}</span>
          <span class="font-display tracking-tight">{$t(item.label)}</span>
        </a>
      {/each}
    </nav>

    <div class="px-3 pt-6 border-t border-slate-100">
      <a class="flex items-center px-3 py-3 text-slate-500 hover:text-blue-600 hover:bg-slate-100 transition-colors duration-200" href="/teacher">
        <span class="material-symbols-outlined mr-3">settings</span>
        <span class="font-display tracking-tight text-sm">Settings</span>
      </a>
      <a class="flex items-center px-3 py-3 text-slate-500 hover:text-blue-600 hover:bg-slate-100 transition-colors duration-200" href="/teacher">
        <span class="material-symbols-outlined mr-3">help_outline</span>
        <span class="font-display tracking-tight text-sm">Support</span>
      </a>
      <div class="mt-6 px-3 flex items-center gap-3">
        <div class="w-8 h-8 rounded-full bg-primary-fixed flex items-center justify-center text-primary text-xs font-bold shrink-0">
          {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
        </div>
        <div class="overflow-hidden">
          <p class="text-xs font-bold truncate">{$currentUser?.username ?? ""}</p>
          <p class="text-[10px] text-slate-500 truncate">{$t("teacher.role")}</p>
        </div>
      </div>
    </div>
  </aside>

  <!-- TopAppBar -->
  <header class="fixed top-0 right-0 left-0 lg:left-64 h-16 z-40 bg-white/80 backdrop-blur-xl border-b border-slate-100/50 shadow-sm shadow-blue-500/5 flex justify-between items-center px-4 md:px-8">
    <div class="flex items-center gap-3 flex-1">
      <button class="lg:hidden p-2 -ml-2 text-slate-600" onclick={() => { sidebarOpen = true; }}>
        <span class="material-symbols-outlined">menu</span>
      </button>
      <div class="relative w-full max-w-md hidden sm:block">
        <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-sm">search</span>
        <input class="w-full bg-slate-100/50 border-none rounded-lg pl-10 pr-4 py-2 text-sm focus:ring-2 focus:ring-blue-500/20 transition-all" placeholder="Search..." type="text" />
      </div>
    </div>
    <div class="flex items-center gap-2 md:gap-4">
      <button class="text-slate-500 hover:text-blue-700 transition-all opacity-80 hover:opacity-100 p-2">
        <span class="material-symbols-outlined">notifications</span>
      </button>
      <button class="text-slate-500 hover:text-blue-700 transition-all opacity-80 hover:opacity-100 p-2">
        <span class="material-symbols-outlined">translate</span>
      </button>
      <div class="h-6 w-px bg-slate-200 mx-1"></div>
      <button
        onclick={() => auth.logout()}
        class="text-blue-600 text-sm font-medium hover:text-blue-700 transition-all opacity-80 hover:opacity-100 px-2"
      >
        {$t("teacher.signout")}
      </button>
    </div>
  </header>

  <!-- Main Content -->
  <main class="lg:ml-64 pt-20 pb-12 px-4 md:px-8 min-h-screen">
    <div class="max-w-7xl mx-auto">
      {@render children()}
    </div>
  </main>
</div>
