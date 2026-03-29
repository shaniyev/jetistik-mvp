<script lang="ts">
  import { goto } from "$app/navigation";
  import { currentUser } from "$lib/stores/auth";
  import { auth } from "$lib/stores/auth";
  import { t } from "$lib/i18n";

  let { children } = $props();

  $effect(() => {
    if ($currentUser && $currentUser.role !== "student") {
      goto("/");
    }
  });
</script>

<div class="min-h-screen bg-surface">
  <!-- Topbar -->
  <header class="bg-surface-lowest/80 backdrop-blur-xl sticky top-0 z-30">
    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between">
      <a href="/student" class="font-display text-lg font-bold text-on-surface">Jetistik</a>

      <nav class="hidden sm:flex items-center gap-6 text-sm text-on-surface-variant">
        <a href="/student" class="hover:text-on-surface transition-colors">{$t("student.title")}</a>
      </nav>

      <div class="flex items-center gap-3">
        <button
          onclick={() => auth.logout()}
          class="text-xs text-on-surface-variant hover:text-error transition-colors"
        >
          {$t("nav.logout")}
        </button>
        <div class="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary text-sm font-bold">
          {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
        </div>
      </div>
    </div>
  </header>

  <!-- Main content -->
  <main class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    {@render children()}
  </main>
</div>
