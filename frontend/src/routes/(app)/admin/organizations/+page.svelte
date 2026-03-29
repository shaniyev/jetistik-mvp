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

  // Create modal state
  let showCreateModal = $state(false);
  let createName = $state("");
  let createDomain = $state("");
  let createLoading = $state(false);
  let createError = $state("");

  // Edit state
  let editingId = $state<number | null>(null);
  let editName = $state("");
  let editDomain = $state("");
  let editStatus = $state("");
  let editLoading = $state(false);

  // More dropdown state
  let openMenuId = $state<number | null>(null);

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

  async function handleCreate() {
    createLoading = true;
    createError = "";
    try {
      await api.post("/api/v1/admin/organizations", { name: createName, domain: createDomain });
      showCreateModal = false;
      createName = "";
      createDomain = "";
      await loadOrgs();
    } catch (e: any) {
      createError = e.message || "Failed to create organization";
    } finally {
      createLoading = false;
    }
  }

  function startEdit(org: Organization) {
    editingId = org.id;
    editName = org.name;
    editDomain = org.domain || "";
    editStatus = org.status;
  }

  function cancelEdit() {
    editingId = null;
  }

  async function saveEdit(orgId: number) {
    editLoading = true;
    try {
      await api.patch(`/api/v1/admin/organizations/${orgId}`, {
        name: editName,
        domain: editDomain,
        status: editStatus,
      });
      editingId = null;
      await loadOrgs();
    } catch (e: any) {
      alert(e.message || "Failed to update organization");
    } finally {
      editLoading = false;
    }
  }

  async function deleteOrg(org: Organization) {
    openMenuId = null;
    if (!confirm(`Delete "${org.name}"? This action cannot be undone.`)) return;
    try {
      await api.delete(`/api/v1/admin/organizations/${org.id}`);
      await loadOrgs();
    } catch (e: any) {
      alert(e.message || "Failed to delete organization");
    }
  }

  function toggleMenu(orgId: number) {
    openMenuId = openMenuId === orgId ? null : orgId;
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "\u2014";
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "2-digit",
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

  const avatarColors = [
    "bg-primary-fixed text-primary",
    "bg-secondary-fixed text-secondary",
    "bg-tertiary-fixed text-tertiary",
  ];

  function getAvatarColor(index: number): string {
    return avatarColors[index % avatarColors.length];
  }

  let totalPages = $derived(Math.ceil(total / perPage));

  const columns = [
    { key: "id", label: "ID", class: "w-28" },
    { key: "name", label: "" },
    { key: "status", label: "" },
    { key: "created_at", label: "" },
    { key: "members", label: "" },
    { key: "actions", label: "", class: "text-right" },
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

<!-- Page Header -->
<header class="flex justify-between items-end mb-10">
  <div class="space-y-1">
    <nav class="flex text-[10px] uppercase tracking-widest text-on-surface-variant/60 gap-2 mb-2">
      <a class="hover:text-primary transition-colors" href="/admin">Admin</a>
      <span>/</span>
      <span class="text-on-surface-variant">{$t("admin.organizations")}</span>
    </nav>
    <h1 class="text-4xl font-extrabold tracking-tight text-on-surface font-display">{$t("admin.orgs.title")}</h1>
    <p class="text-on-surface-variant max-w-2xl">{$t("admin.orgs.subtitle")}</p>
  </div>
  <button
    onclick={() => { showCreateModal = true; }}
    class="bg-gradient-to-br from-primary to-primary-container text-white px-6 py-2.5 rounded-lg font-semibold flex items-center gap-2 shadow-sm hover:shadow-md active:scale-95 transition-all"
  >
    <span class="material-symbols-outlined text-[20px]">add_business</span>
    <span>{$t("admin.orgs.create")}</span>
  </button>
</header>

<!-- Search and Filters -->
<section class="mb-6 flex flex-wrap gap-4 items-center justify-between">
  <div class="relative w-full max-w-md group">
    <span class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-outline-variant group-focus-within:text-primary transition-colors">search</span>
    <input
      type="text"
      bind:value={search}
      onkeydown={(e) => { if (e.key === "Enter") handleSearch(); }}
      placeholder={$t("admin.orgs.searchPlaceholder")}
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
    <button
      class="px-4 py-2.5 bg-surface-container-low text-on-surface-variant rounded-lg border border-outline-variant/20 flex items-center gap-2 hover:bg-surface-container transition-colors text-sm font-medium"
    >
      <span class="material-symbols-outlined text-[18px]">file_download</span>
      <span>Export CSV</span>
    </button>
  </div>
</section>

<!-- Organizations Table -->
<DataTable columns={resolvedColumns} data={orgs} {loading} empty={$t("admin.orgs.empty")}>
  {#snippet row(org: Organization, index: number)}
    <tr class="{index % 2 === 0 ? 'bg-surface-container-lowest' : 'bg-surface-container-low'} hover:bg-white transition-colors group">
      <td class="px-6 py-5 font-mono text-xs text-on-surface-variant">
        {org.code || `#ORG-${String(org.id).padStart(4, '0')}`}
      </td>
      <td class="px-6 py-5">
        {#if editingId === org.id}
          <div class="flex flex-col gap-1">
            <input
              type="text"
              bind:value={editName}
              class="px-2 py-1 rounded bg-surface-container-lowest text-sm text-on-surface border border-outline-variant focus:border-primary outline-none"
            />
            <input
              type="text"
              bind:value={editDomain}
              placeholder={$t("admin.orgs.domain")}
              class="px-2 py-1 rounded bg-surface-container-lowest text-xs text-on-surface-variant border border-outline-variant focus:border-primary outline-none"
            />
          </div>
        {:else}
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded {getAvatarColor(index)} flex items-center justify-center font-bold text-xs">
              {getInitials(org.name)}
            </div>
            <div class="flex flex-col">
              <span class="font-semibold text-on-surface">{org.name}</span>
              <span class="text-xs text-on-surface-variant">{org.domain || ""}</span>
            </div>
          </div>
        {/if}
      </td>
      <td class="px-6 py-5">
        {#if editingId === org.id}
          <select
            bind:value={editStatus}
            class="px-2 py-1 rounded bg-surface-container-lowest text-sm text-on-surface border border-outline-variant focus:border-primary outline-none"
          >
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="pending">Pending</option>
            <option value="archived">Archived</option>
            <option value="suspended">Suspended</option>
          </select>
        {:else}
          <StatusBadge status={org.status} />
        {/if}
      </td>
      <td class="px-6 py-5 text-sm text-on-surface-variant">
        {formatDate(org.created_at)}
      </td>
      <td class="px-6 py-5 text-sm text-on-surface font-medium">
        {org.members_count?.toLocaleString() ?? 0}
      </td>
      <td class="px-6 py-5 text-right">
        {#if editingId === org.id}
          <div class="flex items-center justify-end gap-1">
            <button
              onclick={() => saveEdit(org.id)}
              disabled={editLoading}
              class="p-2 text-primary hover:bg-primary/5 rounded-lg transition-all"
              title="Save"
            >
              <span class="material-symbols-outlined">check</span>
            </button>
            <button
              onclick={cancelEdit}
              class="p-2 text-outline hover:bg-slate-100 rounded-lg transition-all"
              title="Cancel"
            >
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
        {:else}
          <button onclick={() => startEdit(org)} class="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all">
            <span class="material-symbols-outlined">edit</span>
          </button>
          <div class="relative inline-block">
            <button onclick={() => toggleMenu(org.id)} class="p-2 text-outline hover:text-on-surface hover:bg-slate-100 rounded-lg transition-all">
              <span class="material-symbols-outlined">more_vert</span>
            </button>
            {#if openMenuId === org.id}
              <div class="absolute right-0 top-full mt-1 z-50 bg-surface-container-lowest rounded-lg shadow-lg border border-outline-variant/20 py-1 min-w-[140px]">
                <button
                  onclick={() => deleteOrg(org)}
                  class="w-full text-left px-3 py-2 text-sm text-error hover:bg-error-container/30 transition-colors flex items-center gap-2"
                >
                  <span class="material-symbols-outlined text-[18px]">delete</span>
                  {$t("common.delete")}
                </button>
              </div>
            {/if}
          </div>
        {/if}
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
        onclick={() => { page--; loadOrgs(); }}
        class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-surface-container transition-colors text-outline disabled:opacity-30"
      >
        <span class="material-symbols-outlined text-[18px]">chevron_left</span>
      </button>
      {#each Array.from({ length: Math.min(3, totalPages) }, (_, i) => i + 1) as p}
        <button
          onclick={() => { page = p; loadOrgs(); }}
          class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium transition-colors
            {p === page ? 'bg-primary text-white font-bold' : 'hover:bg-surface-container text-on-surface'}"
        >
          {p}
        </button>
      {/each}
      {#if totalPages > 4}
        <span class="px-2 text-outline text-xs">...</span>
        <button
          onclick={() => { page = totalPages; loadOrgs(); }}
          class="w-8 h-8 flex items-center justify-center rounded-lg text-xs font-medium transition-colors
            {totalPages === page ? 'bg-primary text-white font-bold' : 'hover:bg-surface-container text-on-surface'}"
        >
          {totalPages}
        </button>
      {/if}
      <button
        disabled={page * perPage >= total}
        onclick={() => { page++; loadOrgs(); }}
        class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-surface-container transition-colors text-outline disabled:opacity-30"
      >
        <span class="material-symbols-outlined text-[18px]">chevron_right</span>
      </button>
    </div>
  </footer>
{/if}

<!-- System Stats Grid (Bento Style) -->
<section class="mt-12 grid grid-cols-1 md:grid-cols-3 gap-6">
  <div class="bg-surface-container-lowest p-6 rounded-2xl ring-1 ring-outline-variant/10 shadow-sm flex flex-col gap-4">
    <div class="w-12 h-12 rounded-xl bg-blue-50 flex items-center justify-center text-primary">
      <span class="material-symbols-outlined">analytics</span>
    </div>
    <div>
      <h3 class="text-xs font-bold uppercase tracking-widest text-on-surface-variant mb-1">{$t("admin.weeklyGrowth")}</h3>
      <p class="text-3xl font-extrabold text-on-surface font-display">+{stats.weekly_growth}%</p>
      <p class="text-xs text-primary font-semibold mt-2 flex items-center gap-1">
        <span class="material-symbols-outlined text-xs">trending_up</span>
        <span>{stats.new_orgs_this_week} {$t("admin.newOrgsThisWeek")}</span>
      </p>
    </div>
  </div>

  <div class="bg-surface-container-lowest p-6 rounded-2xl ring-1 ring-outline-variant/10 shadow-sm flex flex-col gap-4">
    <div class="w-12 h-12 rounded-xl bg-slate-50 flex items-center justify-center text-secondary">
      <span class="material-symbols-outlined">security</span>
    </div>
    <div>
      <h3 class="text-xs font-bold uppercase tracking-widest text-on-surface-variant mb-1">{$t("admin.ledgerIntegrity")}</h3>
      <p class="text-3xl font-extrabold text-on-surface font-display">{stats.ledger_integrity || 99.9}%</p>
      <p class="text-xs text-on-surface-variant mt-2">{$t("admin.allCertsVerified")}</p>
    </div>
  </div>

  <div class="md:col-span-1 bg-gradient-to-br from-primary/5 to-primary-container/5 p-6 rounded-2xl ring-1 ring-primary/20 flex flex-col justify-between">
    <div>
      <h3 class="text-sm font-bold text-on-surface mb-2">Need bulk import?</h3>
      <p class="text-xs text-on-surface-variant leading-relaxed">Download our CSV template for faster organization onboarding and member mapping.</p>
    </div>
    <button class="mt-4 text-xs font-bold text-primary flex items-center gap-1 hover:underline">
      Get Template <span class="material-symbols-outlined text-xs">arrow_forward</span>
    </button>
  </div>
</section>

<!-- Create Organization Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" onclick={() => { showCreateModal = false; }}></div>
    <div class="relative bg-surface-container-lowest rounded-2xl shadow-xl p-6 w-full max-w-md mx-4 ring-1 ring-outline-variant/10">
      <h2 class="text-lg font-display font-bold text-on-surface mb-4">{$t("admin.orgs.create")}</h2>

      {#if createError}
        <div class="bg-error-container text-on-error-container p-3 rounded-lg text-sm mb-4">{createError}</div>
      {/if}

      <form onsubmit={(e) => { e.preventDefault(); handleCreate(); }} class="space-y-4">
        <div>
          <label for="create-name" class="block text-sm font-medium text-on-surface mb-1">{$t("admin.orgs.name")}</label>
          <input
            id="create-name"
            type="text"
            bind:value={createName}
            required
            class="w-full px-3 py-2.5 rounded-xl bg-surface-container-lowest text-sm text-on-surface ring-1 ring-outline-variant/30 focus:ring-2 focus:ring-primary/50 outline-none transition-all"
          />
        </div>
        <div>
          <label for="create-domain" class="block text-sm font-medium text-on-surface mb-1">{$t("admin.orgs.domain")}</label>
          <input
            id="create-domain"
            type="text"
            bind:value={createDomain}
            placeholder="example.org"
            class="w-full px-3 py-2.5 rounded-xl bg-surface-container-lowest text-sm text-on-surface ring-1 ring-outline-variant/30 focus:ring-2 focus:ring-primary/50 outline-none transition-all"
          />
        </div>
        <div class="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onclick={() => { showCreateModal = false; }}
            class="px-4 py-2.5 rounded-lg text-sm font-medium text-on-surface-variant hover:bg-surface-container transition-colors"
          >
            {$t("common.cancel")}
          </button>
          <button
            type="submit"
            disabled={createLoading || !createName.trim()}
            class="px-6 py-2.5 rounded-lg text-sm font-semibold bg-gradient-to-br from-primary to-primary-container text-white hover:shadow-md transition-all disabled:opacity-50"
          >
            {createLoading ? "..." : $t("admin.orgs.create")}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
