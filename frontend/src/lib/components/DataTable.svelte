<script lang="ts" generics="T">
  import type { Snippet } from "svelte";
  import { t } from "$lib/i18n";

  interface Column {
    key: string;
    label: string;
    class?: string;
  }

  interface Props {
    columns: Column[];
    data: T[];
    loading?: boolean;
    empty?: string;
    row: Snippet<[T, number]>;
  }

  let { columns, data, loading = false, empty = $t("common.no_data"), row }: Props = $props();
</script>

<div class="bg-surface-container-low rounded-2xl overflow-hidden ring-1 ring-outline-variant/10 shadow-sm">
  <table class="w-full text-left border-collapse">
    <thead>
      <tr class="bg-surface-high/50 text-on-surface-variant uppercase tracking-[0.1em] text-[11px] font-bold">
        {#each columns as col}
          <th class="px-6 py-4 {col.class ?? ''}">
            {col.label}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody class="divide-y divide-outline-variant/10">
      {#if loading}
        <tr>
          <td colspan={columns.length} class="px-6 py-12 text-center text-on-surface-variant">
            <div class="flex items-center justify-center gap-2">
              <span class="material-symbols-outlined animate-spin text-primary">progress_activity</span>
              <span>{$t("dataTable.loading")}</span>
            </div>
          </td>
        </tr>
      {:else if data.length === 0}
        <tr>
          <td colspan={columns.length} class="px-6 py-12 text-center text-on-surface-variant">
            {empty}
          </td>
        </tr>
      {:else}
        {#each data as item, i}
          {@render row(item, i)}
        {/each}
      {/if}
    </tbody>
  </table>
</div>
