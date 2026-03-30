<script lang="ts">
  import { goto } from '$app/navigation';
  import { t } from '$lib/i18n';

  let iin = $state('');
  let error = $state('');

  function handleSearch() {
    const cleaned = iin.replace(/\D/g, '');
    if (cleaned.length !== 12) {
      error = $t("verify.iinError");
      return;
    }
    error = '';
    goto(`/verify/${cleaned}`);
  }
</script>

<svelte:head>
  <title>{$t("verify.title")} — Jetistik</title>
</svelte:head>

<div class="min-h-screen bg-surface flex flex-col">
  <!-- Header -->
  <header class="bg-white/80 backdrop-blur-xl border-b border-outline-variant/15 sticky top-0 z-50">
    <div class="max-w-5xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
      <a href="/" class="text-2xl font-display font-bold tracking-tighter text-primary">Jetistik</a>
      <a href="/login" class="text-sm font-medium text-on-surface-variant hover:text-primary transition-colors">{$t("nav.login")}</a>
    </div>
  </header>

  <!-- Main -->
  <main class="flex-1 flex items-center justify-center px-4 py-16">
    <div class="w-full max-w-lg text-center">
      <!-- Icon -->
      <div class="w-20 h-20 mx-auto mb-8 rounded-2xl bg-gradient-to-br from-primary to-primary-container flex items-center justify-center shadow-xl shadow-primary/20">
        <span class="material-symbols-outlined text-white text-4xl">verified_user</span>
      </div>

      <h1 class="font-display text-3xl sm:text-4xl font-extrabold text-on-surface tracking-tight mb-3">
        {$t("verify.pageTitle")}
      </h1>
      <p class="text-on-surface-variant text-base mb-10 max-w-md mx-auto leading-relaxed">
        {$t("verify.pageDesc")}
      </p>

      <!-- Search form -->
      <div class="bg-surface-container-lowest p-2 rounded-2xl shadow-xl border border-outline-variant/10 max-w-md mx-auto">
        <div class="flex gap-2">
          <div class="flex-1 relative">
            <span class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-outline text-xl">badge</span>
            <input
              type="text"
              bind:value={iin}
              oninput={() => { error = ''; }}
              onkeydown={(e) => { if (e.key === 'Enter') handleSearch(); }}
              maxlength="12"
              inputmode="numeric"
              placeholder={$t("verify.iinPlaceholder")}
              class="w-full pl-12 pr-4 py-4 bg-transparent border-none focus:ring-0 text-on-surface font-mono tracking-wider placeholder:text-outline/50 text-base"
            />
          </div>
          <button
            onclick={handleSearch}
            class="bg-gradient-to-br from-primary to-primary-container text-white font-bold px-6 py-4 rounded-xl flex items-center gap-2 hover:shadow-lg hover:shadow-primary/20 transition-all active:scale-95 shrink-0"
          >
            <span class="material-symbols-outlined text-xl">search</span>
            <span class="hidden sm:inline">{$t("verify.search")}</span>
          </button>
        </div>
      </div>

      {#if error}
        <p class="mt-3 text-sm text-error">{error}</p>
      {/if}

      <!-- Or verify by code -->
      <div class="mt-8 pt-8 border-t border-outline-variant/10 max-w-md mx-auto">
        <p class="text-sm text-on-surface-variant mb-4">{$t("verify.orByCode")}</p>
        <div class="flex gap-2">
          <input
            type="text"
            id="code-input"
            placeholder={$t("verify.codePlaceholder")}
            class="flex-1 px-4 py-3 rounded-xl bg-surface-container-low text-on-surface text-sm font-mono border-0 focus:ring-2 focus:ring-primary/20 placeholder:text-outline/40"
            onkeydown={(e) => {
              if (e.key === 'Enter') {
                const val = (e.target as HTMLInputElement).value.trim();
                if (val) goto(`/verify/${val}`);
              }
            }}
          />
          <button
            onclick={() => {
              const el = document.getElementById('code-input') as HTMLInputElement;
              if (el?.value.trim()) goto(`/verify/${el.value.trim()}`);
            }}
            class="px-4 py-3 rounded-xl bg-surface-container-low text-on-surface-variant hover:bg-surface-container-high transition-colors"
          >
            <span class="material-symbols-outlined text-xl">arrow_forward</span>
          </button>
        </div>
      </div>

      <!-- How it works -->
      <div class="mt-12 grid grid-cols-1 sm:grid-cols-3 gap-4 max-w-lg mx-auto text-left">
        <div class="p-4 rounded-xl bg-surface-container-lowest border border-outline-variant/5">
          <div class="w-8 h-8 rounded-lg bg-primary-fixed flex items-center justify-center mb-3">
            <span class="material-symbols-outlined text-primary text-lg">pin</span>
          </div>
          <p class="text-xs font-bold text-on-surface mb-1">{$t("verify.step1Title")}</p>
          <p class="text-[11px] text-on-surface-variant">{$t("verify.step1Desc")}</p>
        </div>
        <div class="p-4 rounded-xl bg-surface-container-lowest border border-outline-variant/5">
          <div class="w-8 h-8 rounded-lg bg-primary-fixed flex items-center justify-center mb-3">
            <span class="material-symbols-outlined text-primary text-lg">list_alt</span>
          </div>
          <p class="text-xs font-bold text-on-surface mb-1">{$t("verify.step2Title")}</p>
          <p class="text-[11px] text-on-surface-variant">{$t("verify.step2Desc")}</p>
        </div>
        <div class="p-4 rounded-xl bg-surface-container-lowest border border-outline-variant/5">
          <div class="w-8 h-8 rounded-lg bg-primary-fixed flex items-center justify-center mb-3">
            <span class="material-symbols-outlined text-primary text-lg">verified</span>
          </div>
          <p class="text-xs font-bold text-on-surface mb-1">{$t("verify.step3Title")}</p>
          <p class="text-[11px] text-on-surface-variant">{$t("verify.step3Desc")}</p>
        </div>
      </div>
    </div>
  </main>

  <!-- Footer -->
  <footer class="border-t border-outline-variant/10 py-6">
    <div class="max-w-5xl mx-auto px-4 sm:px-6 flex items-center justify-between">
      <p class="text-xs text-on-surface-variant">&copy; {new Date().getFullYear()} Jetistik</p>
      <div class="flex gap-4">
        <a href="/privacy" class="text-xs text-on-surface-variant hover:text-primary transition-colors">{$t("landing.footer.privacy")}</a>
        <a href="/terms" class="text-xs text-on-surface-variant hover:text-primary transition-colors">{$t("landing.footer.terms")}</a>
      </div>
    </div>
  </footer>
</div>
