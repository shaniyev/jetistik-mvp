<script lang="ts">
  import { onMount } from "svelte";
  import { api, ApiError, type PaginatedResponse } from "$lib/api/client";
  import { t } from "$lib/i18n";

  interface Event {
    id: number;
    title: string;
    date: string;
    city: string;
    status: string;
    created_at: string;
  }

  let events = $state<Event[]>([]);
  let loading = $state(true);
  let page = $state(1);
  let total = $state(0);
  const perPage = 20;
  let activeFilter = $state("all");

  const filterKeys = [
    { key: "all", labelKey: "staff.events.all" as const },
    { key: "active", labelKey: "staff.events.active" as const },
    { key: "completed", labelKey: "staff.events.completed" as const },
    { key: "draft", labelKey: "staff.events.draft" as const },
  ];

  async function loadEvents() {
    loading = true;
    try {
      const res = await api.get<Event[]>(`/api/v1/staff/events?page=${page}&per_page=${perPage}`) as PaginatedResponse<Event>;
      events = res.data;
      total = res.pagination.total;
    } catch (e) {
      console.error("Failed to load events", e);
    } finally {
      loading = false;
    }
  }

  async function deleteEvent(id: number) {
    if (!confirm($t("staff.events.deleteConfirm"))) return;
    try {
      await api.delete(`/api/v1/staff/events/${id}`);
      loadEvents();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : $t("staff.events.deleteFailed"));
    }
  }

  onMount(loadEvents);

  let filteredEvents = $derived(
    activeFilter === "all"
      ? events
      : events.filter((e) => {
          if (activeFilter === "active") return e.status === "active";
          if (activeFilter === "completed") return e.status === "completed" || e.status === "done";
          if (activeFilter === "draft") return e.status === "draft" || e.status === "inactive";
          return true;
        })
  );

  function statusBadgeClasses(status: string): string {
    switch (status) {
      case "active":
        return "bg-primary-fixed text-on-primary-fixed";
      case "completed":
      case "done":
        return "bg-surface-container-highest text-on-surface-variant";
      case "draft":
      case "inactive":
        return "bg-error-container text-on-error-container";
      default:
        return "bg-surface-container-highest text-on-surface-variant";
    }
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return "—";
    try {
      const d = new Date(dateStr);
      return d.toLocaleDateString("ru-RU", { day: "2-digit", month: "short", year: "numeric" });
    } catch {
      return dateStr;
    }
  }
</script>

<!-- Header / Top Bar -->
<header class="h-20 flex items-center justify-between px-10 bg-surface/80 backdrop-blur-xl sticky top-0 z-30">
  <div>
    <h1 class="text-2xl font-display font-extrabold tracking-tight text-on-surface">{$t("staff.events.title")}</h1>
    <p class="text-sm text-on-surface-variant font-medium">{$t("staff.events.subtitle")}</p>
  </div>
  <div class="flex items-center gap-6">
    <a
      href="/staff/events/create"
      class="bg-gradient-to-br from-primary to-primary-container text-white px-6 py-2.5 rounded-xl font-display font-bold text-sm flex items-center gap-2 shadow-lg shadow-primary/20 active:scale-95 transition-transform"
    >
      <span class="material-symbols-outlined text-lg">add</span>
      {$t("staff.events.create")}
    </a>
  </div>
</header>

<!-- Page Content -->
<div class="px-10 py-8 space-y-8">
  <!-- Status Filter Pill Bar -->
  <div class="flex items-center gap-2 bg-surface-container-low p-1.5 rounded-2xl w-fit">
    {#each filterKeys as filter}
      <button
        onclick={() => { activeFilter = filter.key; }}
        class="px-6 py-2 text-sm rounded-xl font-semibold transition-colors
          {activeFilter === filter.key
            ? 'bg-white text-primary font-bold shadow-sm'
            : 'text-on-surface-variant hover:text-on-surface'}"
      >
        {$t(filter.labelKey)}
      </button>
    {/each}
  </div>

  <!-- Events Grid -->
  {#if loading}
    <div class="text-center py-12 text-on-surface-variant">{$t("staff.events.loading")}</div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
      {#each filteredEvents as event (event.id)}
        <a
          href="/staff/events/{event.id}"
          class="bg-surface-container-lowest rounded-2xl p-6 shadow-sm border border-outline-variant/10 hover:shadow-xl hover:shadow-primary/5 transition-all group flex flex-col h-full"
        >
          <div class="flex justify-between items-start mb-4">
            <span class="px-3 py-1 rounded-full text-[10px] uppercase font-bold tracking-widest {statusBadgeClasses(event.status)}">
              {event.status}
            </span>
            <button
              onclick={(e) => { e.preventDefault(); e.stopPropagation(); deleteEvent(event.id); }}
              class="text-outline hover:text-error transition-colors"
            >
              <span class="material-symbols-outlined">delete</span>
            </button>
          </div>
          <h3 class="text-xl font-display font-bold text-on-surface mb-2 group-hover:text-primary transition-colors leading-tight">
            {event.title}
          </h3>
          <div class="flex items-center gap-2 text-on-surface-variant text-sm mb-6">
            <span class="material-symbols-outlined text-[18px]">location_on</span>
            <span>{event.city || "—"}</span>
          </div>
          <div class="mt-auto pt-6 border-t border-surface-container-low grid grid-cols-2 gap-4">
            <div>
              <p class="text-[10px] text-outline font-bold uppercase tracking-wider mb-1">{$t("common.date")}</p>
              <p class="text-sm font-semibold text-on-surface">{formatDate(event.date)}</p>
            </div>
            <div>
              <p class="text-[10px] text-outline font-bold uppercase tracking-wider mb-1">{$t("common.created")}</p>
              <p class="text-sm font-semibold text-on-surface">{formatDate(event.created_at)}</p>
            </div>
          </div>
        </a>
      {/each}

      <!-- Add New Event Visual Placeholder -->
      <a
        href="/staff/events/create"
        class="border-2 border-dashed border-outline-variant/40 rounded-2xl p-6 flex flex-col items-center justify-center gap-4 hover:border-primary/40 hover:bg-primary/5 transition-all cursor-pointer group h-full min-h-[220px]"
      >
        <div class="w-12 h-12 rounded-full bg-surface-container flex items-center justify-center text-outline group-hover:bg-primary group-hover:text-white transition-all">
          <span class="material-symbols-outlined">add</span>
        </div>
        <p class="text-sm font-bold text-on-surface-variant group-hover:text-primary">{$t("staff.events.addNew")}</p>
      </a>
    </div>

    <!-- Recent Activity Table -->
    {#if events.length > 0}
      <div class="bg-surface-container-lowest rounded-3xl overflow-hidden shadow-sm border border-outline-variant/10 mt-12">
        <div class="p-6 border-b border-surface-container flex items-center justify-between">
          <h2 class="font-display font-bold text-on-surface">{$t("staff.events.recentActivity")}</h2>
          <a href="/staff/events" class="text-primary text-xs font-bold uppercase tracking-widest hover:underline">
            {$t("staff.events.filter.all")}
          </a>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="bg-surface-container-low/50">
                <th class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-outline">{$t("common.event")}</th>
                <th class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-outline">{$t("common.status")}</th>
                <th class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-outline">{$t("common.city")}</th>
                <th class="px-6 py-4 text-[10px] font-bold uppercase tracking-widest text-outline">{$t("common.date")}</th>
                <th class="px-6 py-4 text-right"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-surface-container/30">
              {#each events.slice(0, 5) as event (event.id)}
                <tr class="hover:bg-surface-container-low transition-colors">
                  <td class="px-6 py-4">
                    <div class="flex flex-col">
                      <span class="text-sm font-bold text-on-surface">{event.title}</span>
                      <span class="text-xs text-on-surface-variant">{event.city || ""}</span>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <span class="px-2 py-0.5 rounded text-[10px] font-bold uppercase {statusBadgeClasses(event.status)}">
                      {event.status}
                    </span>
                  </td>
                  <td class="px-6 py-4">
                    <span class="text-sm font-medium text-on-surface">{event.city || "—"}</span>
                  </td>
                  <td class="px-6 py-4 text-sm text-on-surface-variant">{formatDate(event.date)}</td>
                  <td class="px-6 py-4 text-right">
                    <a href="/staff/events/{event.id}">
                      <span class="material-symbols-outlined text-outline cursor-pointer hover:text-primary">chevron_right</span>
                    </a>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {/if}

    <!-- Pagination -->
    {#if total > perPage}
      <div class="flex items-center justify-between text-sm text-on-surface-variant">
        <span>{$t("common.showing")} {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} {$t("common.of")} {total}</span>
        <div class="flex gap-2">
          <button
            disabled={page <= 1}
            onclick={() => { page--; loadEvents(); }}
            class="px-4 py-2 rounded-xl bg-surface-container-low hover:bg-surface-container-high disabled:opacity-50 transition-colors text-sm font-medium"
          >
            {$t("common.previous")}
          </button>
          <button
            disabled={page * perPage >= total}
            onclick={() => { page++; loadEvents(); }}
            class="px-4 py-2 rounded-xl bg-surface-container-low hover:bg-surface-container-high disabled:opacity-50 transition-colors text-sm font-medium"
          >
            {$t("common.next")}
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>
