<script lang="ts">
  import { t } from '$lib/i18n';
</script>

<svelte:head>
  <title>{$t("terms.title")} — Jetistik</title>
</svelte:head>

<div class="min-h-screen bg-surface">
  <header class="bg-white/80 backdrop-blur-xl border-b border-outline-variant/15 sticky top-0 z-50">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
      <a href="/" class="text-2xl font-display font-bold tracking-tighter text-primary">Jetistik</a>
    </div>
  </header>

  <main class="max-w-4xl mx-auto px-4 sm:px-6 py-12">
    <h1 class="font-display text-3xl font-extrabold text-on-surface mb-2">{$t("terms.title")}</h1>
    <p class="text-sm text-on-surface-variant mb-10">{$t("privacy.lastUpdated")}: 30.03.2026</p>

    <div class="prose prose-slate max-w-none space-y-8 text-on-surface text-[15px] leading-relaxed">

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s1.title")}</h2>
        <p>{$t("terms.s1.text")}</p>
      </section>

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s2.title")}</h2>
        <p>{$t("terms.s2.text")}</p>
        <ul class="list-disc pl-6 mt-3 space-y-1 text-on-surface-variant">
          <li>{$t("terms.s2.item1")}</li>
          <li>{$t("terms.s2.item2")}</li>
          <li>{$t("terms.s2.item3")}</li>
          <li>{$t("terms.s2.item4")}</li>
        </ul>
      </section>

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s3.title")}</h2>
        <p>{$t("terms.s3.text")}</p>
      </section>

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s4.title")}</h2>
        <p>{$t("terms.s4.text")}</p>
      </section>

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s5.title")}</h2>
        <p>{$t("terms.s5.text")}</p>
        <ul class="list-disc pl-6 mt-3 space-y-1 text-on-surface-variant">
          <li>{$t("terms.s5.item1")}</li>
          <li>{$t("terms.s5.item2")}</li>
          <li>{$t("terms.s5.item3")}</li>
        </ul>
      </section>

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s6.title")}</h2>
        <p>{$t("terms.s6.text")}</p>
      </section>

      <section>
        <h2 class="font-display text-xl font-bold mb-3">{$t("terms.s7.title")}</h2>
        <p>{$t("terms.s7.text")}</p>
      </section>

      <section class="bg-surface-container-low p-6 rounded-xl">
        <h2 class="font-display text-xl font-bold mb-3">{$t("privacy.contact.title")}</h2>
        <p class="text-on-surface-variant">{$t("privacy.contact.text")}</p>
        <p class="mt-2 font-medium">support@jetistik.kz</p>
      </section>
    </div>
  </main>

  <footer class="border-t border-outline-variant/10 py-8 mt-8">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 text-center">
      <p class="text-xs text-on-surface-variant">&copy; {new Date().getFullYear()} Jetistik</p>
    </div>
  </footer>
</div>
