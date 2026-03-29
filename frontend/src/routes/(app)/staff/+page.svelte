<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";

  onMount(() => {
    goto("/staff/events", { replaceState: true });
  });
</script>

<div class="text-center py-12 text-on-surface-variant">Redirecting...</div>
