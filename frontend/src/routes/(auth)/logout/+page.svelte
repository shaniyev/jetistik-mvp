<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  let done = $state(false);

  onMount(async () => {
    await auth.logout();
    done = true;
  });
</script>

<style>
  .primary-gradient {
    background: linear-gradient(135deg, #004ac6 0%, #2563eb 100%);
  }
  .shadow-ambient {
    box-shadow: 0 20px 40px -10px rgba(0, 74, 198, 0.04);
  }
</style>

<svelte:head>
  <title>{$t("auth.logout")} — Jetistik</title>
</svelte:head>

<div class="max-w-md w-full">
  <div class="relative overflow-hidden">
    <div class="bg-surface-container-lowest rounded-xl p-10 shadow-ambient text-center relative z-10 border border-outline-variant/10">
      {#if done}
        <div class="mb-8 inline-flex items-center justify-center w-20 h-20 rounded-full bg-primary-fixed/30 text-primary">
          <span class="material-symbols-outlined text-4xl">logout</span>
        </div>

        <h1 class="font-display text-3xl font-extrabold text-on-surface mb-2 tracking-tight">
          {$t("auth.logged_out")}
        </h1>
        <p class="text-on-surface-variant font-body mb-10 text-lg leading-relaxed whitespace-pre-line">
          {$t("auth.logged_out_subtitle")}
        </p>

        <div class="space-y-4">
          <a
            class="block w-full primary-gradient text-white py-4 rounded-md font-display font-bold text-base shadow-lg transition-all hover:brightness-110 active:scale-[0.98] flex items-center justify-center gap-2"
            href="/login"
          >
            <span>{$t("auth.login_again")}</span>
            <span class="material-symbols-outlined text-xl">login</span>
          </a>
          <a
            class="block w-full py-4 text-on-surface-variant font-medium text-sm hover:text-primary transition-colors flex items-center justify-center gap-2"
            href="/"
          >
            <span class="material-symbols-outlined text-lg">arrow_back</span>
            {$t("auth.back_to_home")}
          </a>
        </div>
      {:else}
        <div class="mb-8 inline-flex items-center justify-center w-20 h-20 rounded-full bg-primary-fixed/30 text-primary">
          <span class="material-symbols-outlined text-4xl animate-spin">progress_activity</span>
        </div>
        <p class="text-on-surface-variant font-body text-lg">{$t("auth.logging_out")}</p>
      {/if}
    </div>

    <!-- Decorative blurs -->
    <div class="absolute -top-12 -right-12 w-32 h-32 bg-primary/5 rounded-full blur-3xl"></div>
    <div class="absolute -bottom-12 -left-12 w-48 h-48 bg-secondary-container/10 rounded-full blur-3xl"></div>
  </div>

  <!-- Info cards -->
  {#if done}
    <div class="mt-12 grid grid-cols-2 gap-4">
      <div class="bg-surface-container-low p-6 rounded-xl border border-outline-variant/5">
        <span class="material-symbols-outlined text-primary mb-3">verified</span>
        <h3 class="font-display font-bold text-sm text-on-surface mb-1 uppercase tracking-wider">{$t("auth.security")}</h3>
        <p class="text-xs text-on-surface-variant">{$t("auth.security_desc")}</p>
      </div>
      <div class="bg-surface-container-low p-6 rounded-xl border border-outline-variant/5">
        <span class="material-symbols-outlined text-primary mb-3">support_agent</span>
        <h3 class="font-display font-bold text-sm text-on-surface mb-1 uppercase tracking-wider">{$t("auth.support")}</h3>
        <p class="text-xs text-on-surface-variant">{$t("auth.support_desc")}</p>
      </div>
    </div>
  {/if}
</div>
