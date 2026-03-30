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

<div class="p-6 lg:p-10 pb-32 max-w-2xl">
  <!-- Header -->
  <header class="mb-12">
    <a href="/staff/events" class="flex items-center gap-2 text-primary font-semibold text-sm mb-2 hover:underline">
      <span class="material-symbols-outlined text-sm">arrow_back</span>
      <span>{$t("staff.create.backToEvents")}</span>
    </a>
    <h1 class="font-display text-4xl font-extrabold tracking-tight text-on-surface">{$t("staff.create.title")}</h1>
  </header>

  {#if error}
    <div class="p-3 rounded-lg bg-error-container text-on-error-container text-sm mb-6">
      {error}
    </div>
  {/if}

  <!-- Form Card -->
  <section class="bg-surface-container-lowest rounded-xl shadow-sm border border-outline-variant/10 overflow-hidden">
    <form onsubmit={handleSubmit} class="p-8 space-y-6">
      <div>
        <label for="title" class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-2">{$t("staff.create.titleLabel")} *</label>
        <input
          id="title"
          bind:value={title}
          required
          class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm
                 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-shadow"
          placeholder={$t("staff.create.titlePlaceholder")}
        />
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <div>
          <label for="date" class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-2">{$t("staff.create.dateLabel")}</label>
          <input
            id="date"
            type="date"
            bind:value={date}
            class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm
                   focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-shadow"
          />
        </div>
        <div>
          <label for="city" class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-2">{$t("staff.create.cityLabel")}</label>
          <input
            id="city"
            bind:value={city}
            class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm
                   focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-shadow"
            placeholder={$t("staff.create.cityPlaceholder")}
          />
        </div>
      </div>

      <div>
        <label for="desc" class="block text-xs font-bold text-on-surface-variant uppercase tracking-wider mb-2">{$t("staff.create.descLabel")}</label>
        <textarea
          id="desc"
          bind:value={description}
          rows="3"
          class="w-full px-4 py-3 rounded-xl bg-surface border border-outline-variant/20 text-on-surface text-sm
                 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 transition-shadow resize-none"
          placeholder={$t("staff.create.descPlaceholder")}
        ></textarea>
      </div>

      <div class="pt-2">
        <button
          type="submit"
          disabled={submitting || !title}
          class="w-full py-3 rounded-xl text-sm font-semibold
                 bg-gradient-to-br from-primary to-primary-container text-white
                 shadow-lg shadow-primary/20 hover:shadow-xl transition-all disabled:opacity-50 active:scale-[0.98]"
        >
          {submitting ? $t("staff.create.creating") : $t("staff.create.submit")}
        </button>
      </div>
    </form>
  </section>
</div>
