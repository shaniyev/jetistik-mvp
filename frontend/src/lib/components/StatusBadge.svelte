<script lang="ts">
  interface Props {
    status: string;
    size?: "sm" | "md";
  }

  let { status, size = "sm" }: Props = $props();

  const colors: Record<string, string> = {
    valid: "bg-primary-fixed text-on-primary-fixed-variant",
    active: "bg-primary-fixed text-on-primary-fixed-variant",
    revoked: "bg-error-container text-on-error-container",
    inactive: "bg-outline-variant/30 text-on-surface-variant",
    pending: "bg-error-container text-on-error-container",
    uploaded: "bg-blue-50 text-blue-700",
    mapped: "bg-indigo-50 text-indigo-700",
    generating: "bg-purple-50 text-purple-700",
    completed: "bg-primary-fixed text-on-primary-fixed-variant",
    done: "bg-primary-fixed text-on-primary-fixed-variant",
    done_with_errors: "bg-error-container text-on-error-container",
    failed: "bg-error-container text-on-error-container",
    ok: "bg-primary-fixed text-on-primary-fixed-variant",
    archived: "bg-outline-variant/30 text-on-surface-variant",
    suspended: "bg-error-container text-on-error-container",
  };

  const sizeClasses = {
    sm: "px-3 py-1 text-[10px]",
    md: "px-3.5 py-1.5 text-xs",
  };

  let colorClass = $derived(colors[status] ?? "bg-surface-high text-on-surface-variant");
  let sizeClass = $derived(sizeClasses[size]);
</script>

<span class="inline-flex items-center rounded-full font-bold uppercase tracking-wider {colorClass} {sizeClass}">
  {status}
</span>
