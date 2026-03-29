<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/stores/auth';
  import { t } from '$lib/i18n';

  let username = $state('');
  let email = $state('');
  let password = $state('');
  let orgName = $state('');
  let loading = $state(false);
  let error = $state('');

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;

    try {
      const res: any = await api.post('/api/v1/auth/register/org', {
        username,
        email,
        password,
        org_name: orgName,
      });

      if (res.data?.access_token) {
        const { setAccessToken } = await import('$lib/api/client');
        setAccessToken(res.data.access_token);
        await auth.refresh();
        goto('/staff/events');
      }
    } catch (e: any) {
      error = e.message || 'Registration failed';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Register Organization — Jetistik</title>
</svelte:head>

<div class="w-full max-w-md">
  <div class="text-center mb-6">
    <h2 class="font-display text-2xl font-bold text-on-surface">
      {$t("registerOrg.title")}
    </h2>
    <p class="text-sm text-on-surface-variant mt-2">
      {$t("registerOrg.subtitle")}
    </p>
  </div>

  {#if error}
    <div class="bg-error-container text-on-error-container p-3 rounded-lg text-sm mb-4">
      {error}
    </div>
  {/if}

  <form onsubmit={handleSubmit} class="space-y-4">
    <div>
      <label for="orgName" class="block text-sm font-medium text-on-surface mb-1">
        {$t("registerOrg.orgName")}
      </label>
      <input
        id="orgName"
        type="text"
        bind:value={orgName}
        required
        class="w-full px-4 py-2.5 rounded-lg bg-surface-lowest text-on-surface text-sm border-0 border-b-2 border-outline-variant focus:border-primary outline-none transition-colors"
        placeholder={$t("registerOrg.orgPlaceholder")}
      />
    </div>

    <div>
      <label for="username" class="block text-sm font-medium text-on-surface mb-1">
        {$t("registerOrg.username")}
      </label>
      <input
        id="username"
        type="text"
        bind:value={username}
        required
        minlength="3"
        class="w-full px-4 py-2.5 rounded-lg bg-surface-lowest text-on-surface text-sm border-0 border-b-2 border-outline-variant focus:border-primary outline-none transition-colors"
      />
    </div>

    <div>
      <label for="email" class="block text-sm font-medium text-on-surface mb-1">
        {$t("registerOrg.email")}
      </label>
      <input
        id="email"
        type="email"
        bind:value={email}
        required
        class="w-full px-4 py-2.5 rounded-lg bg-surface-lowest text-on-surface text-sm border-0 border-b-2 border-outline-variant focus:border-primary outline-none transition-colors"
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-on-surface mb-1">
        {$t("registerOrg.password")}
      </label>
      <input
        id="password"
        type="password"
        bind:value={password}
        required
        minlength="8"
        class="w-full px-4 py-2.5 rounded-lg bg-surface-lowest text-on-surface text-sm border-0 border-b-2 border-outline-variant focus:border-primary outline-none transition-colors"
      />
    </div>

    <button
      type="submit"
      disabled={loading}
      class="w-full py-3 rounded-lg bg-gradient-to-br from-primary to-primary-container text-on-primary font-display font-semibold text-sm hover:opacity-90 transition-opacity disabled:opacity-50"
    >
      {loading ? $t('registerOrg.submitting') : $t('registerOrg.submit')}
    </button>
  </form>

  <p class="text-center text-sm text-on-surface-variant mt-6">
    {$t("registerOrg.hasAccount")} <a href="/login" class="text-primary hover:underline">{$t("registerOrg.loginLink")}</a>
  </p>
  <p class="text-center text-sm text-on-surface-variant mt-2">
    <a href="/register" class="text-primary hover:underline">{$t("registerOrg.studentTeacherLink")}</a>
  </p>
</div>
