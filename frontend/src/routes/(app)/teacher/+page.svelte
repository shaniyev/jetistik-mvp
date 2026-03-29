<script lang="ts">
  import { onMount } from "svelte";
  import { api } from "$lib/api/client";
  import { currentUser } from "$lib/stores/auth";
  import { t } from "$lib/i18n";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";

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
  let showAddForm = $state(false);
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

  let filteredCertificates = $derived(
    filterStudent
      ? certificates.filter((c) => c.student_iin === filterStudent)
      : certificates
  );

  onMount(() => {
    loadStudents();
    loadCertificates();
  });

  const certColumns = [
    { key: "student", label: "" },
    { key: "event", label: "" },
    { key: "organization", label: "" },
    { key: "date", label: "" },
    { key: "status", label: "" },
  ];
</script>

<div class="space-y-8 max-w-5xl">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
    <div>
      <h1 class="font-display text-2xl sm:text-3xl font-bold text-on-surface">{$t("teacher.title")}</h1>
      <div class="flex items-center gap-3 mt-2">
        <span class="text-xs font-medium text-primary bg-primary/10 px-2 py-0.5 rounded-md uppercase tracking-wide">{$t("teacher.role")}</span>
        <span class="text-xs text-on-surface-variant">{$currentUser?.username ?? ""}</span>
      </div>
    </div>
    <div class="flex items-center gap-3 text-xs text-on-surface-variant">
      <span>{$t("teacher.systemStatus")}</span>
      <span class="flex items-center gap-1.5 text-emerald-600">
        <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
        {$t("teacher.nodesSynchronized")}
      </span>
    </div>
  </div>

  <!-- My Students Section -->
  <section class="space-y-4">
    <h2 class="font-display text-lg font-bold text-on-surface">{$t("teacher.myStudents")}</h2>

    <!-- Add Student Toggle -->
    <button
      onclick={() => { showAddForm = !showAddForm; addError = ""; addSuccess = ""; }}
      class="flex items-center gap-2 text-sm text-primary hover:text-primary-container transition-colors font-medium"
    >
      <svg class="w-4 h-4 transition-transform {showAddForm ? 'rotate-45' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
      </svg>
      {$t("teacher.addStudent")}
    </button>

    <!-- Add Student Form -->
    {#if showAddForm}
      <div class="bg-surface-lowest rounded-lg p-5 space-y-4 max-w-md">
        <div>
          <label for="iin-input" class="block text-xs font-medium text-on-surface-variant mb-1.5 uppercase tracking-wide">
            {$t("teacher.addStudentPlaceholder")}
          </label>
          <input
            id="iin-input"
            type="text"
            value={iinInput}
            oninput={handleIinInput}
            placeholder="0000  0000  0000"
            class="w-full px-3 py-2.5 bg-surface-lowest text-on-surface text-lg font-mono tracking-widest
                   border-b-2 border-outline-variant/30 focus:border-primary outline-none transition-colors"
          />
        </div>

        {#if addError}
          <div class="flex items-start gap-2 bg-error-container/30 rounded-md px-3 py-2">
            <svg class="w-4 h-4 text-error mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
            </svg>
            <div>
              <p class="text-sm font-medium text-error">{addError}</p>
              <p class="text-xs text-on-surface-variant">{$t("teacher.invalidIinDesc")}</p>
            </div>
          </div>
        {/if}

        {#if addSuccess}
          <div class="flex items-center gap-2 bg-emerald-50 rounded-md px-3 py-2">
            <svg class="w-4 h-4 text-emerald-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
            </svg>
            <p class="text-sm text-emerald-700">{addSuccess}</p>
          </div>
        {/if}

        <button
          onclick={addStudent}
          disabled={adding}
          class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium
                 bg-gradient-to-br from-primary to-primary-container text-on-primary
                 hover:shadow-lg transition-shadow disabled:opacity-50"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          {adding ? $t("common.loading") : $t("teacher.addStudentBtn")}
        </button>
      </div>
    {/if}

    <!-- Active Students List -->
    {#if loadingStudents}
      <p class="text-sm text-on-surface-variant">{$t("common.loading")}</p>
    {:else if students.length === 0}
      <p class="text-sm text-on-surface-variant">{$t("teacher.noStudents")}</p>
    {:else}
      <div class="space-y-2">
        <p class="text-xs text-on-surface-variant uppercase tracking-wide font-medium">{$t("teacher.activeIdentity")}</p>
        {#each students as student}
          <div class="flex items-center justify-between bg-surface-lowest rounded-lg px-4 py-3">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary text-xs font-bold">
                {student.username?.[0]?.toUpperCase() ?? "?"}
              </div>
              <div>
                <span class="text-sm font-mono text-on-surface">{maskIin(student.iin)}</span>
                {#if student.username}
                  <p class="text-xs text-on-surface-variant">{student.username}</p>
                {/if}
              </div>
              {#if student.status === "verified" || student.status === "active"}
                <span class="text-xs text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full uppercase tracking-wide font-medium">
                  {$t("teacher.verifiedIdentity")}
                </span>
              {/if}
            </div>
            <button
              onclick={() => removeStudent(student.iin)}
              aria-label={$t("teacher.removeStudent")}
              class="text-xs text-on-surface-variant hover:text-error transition-colors"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <!-- Student Certificates Section -->
  <section class="space-y-4">
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <div>
        <h2 class="font-display text-lg font-bold text-on-surface">{$t("teacher.certificates")}</h2>
        <p class="text-xs text-on-surface-variant mt-0.5">{$t("teacher.certificatesDesc")}</p>
      </div>
      <div class="flex items-center gap-2">
        <select
          bind:value={filterStudent}
          class="text-sm bg-surface-lowest text-on-surface rounded-lg px-3 py-2 border-b-2 border-outline-variant/30 focus:border-primary outline-none"
        >
          <option value="">{$t("teacher.filterByStudent")}</option>
          {#each students as student}
            <option value={student.iin}>{student.username ?? maskIin(student.iin)}</option>
          {/each}
        </select>
      </div>
    </div>

    <!-- Desktop table -->
    <div class="hidden sm:block">
      <DataTable columns={certColumns} data={filteredCertificates} loading={loadingCerts} empty={$t("teacher.noCertificates")}>
        {#snippet row(cert: Certificate)}
          <tr class="hover:bg-surface-low/50 transition-colors">
            <td class="px-4 py-3">
              <div>
                <p class="text-sm font-medium text-on-surface">{cert.student_name}</p>
                <p class="text-xs text-on-surface-variant font-mono">{maskIin(cert.student_iin)}</p>
              </div>
            </td>
            <td class="px-4 py-3 text-sm text-on-surface">{cert.event_title}</td>
            <td class="px-4 py-3 text-sm text-on-surface-variant">{cert.organization_name}</td>
            <td class="px-4 py-3 text-sm text-on-surface-variant">{cert.event_date ?? "---"}</td>
            <td class="px-4 py-3">
              <StatusBadge status={cert.status} />
            </td>
          </tr>
        {/snippet}
      </DataTable>
    </div>

    <!-- Mobile cards -->
    <div class="sm:hidden space-y-3">
      {#if loadingCerts}
        <p class="text-sm text-on-surface-variant text-center py-8">{$t("common.loading")}</p>
      {:else if filteredCertificates.length === 0}
        <p class="text-sm text-on-surface-variant text-center py-8">{$t("teacher.noCertificates")}</p>
      {:else}
        {#each filteredCertificates as cert}
          <div class="bg-surface-lowest rounded-lg p-4 space-y-2">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-on-surface">{cert.student_name}</p>
                <p class="text-xs text-on-surface-variant font-mono">{maskIin(cert.student_iin)}</p>
              </div>
              <StatusBadge status={cert.status} />
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs">
              <div>
                <p class="text-on-surface-variant uppercase tracking-wide">{$t("teacher.col.event")}</p>
                <p class="text-on-surface mt-0.5">{cert.event_title}</p>
              </div>
              <div>
                <p class="text-on-surface-variant uppercase tracking-wide">{$t("teacher.col.date")}</p>
                <p class="text-on-surface mt-0.5">{cert.event_date ?? "---"}</p>
              </div>
            </div>
            <div class="text-xs">
              <p class="text-on-surface-variant uppercase tracking-wide">{$t("teacher.col.organization")}</p>
              <p class="text-on-surface mt-0.5">{cert.organization_name}</p>
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </section>
</div>
