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

  onMount(loadUsers);

  function handleSearch() {
    page = 1;
    loadUsers();
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "\u2014";
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
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

  const columns = [
    { key: "id", label: "ID", class: "w-20" },
    { key: "username", label: "" },
    { key: "email", label: "" },
    { key: "iin", label: "" },
    { key: "role", label: "" },
    { key: "status", label: "" },
    { key: "created_at", label: "" },
    { key: "actions", label: "", class: "w-20" },
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

<div class="space-y-6">
  <!-- Header -->
  <div>
    <h1 class="font-display text-2xl font-bold text-on-surface">{$t("admin.users.title")}</h1>
    <p class="text-sm text-on-surface-variant mt-1">{$t("admin.users.subtitle")}</p>
  </div>

  <!-- Search -->
  <div class="flex items-center gap-3">
    <div class="relative flex-1 max-w-md">
      <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-on-surface-variant" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
      </svg>
      <input
        type="text"
        bind:value={search}
        onkeydown={(e) => { if (e.key === "Enter") handleSearch(); }}
        placeholder={$t("admin.users.searchPlaceholder")}
        class="w-full pl-10 pr-4 py-2.5 rounded-lg bg-surface-lowest text-sm text-on-surface placeholder:text-on-surface-variant/60 focus:outline-none focus:ring-2 focus:ring-primary/20 transition-shadow"
      />
    </div>
    <button
      onclick={handleSearch}
      class="px-4 py-2.5 rounded-lg text-sm font-medium text-on-surface-variant bg-surface-low hover:bg-surface-high transition-colors"
    >
      {$t("common.filter")}
    </button>
  </div>

  <!-- Table -->
  <DataTable columns={resolvedColumns} data={users} {loading} empty={$t("admin.users.empty")}>
    {#snippet row(user: User)}
      <tr class="hover:bg-surface-low/50 transition-colors">
        <td class="px-4 py-3 text-xs text-on-surface-variant font-mono">
          {user.id}
        </td>
        <td class="px-4 py-3">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary text-xs font-bold shrink-0">
              {user.username?.[0]?.toUpperCase() ?? "?"}
            </div>
            <span class="font-medium text-on-surface">{user.username}</span>
          </div>
        </td>
        <td class="px-4 py-3 text-on-surface-variant text-sm">
          {user.email || "\u2014"}
        </td>
        <td class="px-4 py-3 text-on-surface-variant text-sm font-mono">
          {maskIin(user.iin)}
        </td>
        <td class="px-4 py-3">
          <span class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium capitalize {roleBadgeColors[user.role] ?? 'bg-surface-high text-on-surface-variant'}">
            {user.role}
          </span>
        </td>
        <td class="px-4 py-3">
          <StatusBadge status={user.status ?? "active"} />
        </td>
        <td class="px-4 py-3 text-on-surface-variant text-sm">
          {formatDate(user.created_at)}
        </td>
        <td class="px-4 py-3">
          <button class="p-1.5 rounded-md hover:bg-surface-high transition-colors text-on-surface-variant hover:text-on-surface" title={$t("common.edit")}>
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
            </svg>
          </button>
        </td>
      </tr>
    {/snippet}
  </DataTable>

  <!-- Pagination -->
  {#if total > perPage}
    <div class="flex items-center justify-between text-sm text-on-surface-variant">
      <span>{$t("common.showing")} {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} {$t("common.of")} {total}</span>
      <div class="flex gap-1">
        <button
          disabled={page <= 1}
          onclick={() => { page--; loadUsers(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          {$t("common.previous")}
        </button>
        {#each Array.from({ length: Math.min(5, Math.ceil(total / perPage)) }, (_, i) => i + 1) as p}
          <button
            onclick={() => { page = p; loadUsers(); }}
            class="w-8 h-8 rounded-md text-sm transition-colors {p === page ? 'bg-primary text-on-primary font-medium' : 'hover:bg-surface-high'}"
          >
            {p}
          </button>
        {/each}
        <button
          disabled={page * perPage >= total}
          onclick={() => { page++; loadUsers(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          {$t("common.next")}
        </button>
      </div>
    </div>
  {/if}
</div>
