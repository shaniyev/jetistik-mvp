<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";

  onMount(() => {
    goto("/admin/organizations", { replaceState: true });
  });
</script>

<div class="flex items-center justify-center py-20">
  <p class="text-on-surface-variant">Redirecting...</p>
</div>
