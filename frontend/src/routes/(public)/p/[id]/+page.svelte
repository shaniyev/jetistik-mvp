<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '$lib/i18n';

  let { data }: { data: { id: string } } = $props();

  interface PublicProfile {
    id: number;
    username: string;
    role: string;
    member_since: string;
    certificates: Certificate[];
    stats: {
      total_certificates: number;
      valid_certificates: number;
      organizations: number;
    };
  }

  interface Certificate {
    code: string;
    name: string;
    event_title: string;
    org_name: string;
    status: string;
    created_at: string;
  }

  let profile = $state<PublicProfile | null>(null);
  let loading = $state(true);
  let error = $state(false);
  let linkCopied = $state(false);

  const API_BASE = import.meta.env.VITE_API_URL || '';

  onMount(async () => {
    try {
      const res = await fetch(`${API_BASE}/api/v1/p/${data.id}`);
      if (!res.ok) { error = true; return; }
      const body = await res.json();
      profile = body.data;
    } catch {
      error = true;
    } finally {
      loading = false;
    }
  });

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(window.location.href);
      linkCopied = true;
      setTimeout(() => { linkCopied = false; }, 2000);
    } catch {}
  }

  function formatDate(d: string) {
    return new Date(d).toLocaleDateString('ru-RU', { year: 'numeric', month: 'long', day: 'numeric' });
  }

  function formatMonth(d: string) {
    return new Date(d).toLocaleDateString('ru-RU', { year: 'numeric', month: 'short' });
  }

  let validCerts = $derived(profile?.certificates?.filter(c => c.status === 'valid') ?? []);
  let revokedCerts = $derived(profile?.certificates?.filter(c => c.status === 'revoked') ?? []);
</script>

<svelte:head>
  <title>{profile?.username ?? 'Profile'} — Jetistik</title>
  <meta name="description" content="Certificate portfolio on Jetistik" />
</svelte:head>

<div class="min-h-screen bg-surface">
  <!-- Header -->
  <header class="bg-white/80 backdrop-blur-xl border-b border-outline-variant/15 sticky top-0 z-50">
    <div class="max-w-5xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
      <a href="/" class="text-2xl font-display font-bold tracking-tighter text-primary">Jetistik</a>
      <button
        onclick={copyLink}
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-surface-container-low hover:bg-surface-container-high transition-colors"
      >
        <span class="material-symbols-outlined text-[18px]">{linkCopied ? 'check' : 'share'}</span>
        {linkCopied ? $t("common.copied") : $t("student.shareProfile")}
      </button>
    </div>
  </header>

  {#if loading}
    <div class="flex items-center justify-center py-32">
      <div class="text-on-surface-variant">{$t("common.loading")}</div>
    </div>

  {:else if error || !profile}
    <div class="flex items-center justify-center py-32">
      <div class="text-center">
        <span class="material-symbols-outlined text-6xl text-outline-variant mb-4 block">person_off</span>
        <h2 class="font-display text-2xl font-bold text-on-surface mb-2">{$t("verify.notFoundTitle")}</h2>
        <a href="/" class="text-primary hover:underline text-sm">{$t("nav.home")}</a>
      </div>
    </div>

  {:else}
    <main class="max-w-5xl mx-auto px-4 sm:px-6 py-8 sm:py-12">

      <!-- Profile Hero -->
      <div class="bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 rounded-[2rem] p-8 sm:p-12 text-white mb-8 relative overflow-hidden">
        <div class="absolute -right-20 -top-20 w-80 h-80 bg-primary/20 rounded-full blur-[100px]"></div>
        <div class="absolute -left-10 -bottom-10 w-60 h-60 bg-blue-500/10 rounded-full blur-[80px]"></div>

        <div class="relative z-10 flex flex-col sm:flex-row gap-6 items-center sm:items-start">
          <!-- Avatar -->
          <div class="w-24 h-24 sm:w-28 sm:h-28 rounded-[1.5rem] bg-gradient-to-br from-primary to-primary-container flex items-center justify-center text-white font-display text-4xl sm:text-5xl font-bold shadow-2xl shrink-0">
            {profile.username[0]?.toUpperCase() ?? '?'}
          </div>

          <div class="text-center sm:text-left flex-1">
            <h1 class="font-display text-3xl sm:text-4xl font-extrabold tracking-tight mb-1">
              {profile.username}
            </h1>
            <div class="flex items-center justify-center sm:justify-start gap-3 mt-2 flex-wrap">
              <span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/10 text-xs font-semibold uppercase tracking-wider backdrop-blur-sm">
                <span class="material-symbols-outlined text-[14px]">school</span>
                {profile.role}
              </span>
              <span class="text-white/50 text-xs">
                {$t("student.memberSince")} {formatMonth(profile.member_since)}
              </span>
            </div>

            <!-- Stats -->
            <div class="flex gap-6 mt-6 justify-center sm:justify-start">
              <div>
                <div class="text-3xl font-display font-extrabold">{profile.stats.total_certificates}</div>
                <div class="text-white/50 text-xs uppercase tracking-wider mt-0.5">{$t("student.totalCerts")}</div>
              </div>
              <div class="w-px bg-white/10"></div>
              <div>
                <div class="text-3xl font-display font-extrabold text-emerald-400">{profile.stats.valid_certificates}</div>
                <div class="text-white/50 text-xs uppercase tracking-wider mt-0.5">{$t("student.filter.valid")}</div>
              </div>
              <div class="w-px bg-white/10"></div>
              <div>
                <div class="text-3xl font-display font-extrabold">{profile.stats.organizations}</div>
                <div class="text-white/50 text-xs uppercase tracking-wider mt-0.5">{$t("student.orgsCount")}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Certificates -->
      <div>
        <div class="flex items-center justify-between mb-6">
          <h2 class="font-display text-2xl font-bold text-on-surface">{$t("student.title")}</h2>
          <span class="text-sm text-on-surface-variant">{profile.certificates.length} {$t("staff.certs.total")}</span>
        </div>

        {#if profile.certificates.length === 0}
          <div class="bg-surface-container-low rounded-2xl p-16 text-center border-2 border-dashed border-outline-variant/20">
            <span class="material-symbols-outlined text-5xl text-outline-variant mb-3 block">workspace_premium</span>
            <p class="text-on-surface-variant">{$t("student.noCertificates")}</p>
          </div>
        {:else}
          <div class="grid gap-4">
            {#each profile.certificates as cert}
              <a
                href="/verify/{cert.code}"
                class="bg-surface-container-lowest rounded-2xl p-6 flex flex-col sm:flex-row sm:items-center gap-4 hover:shadow-md transition-all group border border-outline-variant/5"
              >
                <!-- Icon -->
                <div class="w-14 h-14 rounded-xl {cert.status === 'valid' ? 'bg-primary-fixed' : 'bg-surface-container-high'} flex items-center justify-center shrink-0">
                  <span class="material-symbols-outlined text-[28px] {cert.status === 'valid' ? 'text-primary' : 'text-outline'}">
                    {cert.status === 'valid' ? 'workspace_premium' : 'description'}
                  </span>
                </div>

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <h3 class="font-display font-bold text-on-surface group-hover:text-primary transition-colors">{cert.event_title}</h3>
                  <div class="flex flex-wrap gap-x-3 gap-y-1 mt-1">
                    {#if cert.org_name}
                      <span class="text-xs text-on-surface-variant flex items-center gap-1">
                        <span class="material-symbols-outlined text-[14px]">corporate_fare</span>
                        {cert.org_name}
                      </span>
                    {/if}
                    <span class="text-xs text-on-surface-variant flex items-center gap-1">
                      <span class="material-symbols-outlined text-[14px]">calendar_today</span>
                      {formatDate(cert.created_at)}
                    </span>
                  </div>
                  {#if cert.name}
                    <p class="text-xs text-outline mt-1">{cert.name}</p>
                  {/if}
                </div>

                <!-- Status -->
                <div class="flex items-center gap-3 shrink-0">
                  {#if cert.status === 'valid'}
                    <span class="inline-flex items-center px-3 py-1.5 rounded-full bg-emerald-50 text-emerald-700 text-[10px] font-bold uppercase tracking-wider border border-emerald-200/50">
                      <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-2"></span>
                      {$t("student.filter.valid")}
                    </span>
                  {:else}
                    <span class="inline-flex items-center px-3 py-1.5 rounded-full bg-error-container text-on-error-container text-[10px] font-bold uppercase tracking-wider">
                      {$t("student.filter.revoked")}
                    </span>
                  {/if}
                  <span class="material-symbols-outlined text-outline-variant group-hover:text-primary transition-colors">chevron_right</span>
                </div>
              </a>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Verified Badge -->
      <div class="mt-8 bg-surface-container-low rounded-2xl p-6 flex items-center gap-4 border border-outline-variant/10">
        <div class="w-12 h-12 rounded-xl bg-primary-fixed flex items-center justify-center shrink-0">
          <span class="material-symbols-outlined text-primary text-[28px]">verified_user</span>
        </div>
        <div>
          <p class="font-display font-bold text-on-surface text-sm">{$t("student.identityVerified")}</p>
          <p class="text-xs text-on-surface-variant mt-0.5">{$t("student.verifiedDesc")}</p>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="border-t border-outline-variant/10 py-8 mt-8">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 flex flex-col sm:flex-row items-center justify-between gap-3">
        <p class="text-xs text-on-surface-variant">&copy; {new Date().getFullYear()} Jetistik</p>
        <a href="/" class="text-xs text-primary hover:underline">jetistik.kz</a>
      </div>
    </footer>
  {/if}
</div>
