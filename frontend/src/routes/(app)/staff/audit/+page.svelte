<script lang="ts">
  import { onMount } from "svelte";
  import { api, type PaginatedResponse } from "$lib/api/client";
  import DataTable from "$lib/components/DataTable.svelte";

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

  onMount(loadLogs);

  const columns = [
    { key: "time", label: "Time" },
    { key: "actor", label: "Actor" },
    { key: "action", label: "Action" },
    { key: "object", label: "Object" },
  ];

  const actionOptions = [
    "", "event.create", "event.update", "event.delete",
    "template.upload", "template.delete",
    "batch.upload", "batch.mapping", "batch.generate",
    "certificate.revoke", "certificate.unrevoke", "certificate.delete",
  ];
</script>

<div class="space-y-6">
  <div>
    <h1 class="font-display text-2xl font-bold text-on-surface">Audit Log</h1>
    <p class="text-sm text-on-surface-variant mt-1">Activity history for your organization</p>
  </div>

  <!-- Filter -->
  <div class="flex items-center gap-3">
    <select
      bind:value={actionFilter}
      onchange={() => { currentPage = 1; loadLogs(); }}
      class="px-3 py-2 rounded-md bg-surface-lowest text-on-surface text-sm
             focus:outline-none focus:ring-2 focus:ring-primary/30"
    >
      <option value="">All actions</option>
      {#each actionOptions.filter(Boolean) as action}
        <option value={action}>{formatAction(action)}</option>
      {/each}
    </select>
  </div>

  <DataTable {columns} data={logs} {loading} empty="No audit logs yet.">
    {#snippet row(log: AuditLog)}
      <tr class="hover:bg-surface-low/50 transition-colors">
        <td class="px-4 py-3 text-xs text-on-surface-variant whitespace-nowrap">
          {new Date(log.created_at).toLocaleString()}
        </td>
        <td class="px-4 py-3 text-sm text-on-surface">
          {log.actor_username || "system"}
        </td>
        <td class="px-4 py-3">
          <span class="px-2 py-0.5 rounded bg-surface-low text-xs font-mono text-on-surface-variant">
            {log.action}
          </span>
        </td>
        <td class="px-4 py-3 text-sm text-on-surface-variant">
          {#if log.object_type}
            {log.object_type} #{log.object_id}
          {:else}
            —
          {/if}
        </td>
      </tr>
    {/snippet}
  </DataTable>

  {#if total > perPage}
    <div class="flex items-center justify-between text-sm text-on-surface-variant">
      <span>Page {currentPage} of {Math.ceil(total / perPage)}</span>
      <div class="flex gap-2">
        <button
          disabled={currentPage <= 1}
          onclick={() => { currentPage--; loadLogs(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          Previous
        </button>
        <button
          disabled={currentPage * perPage >= total}
          onclick={() => { currentPage++; loadLogs(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  {/if}
</div>
