<script lang="ts">
  import { onMount } from "svelte";
  import { api } from "$lib/api/client";
  import { currentUser } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  interface Certificate {
    id: number;
    event_title: string;
    organization_name: string;
    event_date: string;
    status: string;
    code: string;
    description?: string;
  }

  let certificates = $state<Certificate[]>([]);
  let loading = $state(true);
  let linkCopied = $state(false);
  let editingIIN = $state(false);
  let newIIN = $state("");
  let savingIIN = $state(false);
  let statusFilter = $state<string>("");

  async function loadCertificates() {
    loading = true;
    try {
      const res = await api.get<Certificate[]>("/api/v1/student/certificates");
      certificates = res.data;
    } catch (e) {
      console.error("Failed to load certificates", e);
    } finally {
      loading = false;
    }
  }

  async function downloadPdf(id: number) {
    try {
      const token = (await import("$lib/api/client")).getAccessToken();
      const apiBase = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
      const url = `${apiBase}/api/v1/student/certificates/${id}/download`;
      const a = document.createElement("a");
      a.href = url;
      a.setAttribute("download", "");
      if (token) {
        const res = await fetch(url, {
          headers: { Authorization: `Bearer ${token}` },
          credentials: "include",
        });
        if (!res.ok) throw new Error("Download failed");
        const blob = await res.blob();
        a.href = URL.createObjectURL(blob);
      }
      document.body.appendChild(a);
      a.click();
      a.remove();
    } catch (e) {
      console.error("Download failed", e);
    }
  }

  function maskIin(iin: string | undefined): string {
    if (!iin || iin.length < 12) return iin ?? "---";
    return iin.slice(0, 3) + "***" + iin.slice(6, 9) + "***";
  }

  async function copyProfileLink() {
    // Share link to the platform, never expose IIN
    const url = `${window.location.origin}`;
    try {
      await navigator.clipboard.writeText(url);
      linkCopied = true;
      setTimeout(() => { linkCopied = false; }, 2000);
    } catch {
      // Fallback
    }
  }

  function startEditIIN() {
    newIIN = $currentUser?.iin ?? "";
    editingIIN = true;
  }

  async function saveIIN() {
    if (!newIIN || newIIN.length !== 12) return;
    savingIIN = true;
    try {
      await api.patch("/api/v1/profile", { iin: newIIN });
      const { auth } = await import("$lib/stores/auth");
      await auth.refresh();
      editingIIN = false;
      loadCertificates();
    } catch (e) {
      console.error("Failed to update IIN", e);
    } finally {
      savingIIN = false;
    }
  }

  let filteredCertificates = $derived(
    statusFilter
      ? certificates.filter((c) => c.status === statusFilter)
      : certificates
  );

  onMount(loadCertificates);
</script>

<!-- Header Section -->
<header class="mb-10 md:mb-16">
  <div class="flex flex-col md:flex-row md:items-end justify-between gap-6">
    <div>
      <h1 class="font-display font-extrabold text-3xl md:text-5xl tracking-tight text-on-surface mb-2">
        {$t("student.title")}
      </h1>
      <p class="font-body text-on-surface-variant text-base md:text-lg max-w-xl leading-snug">
        {$t("student.subtitle")}
      </p>
    </div>
    <div class="flex gap-3">
      <div class="flex-1 md:flex-none relative">
        <select
          bind:value={statusFilter}
          class="w-full appearance-none bg-surface-container-lowest text-on-surface border border-outline-variant/30 px-5 py-3 rounded-xl text-sm font-semibold pr-10 focus:ring-primary focus:border-primary transition-all cursor-pointer"
        >
          <option value="">{$t("common.filter")}: All</option>
          <option value="valid">Valid</option>
          <option value="revoked">Revoked</option>
          <option value="completed">Completed</option>
        </select>
        <span class="material-symbols-outlined absolute right-3 top-1/2 -translate-y-1/2 text-on-surface-variant pointer-events-none text-[20px]">filter_list</span>
      </div>
      <button
        onclick={copyProfileLink}
        class="flex-1 md:flex-none bg-gradient-to-br from-primary to-primary-container text-white px-6 py-3 rounded-xl flex items-center justify-center gap-2 hover:shadow-lg hover:shadow-primary/20 transition-all active:scale-95"
      >
        <span class="material-symbols-outlined text-[20px]">share</span>
        <span class="text-sm font-semibold">{$t("student.shareProfile")}</span>
      </button>
    </div>
  </div>
</header>

<!-- Profile Bento Grid -->
<div class="grid grid-cols-1 lg:grid-cols-12 gap-6 mb-12">
  <!-- Profile Info Card -->
  <div class="lg:col-span-4 bg-surface-container-lowest p-6 md:p-8 rounded-[2rem] shadow-sm border border-outline-variant/10 h-fit">
    <div class="flex flex-col items-center text-center">
      <div class="relative mb-4">
        <div class="w-24 h-24 rounded-[2rem] overflow-hidden bg-primary-fixed shadow-inner flex items-center justify-center text-primary font-bold text-3xl">
          {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
        </div>
        <button class="absolute -bottom-2 -right-2 bg-white p-2 rounded-full shadow-md hover:bg-slate-50 transition-colors border border-slate-100">
          <span class="material-symbols-outlined text-[18px] text-primary">edit</span>
        </button>
      </div>
      <h2 class="font-display font-bold text-xl text-on-surface mb-1">{$currentUser?.username ?? ""}</h2>
      <p class="text-on-surface-variant text-sm mb-6 uppercase tracking-wider font-semibold">{$t("student.role")} / Student</p>

      <div class="w-full space-y-4 text-left">
        <!-- IIN Section -->
        <div class="group">
          <div class="flex justify-between items-center mb-1 px-1">
            <label class="text-[11px] font-bold text-outline uppercase tracking-widest">{$t("student.iin")} / IIN</label>
            {#if !editingIIN}
              <button onclick={startEditIIN} class="text-primary text-[11px] font-bold uppercase tracking-widest hover:underline">{$t("common.edit")}</button>
            {/if}
          </div>

          {#if editingIIN}
            <div class="space-y-3">
              <input
                type="text"
                bind:value={newIIN}
                maxlength={12}
                placeholder="123456789012"
                class="w-full bg-surface p-4 rounded-xl font-mono text-on-surface border border-outline-variant/20 shadow-inner focus:outline-none focus:border-primary transition-colors"
              />
              <div class="flex items-center gap-2">
                <button
                  onclick={saveIIN}
                  disabled={savingIIN || newIIN.length !== 12}
                  class="px-4 py-2 rounded-lg text-xs font-bold bg-primary text-on-primary hover:bg-primary/90 disabled:opacity-50 transition-colors"
                >
                  {savingIIN ? "..." : $t("common.save")}
                </button>
                <button
                  onclick={() => { editingIIN = false; }}
                  class="px-4 py-2 rounded-lg text-xs font-bold text-on-surface-variant hover:bg-surface-container-low transition-colors"
                >
                  {$t("common.cancel")}
                </button>
              </div>
            </div>
          {:else}
            <div class="bg-surface p-4 rounded-xl font-mono text-on-surface flex justify-between items-center border border-outline-variant/20 shadow-inner">
              <span class="tracking-widest">{$currentUser?.iin ?? "---"}</span>
              <button class="text-outline hover:text-primary transition-colors">
                <span class="material-symbols-outlined text-[20px]">content_copy</span>
              </button>
            </div>
          {/if}
        </div>

        <!-- Verified Badge -->
        <div class="bg-blue-50/50 p-4 rounded-xl border border-blue-100/50">
          <div class="flex gap-3 items-start">
            <span class="material-symbols-outlined text-blue-600 shrink-0">verified_user</span>
            <div class="text-[12px] leading-tight text-blue-900 font-medium">
              {$t("student.identityVerified")}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Certificates List Section -->
  <div class="lg:col-span-8">
    <!-- Data Row Header (Hidden on Mobile) -->
    <div class="hidden md:grid grid-cols-12 gap-4 px-6 mb-4 text-[11px] font-bold text-outline uppercase tracking-[0.15em]">
      <div class="col-span-5">{$t("student.certificateDetails")}</div>
      <div class="col-span-3">Status</div>
      <div class="col-span-4 text-right">Actions</div>
    </div>

    <!-- Certificate Cards Container -->
    <div class="grid grid-cols-1 gap-4">
      {#if loading}
        <div class="bg-surface-container-lowest p-12 rounded-[1.5rem] shadow-sm text-center text-on-surface-variant">
          {$t("common.loading")}
        </div>
      {:else if filteredCertificates.length === 0}
        <!-- Empty State -->
        <div class="bg-surface-container-low border-2 border-dashed border-outline-variant/30 rounded-[2rem] p-16 text-center">
          <div class="w-20 h-20 bg-surface-container-lowest rounded-full flex items-center justify-center mx-auto mb-6 shadow-sm">
            <span class="material-symbols-outlined text-4xl text-outline">search_off</span>
          </div>
          <h3 class="font-display font-bold text-xl text-on-surface mb-2">{$t("student.noCertificates")}</h3>
          <button onclick={loadCertificates} class="mt-4 bg-primary text-white px-6 py-2 rounded-xl text-sm font-bold shadow-lg shadow-primary/20 hover:scale-105 active:scale-95 transition-all">
            Refresh Data
          </button>
        </div>
      {:else}
        {#each filteredCertificates as cert}
          <div class="bg-surface-container-lowest p-5 md:p-6 rounded-[1.5rem] shadow-sm hover:shadow-md transition-shadow group flex flex-col md:grid md:grid-cols-12 md:items-center gap-5">
            <!-- Certificate Details -->
            <div class="md:col-span-5 flex gap-4 items-center">
              <div class="w-14 h-14 md:w-12 md:h-12 {cert.status === 'revoked' ? 'bg-surface-container' : 'bg-primary-fixed'} rounded-xl flex items-center justify-center shrink-0">
                <span class="material-symbols-outlined {cert.status === 'revoked' ? 'text-outline' : 'text-primary'} text-[28px]">
                  {cert.status === 'revoked' ? 'description' : 'workspace_premium'}
                </span>
              </div>
              <div>
                <h3 class="font-display font-bold text-on-surface leading-tight text-lg md:text-base">{cert.event_title}</h3>
                <div class="flex flex-col mt-0.5">
                  <span class="text-xs text-on-surface-variant font-medium">{cert.organization_name} / {cert.event_date ?? ""}</span>
                  {#if cert.description}
                    <span class="text-[10px] text-outline mt-0.5">{cert.description}</span>
                  {/if}
                </div>
              </div>
            </div>

            <!-- Status -->
            <div class="md:col-span-3">
              {#if cert.status === "valid" || cert.status === "completed"}
                <span class="inline-flex items-center px-3 py-1.5 rounded-full bg-primary-fixed text-blue-800 text-[10px] font-bold uppercase tracking-wider border border-blue-200">
                  <span class="w-1.5 h-1.5 rounded-full bg-blue-600 mr-2 animate-pulse"></span>
                  VALID
                </span>
              {:else if cert.status === "revoked"}
                <span class="inline-flex items-center px-3 py-1.5 rounded-full bg-error-container text-on-error-container text-[10px] font-bold uppercase tracking-wider border border-red-200/50">
                  <span class="w-1.5 h-1.5 rounded-full bg-error mr-2"></span>
                  REVOKED
                </span>
              {:else}
                <span class="inline-flex items-center px-3 py-1.5 rounded-full bg-surface-container-high text-on-surface-variant text-[10px] font-bold uppercase tracking-wider">
                  <span class="w-1.5 h-1.5 rounded-full bg-outline mr-2"></span>
                  {cert.status.toUpperCase()}
                </span>
              {/if}
            </div>

            <!-- Actions -->
            <div class="md:col-span-4 flex gap-3">
              {#if cert.status === "revoked"}
                <button class="w-full md:w-auto text-outline border border-outline-variant/30 px-4 py-3 md:py-2 rounded-xl text-sm font-semibold cursor-not-allowed opacity-50 flex items-center justify-center gap-2 bg-surface">
                  <span class="material-symbols-outlined text-[20px] md:text-[18px]">lock</span>
                  Unavailable
                </button>
              {:else}
                <a
                  href="/verify/{cert.code}"
                  class="flex-1 md:flex-none text-on-surface bg-surface-container-low px-4 py-3 md:py-2 rounded-xl text-sm font-semibold hover:bg-surface-container-high transition-colors flex items-center justify-center gap-2"
                >
                  <span class="material-symbols-outlined text-[20px] md:text-[18px]">visibility</span>
                  {$t("student.viewCert")}
                </a>
                <button
                  onclick={() => downloadPdf(cert.id)}
                  class="flex-1 md:flex-none bg-primary text-white px-4 py-3 md:py-2 rounded-xl text-sm font-semibold hover:bg-primary-container transition-colors flex items-center justify-center gap-2 shadow-sm"
                >
                  <span class="material-symbols-outlined text-[20px] md:text-[18px]">download</span>
                  {$t("student.downloadPdf")}
                </button>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </div>
</div>

<!-- Featured QR Card -->
<div class="bg-[#2d3133] text-[#eff1f3] rounded-[2rem] md:rounded-[2.5rem] p-8 md:p-12 overflow-hidden relative shadow-2xl">
  <div class="relative z-10 flex flex-col md:flex-row gap-10 items-center">
    <div class="flex-1 text-center md:text-left">
      <h3 class="font-display text-3xl md:text-4xl font-bold mb-4 tracking-tight">{$t("student.portfolioQr")}</h3>
      <p class="text-blue-100/70 text-base md:text-lg mb-8 max-w-md leading-relaxed">
        {$t("student.portfolioQrDesc")}
      </p>
      <div class="flex flex-col sm:flex-row gap-4 justify-center md:justify-start">
        <button
          onclick={copyProfileLink}
          class="bg-white text-on-surface px-8 py-3.5 rounded-xl font-bold text-sm shadow-sm hover:bg-slate-50 transition-colors active:scale-95"
        >
          {linkCopied ? $t("common.copied") : $t("student.copyLink")}
        </button>
        <a href="/student" class="bg-blue-600/30 border border-blue-400/30 text-white px-8 py-3.5 rounded-xl font-bold text-sm hover:bg-blue-600/50 transition-colors active:scale-95 inline-block">
          {$t("staff.nav.settings")}
        </a>
      </div>
    </div>
    <div class="shrink-0 w-56 h-56 md:w-64 md:h-64 bg-white p-5 rounded-[2rem] shadow-2xl rotate-3 hover:rotate-0 transition-transform duration-500">
      <div class="w-full h-full bg-surface-container-low rounded-xl flex items-center justify-center text-on-surface-variant text-sm font-mono">
        QR Code
      </div>
    </div>
  </div>
  <!-- Decorative Background Elements -->
  <div class="absolute -right-20 -top-20 w-80 h-80 bg-primary/20 rounded-full blur-[100px] pointer-events-none"></div>
  <div class="absolute -left-20 -bottom-20 w-80 h-80 bg-blue-500/10 rounded-full blur-[100px] pointer-events-none"></div>
</div>
