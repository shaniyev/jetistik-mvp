<script lang="ts">
  interface Props {
    status: string;
    size?: "sm" | "md";
  }

  let { status, size = "sm" }: Props = $props();

  const colors: Record<string, string> = {
    valid: "bg-emerald-50 text-emerald-700",
    active: "bg-emerald-50 text-emerald-700",
    revoked: "bg-red-50 text-red-700",
    inactive: "bg-on-surface-variant/10 text-on-surface-variant",
    pending: "bg-amber-50 text-amber-700",
    uploaded: "bg-blue-50 text-blue-700",
    mapped: "bg-indigo-50 text-indigo-700",
    generating: "bg-purple-50 text-purple-700",
    completed: "bg-emerald-50 text-emerald-700",
    done: "bg-emerald-50 text-emerald-700",
    done_with_errors: "bg-amber-50 text-amber-700",
    failed: "bg-red-50 text-red-700",
    ok: "bg-emerald-50 text-emerald-700",
    archived: "bg-on-surface-variant/10 text-on-surface-variant",
  };

  const sizeClasses = {
    sm: "px-2 py-0.5 text-xs",
    md: "px-2.5 py-1 text-sm",
  };

  let colorClass = $derived(colors[status] ?? "bg-surface-high text-on-surface-variant");
  let sizeClass = $derived(sizeClasses[size]);
</script>

<span class="inline-flex items-center rounded-md font-medium capitalize {colorClass} {sizeClass}">
  {status}
</span>
