<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import { ApiError } from "$lib/api/client";

  let username = $state("");
  let password = $state("");
  let error = $state("");
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = "";
    loading = true;

    try {
      await auth.login(username, password);
      goto("/");
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

<form onsubmit={handleSubmit}>
  <h2 class="font-display text-xl font-semibold text-on-surface mb-6">
    {$t("auth.login")}
  </h2>

  {#if error}
    <div class="bg-error-container text-on-error-container rounded-md p-3 mb-4 text-sm">
      {error}
    </div>
  {/if}

  <div class="space-y-4">
    <div>
      <label for="username" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.username")}
      </label>
      <input
        id="username"
        type="text"
        bind:value={username}
        required
        autocomplete="username"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
        placeholder={$t("auth.username")}
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.password")}
      </label>
      <input
        id="password"
        type="password"
        bind:value={password}
        required
        autocomplete="current-password"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
        placeholder={$t("auth.password")}
      />
    </div>
  </div>

  <button
    type="submit"
    disabled={loading}
    class="w-full mt-6 py-2.5 px-4 rounded-md text-on-primary font-medium text-sm
           bg-gradient-to-br from-primary to-primary-container
           hover:opacity-90 disabled:opacity-50
           transition-opacity cursor-pointer disabled:cursor-not-allowed"
  >
    {loading ? "..." : $t("auth.login_action")}
  </button>

  <p class="text-center text-sm text-on-surface-variant mt-4">
    {$t("auth.no_account")}
    <a href="/register" class="text-primary font-medium hover:underline">
      {$t("auth.register")}
    </a>
  </p>
</form>
