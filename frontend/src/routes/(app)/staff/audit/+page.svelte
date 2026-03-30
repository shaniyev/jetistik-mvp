<script lang="ts">
  import { onMount } from "svelte";
  import { api, type PaginatedResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";

  interface AuditLog {
    id: number;
    actor_username: string;
    action: string;
    object_type: string;
    object_id: string;
    meta: Record<string, unknown>;
    created_at: string;
  }

  let logs = $state<AuditLog[]>([]);
  let loading = $state(true);
  let currentPage = $state(1);
  let total = $state(0);
  let actionFilter = $state("");
  const perPage = 30;

  async function loadLogs() {
    loading = true;
    try {
      let url = `/api/v1/staff/audit-log?page=${currentPage}&per_page=${perPage}`;
      if (actionFilter) url += `&action=${encodeURIComponent(actionFilter)}`;
      const res = await api.get<AuditLog[]>(url) as PaginatedResponse<AuditLog>;
      logs = res.data;
      total = res.pagination.total;
    } catch (e) {
      console.error("Failed to load audit logs", e);
    } finally {
      loading = false;
    }
  }

  function formatAction(action: string): string {
    return action.replace(".", " / ");
  }

  function actionBadgeColor(action: string): string {
    if (action.includes("delete") || action.includes("revoke")) return "bg-error-container text-on-error-container";
    if (action.includes("create") || action.includes("generate")) return "bg-primary-fixed text-on-primary-fixed-variant";
    if (action.includes("update") || action.includes("mapping")) return "bg-blue-50 text-blue-700";
    if (action.includes("upload")) return "bg-indigo-50 text-indigo-700";
    return "bg-surface-container-highest text-on-surface-variant";
  }

  onMount(loadLogs);

  const actionOptions = [
    "", "event.create", "event.update", "event.delete",
    "template.upload", "template.delete",
    "batch.upload", "batch.mapping", "batch.generate",
    "certificate.revoke", "certificate.unrevoke", "certificate.delete",
  ];

  let totalPages = $derived(Math.ceil(total / perPage));
</script>

<div class="p-6 lg:p-10 pb-32">
  <!-- Header -->
  <header class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12">
    <div>
      <h1 class="font-display text-4xl font-extrabold tracking-tight text-on-surface">{$t("staff.audit.title")}</h1>
      <p class="text-on-surface-variant mt-1">{$t("staff.audit.subtitle")}</p>
    </div>
    <div>
      <select
        bind:value={actionFilter}
        onchange={() => { currentPage = 1; loadLogs(); }}
        class="px-4 py-2.5 rounded-xl bg-surface-container-lowest border border-outline-variant/20 text-on-surface text-sm font-medium
               focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-shadow"
      >
        <option value="">{$t("staff.audit.allActions")}</option>
        {#each actionOptions.filter(Boolean) as action}
          <option value={action}>{formatAction(action)}</option>
        {/each}
      </select>
    </div>
  </header>

  <!-- Table Card -->
  <section class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full text-left">
        <thead class="bg-surface-container-low">
          <tr>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.time")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.actor")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.action")}</th>
            <th class="px-6 py-4 text-[11px] font-bold uppercase tracking-[0.1em] text-on-surface-variant">{$t("common.object")}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant/10">
          {#if loading}
            <tr>
              <td colspan="4" class="px-6 py-16 text-center text-on-surface-variant">
                <div class="flex items-center justify-center gap-2">
                  <span class="material-symbols-outlined animate-spin text-primary">progress_activity</span>
                  <span>{$t("common.loading")}</span>
                </div>
              </td>
            </tr>
          {:else if logs.length === 0}
            <tr>
              <td colspan="4" class="px-6 py-16 text-center text-on-surface-variant">{$t("staff.audit.empty")}</td>
            </tr>
          {:else}
            {#each logs as log (log.id)}
              <tr class="hover:bg-surface-container-low/30 transition-colors">
                <td class="px-6 py-5 text-sm text-on-surface-variant whitespace-nowrap">
                  {new Date(log.created_at).toLocaleString()}
                </td>
                <td class="px-6 py-5">
                  <span class="text-sm font-medium text-on-surface">{log.actor_username || "system"}</span>
                </td>
                <td class="px-6 py-5">
                  <span class="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider {actionBadgeColor(log.action)}">
                    {log.action}
                  </span>
                </td>
                <td class="px-6 py-5 text-sm text-on-surface-variant">
                  {#if log.object_type}
                    <span class="font-mono">{log.object_type} #{log.object_id}</span>
                  {:else}
                    —
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>

  <!-- Pagination -->
  {#if total > perPage}
    <div class="flex items-center justify-between mt-6">
      <p class="text-sm text-on-surface-variant">
        {$t("staff.audit.page_of")} {currentPage} / {totalPages}
      </p>
      <div class="flex items-center gap-2">
        <button
          disabled={currentPage <= 1}
          onclick={() => { currentPage--; loadLogs(); }}
          class="px-4 py-2 rounded-lg border border-outline-variant/20 text-sm font-semibold text-on-surface hover:bg-surface-container transition-colors disabled:opacity-40"
        >
          {$t("common.previous")}
        </button>
        <button
          disabled={currentPage * perPage >= total}
          onclick={() => { currentPage++; loadLogs(); }}
          class="px-4 py-2 rounded-lg border border-outline-variant/20 text-sm font-semibold text-on-surface hover:bg-surface-container transition-colors disabled:opacity-40"
        >
          {$t("common.next")}
        </button>
      </div>
    </div>
  {/if}
</div>
