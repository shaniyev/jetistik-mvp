<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { t, language, setLanguage, type Language } from "$lib/i18n";
  import { auth, isAuthenticated, currentUser } from "$lib/stores/auth";

  let loggedIn = $state(false);
  let userRole = $state("");

  onMount(() => {
    const unsub = isAuthenticated.subscribe(val => { loggedIn = val; });
    const unsub2 = currentUser.subscribe(u => { userRole = u?.role ?? ""; });
    auth.refresh();
    return () => { unsub(); unsub2(); };
  });

  function getDashboardPath(role: string): string {
    switch (role) {
      case "admin": return "/admin/organizations";
      case "staff": return "/staff/events";
      case "teacher": return "/teacher";
      case "student": return "/student";
      default: return "/";
    }
  }

  let iin = $state("");
  let iinError = $state("");
  let openFaq = $state<number | null>(null);

  // Organizer form
  let orgFullName = $state("");
  let orgOrganization = $state("");
  let orgPhone = $state("");
  let orgSending = $state(false);
  let orgMessage = $state("");
  let orgMessageType = $state<"success" | "error" | "">("");

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

  async function submitOrgRequest() {
    if (!orgFullName.trim() || !orgOrganization.trim() || !orgPhone.trim()) return;
    orgSending = true;
    orgMessage = "";
    try {
      const res = await fetch("/api/v1/organizer-request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          fullName: orgFullName.trim(),
          organization: orgOrganization.trim(),
          phone: orgPhone.trim(),
        }),
      });
      if (res.ok) {
        orgMessage = $t("landing.organizer.success");
        orgMessageType = "success";
        orgFullName = "";
        orgOrganization = "";
        orgPhone = "";
      } else {
        orgMessage = $t("landing.organizer.error");
        orgMessageType = "error";
      }
    } catch {
      orgMessage = $t("landing.organizer.error");
      orgMessageType = "error";
    } finally {
      orgSending = false;
    }
  }

  const faqItems = [
    { q: "landing.faq.q1New" as const, a: "landing.faq.a1New" as const },
    { q: "landing.faq.q2New" as const, a: "landing.faq.a2New" as const },
    { q: "landing.faq.q3New" as const, a: "landing.faq.a3New" as const },
  ];
</script>

<svelte:head>
  <style>
    .glass-header {
      backdrop-filter: blur(20px);
      background: rgba(255, 255, 255, 0.8);
    }
    .hero-gradient {
      background: linear-gradient(135deg, #004ac6 0%, #2563eb 100%);
    }
    .hero-gradient-text {
      background: linear-gradient(135deg, #004ac6 0%, #2563eb 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
  </style>
</svelte:head>

<div class="bg-surface text-on-surface selection:bg-primary-fixed selection:text-on-primary-fixed">
  <!-- Header -->
  <header class="fixed top-0 w-full z-50 glass-header border-b border-slate-200/50">
    <div class="flex justify-between items-center px-6 py-4 max-w-screen-2xl mx-auto">
      <div class="flex items-center gap-12">
        <a class="text-2xl font-bold tracking-tighter text-blue-700 font-display" href="/">Jetistik</a>
        <nav class="hidden md:flex gap-8 items-center font-display font-semibold tracking-tight text-sm">
          <a class="text-slate-600 hover:text-blue-500 transition-colors" href="#how-it-works">{$t("nav.howItWorks")}</a>
          <a class="text-slate-600 hover:text-blue-500 transition-colors" href="#faq">{$t("nav.faq")}</a>
          <a class="text-slate-600 hover:text-blue-500 transition-colors" href="#organizers">{$t("nav.forOrganizers")}</a>
        </nav>
      </div>
      <div class="flex items-center gap-4">
        <!-- Language switcher -->
        <div class="flex items-center gap-1 mr-4 px-3 py-1 bg-surface-container rounded-full text-[10px] font-bold tracking-widest uppercase">
          {#each languages as lang, i}
            {#if i > 0}
              <span class="text-outline-variant">/</span>
            {/if}
            <button
              onclick={() => setLanguage(lang.code)}
              class="transition-colors {$language === lang.code ? 'text-primary' : 'text-outline'}"
            >
              {lang.label}
            </button>
          {/each}
        </div>
        {#if loggedIn}
          <a href={getDashboardPath(userRole)} class="bg-primary text-white text-sm font-semibold px-5 py-2 rounded-lg shadow-sm hover:opacity-90 transition-all active:scale-95">
            Dashboard
          </a>
        {:else}
          <a href="/login" class="hidden sm:inline-flex text-sm font-semibold text-slate-600 hover:bg-slate-50/50 px-4 py-2 rounded-lg transition-all active:scale-95">
            {$t("nav.login")}
          </a>
          <a href="/register" class="bg-primary text-white text-sm font-semibold px-5 py-2 rounded-lg shadow-sm hover:opacity-90 transition-all active:scale-95">
          {$t("nav.register")}
        </a>
        {/if}
      </div>
    </div>
  </header>

  <main class="pt-24">
    <!-- Hero Section -->
    <section class="relative overflow-hidden px-6 pt-16 pb-24 md:pt-32 md:pb-40">
      <div class="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
        <div class="z-10">
          <!-- Badge -->
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary-fixed text-on-primary-fixed-variant text-xs font-bold uppercase tracking-widest mb-6">
            <!-- verified icon -->
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
            </svg>
            {$t("landing.hero.badge")}
          </div>

          <h1 class="text-5xl md:text-7xl font-extrabold tracking-tighter leading-[1.1] mb-6 text-on-surface font-display">
            {$t("landing.hero.titleStart")} <br/>
            <span class="hero-gradient-text">{$t("landing.hero.titleGradient")}</span>
          </h1>

          <p class="text-lg text-on-surface-variant mb-10 max-w-lg leading-relaxed">
            {$t("landing.hero.subtitleMain")}
            <span class="block mt-2 text-sm italic opacity-75">{$t("landing.hero.subtitleKz")}</span>
          </p>

          <!-- IIN Search Card -->
          <div class="bg-surface-container-lowest p-2 rounded-xl shadow-xl border border-outline-variant/10 flex flex-col md:flex-row gap-2 max-w-xl">
            <div class="flex-1 relative">
              <!-- badge icon -->
              <svg class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-outline" fill="currentColor" viewBox="0 0 24 24">
                <path d="M20 7h-5V4c0-1.1-.9-2-2-2h-2c-1.1 0-2 .9-2 2v3H4c-1.1 0-2 .9-2 2v11c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V9c0-1.1-.9-2-2-2zm-9-3h2v5h-2V4zm0 12H9v-2h2v2zm4 0h-2v-2h2v2zm4 0h-2v-2h2v2z"/>
              </svg>
              <input
                type="text"
                bind:value={iin}
                oninput={() => { iinError = ""; }}
                onkeydown={(e) => { if (e.key === "Enter") handleSearch(); }}
                maxlength={12}
                placeholder={$t("landing.hero.iinPlaceholder")}
                class="w-full pl-12 pr-4 py-4 bg-transparent border-none focus:ring-0 text-on-surface font-medium placeholder:text-outline/60"
              />
            </div>
            <button
              onclick={handleSearch}
              class="hero-gradient text-white font-bold px-8 py-4 rounded-lg flex items-center justify-center gap-2 transition-transform hover:scale-[1.02] active:scale-95"
            >
              {$t("landing.hero.showCerts")}
              <!-- arrow_forward icon -->
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 4l-1.41 1.41L16.17 11H4v2h12.17l-5.58 5.59L12 20l8-8z"/>
              </svg>
            </button>
          </div>
          {#if iinError}
            <p class="mt-2 text-sm text-error">{iinError}</p>
          {/if}
        </div>

        <!-- Certificate Mockup -->
        <div class="relative hidden lg:block">
          <div class="absolute -top-20 -right-20 w-96 h-96 bg-primary/5 rounded-full blur-3xl"></div>
          <div class="relative z-10 bg-surface-container-lowest p-8 rounded-3xl shadow-2xl border border-outline-variant/20 rotate-3">
            <div class="flex justify-between items-start mb-8">
              <div class="h-12 w-32 bg-slate-100 rounded animate-pulse"></div>
              <div class="p-2 bg-surface rounded-lg border border-primary/20">
                <!-- QR placeholder -->
                <svg class="w-16 h-16 text-primary/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0 1 3.75 9.375v-4.5ZM3.75 14.625c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5a1.125 1.125 0 0 1-1.125-1.125v-4.5ZM13.5 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0 1 13.5 9.375v-4.5Z" />
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 6.75h.75v.75h-.75v-.75ZM6.75 16.5h.75v.75h-.75v-.75ZM16.5 6.75h.75v.75h-.75v-.75ZM13.5 13.5h.75v.75h-.75v-.75ZM13.5 19.5h.75v.75h-.75v-.75ZM19.5 13.5h.75v.75h-.75v-.75ZM19.5 19.5h.75v.75h-.75v-.75ZM16.5 16.5h.75v.75h-.75v-.75Z" />
                </svg>
              </div>
            </div>
            <div class="space-y-4 mb-8">
              <div class="h-8 w-3/4 bg-slate-100 rounded"></div>
              <div class="h-4 w-1/2 bg-slate-50 rounded"></div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="h-12 bg-slate-50 rounded-lg"></div>
              <div class="h-12 bg-slate-50 rounded-lg"></div>
            </div>
            <div class="mt-8 pt-8 border-t border-slate-100 flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-primary-fixed"></div>
              <div>
                <div class="h-4 w-24 bg-slate-100 rounded mb-1"></div>
                <div class="h-3 w-16 bg-slate-50 rounded"></div>
              </div>
            </div>
          </div>

          <!-- Floating Verified badge -->
          <div class="absolute -bottom-10 -left-10 z-20 bg-white p-6 rounded-2xl shadow-xl border border-outline-variant/20 -rotate-6">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-primary-fixed rounded-lg">
                <svg class="w-5 h-5 text-primary" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z"/>
                </svg>
              </div>
              <div>
                <p class="text-xs font-bold text-on-surface">{$t("landing.hero.verifiedSystem")}</p>
                <p class="text-[10px] text-on-surface-variant">{$t("landing.hero.blockchainSecured")}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- How it works -->
    <section class="py-24 bg-surface-container-low" id="how-it-works">
      <div class="max-w-7xl mx-auto px-6 text-center mb-16">
        <h2 class="text-3xl md:text-5xl font-extrabold tracking-tight mb-4 font-display">{$t("landing.howItWorks")}</h2>
        <p class="text-on-surface-variant max-w-2xl mx-auto">{$t("landing.howItWorksSubtitle")}</p>
      </div>
      <div class="max-w-7xl mx-auto px-6 grid grid-cols-1 md:grid-cols-3 gap-8">
        <!-- Step 1: Upload Template -->
        <div class="bg-surface-container-lowest p-8 rounded-2xl border border-outline-variant/5 shadow-sm hover:shadow-md transition-shadow">
          <div class="w-14 h-14 bg-primary-fixed text-primary rounded-xl flex items-center justify-center mb-6">
            <svg class="w-7 h-7" fill="currentColor" viewBox="0 0 24 24">
              <path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6zm4 18H6V4h7v5h5v11zm-4-6v4h-4v-4H7l5-5 5 5h-3z"/>
            </svg>
          </div>
          <h3 class="text-xl font-bold mb-3 font-display">{$t("landing.step1.titleNew")}</h3>
          <p class="text-on-surface-variant text-sm leading-relaxed mb-4">{$t("landing.step1.descNew")}</p>
          <p class="text-[11px] font-bold text-primary/60 uppercase tracking-widest">{$t("landing.step1.kz")}</p>
        </div>

        <!-- Step 2: Import Data -->
        <div class="bg-surface-container-lowest p-8 rounded-2xl border border-outline-variant/5 shadow-sm hover:shadow-md transition-shadow">
          <div class="w-14 h-14 bg-primary-fixed text-primary rounded-xl flex items-center justify-center mb-6">
            <svg class="w-7 h-7" fill="currentColor" viewBox="0 0 24 24">
              <path d="M22 9V7h-2V5c0-1.1-.9-2-2-2H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2v-2h2v-2h-2v-2h2v-2h-2V9h2zm-4 10H4V5h14v14zM6 13h5v4H6v-4zm6-6h4v3h-4V7zM6 7h5v5H6V7zm6 4h4v6h-4v-6z"/>
            </svg>
          </div>
          <h3 class="text-xl font-bold mb-3 font-display">{$t("landing.step2.titleNew")}</h3>
          <p class="text-on-surface-variant text-sm leading-relaxed mb-4">{$t("landing.step2.descNew")}</p>
          <p class="text-[11px] font-bold text-primary/60 uppercase tracking-widest">{$t("landing.step2.kz")}</p>
        </div>

        <!-- Step 3: Generate QR -->
        <div class="bg-surface-container-lowest p-8 rounded-2xl border border-outline-variant/5 shadow-sm hover:shadow-md transition-shadow">
          <div class="w-14 h-14 bg-primary-fixed text-primary rounded-xl flex items-center justify-center mb-6">
            <svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0 1 3.75 9.375v-4.5ZM3.75 14.625c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5a1.125 1.125 0 0 1-1.125-1.125v-4.5ZM13.5 4.875c0-.621.504-1.125 1.125-1.125h4.5c.621 0 1.125.504 1.125 1.125v4.5c0 .621-.504 1.125-1.125 1.125h-4.5A1.125 1.125 0 0 1 13.5 9.375v-4.5Z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 6.75h.75v.75h-.75v-.75ZM6.75 16.5h.75v.75h-.75v-.75ZM16.5 6.75h.75v.75h-.75v-.75ZM13.5 13.5h.75v.75h-.75v-.75ZM13.5 19.5h.75v.75h-.75v-.75ZM19.5 13.5h.75v.75h-.75v-.75ZM19.5 19.5h.75v.75h-.75v-.75ZM16.5 16.5h.75v.75h-.75v-.75Z" />
            </svg>
          </div>
          <h3 class="text-xl font-bold mb-3 font-display">{$t("landing.step3.titleNew")}</h3>
          <p class="text-on-surface-variant text-sm leading-relaxed mb-4">{$t("landing.step3.descNew")}</p>
          <p class="text-[11px] font-bold text-primary/60 uppercase tracking-widest">{$t("landing.step3.kz")}</p>
        </div>
      </div>
    </section>

    <!-- For Organizers -->
    <section class="py-24 overflow-hidden" id="organizers">
      <div class="max-w-7xl mx-auto px-6 grid grid-cols-1 lg:grid-cols-2 gap-16 items-center">
        <div>
          <h2 class="text-3xl md:text-5xl font-extrabold tracking-tight mb-6 font-display">{$t("landing.organizer.titleNew")}</h2>
          <p class="text-on-surface-variant text-lg leading-relaxed mb-8">
            {$t("landing.organizer.descNew")}
          </p>
          <ul class="space-y-4">
            <li class="flex items-center gap-3">
              <svg class="w-6 h-6 text-primary shrink-0" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
              <span class="font-medium">{$t("landing.organizer.check1")}</span>
            </li>
            <li class="flex items-center gap-3">
              <svg class="w-6 h-6 text-primary shrink-0" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
              <span class="font-medium">{$t("landing.organizer.check2")}</span>
            </li>
            <li class="flex items-center gap-3">
              <svg class="w-6 h-6 text-primary shrink-0" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
              <span class="font-medium">{$t("landing.organizer.check3")}</span>
            </li>
          </ul>
        </div>

        <!-- Application Form -->
        <div class="bg-surface-container-low p-8 md:p-12 rounded-3xl border border-outline-variant/20">
          <h3 class="text-2xl font-bold mb-8 font-display">{$t("landing.organizer.formTitle")}</h3>
          <form class="space-y-6" onsubmit={(e) => { e.preventDefault(); submitOrgRequest(); }}>
            <div class="space-y-1">
              <label class="text-xs font-bold uppercase tracking-wider text-outline">{$t("landing.organizer.nameLabel")}</label>
              <input
                bind:value={orgFullName}
                type="text"
                placeholder={$t("landing.organizer.namePlaceholder")}
                class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 focus:border-primary focus:ring-0 px-0 py-3 transition-colors"
              />
            </div>
            <div class="space-y-1">
              <label class="text-xs font-bold uppercase tracking-wider text-outline">{$t("landing.organizer.orgLabel")}</label>
              <input
                bind:value={orgOrganization}
                type="text"
                placeholder={$t("landing.organizer.orgPlaceholder")}
                class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 focus:border-primary focus:ring-0 px-0 py-3 transition-colors"
              />
            </div>
            <div class="space-y-1">
              <label class="text-xs font-bold uppercase tracking-wider text-outline">{$t("landing.organizer.phoneLabel")}</label>
              <input
                bind:value={orgPhone}
                type="tel"
                placeholder={$t("landing.organizer.phonePlaceholder")}
                class="w-full bg-surface-container-lowest border-0 border-b-2 border-outline-variant/30 focus:border-primary focus:ring-0 px-0 py-3 transition-colors"
              />
            </div>
            <button
              type="submit"
              disabled={orgSending}
              class="w-full bg-on-surface text-white font-bold py-4 rounded-xl shadow-lg hover:bg-black transition-colors disabled:opacity-50"
            >
              {orgSending ? $t("landing.organizer.sending") : $t("landing.organizer.submit")}
            </button>
            {#if orgMessage}
              <p class="text-sm {orgMessageType === 'success' ? 'text-green-600' : 'text-error'}">{orgMessage}</p>
            {/if}
          </form>
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="py-24 bg-surface" id="faq">
      <div class="max-w-4xl mx-auto px-6">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-5xl font-extrabold tracking-tight mb-4 font-display">{$t("landing.faq.title")}</h2>
          <p class="text-on-surface-variant uppercase tracking-[0.2em] text-[10px] font-bold">{$t("landing.faq.subtitleKz")}</p>
        </div>
        <div class="space-y-4">
          {#each faqItems as item, i}
            <div class="bg-surface-container-lowest rounded-2xl border border-outline-variant/10 overflow-hidden">
              <button
                onclick={() => toggleFaq(i)}
                class="w-full px-8 py-6 text-left flex justify-between items-center hover:bg-slate-50 transition-colors"
              >
                <span class="font-bold text-on-surface">{$t(item.q)}</span>
                <svg
                  class="w-5 h-5 text-outline shrink-0 ml-4 transition-transform duration-200 {openFaq === i ? 'rotate-180' : ''}"
                  fill="currentColor" viewBox="0 0 24 24"
                >
                  <path d="M16.59 8.59L12 13.17 7.41 8.59 6 10l6 6 6-6z"/>
                </svg>
              </button>
              {#if openFaq === i}
                <div class="px-8 pb-6 text-on-surface-variant text-sm leading-relaxed">
                  {$t(item.a)}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    </section>
  </main>

  <!-- Footer -->
  <footer class="bg-slate-950 text-white w-full py-16 px-6 mt-auto">
    <div class="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-4 gap-12 mb-12">
      <div class="col-span-1">
        <a class="text-3xl font-bold tracking-tighter text-white mb-6 block font-display" href="/">Jetistik</a>
        <p class="text-slate-400 text-sm leading-relaxed">
          {$t("landing.footer.brandDesc")}
        </p>
      </div>
      <div>
        <h4 class="text-xs font-bold uppercase tracking-widest text-slate-500 mb-6">{$t("landing.footer.platform")}</h4>
        <ul class="space-y-4 text-sm text-slate-300">
          <li><a class="hover:text-primary transition-colors" href="#how-it-works">{$t("nav.howItWorks")}</a></li>
          <li><a class="hover:text-primary transition-colors" href="#organizers">{$t("nav.forOrganizers")}</a></li>
          <li><a class="hover:text-primary transition-colors" href="/verify">{$t("landing.footer.verification")}</a></li>
        </ul>
      </div>
      <div>
        <h4 class="text-xs font-bold uppercase tracking-widest text-slate-500 mb-6">{$t("landing.footer.support")}</h4>
        <ul class="space-y-4 text-sm text-slate-300">
          <li><a class="hover:text-primary transition-colors" href="#faq">{$t("nav.faq")}</a></li>
          <li><a class="hover:text-primary transition-colors" href="mailto:support@jetistik.kz">{$t("landing.footer.contact")}</a></li>
          <li><a class="hover:text-primary transition-colors" href="#">{$t("landing.footer.help")}</a></li>
        </ul>
      </div>
      <div>
        <h4 class="text-xs font-bold uppercase tracking-widest text-slate-500 mb-6">{$t("landing.footer.legal")}</h4>
        <ul class="space-y-4 text-sm text-slate-300">
          <li><a class="hover:text-primary transition-colors" href="/privacy">{$t("landing.footer.privacy")}</a></li>
          <li><a class="hover:text-primary transition-colors" href="/terms">{$t("landing.footer.terms")}</a></li>
        </ul>
      </div>
    </div>
    <div class="max-w-7xl mx-auto pt-8 border-t border-slate-900 flex flex-col md:flex-row justify-between items-center gap-4">
      <p class="text-slate-500 text-xs">&copy; {new Date().getFullYear()} Jetistik Certificate Platform. {$t("landing.footer.rights")}</p>
      <div class="flex gap-6">
        <a class="text-slate-500 hover:text-white transition-colors" href="#">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>
        </a>
        <a class="text-slate-500 hover:text-white transition-colors" href="mailto:support@jetistik.kz">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/></svg>
        </a>
        <a class="text-slate-500 hover:text-white transition-colors" href="#">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M18 16.08c-.76 0-1.44.3-1.96.77L8.91 12.7c.05-.23.09-.46.09-.7s-.04-.47-.09-.7l7.05-4.11c.54.5 1.25.81 2.04.81 1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3c0 .24.04.47.09.7L8.04 9.81C7.5 9.31 6.79 9 6 9c-1.66 0-3 1.34-3 3s1.34 3 3 3c.79 0 1.5-.31 2.04-.81l7.12 4.16c-.05.21-.08.43-.08.65 0 1.61 1.31 2.92 2.92 2.92 1.61 0 2.92-1.31 2.92-2.92s-1.31-2.92-2.92-2.92z"/></svg>
        </a>
      </div>
    </div>
  </footer>

  <!-- Floating Chat Button -->
  <button class="fixed bottom-8 right-8 w-14 h-14 bg-primary text-white rounded-full shadow-2xl flex items-center justify-center hover:scale-110 active:scale-90 transition-transform z-40">
    <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
      <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/>
    </svg>
  </button>
</div>
