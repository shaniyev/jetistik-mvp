<script>
  let { children } = $props();
</script>

<style>
  .glass-effect {
    backdrop-filter: blur(12px);
    background: rgba(255, 255, 255, 0.8);
  }
  .material-symbols-outlined {
    font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
    display: inline-block;
    line-height: 1;
    text-transform: none;
    letter-spacing: normal;
    word-wrap: normal;
    white-space: nowrap;
    direction: ltr;
  }
</style>

<div class="bg-surface font-body text-on-surface min-h-screen flex flex-col">
  <!-- Header -->
  <header class="fixed top-0 w-full z-50 glass-effect border-b border-outline-variant/15">
    <div class="flex justify-between items-center px-6 py-4 max-w-screen-2xl mx-auto">
      <a href="/" class="text-2xl font-display font-extrabold tracking-tighter text-primary hover:opacity-80 transition-opacity">
        Jetistik
      </a>
      <div class="flex items-center gap-4">
        <span class="text-xs font-body font-semibold text-on-surface-variant tracking-wider uppercase">Merit Ledger</span>
      </div>
    </div>
  </header>

  <!-- Main content -->
  <main class="flex-grow flex items-center justify-center px-4 pt-24 pb-12">
    {@render children()}
  </main>

  <!-- Footer -->
  <footer class="w-full py-8 px-6 mt-auto border-t border-outline-variant/15 bg-surface-container-low">
    <div class="max-w-7xl mx-auto flex flex-col md:flex-row justify-between items-center gap-4">
      <p class="font-body text-xs text-on-surface-variant opacity-70">
        &copy; 2024 Jetistik Certificate Platform. All rights reserved.
      </p>
      <div class="flex gap-6">
        <a class="text-xs font-semibold text-on-surface-variant hover:text-primary transition-colors" href="/verify">Verification</a>
      </div>
    </div>
  </footer>

  <!-- Background decoration -->
  <div class="fixed top-0 left-0 w-full h-full -z-10 pointer-events-none overflow-hidden">
    <div class="absolute top-[-10%] right-[-10%] w-[60%] h-[60%] bg-primary/5 blur-[120px] rounded-full"></div>
    <div class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-secondary-container/10 blur-[100px] rounded-full"></div>
  </div>
</div>
