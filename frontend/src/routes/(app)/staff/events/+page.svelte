<script lang="ts">
  import { onMount } from "svelte";
  import { api, ApiError, type PaginatedResponse } from "$lib/api/client";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import DataTable from "$lib/components/DataTable.svelte";

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
    if (!confirm("Delete this event? This cannot be undone.")) return;
    try {
      await api.delete(`/api/v1/staff/events/${id}`);
      loadEvents();
    } catch (err) {
      alert(err instanceof ApiError ? err.message : "Failed to delete event");
    }
  }

  onMount(loadEvents);

  const columns = [
    { key: "title", label: "Title" },
    { key: "date", label: "Date" },
    { key: "city", label: "City" },
    { key: "status", label: "Status" },
    { key: "actions", label: "", class: "w-32" },
  ];
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="font-display text-2xl font-bold text-on-surface">Events</h1>
      <p class="text-sm text-on-surface-variant mt-1">Manage your organization's events</p>
    </div>
    <a
      href="/staff/events/create"
      class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium
             bg-gradient-to-br from-primary to-primary-container text-on-primary
             hover:shadow-lg transition-shadow"
    >
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
      </svg>
      New Event
    </a>
  </div>

  <DataTable {columns} data={events} {loading} empty="No events yet. Create your first event.">
    {#snippet row(event: Event)}
      <tr class="hover:bg-surface-low/50 transition-colors">
        <td class="px-4 py-3">
          <a href="/staff/events/{event.id}" class="font-medium text-on-surface hover:text-primary transition-colors">
            {event.title}
          </a>
        </td>
        <td class="px-4 py-3 text-on-surface-variant">
          {event.date || "—"}
        </td>
        <td class="px-4 py-3 text-on-surface-variant">
          {event.city || "—"}
        </td>
        <td class="px-4 py-3">
          <StatusBadge status={event.status} />
        </td>
        <td class="px-4 py-3">
          <div class="flex items-center gap-2">
            <a href="/staff/events/{event.id}" class="text-xs text-primary hover:underline">View</a>
            <button onclick={() => deleteEvent(event.id)} class="text-xs text-error hover:underline">Delete</button>
          </div>
        </td>
      </tr>
    {/snippet}
  </DataTable>

  {#if total > perPage}
    <div class="flex items-center justify-between text-sm text-on-surface-variant">
      <span>Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, total)} of {total}</span>
      <div class="flex gap-2">
        <button
          disabled={page <= 1}
          onclick={() => { page--; loadEvents(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          Previous
        </button>
        <button
          disabled={page * perPage >= total}
          onclick={() => { page++; loadEvents(); }}
          class="px-3 py-1.5 rounded-md bg-surface-low hover:bg-surface-high disabled:opacity-50 transition-colors"
        >
          Next
        </button>
      </div>
    </div>
  {/if}
</div>
