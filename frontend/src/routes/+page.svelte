<script lang="ts">
  import { goto } from "$app/navigation";
  import { t, language, setLanguage, type Language } from "$lib/i18n";

  let iin = $state("");
  let iinError = $state("");
  let openFaq = $state<number | null>(null);

  const languages: { code: Language; label: string }[] = [
    { code: "kz", label: "KZ" },
    { code: "ru", label: "RU" },
    { code: "en", label: "EN" },
  ];

  function handleSearch() {
    const cleaned = iin.replace(/\s/g, "");
    if (!/^\d{12}$/.test(cleaned)) {
      iinError = $t("landing.hero.iinError");
      return;
    }
    iinError = "";
    goto(`/verify/${cleaned}`);
  }

  function toggleFaq(index: number) {
    openFaq = openFaq === index ? null : index;
  }

  const faqItems = [
    { q: "landing.faq.q1" as const, a: "landing.faq.a1" as const },
    { q: "landing.faq.q2" as const, a: "landing.faq.a2" as const },
    { q: "landing.faq.q3" as const, a: "landing.faq.a3" as const },
    { q: "landing.faq.q4" as const, a: "landing.faq.a4" as const },
  ];

  const steps = [
    {
      key: 1,
      title: "landing.step1.title" as const,
      desc: "landing.step1.desc" as const,
      icon: "keyboard",
    },
    {
      key: 2,
      title: "landing.step2.title" as const,
      desc: "landing.step2.desc" as const,
      icon: "list",
    },
    {
      key: 3,
      title: "landing.step3.title" as const,
      desc: "landing.step3.desc" as const,
      icon: "download",
    },
    {
      key: 4,
      title: "landing.step4.title" as const,
      desc: "landing.step4.desc" as const,
      icon: "qr",
    },
  ];
</script>

<div class="min-h-screen bg-surface-lowest">
  <!-- Header -->
  <header class="sticky top-0 z-50 bg-surface-lowest/80 backdrop-blur-xl">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <a href="/" class="font-display text-xl font-bold text-primary">Jetistik</a>

        <nav class="hidden md:flex items-center gap-6">
          <a href="#how-it-works" class="text-sm text-on-surface-variant hover:text-on-surface transition-colors">{$t("nav.howItWorks")}</a>
          <a href="#verification" class="text-sm text-on-surface-variant hover:text-on-surface transition-colors">{$t("nav.verification")}</a>
          <a href="#faq" class="text-sm text-on-surface-variant hover:text-on-surface transition-colors">{$t("nav.faq")}</a>
          <a href="#organizers" class="text-sm text-on-surface-variant hover:text-on-surface transition-colors">{$t("nav.forOrganizers")}</a>
        </nav>

        <div class="flex items-center gap-3">
          <!-- Language switcher -->
          <div class="flex items-center rounded-md bg-surface-low p-0.5">
            {#each languages as lang}
              <button
                onclick={() => setLanguage(lang.code)}
                class="px-2 py-1 rounded text-xs font-medium transition-colors
                  {$language === lang.code
                    ? 'bg-surface-lowest text-on-surface shadow-sm'
                    : 'text-on-surface-variant hover:text-on-surface'}"
              >
                {lang.label}
              </button>
            {/each}
          </div>

          <a href="/login" class="hidden sm:inline-flex text-sm font-medium text-on-surface-variant hover:text-on-surface transition-colors">
            {$t("nav.login")}
          </a>
          <a
            href="/register"
            class="inline-flex items-center px-4 py-2 rounded-lg text-sm font-medium bg-gradient-to-br from-primary to-primary-container text-on-primary hover:shadow-lg transition-shadow"
          >
            {$t("nav.register")}
          </a>
        </div>
      </div>
    </div>
  </header>

  <!-- Hero -->
  <section id="verification" class="relative overflow-hidden">
    <div class="absolute inset-0 bg-gradient-to-br from-primary/[0.03] to-transparent"></div>
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 pt-16 pb-24 sm:pt-24 sm:pb-32 relative">
      <div class="max-w-2xl">
        <h1 class="font-display text-4xl sm:text-5xl lg:text-6xl font-bold text-on-surface leading-tight tracking-tight">
          {$t("landing.hero.title")}
        </h1>
        <p class="mt-6 text-lg text-on-surface-variant leading-relaxed max-w-xl">
          {$t("landing.hero.subtitle")}
        </p>

        <!-- IIN Search -->
        <div class="mt-10 max-w-lg">
          <div class="flex gap-2">
            <div class="flex-1 relative">
              <input
                type="text"
                bind:value={iin}
                oninput={() => { iinError = ""; }}
                onkeydown={(e) => { if (e.key === "Enter") handleSearch(); }}
                maxlength={12}
                placeholder={$t("landing.hero.iinPlaceholder")}
                class="w-full px-4 py-3.5 rounded-lg bg-surface text-on-surface placeholder:text-on-surface-variant/50 text-base font-mono tracking-wider focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow {iinError ? 'ring-2 ring-error/30' : ''}"
              />
            </div>
            <button
              onclick={handleSearch}
              class="px-6 py-3.5 rounded-lg text-sm font-semibold bg-gradient-to-br from-primary to-primary-container text-on-primary hover:shadow-lg hover:shadow-primary/20 transition-all active:scale-[0.98] shrink-0"
            >
              {$t("landing.hero.search")}
            </button>
          </div>
          {#if iinError}
            <p class="mt-2 text-sm text-error">{iinError}</p>
          {/if}
        </div>
      </div>
    </div>
  </section>

  <!-- How it works -->
  <section id="how-it-works" class="py-20 sm:py-28 bg-surface">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
      <h2 class="font-display text-2xl sm:text-3xl font-bold text-on-surface text-center">
        {$t("landing.howItWorks")}
      </h2>

      <div class="mt-14 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {#each steps as step}
          <div class="bg-surface-lowest rounded-lg p-6 relative">
            <div class="w-10 h-10 rounded-lg bg-primary-fixed flex items-center justify-center mb-4">
              {#if step.icon === "keyboard"}
                <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1 1 21.75 8.25Z" />
                </svg>
              {:else if step.icon === "list"}
                <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
                </svg>
              {:else if step.icon === "download"}
                <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
                </svg>
              {:else if step.icon === "qr"}
                <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0 1 3.75 9.375v-4.5ZM3.75 14.625c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5a1.125 1.125 0 0 1-1.125-1.125v-4.5ZM13.5 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0 1 13.5 9.375v-4.5Z" />
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 6.75h.75v.75h-.75v-.75ZM6.75 16.5h.75v.75h-.75v-.75ZM16.5 6.75h.75v.75h-.75v-.75ZM13.5 13.5h.75v.75h-.75v-.75ZM13.5 19.5h.75v.75h-.75v-.75ZM19.5 13.5h.75v.75h-.75v-.75ZM19.5 19.5h.75v.75h-.75v-.75ZM16.5 16.5h.75v.75h-.75v-.75Z" />
                </svg>
              {/if}
            </div>
            <span class="absolute top-4 right-5 text-3xl font-display font-bold text-outline-variant/20">{step.key}</span>
            <h3 class="font-display text-base font-semibold text-on-surface mb-2">{$t(step.title)}</h3>
            <p class="text-sm text-on-surface-variant leading-relaxed">{$t(step.desc)}</p>
          </div>
        {/each}
      </div>
    </div>
  </section>

  <!-- Become organizer -->
  <section id="organizers" class="py-20 sm:py-28">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="bg-surface rounded-2xl p-8 sm:p-12 text-center max-w-2xl mx-auto">
        <h2 class="font-display text-2xl sm:text-3xl font-bold text-on-surface">
          {$t("landing.organizer.title")}
        </h2>
        <p class="mt-4 text-on-surface-variant leading-relaxed max-w-lg mx-auto">
          {$t("landing.organizer.desc")}
        </p>
        <a
          href="/register"
          class="mt-8 inline-flex items-center gap-2 px-6 py-3 rounded-lg text-sm font-semibold bg-gradient-to-br from-primary to-primary-container text-on-primary hover:shadow-lg hover:shadow-primary/20 transition-all"
        >
          {$t("landing.organizer.cta")}
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
          </svg>
        </a>
      </div>
    </div>
  </section>

  <!-- FAQ -->
  <section id="faq" class="py-20 sm:py-28 bg-surface">
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
      <h2 class="font-display text-2xl sm:text-3xl font-bold text-on-surface text-center mb-12">
        {$t("landing.faq.title")}
      </h2>

      <div class="space-y-2">
        {#each faqItems as item, i}
          <div class="bg-surface-lowest rounded-lg overflow-hidden">
            <button
              onclick={() => toggleFaq(i)}
              class="w-full flex items-center justify-between px-6 py-4 text-left"
            >
              <span class="font-medium text-on-surface text-sm sm:text-base">{$t(item.q)}</span>
              <svg
                class="w-5 h-5 text-on-surface-variant shrink-0 ml-4 transition-transform duration-200 {openFaq === i ? 'rotate-180' : ''}"
                fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" />
              </svg>
            </button>
            {#if openFaq === i}
              <div class="px-6 pb-4">
                <p class="text-sm text-on-surface-variant leading-relaxed">{$t(item.a)}</p>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  </section>

  <!-- Footer -->
  <footer class="bg-on-surface py-12 sm:py-16">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-8">
        <div>
          <span class="font-display text-lg font-bold text-surface-lowest">Jetistik</span>
          <p class="mt-2 text-sm text-on-surface-variant/80">{$t("app.tagline")}</p>
        </div>

        <div>
          <h4 class="text-sm font-semibold text-surface-lowest mb-3">{$t("landing.footer.platform")}</h4>
          <ul class="space-y-2">
            <li><a href="#how-it-works" class="text-sm text-on-surface-variant/70 hover:text-surface-lowest transition-colors">{$t("nav.howItWorks")}</a></li>
            <li><a href="#verification" class="text-sm text-on-surface-variant/70 hover:text-surface-lowest transition-colors">{$t("nav.verification")}</a></li>
            <li><a href="#organizers" class="text-sm text-on-surface-variant/70 hover:text-surface-lowest transition-colors">{$t("nav.forOrganizers")}</a></li>
          </ul>
        </div>

        <div>
          <h4 class="text-sm font-semibold text-surface-lowest mb-3">{$t("landing.footer.support")}</h4>
          <ul class="space-y-2">
            <li><a href="#faq" class="text-sm text-on-surface-variant/70 hover:text-surface-lowest transition-colors">{$t("nav.faq")}</a></li>
            <li><a href="mailto:support@jetistik.kz" class="text-sm text-on-surface-variant/70 hover:text-surface-lowest transition-colors">{$t("landing.footer.contact")}</a></li>
            <li><a href="/privacy" class="text-sm text-on-surface-variant/70 hover:text-surface-lowest transition-colors">{$t("landing.footer.privacy")}</a></li>
          </ul>
        </div>
      </div>

      <div class="mt-10 pt-6 border-t border-surface-lowest/10 flex flex-col sm:flex-row items-center justify-between gap-3">
        <p class="text-xs text-on-surface-variant/50">&copy; {new Date().getFullYear()} Jetistik. {$t("landing.footer.rights")}</p>
      </div>
    </div>
  </footer>
</div>
