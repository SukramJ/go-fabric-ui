<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Spinner.svelte:1-100 at b40b289a: three threads with a
     sweeping segment; decorative, collapses to a static mark under reduced
     motion. -->
<script lang="ts">
  import { cn } from "./utils";

  type Props = { size?: number; class?: string };
  let { size = 18, class: className }: Props = $props();
</script>

<span class={cn("fui-loader text-[var(--ha-secondary-text-color)]", className)} style="--fui-size:{size}px" aria-hidden="true">
  <span class="fui-thread"></span>
  <span class="fui-thread"></span>
  <span class="fui-thread"></span>
</span>

<style>
  .fui-loader {
    display: inline-block;
    position: relative;
    width: calc(var(--fui-size) * 1.5);
    height: var(--fui-size);
    flex: none;
  }
  .fui-thread {
    position: absolute;
    left: 0;
    right: 0;
    height: max(1.5px, calc(var(--fui-size) / 9));
    border-radius: 9999px;
    background-color: currentColor;
    opacity: 0.25;
    overflow: hidden;
  }
  .fui-thread:nth-child(1) { top: 12%; }
  .fui-thread:nth-child(2) { top: 44%; }
  .fui-thread:nth-child(3) { top: 76%; }
  .fui-thread::after {
    content: "";
    position: absolute;
    top: 0;
    bottom: 0;
    width: 40%;
    border-radius: inherit;
    background-color: currentColor;
    opacity: 0.9;
    animation: fui-sweep 1.1s ease-in-out infinite;
  }
  .fui-thread:nth-child(2)::after { animation-delay: 0.16s; }
  .fui-thread:nth-child(3)::after { animation-delay: 0.32s; }
  @keyframes fui-sweep {
    0% { transform: translateX(-110%); }
    55%, 100% { transform: translateX(360%); }
  }
  @media (prefers-reduced-motion: reduce) {
    .fui-thread { opacity: 0.5; }
    .fui-thread::after { animation: none; opacity: 0; }
  }
</style>
