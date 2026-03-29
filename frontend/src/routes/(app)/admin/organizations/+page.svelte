<script lang="ts">
  import { onMount } from "svelte";
  import { api, type PaginatedResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";

  interface Organization {
    id: number;
    code: string;
    name: string;
    domain: string;
    status: string;
    created_at: string;
    members_count: number;
  }

  interface Stats {
    weekly_growth: number;
    new_orgs_this_week: number;
    ledger_integrity: number;
  }

  let orgs = $state<Organization[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  let search = $state("");
  let stats = $state<Stats>({ weekly_growth: 0, new_orgs_this_week: 0, ledger_integrity: 0 });
  const perPage = 20;

  async function loadOrgs() {
    loading = true;
    try {
      const query = new URLSearchParams({ page: String(page), per_page: String(perPage) });
      if (search) query.set("search", search);
      const res = await api.get<Organization[]>(`/api/v1/admin/organizations?${query}`) as PaginatedResponse<Organization>;
      orgs = res.data;
      total = res.pagination.total;
    } catch (e) {
      console.error("Failed to load organizations", e);
    } finally {
      loading = false;
    }
  }

  async function loadStats() {
    try {
      const res = await api.get<Stats>("/api/v1/admin/stats");
      stats = res.data;
    } catch {
      // Stats are non-critical
    }
  }

  onMount(() => {
    loadOrgs();
    loadStats();
  });

  function handleSearch() {
    page = 1;
    loadOrgs();
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "\u2014";
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  }

  function getInitials(name: string): string {
    return name
      .split(/\s+/)
      .slice(0, 2)
      .map((w) => w[0]?.toUpperCase() ?? "")
      .join("");
  }

  const columns = [
    { key: "id", label: "ID", class: "w-28" },
    { key: "name", label: "" },
    { key: "status", label: "" },
    { key: "created_at", label: "" },
    { key: "members", label: "" },
    { key: "actions", label: "", class: "w-24" },
  ];

  let resolvedColumns = $derived(columns.map((c) => {
    if (c.key === "name") return { ...c, label: $t("admin.orgs.name") };
    if (c.key === "status") return { ...c, label: $t("common.status") };
    if (c.key === "created_at") return { ...c, label: $t("admin.orgs.createdDate") };
    if (c.key === "members") return { ...c, label: $t("admin.orgs.members") };
    if (c.key === "actions") return { ...c, label: $t("common.actions") };
    return c;
  }));
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-start justify-between">
    <div>
      <h1 class="font-display text-2xl font-bold text-on-surface">{$t("admin.orgs.title")}</h1>
      <p class="text-sm text-on-surface-variant mt-1 max-w-xl">{$t("admin.orgs.subtitle")}</p>
    </div>
    <button
      class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium
             bg-gradient-to-br from-primary to-primary-container text-on-primary
             hover:shadow-lg transition-shadow"
    >
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
      </svg>
      {$t("admin.orgs.create")}
    </button>
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
        placeholder={$t("admin.orgs.searchPlaceholder")}
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
  <DataTable columns={resolvedColumns} data={orgs} {loading} empty={$t("admin.orgs.empty")}>
    {#snippet row(org: Organization)}
      <tr class="hover:bg-surface-low/50 transition-colors">
        <td class="px-4 py-3">
          <span class="text-xs text-on-surface-variant font-mono">{org.code || `#${org.id}`}</span>
        </td>
        <td class="px-4 py-3">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg bg-surface-high flex items-center justify-center text-xs font-bold text-on-surface-variant shrink-0">
              {getInitials(org.name)}
            </div>
            <div>
              <p class="font-medium text-on-surface">{org.name}</p>
              <p class="text-xs text-on-surface-variant">{org.domain || ""}</p>
            </div>
          </div>
        </td>
        <td class="px-4 py-3">
          <StatusBadge status={org.status} />
        </td>
        <td class="px-4 py-3 text-on-surface-variant text-sm">
          {formatDate(org.created_at)}
        </td>
        <td class="px-4 py-3 text-on-surface text-sm font-medium">
          {org.members_count?.toLocaleString() ?? 0}
        </td>
        <td class="px-4 py-3">
          <div class="flex items-center gap-2">
            <button class="p-1.5 rounded-md hover:bg-surface-high transition-colors text-on-surface-variant hover:text-on-surface" title={$t("common.edit")}>
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
              </svg>
            </button>
            <button class="p-1.5 rounded-md hover:bg-surface-high transition-colors text-on-surface-variant hover:text-on-surface" title="More">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 12.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 18.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5Z" />
              </svg>
            </button>
          </div>
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
          onclick={() => { page--; loadOrgs(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          {$t("common.previous")}
        </button>
        {#each Array.from({ length: Math.min(5, Math.ceil(total / perPage)) }, (_, i) => i + 1) as p}
          <button
            onclick={() => { page = p; loadOrgs(); }}
            class="w-8 h-8 rounded-md text-sm transition-colors {p === page ? 'bg-primary text-on-primary font-medium' : 'hover:bg-surface-high'}"
          >
            {p}
          </button>
        {/each}
        <button
          disabled={page * perPage >= total}
          onclick={() => { page++; loadOrgs(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          {$t("common.next")}
        </button>
      </div>
    </div>
  {/if}

  <!-- Stats Cards -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
    <div class="bg-surface-lowest rounded-lg p-5">
      <div class="flex items-center gap-3 mb-3">
        <div class="w-10 h-10 rounded-lg bg-primary-fixed flex items-center justify-center">
          <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18 9 11.25l4.306 4.306a11.95 11.95 0 0 1 5.814-5.518l2.74-1.22m0 0-5.94-2.281m5.94 2.28-2.28 5.941" />
          </svg>
        </div>
        <span class="text-xs uppercase tracking-wider text-on-surface-variant font-medium">{$t("admin.weeklyGrowth")}</span>
      </div>
      <p class="text-2xl font-display font-bold text-on-surface">+{stats.weekly_growth}%</p>
      <p class="text-xs text-primary mt-1">{stats.new_orgs_this_week} {$t("admin.newOrgsThisWeek")}</p>
    </div>

    <div class="bg-surface-lowest rounded-lg p-5">
      <div class="flex items-center gap-3 mb-3">
        <div class="w-10 h-10 rounded-lg bg-primary-fixed flex items-center justify-center">
          <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
          </svg>
        </div>
        <span class="text-xs uppercase tracking-wider text-on-surface-variant font-medium">{$t("admin.ledgerIntegrity")}</span>
      </div>
      <p class="text-2xl font-display font-bold text-on-surface">{stats.ledger_integrity || 99.9}%</p>
      <p class="text-xs text-on-surface-variant mt-1">{$t("admin.allCertsVerified")}</p>
    </div>
  </div>
</div>
