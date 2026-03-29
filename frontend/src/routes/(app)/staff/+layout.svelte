<script lang="ts">
  import { page } from "$app/stores";
  import { auth, currentUser } from "$lib/stores/auth";

  let { children } = $props();

  const navItems = [
    { href: "/staff/events", label: "Events", icon: "event" },
    { href: "/staff/certificates", label: "Certificates", icon: "verified" },
    { href: "/staff/audit", label: "Audit Log", icon: "receipt_long" },
  ];

  let currentPath = $derived($page.url.pathname);
</script>

<div class="min-h-screen bg-surface flex">
  <!-- Sidebar -->
  <aside class="h-screen w-64 fixed left-0 top-0 bg-slate-50 flex flex-col p-4 gap-2 z-40 border-r border-slate-100">
    <div class="mb-8 px-2 flex items-center gap-3">
      <div class="w-10 h-10 rounded-xl bg-primary flex items-center justify-center text-white shadow-lg">
        <span class="material-symbols-outlined active-nav-icon">architecture</span>
      </div>
      <div>
        <h2 class="font-display text-lg font-extrabold tracking-tight text-on-surface">Staff Portal</h2>
        <p class="text-[10px] text-on-surface-variant uppercase tracking-widest font-semibold">Management Console</p>
      </div>
    </div>

    <nav class="flex-1 flex flex-col gap-1">
      {#each navItems as item}
        {@const isActive = currentPath.startsWith(item.href)}
        <a
          href={item.href}
          class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200 ease-in-out
            {isActive
              ? 'bg-blue-50 text-blue-700'
              : 'text-slate-500 hover:text-slate-900 hover:bg-slate-100'}"
        >
          <span class="material-symbols-outlined {isActive ? 'active-nav-icon' : ''}">{item.icon}</span>
          <span>{item.label}</span>
        </a>
      {/each}
    </nav>

    <div class="mt-auto flex flex-col gap-1 pt-4 border-t border-slate-100">
      <a
        href="/staff/settings"
        class="text-slate-500 hover:text-slate-900 hover:bg-slate-100 rounded-lg flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-all duration-200 ease-in-out"
      >
        <span class="material-symbols-outlined">settings</span>
        <span>Settings</span>
      </a>
      <button
        onclick={() => auth.logout()}
        class="text-slate-500 hover:text-slate-900 hover:bg-slate-100 rounded-lg flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-all duration-200 ease-in-out w-full text-left"
      >
        <span class="material-symbols-outlined">logout</span>
        <span>Logout</span>
      </button>
    </div>
  </aside>

  <!-- Main content -->
  <main class="ml-64 flex-1 flex flex-col min-h-screen">
    {@render children()}
  </main>
</div>
