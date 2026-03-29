<script lang="ts">
  import { goto } from "$app/navigation";
  import { api, ApiError } from "$lib/api/client";
  import { t } from "$lib/i18n";

  let title = $state("");
  let date = $state("");
  let city = $state("");
  let description = $state("");
  let error = $state("");
  let submitting = $state(false);

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    error = "";
    submitting = true;

    try {
      const res = await api.post<{ id: number }>("/api/v1/staff/events", {
        title,
        date,
        city,
        description,
      });
      goto(`/staff/events/${res.data.id}`);
    } catch (err) {
      if (err instanceof ApiError) {
        error = err.message;
      } else {
        error = $t("staff.create.failed");
      }
    } finally {
      submitting = false;
    }
  }
</script>

<div class="max-w-xl space-y-6">
  <div>
    <a href="/staff/events" class="text-sm text-on-surface-variant hover:text-primary transition-colors">
      &larr; {$t("staff.create.backToEvents")}
    </a>
    <h1 class="font-display text-2xl font-bold text-on-surface mt-2">{$t("staff.create.title")}</h1>
  </div>

  {#if error}
    <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm">
      {error}
    </div>
  {/if}

  <form onsubmit={handleSubmit} class="space-y-5 bg-surface-lowest rounded-lg p-6">
    <div>
      <label for="title" class="block text-sm font-medium text-on-surface mb-1.5">{$t("staff.create.titleLabel")} *</label>
      <input
        id="title"
        bind:value={title}
        required
        class="w-full px-3 py-2.5 rounded-md bg-surface text-on-surface text-sm
               focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow"
        placeholder={$t("staff.create.titlePlaceholder")}
      />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label for="date" class="block text-sm font-medium text-on-surface mb-1.5">{$t("staff.create.dateLabel")}</label>
        <input
          id="date"
          type="date"
          bind:value={date}
          class="w-full px-3 py-2.5 rounded-md bg-surface text-on-surface text-sm
                 focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow"
        />
      </div>
      <div>
        <label for="city" class="block text-sm font-medium text-on-surface mb-1.5">{$t("staff.create.cityLabel")}</label>
        <input
          id="city"
          bind:value={city}
          class="w-full px-3 py-2.5 rounded-md bg-surface text-on-surface text-sm
                 focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow"
          placeholder={$t("staff.create.cityPlaceholder")}
        />
      </div>
    </div>

    <div>
      <label for="desc" class="block text-sm font-medium text-on-surface mb-1.5">{$t("staff.create.descLabel")}</label>
      <textarea
        id="desc"
        bind:value={description}
        rows="3"
        class="w-full px-3 py-2.5 rounded-md bg-surface text-on-surface text-sm
               focus:outline-none focus:ring-2 focus:ring-primary/30 transition-shadow resize-none"
        placeholder={$t("staff.create.descPlaceholder")}
      ></textarea>
    </div>

    <button
      type="submit"
      disabled={submitting || !title}
      class="w-full py-2.5 rounded-lg text-sm font-medium
             bg-gradient-to-br from-primary to-primary-container text-on-primary
             hover:shadow-lg disabled:opacity-50 transition-all"
    >
      {submitting ? $t("staff.create.creating") : $t("staff.create.submit")}
    </button>
  </form>
</div>
