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

<div class="min-h-screen bg-surface font-body text-on-surface">
  <!-- Top Navigation Bar -->
  <nav class="fixed top-0 w-full z-50 bg-white/80 backdrop-blur-xl border-b border-slate-200/50 shadow-sm">
    <div class="flex justify-between items-center px-4 sm:px-6 py-3 sm:py-4 max-w-screen-2xl mx-auto">
      <a href="/student" class="text-2xl font-display font-bold tracking-tighter text-blue-700">Jetistik</a>

      <!-- Desktop Links -->
      <div class="hidden md:flex gap-8 items-center font-display font-semibold tracking-tight">
        <a class="text-slate-600 hover:text-blue-500 transition-colors duration-200" href="/">How it works</a>
        <a class="text-slate-600 hover:text-blue-500 transition-colors duration-200" href="/">FAQ</a>
        <a class="text-slate-600 hover:text-blue-500 transition-colors duration-200" href="/">For Organizers</a>
        <a class="text-blue-600 border-b-2 border-blue-600" href="/student">Dashboard</a>
      </div>

      <div class="flex items-center gap-4">
        <button
          onclick={() => auth.logout()}
          class="hidden md:block text-slate-600 font-medium hover:bg-slate-50/50 px-4 py-2 rounded-lg transition-all active:scale-95"
        >
          {$t("nav.logout")}
        </button>
        <div class="w-10 h-10 rounded-full bg-primary-fixed flex items-center justify-center overflow-hidden border-2 border-white shadow-sm text-primary font-bold">
          {$currentUser?.username?.[0]?.toUpperCase() ?? "?"}
        </div>
      </div>
    </div>
  </nav>

  <!-- Main Content Canvas -->
  <main class="pt-20 sm:pt-24 pb-32 md:pb-12 px-4 md:px-8 max-w-7xl mx-auto min-h-screen">
    {@render children()}
  </main>

  <!-- Bottom Navigation (Mobile Only) -->
  <nav class="md:hidden fixed bottom-0 left-0 w-full bg-white/95 backdrop-blur-xl flex justify-around items-center px-4 pb-8 pt-4 border-t border-slate-200/60 z-50">
    <a class="flex flex-col items-center justify-center text-slate-500 group" href="/">
      <span class="material-symbols-outlined transition-colors group-hover:text-primary">home</span>
      <span class="text-[10px] font-semibold uppercase tracking-widest mt-1">Home</span>
    </a>
    <a class="flex flex-col items-center justify-center text-slate-500 group" href="/student">
      <span class="material-symbols-outlined transition-colors group-hover:text-primary">calendar_month</span>
      <span class="text-[10px] font-semibold uppercase tracking-widest mt-1">Events</span>
    </a>
    <a class="flex flex-col items-center justify-center bg-blue-50 text-blue-700 rounded-2xl px-5 py-2.5 -mt-2 shadow-sm border border-blue-100" href="/student">
      <span class="material-symbols-outlined">workspace_premium</span>
      <span class="text-[10px] font-bold uppercase tracking-widest mt-1">Certs</span>
    </a>
    <a class="flex flex-col items-center justify-center text-slate-500 group" href="/student">
      <span class="material-symbols-outlined transition-colors group-hover:text-primary">person</span>
      <span class="text-[10px] font-semibold uppercase tracking-widest mt-1">Profile</span>
    </a>
  </nav>

  <!-- Footer -->
  <footer class="bg-slate-50 border-t border-slate-200 py-12 sm:py-16 px-6">
    <div class="max-w-7xl mx-auto grid grid-cols-1 md:grid-cols-4 gap-12">
      <div class="col-span-1">
        <div class="font-display font-bold text-2xl text-slate-900 mb-4">Jetistik</div>
        <p class="text-slate-500 text-sm leading-relaxed max-w-xs">
          Elevating the standard of professional certification in Kazakhstan through the Sovereign Ledger system.
        </p>
      </div>
      <div class="space-y-4">
        <h4 class="font-bold text-slate-900 text-xs uppercase tracking-[0.2em]">Platform</h4>
        <nav class="flex flex-col gap-3">
          <a class="text-slate-500 hover:text-blue-600 text-sm transition-colors" href="/">Privacy Policy</a>
          <a class="text-slate-500 hover:text-blue-600 text-sm transition-colors" href="/">Terms of Service</a>
          <a class="text-slate-500 hover:text-blue-600 text-sm transition-colors" href="/verify">Verification</a>
        </nav>
      </div>
      <div class="space-y-4">
        <h4 class="font-bold text-slate-900 text-xs uppercase tracking-[0.2em]">Support</h4>
        <nav class="flex flex-col gap-3">
          <a class="text-slate-500 hover:text-blue-600 text-sm transition-colors" href="/">Contact Us</a>
          <a class="text-slate-500 hover:text-blue-600 text-sm transition-colors" href="/">FAQ</a>
          <a class="text-slate-500 hover:text-blue-600 text-sm transition-colors" href="/">Guidelines</a>
        </nav>
      </div>
      <div class="space-y-4">
        <h4 class="font-bold text-slate-900 text-xs uppercase tracking-[0.2em]">Language / Til</h4>
        <div class="flex gap-2">
          <button class="bg-white border border-slate-200 px-4 py-2 rounded-lg text-xs font-bold hover:border-primary transition-colors shadow-sm">KZ</button>
          <button class="bg-white border border-slate-200 px-4 py-2 rounded-lg text-xs font-bold hover:border-primary transition-colors shadow-sm">RU</button>
          <button class="bg-primary text-white px-4 py-2 rounded-lg text-xs font-bold shadow-md shadow-primary/20">EN</button>
        </div>
      </div>
    </div>
    <div class="max-w-7xl mx-auto mt-12 sm:mt-16 pt-8 border-t border-slate-200 flex flex-col md:flex-row justify-between items-center gap-6">
      <p class="text-slate-400 text-xs">2024 Jetistik Certificate Platform. All rights reserved.</p>
      <div class="flex gap-8">
        <span class="material-symbols-outlined text-slate-400 hover:text-primary cursor-pointer transition-colors">public</span>
        <span class="material-symbols-outlined text-slate-400 hover:text-primary cursor-pointer transition-colors">shield</span>
        <span class="material-symbols-outlined text-slate-400 hover:text-primary cursor-pointer transition-colors">history</span>
      </div>
    </div>
  </footer>
</div>
