<script lang="ts" generics="T">
  import type { Snippet } from "svelte";

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

  let { columns, data, loading = false, empty = "No data found.", row }: Props = $props();
</script>

<div class="overflow-x-auto rounded-lg bg-surface-lowest">
  <table class="w-full text-sm text-left">
    <thead>
      <tr class="bg-surface-low">
        {#each columns as col}
          <th class="px-4 py-3 font-medium text-on-surface-variant {col.class ?? ''}">
            {col.label}
          </th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#if loading}
        <tr>
          <td colspan={columns.length} class="px-4 py-12 text-center text-on-surface-variant">
            Loading...
          </td>
        </tr>
      {:else if data.length === 0}
        <tr>
          <td colspan={columns.length} class="px-4 py-12 text-center text-on-surface-variant">
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
