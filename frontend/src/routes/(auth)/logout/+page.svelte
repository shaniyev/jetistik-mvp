<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  let done = $state(false);

  onMount(async () => {
    await auth.logout();
    done = true;
    setTimeout(() => goto("/login"), 1500);
  });
</script>

<svelte:head>
  <title>{$t("auth.logout")} — Jetistik</title>
</svelte:head>

<div class="text-center py-8">
  {#if done}
    <p class="text-on-surface font-medium">{$t("auth.logged_out")}</p>
    <p class="text-on-surface-variant text-sm mt-2">
      Redirecting to login...
    </p>
  {:else}
    <p class="text-on-surface-variant">{$t("auth.logging_out")}</p>
  {/if}
</div>
