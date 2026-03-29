<script lang="ts">
  import { onMount } from "svelte";
  import { api, type PaginatedResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";

  interface User {
    id: number;
    username: string;
    email: string;
    iin: string;
    role: string;
    is_active: boolean;
    status: string;
    created_at: string;
  }

  let users = $state<User[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  let search = $state("");
  const perPage = 20;

  async function loadUsers() {
    loading = true;
    try {
      const query = new URLSearchParams({ page: String(page), per_page: String(perPage) });
      if (search) query.set("search", search);
      const res = await api.get<User[]>(`/api/v1/admin/users?${query}`) as PaginatedResponse<User>;
      users = res.data;
      total = res.pagination.total;
    } catch (e) {
      console.error("Failed to load users", e);
    } finally {
      loading = false;
    }
  }

  async function changeRole(userId: number, newRole: string) {
    try {
      await api.patch(`/api/v1/admin/users/${userId}`, { role: newRole });
      loadUsers();
    } catch (e: any) {
      alert(e.message || 'Failed to update role');
    }
  }

  async function toggleActive(user: User) {
    try {
      await api.patch(`/api/v1/admin/users/${user.id}`, { is_active: !user.is_active });
      loadUsers();
    } catch (e: any) {
      alert(e.message || 'Failed to update user');
    }
  }

  onMount(loadUsers);

  function handleSearch() {
    page = 1;
    loadUsers();
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "\u2014";
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "2-digit",
      year: "numeric",
    });
  }

  function maskIin(iin: string): string {
    if (!iin || iin.length < 12) return iin || "\u2014";
    return iin.slice(0, 4) + "****" + iin.slice(8);
  }

  const roleBadgeColors: Record<string, string> = {
    admin: "bg-purple-50 text-purple-700",
    staff: "bg-blue-50 text-blue-700",
    teacher: "bg-indigo-50 text-indigo-700",
    student: "bg-emerald-50 text-emerald-700",
  };

  let totalPages = $derived(Math.ceil(total / perPage));

  const columns = [
    { key: "id", label: "ID", class: "w-20" },
    { key: "username", label: "" },
    { key: "email", label: "" },
    { key: "iin", label: "" },
    { key: "role", label: "" },
    { key: "status", label: "" },
    { key: "created_at", label: "" },
    { key: "actions", label: "", class: "text-right" },
  ];

  let resolvedColumns = $derived(columns.map((c) => {
    if (c.key === "username") return { ...c, label: $t("admin.users.username") };
    if (c.key === "email") return { ...c, label: $t("admin.users.email") };
    if (c.key === "iin") return { ...c, label: $t("admin.users.iin") };
    if (c.key === "role") return { ...c, label: $t("admin.users.role") };
    if (c.key === "status") return { ...c, label: $t("common.status") };
    if (c.key === "created_at") return { ...c, label: $t("common.created") };
    if (c.key === "actions") return { ...c, label: $t("common.actions") };
    return c;
  }));
</script>

<!-- Page Header -->
<header class="flex justify-between items-end mb-10">
  <div class="space-y-1">
    <nav class="flex text-[10px] uppercase tracking-widest text-on-surface-variant/60 gap-2 mb-2">
      <a class="hover:text-primary transition-colors" href="/admin">Admin</a>
      <span>/</span>
      <span class="text-on-surface-variant">{$t("admin.users")}</span>
    </nav>
    <h1 class="text-4xl font-extrabold tracking-tight text-on-surface font-display">{$t("admin.users.title")}</h1>
    <p class="text-on-surface-variant max-w-2xl">{$t("admin.users.subtitle")}</p>
  </div>
</header>

<!-- Search and Filters -->
<section class="mb-6 flex flex-wrap gap-4 items-center justify-between">
  <div class="relative w-full max-w-md group">
    <span class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-outline-variant group-focus-within:text-primary transition-colors">search</span>
    <input
      type="text"
      bind:value={search}
      onkeydown={(e) => { if (e.key === "Enter") handleSearch(); }}
      placeholder={$t("admin.users.searchPlaceholder")}
      class="w-full pl-12 pr-4 py-3 bg-surface-container-lowest rounded-xl border-none ring-1 ring-outline-variant/30 focus:ring-2 focus:ring-primary/50 transition-all outline-none text-sm placeholder:text-on-surface-variant/40 shadow-sm"
    />
  </div>
  <div class="flex gap-2">
    <button
      onclick={handleSearch}
      class="px-4 py-2.5 bg-surface-container-low text-on-surface-variant rounded-lg border border-outline-variant/20 flex items-center gap-2 hover:bg-surface-container transition-colors text-sm font-medium"
    >
      <span class="material-symbols-outlined text-[18px]">filter_list</span>
      <span>{$t("common.filter")}</span>
    </button>
  </div>
</section>

<!-- Table -->
<DataTable columns={resolvedColumns} data={users} {loading} empty={$t("admin.users.empty")}>
  {#snippet row(user: User, index: number)}
    <tr class="{index % 2 === 0 ? 'bg-surface-container-lowest' : 'bg-surface-container-low'} hover:bg-white transition-colors group">
      <td class="px-6 py-5 text-xs text-on-surface-variant font-mono">
        {user.id}
      </td>
      <td class="px-6 py-5">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary text-xs font-bold shrink-0">
            {user.username?.[0]?.toUpperCase() ?? "?"}
          </div>
          <span class="font-semibold text-on-surface">{user.username}</span>
        </div>
      </td>
      <td class="px-6 py-5 text-on-surface-variant text-sm">
        {user.email || "\u2014"}
      </td>
      <td class="px-6 py-5 text-on-surface-variant text-sm font-mono">
        {maskIin(user.iin)}
      </td>
      <td class="px-6 py-5">
        <span class="inline-flex items-center rounded-full px-3 py-1 text-[10px] font-bold uppercase tracking-wider capitalize {roleBadgeColors[user.role] ?? 'bg-surface-high text-on-surface-variant'}">
          {user.role}
        </span>
      </td>
      <td class="px-6 py-5">
        <StatusBadge status={user.status ?? "active"} />
      </td>
      <td class="px-6 py-5 text-on-surface-variant text-sm">
        {formatDate(user.created_at)}
      </td>
      <td class="px-6 py-5 text-right">
        <div class="flex items-center justify-end gap-1">
          <select
            value={user.role}
            onchange={(e) => changeRole(user.id, (e.target as HTMLSelectElement).value)}
            class="text-xs px-2 py-1.5 rounded-lg bg-surface-container-lowest ring-1 ring-outline-variant/20 text-on-surface cursor-pointer outline-none focus:ring-primary/50"
          >
            <option value="student">student</option>
            <option value="teacher">teacher</option>
            <option value="staff">staff</option>
            <option value="admin">admin</option>
          </select>
          <button
            onclick={() => toggleActive(user)}
            class="p-2 rounded-lg transition-all text-sm font-medium {user.is_active ? 'text-error hover:bg-error-container/30' : 'text-emerald-600 hover:bg-emerald-50'}"
            title={user.is_active ? 'Deactivate' : 'Activate'}
          >
            <span class="material-symbols-outlined text-[20px]">{user.is_active ? 'block' : 'check_circle'}</span>
          </button>
        </div>
      </td>
    </tr>
  {/snippet}
</DataTable>

<!-- Pagination -->
{#if total > 0}
  <footer class="px-6 py-4 bg-surface-container-high/30 border-t border-outline-variant/10 flex items-center justify-between -mt-[1px] rounded-b-2xl">
    <p class="text-xs text-on-surface-variant">
      {$t("common.showing")} {(page - 1) * perPage + 1} to {Math.min(page * perPage, total)} of {total} entries
    </p>
    <div class="flex items-center gap-1">
      <button
        disabled={page <= 1}
        onclick={() => { page--; loadUsers(); }}
        class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-surface-container transition-colors text-outline disabled:opacity-30"
      >
        <span class="material-symbols-outlined text-[18px]">chevron_left</span>
      </button>
      {#each Array.from({ length: Math.min(3, totalPages) }, (_, i) => i + 1) as p}
        <button
          onclick={() => { page = p; loadUsers(); }}
          class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium transition-colors
            {p === page ? 'bg-primary text-white font-bold' : 'hover:bg-surface-container text-on-surface'}"
        >
          {p}
        </button>
      {/each}
      {#if totalPages > 4}
        <span class="px-2 text-outline text-xs">...</span>
        <button
          onclick={() => { page = totalPages; loadUsers(); }}
          class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium transition-colors
            {totalPages === page ? 'bg-primary text-white font-bold' : 'hover:bg-surface-container text-on-surface'}"
        >
          {totalPages}
        </button>
      {/if}
      <button
        disabled={page * perPage >= total}
        onclick={() => { page++; loadUsers(); }}
        class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-surface-container transition-colors text-outline disabled:opacity-30"
      >
        <span class="material-symbols-outlined text-[18px]">chevron_right</span>
      </button>
    </div>
  </footer>
{/if}
