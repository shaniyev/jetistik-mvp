<script lang="ts">
  import { onMount } from "svelte";
  import { api } from "$lib/api/client";
  import { currentUser } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";

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
        // Fetch as blob with auth header
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
    const iin = $currentUser?.iin;
    const url = iin
      ? `${window.location.origin}/verify/${iin}`
      : `${window.location.origin}/verify`;
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
      // Refresh user data
      const { auth } = await import("$lib/stores/auth");
      await auth.refresh();
      editingIIN = false;
      // Reload certificates for new IIN
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

  const columns = [
    { key: "details", label: "" },
    { key: "status", label: "" },
    { key: "actions", label: "", class: "w-40" },
  ];
</script>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
    <div>
      <h1 class="font-display text-2xl sm:text-3xl font-bold text-on-surface">{$t("student.title")}</h1>
      <p class="text-sm text-on-surface-variant mt-1">{$t("student.subtitle")}</p>
    </div>
    <div class="flex items-center gap-2">
      <select
        bind:value={statusFilter}
        class="px-3 py-2 rounded-lg text-sm text-on-surface-variant bg-surface-lowest hover:bg-surface-low transition-colors border-0"
      >
        <option value="">{$t("common.filter")}: All</option>
        <option value="valid">Valid</option>
        <option value="revoked">Revoked</option>
        <option value="completed">Completed</option>
      </select>
      <button
        onclick={copyProfileLink}
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium
               bg-gradient-to-br from-primary to-primary-container text-on-primary
               hover:shadow-lg transition-shadow"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0-12.814a2.25 2.25 0 1 0 0-2.186m0 2.186a2.25 2.25 0 1 0 0 2.186" />
        </svg>
        {$t("student.shareProfile")}
      </button>
    </div>
  </div>

  <!-- Main grid: Profile + Certificates -->
  <div class="grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-6">
    <!-- Profile Card -->
    <div class="bg-surface-lowest rounded-lg p-6 space-y-4 h-fit">
      <div class="flex flex-col items-center text-center">
        <div class="w-20 h-20 rounded-full bg-surface-low flex items-center justify-center text-on-surface-variant text-2xl font-bold mb-3">
          {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
        </div>
        <h2 class="font-display font-bold text-on-surface">{$currentUser?.username ?? ""}</h2>
        <span class="text-xs text-on-surface-variant uppercase tracking-wide mt-1">{$t("student.role")}</span>
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between text-sm">
          <span class="text-on-surface-variant">{$t("student.iin")}</span>
          {#if !editingIIN}
            <button onclick={startEditIIN} class="text-xs text-primary hover:underline">{$t("common.edit")}</button>
          {/if}
        </div>
        {#if editingIIN}
          <div class="flex items-center gap-2">
            <input
              type="text"
              bind:value={newIIN}
              maxlength={12}
              placeholder="123456789012"
              class="flex-1 px-3 py-2 rounded-md bg-surface-low text-sm font-mono text-on-surface border border-primary/30 focus:outline-none focus:border-primary"
            />
          </div>
          <div class="flex items-center gap-2">
            <button
              onclick={saveIIN}
              disabled={savingIIN || newIIN.length !== 12}
              class="px-3 py-1.5 rounded-md text-xs font-medium bg-primary text-on-primary hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              {savingIIN ? "..." : $t("common.save")}
            </button>
            <button
              onclick={() => { editingIIN = false; }}
              class="px-3 py-1.5 rounded-md text-xs font-medium text-on-surface-variant hover:bg-surface-low transition-colors"
            >
              {$t("common.cancel")}
            </button>
          </div>
        {:else}
          <div class="flex items-center gap-2 bg-surface-low rounded-md px-3 py-2">
            <span class="text-sm font-mono text-on-surface">{maskIin($currentUser?.iin)}</span>
            <svg class="w-4 h-4 text-on-surface-variant ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25Z" />
            </svg>
          </div>
        {/if}
      </div>

      <div class="flex items-center gap-2 bg-emerald-50 rounded-md px-3 py-2">
        <svg class="w-4 h-4 text-emerald-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
        <span class="text-xs text-emerald-700">{$t("student.identityVerified")}</span>
      </div>
    </div>

    <!-- Certificates Table -->
    <div class="space-y-4">
      <DataTable {columns} data={filteredCertificates} {loading} empty={$t("student.noCertificates")}>
        {#snippet row(cert: Certificate)}
          <tr class="hover:bg-surface-low/50 transition-colors">
            <td class="px-4 py-4">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
                  <svg class="w-5 h-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 7.74-3.342" />
                  </svg>
                </div>
                <div>
                  <p class="font-medium text-on-surface">{cert.event_title}</p>
                  <p class="text-xs text-on-surface-variant mt-0.5">{cert.organization_name} / {cert.event_date ?? ""}</p>
                  {#if cert.description}
                    <p class="text-xs text-on-surface-variant mt-0.5">{cert.description}</p>
                  {/if}
                </div>
              </div>
            </td>
            <td class="px-4 py-4">
              <StatusBadge status={cert.status} />
            </td>
            <td class="px-4 py-4">
              <div class="flex items-center gap-2">
                <a
                  href="/verify/{cert.code}"
                  class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-md text-xs font-medium text-primary bg-primary/5 hover:bg-primary/10 transition-colors"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                  </svg>
                  {$t("student.viewCert")}
                </a>
                {#if cert.status === "valid" || cert.status === "completed"}
                  <button
                    onclick={() => downloadPdf(cert.id)}
                    class="inline-flex items-center gap-1 px-2.5 py-1.5 rounded-md text-xs font-medium text-on-primary bg-gradient-to-br from-primary to-primary-container hover:shadow-md transition-shadow"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
                    </svg>
                    {$t("student.downloadPdf")}
                  </button>
                {/if}
              </div>
            </td>
          </tr>
        {/snippet}
      </DataTable>
    </div>
  </div>

  <!-- Portfolio QR Section -->
  <div class="bg-gradient-to-br from-on-surface to-on-surface/90 rounded-xl p-6 sm:p-8 flex flex-col sm:flex-row items-center gap-6 sm:gap-12">
    <div class="flex-1 space-y-3">
      <h2 class="font-display text-xl sm:text-2xl font-bold text-white">{$t("student.portfolioQr")}</h2>
      <p class="text-sm text-white/60 max-w-md">{$t("student.portfolioQrDesc")}</p>
      <div class="flex flex-wrap gap-3 pt-2">
        <button
          onclick={copyProfileLink}
          class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium bg-white/10 text-white hover:bg-white/20 transition-colors"
        >
          {linkCopied ? $t("common.copied") : $t("student.copyLink")}
        </button>
      </div>
    </div>
    <div class="w-32 h-32 sm:w-36 sm:h-36 bg-white rounded-lg p-2 shrink-0">
      <div class="w-full h-full bg-surface-low rounded flex items-center justify-center text-on-surface-variant text-xs">
        QR
      </div>
    </div>
  </div>
</div>
