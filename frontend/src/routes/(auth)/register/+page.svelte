<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth, currentUser } from "$lib/stores/auth";
  import { t, language } from "$lib/i18n";
  import { ApiError } from "$lib/api/client";
  import { get } from "svelte/store";

  let username = $state("");
  let email = $state("");
  let password = $state("");
  let iin = $state("");
  let role = $state<"student" | "teacher">("student");
  let error = $state("");
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = "";
    loading = true;

    try {
      let currentLang: string = "kz";
      language.subscribe((v) => (currentLang = v))();

      await auth.register({
        username,
        password,
        email: email || undefined,
        iin: iin || undefined,
        role,
        language: currentLang,
      });
      const user = get(currentUser);
      const path = user?.role === "teacher" ? "/teacher" : "/student";
      goto(path);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "USERNAME_EXISTS") {
          error = $t("auth.error.username_exists");
        } else if (err.code === "EMAIL_EXISTS") {
          error = $t("auth.error.email_exists");
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
  <title>{$t("auth.register")} — Jetistik</title>
</svelte:head>

<form onsubmit={handleSubmit}>
  <h2 class="font-display text-xl font-semibold text-on-surface mb-6">
    {$t("auth.register")}
  </h2>

  {#if error}
    <div class="bg-error-container text-on-error-container rounded-md p-3 mb-4 text-sm">
      {error}
    </div>
  {/if}

  <div class="space-y-4">
    <div>
      <label for="username" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.username")} *
      </label>
      <input
        id="username"
        type="text"
        bind:value={username}
        required
        minlength="3"
        autocomplete="username"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
      />
    </div>

    <div>
      <label for="email" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.email")}
      </label>
      <input
        id="email"
        type="email"
        bind:value={email}
        autocomplete="email"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.password")} *
      </label>
      <input
        id="password"
        type="password"
        bind:value={password}
        required
        minlength="8"
        autocomplete="new-password"
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
      />
    </div>

    <div>
      <label for="iin" class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.iin")}
      </label>
      <input
        id="iin"
        type="text"
        bind:value={iin}
        maxlength="12"
        inputmode="numeric"
        oninput={(e) => { iin = (e.target as HTMLInputElement).value.replace(/\D/g, ''); }}
        class="w-full px-3 py-2.5 bg-surface-low rounded-md text-on-surface font-mono tracking-wider
               placeholder:text-on-surface-variant/50
               focus:outline-2 focus:outline-primary focus:outline-offset-0
               transition-colors"
        placeholder="123456789012"
      />
    </div>

    <div>
      <label class="block text-sm font-medium text-on-surface mb-1.5">
        {$t("auth.register_as")} *
      </label>
      <div class="flex gap-3">
        <button
          type="button"
          onclick={() => (role = "student")}
          class="flex-1 py-2.5 rounded-md text-sm font-medium transition-colors
                 {role === 'student'
                   ? 'bg-gradient-to-br from-primary to-primary-container text-on-primary'
                   : 'bg-surface-low text-on-surface-variant hover:text-on-surface'}"
        >
          {$t("auth.role.student")}
        </button>
        <button
          type="button"
          onclick={() => (role = "teacher")}
          class="flex-1 py-2.5 rounded-md text-sm font-medium transition-colors
                 {role === 'teacher'
                   ? 'bg-gradient-to-br from-primary to-primary-container text-on-primary'
                   : 'bg-surface-low text-on-surface-variant hover:text-on-surface'}"
        >
          {$t("auth.role.teacher")}
        </button>
      </div>
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
    {loading ? "..." : $t("auth.register_action")}
  </button>

  <p class="text-center text-sm text-on-surface-variant mt-4">
    {$t("auth.have_account")}
    <a href="/login" class="text-primary font-medium hover:underline">
      {$t("auth.login")}
    </a>
  </p>
</form>
