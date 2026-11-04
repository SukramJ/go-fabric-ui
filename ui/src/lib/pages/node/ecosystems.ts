// The pairing hints per controller ecosystem: static copy taken from
// matter.js docs/ECOSYSTEMS.md at 862ad75f (Apple :89-99, Google :101-114,
// Amazon :116-129, SmartThings :136-146, Home Assistant :148-153), with
// the links the document gives and without its app version numbers. The
// text is in the catalogues (node.pair.hints.<ecosystem>), the link labels
// under node.pair.hints.<ecosystem>.<link>.
import type { Ecosystem } from "../../api/client";

export type EcosystemHint = { ecosystem: Ecosystem["ecosystem"]; links: { key: string; href: string }[] };

export const ECOSYSTEM_HINTS: EcosystemHint[] = [
  {
    ecosystem: "apple",
    links: [
      { key: "hub", href: "https://support.apple.com/en-us/HT207057" },
      { key: "pair", href: "https://support.apple.com/en-asia/102135" },
    ],
  },
  {
    ecosystem: "google",
    links: [
      { key: "hubs", href: "https://support.google.com/googlenest/answer/12391458?hl=en" },
      { key: "troubleshooting", href: "https://developers.home.google.com/matter/build/troubleshooting?hl=en#verify_your_google_play_services_gps_matter_modules" },
    ],
  },
  { ecosystem: "amazon", links: [{ key: "hubs", href: "https://www.amazon.com/b?ie=UTF8&node=37490568011" }] },
  { ecosystem: "smartthings", links: [] },
  {
    ecosystem: "home_assistant",
    links: [
      { key: "integration", href: "https://www.home-assistant.io/integrations/matter/" },
      { key: "thread", href: "https://www.home-assistant.io/integrations/thread/" },
    ],
  },
];
