<script>
  let { children } = $props();
</script>

<div class="min-h-screen bg-surface flex items-center justify-center p-4">
  <div class="w-full max-w-md">
    <div class="text-center mb-8">
      <h1 class="font-display text-2xl font-bold text-on-surface">Jetistik</h1>
      <p class="text-on-surface-variant text-sm mt-1">
        Digital certificate platform
      </p>
    </div>
    <div class="bg-surface-lowest rounded-lg p-8 shadow-[0_4px_40px_rgba(0,74,198,0.04)]">
      {@render children()}
    </div>
  </div>
</div>
