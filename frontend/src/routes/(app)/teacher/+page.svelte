<script lang="ts">
  import { onMount } from "svelte";
  import { api } from "$lib/api/client";
  import { currentUser } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  interface Student {
    iin: string;
    username?: string;
    status?: string;
  }

  interface Certificate {
    id: number;
    student_name: string;
    student_iin: string;
    event_title: string;
    organization_name: string;
    event_date: string;
    status: string;
  }

  let students = $state<Student[]>([]);
  let certificates = $state<Certificate[]>([]);
  let loadingStudents = $state(true);
  let loadingCerts = $state(true);

  // Add student form
  let iinInput = $state("");
  let addError = $state("");
  let addSuccess = $state("");
  let adding = $state(false);

  // Filter
  let filterStudent = $state("");

  async function loadStudents() {
    loadingStudents = true;
    try {
      const res = await api.get<Student[]>("/api/v1/teacher/students");
      students = res.data ?? [];
    } catch (e) {
      console.error("Failed to load students", e);
    } finally {
      loadingStudents = false;
    }
  }

  async function loadCertificates() {
    loadingCerts = true;
    try {
      const res = await api.get<Certificate[]>("/api/v1/teacher/certificates");
      certificates = res.data ?? [];
    } catch (e) {
      console.error("Failed to load certificates", e);
    } finally {
      loadingCerts = false;
    }
  }

  async function addStudent() {
    const cleanIin = iinInput.replace(/\s/g, "");
    if (!/^\d{12}$/.test(cleanIin)) {
      addError = $t("teacher.invalidIin");
      return;
    }

    adding = true;
    addError = "";
    addSuccess = "";
    try {
      await api.post("/api/v1/teacher/students", { iin: cleanIin });
      addSuccess = $t("teacher.studentAdded");
      iinInput = "";
      await loadStudents();
      await loadCertificates();
    } catch (e: unknown) {
      const err = e as { status?: number; message?: string };
      if (err.status === 409) {
        addError = $t("teacher.studentAlreadyAdded");
      } else {
        addError = err.message ?? "Failed to add student";
      }
    } finally {
      adding = false;
    }
  }

  async function removeStudent(iin: string) {
    try {
      await api.delete(`/api/v1/teacher/students/${iin}`);
      await loadStudents();
      await loadCertificates();
    } catch (e) {
      console.error("Failed to remove student", e);
    }
  }

  function maskIin(iin: string): string {
    if (iin.length < 12) return iin;
    return iin.slice(0, 4) + "****" + iin.slice(8);
  }

  function formatIinInput(value: string): string {
    const digits = value.replace(/\D/g, "").slice(0, 12);
    const groups = [];
    for (let i = 0; i < digits.length; i += 4) {
      groups.push(digits.slice(i, i + 4));
    }
    return groups.join("  ");
  }

  function handleIinInput(e: Event) {
    const target = e.target as HTMLInputElement;
    iinInput = formatIinInput(target.value);
  }

  function getIinPrefix(iin: string): string {
    return iin.slice(0, 2);
  }

  let filteredCertificates = $derived(
    filterStudent
      ? certificates.filter((c) => c.student_iin === filterStudent)
      : certificates
  );

  async function downloadPdf(certId: number) {
    try {
      const token = (await import("$lib/api/client")).getAccessToken();
      const apiBase = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
      const url = `${apiBase}/api/v1/teacher/certificates/${certId}/download`;
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

  onMount(() => {
    loadStudents();
    loadCertificates();
  });
</script>

<div class="space-y-8 md:space-y-12">
  <!-- Hero / Greeting -->
  <section class="flex flex-col md:flex-row justify-between items-start md:items-end gap-4">
    <div>
      <h2 class="text-2xl md:text-4xl font-extrabold font-display tracking-tight text-on-surface mb-2">{$t("teacher.title")}</h2>
      <div class="flex flex-wrap items-center gap-2">
        <span class="px-2 py-0.5 bg-primary-fixed text-on-primary-fixed-variant text-[10px] font-bold rounded uppercase tracking-wider">{$t("teacher.role")}</span>
        <span class="text-on-surface-variant text-xs md:text-sm font-medium">Jetistik Central Node</span>
      </div>
    </div>
    <div class="flex flex-col items-start md:items-end w-full md:w-auto pt-2 md:pt-0 border-t md:border-t-0 border-slate-100 mt-2 md:mt-0">
      <p class="text-[10px] md:text-xs text-on-surface-variant font-medium uppercase tracking-tight">{$t("teacher.systemStatus")}</p>
      <p class="text-xs md:text-sm font-bold text-primary flex items-center gap-2 mt-1">
        <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
        {$t("teacher.nodesSynchronized")}
      </p>
    </div>
  </section>

  <!-- Section 1: My Students -->
  <section id="students">
    <div class="mb-6">
      <h3 class="text-lg md:text-xl font-bold font-display text-on-surface">{$t("teacher.myStudents")}</h3>
      <p class="text-[10px] md:text-xs text-on-surface-variant mt-1">{$t("teacher.myStudentsDesc")}</p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
      <!-- Add Student Form -->
      <div class="lg:col-span-4">
        <div class="bg-surface-container-lowest p-6 md:p-8 rounded-xl shadow-sm border border-outline-variant/10">
          <h4 class="text-sm font-bold text-on-surface mb-6 flex items-center gap-2">
            <span class="material-symbols-outlined text-primary text-lg">person_add</span>
            {$t("teacher.addStudent")}
          </h4>
          <div class="space-y-4">
            <div>
              <label class="block text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-2">
                {$t("teacher.addStudentPlaceholder")}
              </label>
              <div class="relative">
                <input
                  type="text"
                  value={iinInput}
                  oninput={handleIinInput}
                  placeholder="0000  0000  0000"
                  class="w-full bg-surface-container-low border-b-2 border-outline-variant focus:border-primary border-t-0 border-l-0 border-r-0 px-4 py-3 text-sm focus:ring-0 transition-all font-mono tracking-widest"
                />
              </div>
            </div>
            <button
              onclick={addStudent}
              disabled={adding}
              class="w-full bg-gradient-to-br from-primary to-primary-container text-white font-bold py-3 px-6 rounded-md shadow-lg shadow-primary/20 hover:scale-[1.02] active:scale-95 transition-all text-sm flex justify-center items-center gap-2 disabled:opacity-50"
            >
              <span class="material-symbols-outlined text-sm">add</span>
              {adding ? $t("common.loading") : $t("teacher.addStudentBtn")}
            </button>

            <!-- Error Message -->
            {#if addError}
              <div class="p-3 bg-error-container/50 rounded-lg flex items-start gap-3">
                <span class="material-symbols-outlined text-error text-sm mt-0.5">error</span>
                <div>
                  <p class="text-[11px] font-bold text-on-error-container">{addError}</p>
                  <p class="text-[10px] text-on-error-container/80 mt-1">{$t("teacher.invalidIinDesc")}</p>
                </div>
              </div>
            {/if}

            <!-- Success Message -->
            {#if addSuccess}
              <div class="p-3 bg-secondary-container/30 rounded-lg flex items-start gap-3">
                <span class="material-symbols-outlined text-on-secondary-container text-sm mt-0.5">check_circle</span>
                <div>
                  <p class="text-[11px] font-bold text-on-secondary-container">{addSuccess}</p>
                </div>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- Students List -->
      <div class="lg:col-span-8">
        <div class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 overflow-hidden">
          {#if loadingStudents}
            <div class="p-8 text-center text-on-surface-variant text-sm">{$t("common.loading")}</div>
          {:else if students.length === 0}
            <div class="p-8 text-center text-on-surface-variant text-sm">{$t("teacher.noStudents")}</div>
          {:else}
            <!-- Desktop Table -->
            <div class="hidden lg:block">
              <table class="w-full text-left border-collapse">
                <thead class="bg-surface-container-low border-b border-outline-variant/10">
                  <tr>
                    <th class="px-6 py-4 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">IIN</th>
                    <th class="px-6 py-4 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest text-right">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-surface-container-low">
                  {#each students as student}
                    <tr class="hover:bg-surface-container-low/30 transition-colors">
                      <td class="px-6 py-4">
                        <div class="flex items-center gap-3">
                          <div class="w-8 h-8 rounded-full bg-primary-fixed flex items-center justify-center text-primary font-bold text-xs">
                            {getIinPrefix(student.iin)}
                          </div>
                          <div>
                            <span class="font-mono text-sm tracking-widest text-on-surface">{maskIin(student.iin)}</span>
                            {#if student.username}
                              <p class="text-[10px] text-on-surface-variant mt-0.5">{student.username}</p>
                            {/if}
                          </div>
                        </div>
                      </td>
                      <td class="px-6 py-4 text-right">
                        <button
                          onclick={() => removeStudent(student.iin)}
                          class="p-2 text-on-surface-variant hover:text-error hover:bg-error-container/20 rounded-lg transition-all"
                        >
                          <span class="material-symbols-outlined text-lg">delete</span>
                        </button>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>

            <!-- Mobile List -->
            <div class="lg:hidden">
              <div class="p-4 bg-surface-container-low border-b border-outline-variant/10">
                <span class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("teacher.activeIdentity")}</span>
              </div>
              <div class="divide-y divide-surface-container-low">
                {#each students as student}
                  <div class="p-4 flex justify-between items-center hover:bg-surface-container-low/30 transition-colors">
                    <div class="flex items-center gap-3">
                      <div class="w-8 h-8 rounded-full bg-primary-fixed flex items-center justify-center text-primary font-bold text-xs shrink-0">
                        {getIinPrefix(student.iin)}
                      </div>
                      <div>
                        <p class="font-mono text-sm tracking-widest text-on-surface">{maskIin(student.iin)}</p>
                        {#if student.username}
                          <p class="text-[10px] text-on-surface-variant mt-0.5">{student.username}</p>
                        {:else}
                          <p class="text-[10px] text-on-surface-variant uppercase mt-0.5">{$t("teacher.verifiedIdentity")}</p>
                        {/if}
                      </div>
                    </div>
                    <button
                      onclick={() => removeStudent(student.iin)}
                      class="p-2 text-on-surface-variant hover:text-error hover:bg-error-container/20 rounded-lg transition-all"
                    >
                      <span class="material-symbols-outlined text-lg">delete</span>
                    </button>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </section>

  <!-- Section 2: Student Certificates -->
  <section class="pb-12" id="certificates">
    <div class="flex flex-col gap-6 mb-8">
      <div>
        <h3 class="text-lg md:text-xl font-bold font-display text-on-surface">{$t("teacher.certificates")}</h3>
        <p class="text-[10px] md:text-xs text-on-surface-variant mt-1">{$t("teacher.certificatesDesc")}</p>
      </div>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-end gap-3">
        <div class="relative flex-1">
          <label class="block text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">{$t("teacher.filterByStudent")}</label>
          <select
            bind:value={filterStudent}
            class="w-full appearance-none bg-surface-container-lowest border border-outline-variant/30 rounded-lg px-4 py-2 text-sm pr-10 focus:ring-primary focus:border-primary transition-all"
          >
            <option value="">{$t("teacher.filterByStudent")}</option>
            {#each students as student}
              <option value={student.iin}>{student.username ?? maskIin(student.iin)}</option>
            {/each}
          </select>
          <span class="material-symbols-outlined absolute right-3 bottom-2 text-slate-400 pointer-events-none">expand_more</span>
        </div>
        <button class="bg-surface-container-lowest border border-outline-variant/30 px-4 py-2 rounded-lg text-sm font-bold flex items-center justify-center gap-2 hover:bg-slate-50 transition-all shadow-sm">
          <span class="material-symbols-outlined text-sm">filter_list</span>
          {$t("teacher.moreFilters")}
        </button>
      </div>
    </div>

    <!-- Certificates - Responsive Layout -->
    {#if loadingCerts}
      <div class="bg-surface-container-lowest rounded-2xl p-12 text-center text-on-surface-variant text-sm shadow-sm border border-outline-variant/10">
        {$t("common.loading")}
      </div>
    {:else if filteredCertificates.length === 0}
      <div class="bg-surface-container-lowest rounded-2xl p-12 text-center text-on-surface-variant text-sm shadow-sm border border-outline-variant/10">
        {$t("teacher.noCertificates")}
      </div>
    {:else}
      <div class="lg:bg-surface-container-lowest lg:rounded-2xl lg:shadow-sm lg:border lg:border-outline-variant/10 lg:overflow-hidden">
        <!-- Desktop Table (lg and up) -->
        <div class="hidden lg:block overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-surface-container-low/50">
                <th class="px-6 py-5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("teacher.col.student")}</th>
                <th class="px-6 py-5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("teacher.col.event")}</th>
                <th class="px-6 py-5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("teacher.col.organization")}</th>
                <th class="px-6 py-5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("teacher.col.date")}</th>
                <th class="px-6 py-5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest">{$t("teacher.col.status")}</th>
                <th class="px-6 py-5 text-[10px] font-bold text-on-surface-variant uppercase tracking-widest text-right">PDF</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-surface-container-low">
              {#each filteredCertificates as cert}
                <tr class="hover:bg-surface-container-low/30 transition-all">
                  <td class="px-6 py-5">
                    <div class="flex flex-col">
                      <span class="text-sm font-bold text-on-surface">{cert.student_name}</span>
                      <span class="text-[10px] font-mono text-on-surface-variant tracking-wider">{maskIin(cert.student_iin)}</span>
                    </div>
                  </td>
                  <td class="px-6 py-5">
                    <span class="text-sm font-medium text-on-surface">{cert.event_title}</span>
                  </td>
                  <td class="px-6 py-5">
                    <span class="text-xs text-on-surface-variant">{cert.organization_name}</span>
                  </td>
                  <td class="px-6 py-5">
                    <span class="text-xs text-on-surface-variant">{cert.event_date ?? "---"}</span>
                  </td>
                  <td class="px-6 py-5">
                    {#if cert.status === "valid" || cert.status === "completed"}
                      <span class="px-3 py-1 bg-primary-fixed text-on-primary-fixed-variant text-[10px] font-bold rounded-full uppercase tracking-wider">VALID</span>
                    {:else if cert.status === "revoked"}
                      <span class="px-3 py-1 bg-error-container text-on-error-container text-[10px] font-bold rounded-full uppercase tracking-wider">REVOKED</span>
                    {:else}
                      <span class="px-3 py-1 bg-surface-container-high text-on-surface-variant text-[10px] font-bold rounded-full uppercase tracking-wider">{cert.status.toUpperCase()}</span>
                    {/if}
                  </td>
                  <td class="px-6 py-5 text-right">
                    {#if cert.status === "valid" || cert.status === "completed"}
                      <button
                        onclick={() => downloadPdf(cert.id)}
                        class="p-2 text-primary hover:bg-primary-container/10 rounded-lg transition-all"
                        title="Download PDF"
                      >
                        <span class="material-symbols-outlined">download</span>
                      </button>
                    {:else}
                      <button class="p-2 text-slate-300 cursor-not-allowed" disabled>
                        <span class="material-symbols-outlined">download_for_offline</span>
                      </button>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <!-- Mobile Cards View (< lg) -->
        <div class="lg:hidden space-y-4">
          {#each filteredCertificates as cert}
            <div class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 p-5 space-y-4">
              <div class="flex justify-between items-start">
                <div>
                  <p class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest mb-1">{$t("teacher.col.student")}</p>
                  <p class="text-sm font-bold text-on-surface">{cert.student_name}</p>
                  <p class="text-[10px] font-mono text-on-surface-variant tracking-wider">{maskIin(cert.student_iin)}</p>
                </div>
                {#if cert.status === "valid" || cert.status === "completed"}
                  <span class="px-3 py-1 bg-primary-fixed text-on-primary-fixed-variant text-[10px] font-bold rounded-full uppercase tracking-wider">VALID</span>
                {:else if cert.status === "revoked"}
                  <span class="px-3 py-1 bg-error-container text-on-error-container text-[10px] font-bold rounded-full uppercase tracking-wider">REVOKED</span>
                {:else}
                  <span class="px-3 py-1 bg-surface-container-high text-on-surface-variant text-[10px] font-bold rounded-full uppercase tracking-wider">{cert.status.toUpperCase()}</span>
                {/if}
              </div>
              <div class="grid grid-cols-2 gap-4 pt-2 border-t border-surface-container-low">
                <div>
                  <p class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest mb-1">{$t("teacher.col.event")}</p>
                  <p class="text-xs font-medium text-on-surface">{cert.event_title}</p>
                </div>
                <div>
                  <p class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest mb-1">{$t("teacher.col.date")}</p>
                  <p class="text-xs text-on-surface-variant">{cert.event_date ?? "---"}</p>
                </div>
              </div>
              <div class="flex justify-between items-center pt-2">
                <div>
                  <p class="text-[10px] font-bold text-on-surface-variant uppercase tracking-widest mb-1">{$t("teacher.col.organization")}</p>
                  <p class="text-xs text-on-surface-variant">{cert.organization_name}</p>
                </div>
                {#if cert.status === "valid" || cert.status === "completed"}
                  <button
                    onclick={() => downloadPdf(cert.id)}
                    class="p-3 bg-primary/10 text-primary rounded-lg transition-all"
                  >
                    <span class="material-symbols-outlined">download</span>
                  </button>
                {:else}
                  <button class="p-3 bg-slate-100 text-slate-300 rounded-lg cursor-not-allowed" disabled>
                    <span class="material-symbols-outlined">download_for_offline</span>
                  </button>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </section>
</div>
