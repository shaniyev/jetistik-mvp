<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth, currentUser } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import { ApiError } from "$lib/api/client";
  import { get } from "svelte/store";

  let username = $state("");
  let password = $state("");
  let error = $state("");
  let loading = $state(false);

  function getDashboardPath(role: string): string {
    switch (role) {
      case "admin": return "/admin/organizations";
      case "staff": return "/staff/events";
      case "teacher": return "/teacher";
      case "student": return "/student";
      default: return "/";
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = "";
    loading = true;

    try {
      await auth.login(username, password);
      const user = get(currentUser);
      goto(getDashboardPath(user?.role ?? ""));
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "INVALID_CREDENTIALS") {
          error = $t("auth.error.invalid_credentials");
        } else {
          error = err.message;
        }
      } else {
        error = "An unexpected error occurred";
      }
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>{$t("auth.login")} — Jetistik</title>
</svelte:head>

<div class="w-full max-w-[480px]">
  <!-- Error alert -->
  {#if error}
    <div class="mb-6 space-y-3">
      <div class="bg-error-container text-on-error-container p-4 rounded-xl flex items-start gap-3 shadow-sm">
        <span class="material-symbols-outlined shrink-0">error</span>
        <div>
          <p class="font-body text-sm font-semibold">{error}</p>
        </div>
      </div>
    </div>
  {/if}

  <!-- Login card -->
  <div class="bg-surface-container-lowest rounded-2xl shadow-xl shadow-primary/5 p-8 md:p-10 border border-outline-variant/10">
    <div class="mb-8">
      <h1 class="font-display text-3xl font-bold text-on-surface tracking-tight">{$t("auth.login_title")}</h1>
      <p class="font-body text-on-surface-variant text-sm mt-2">{$t("auth.login_subtitle")}</p>
    </div>

    <form onsubmit={handleSubmit} class="space-y-6">
      <!-- Username field -->
      <div class="group">
        <label class="block font-body text-xs font-bold text-on-surface-variant uppercase tracking-widest mb-2 px-1" for="username">
          {$t("auth.username")}
        </label>
        <div class="relative">
          <input
            class="w-full bg-surface-container-low border-0 border-b-2 border-outline-variant focus:border-primary focus:ring-0 transition-all px-4 py-3 font-body text-on-surface rounded-t-lg placeholder:text-outline-variant/60"
            id="username"
            type="text"
            bind:value={username}
            required
            autocomplete="username"
            placeholder={$t("auth.username")}
          />
          <span class="absolute right-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-outline-variant">person</span>
        </div>
      </div>

      <!-- Password field -->
      <div class="group">
        <div class="flex justify-between items-end mb-2 px-1">
          <label class="block font-body text-xs font-bold text-on-surface-variant uppercase tracking-widest" for="password">
            {$t("auth.password")}
          </label>
        </div>
        <div class="relative">
          <input
            class="w-full bg-surface-container-low border-0 border-b-2 border-outline-variant focus:border-primary focus:ring-0 transition-all px-4 py-3 font-body text-on-surface rounded-t-lg placeholder:text-outline-variant/60"
            id="password"
            type="password"
            bind:value={password}
            required
            autocomplete="current-password"
            placeholder="••••••••"
          />
          <span class="absolute right-4 top-1/2 -translate-y-1/2 material-symbols-outlined text-outline-variant">lock</span>
        </div>
      </div>

      <!-- Submit button -->
      <button
        class="w-full py-4 bg-gradient-to-r from-primary to-primary-container text-white font-display font-bold rounded-xl shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:hover:scale-100 cursor-pointer disabled:cursor-not-allowed"
        type="submit"
        disabled={loading}
      >
        <span>{loading ? "..." : $t("auth.login_submit")}</span>
        {#if !loading}
          <span class="material-symbols-outlined text-sm">login</span>
        {/if}
      </button>
    </form>

    <!-- Footnote links -->
    <div class="mt-10 pt-8 border-t border-outline-variant/15 text-center">
      <p class="text-sm font-body text-on-surface-variant">
        {$t("auth.no_account")}
        <a class="text-primary font-bold hover:underline ml-1" href="/register">{$t("auth.register")}</a>
      </p>
    </div>
  </div>

  <!-- Decorative branding element -->
  <div class="mt-8 text-center">
    <div class="inline-flex items-center gap-2 px-4 py-2 bg-surface-high/50 rounded-full">
      <span class="material-symbols-outlined text-xs text-primary" style="font-variation-settings: 'FILL' 1;">verified_user</span>
      <span class="text-[10px] font-bold text-on-surface-variant tracking-widest uppercase">Verified Certificate Infrastructure</span>
    </div>
  </div>
</div>
