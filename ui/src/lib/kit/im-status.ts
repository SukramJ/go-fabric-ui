// The Interaction Model status codes an attribute read can answer with,
// by the name go-fabric gives them (im/status.go:64-101 at 82591648,
// Matter Core §8.10), each with its i18n key under inventory.status. A
// status outside the map shows as hex.
export const IM_STATUS: Readonly<Record<number, string>> = {
  0x01: "failure",
  0x7e: "unsupported_access",
  0x7f: "unsupported_endpoint",
  0x86: "unsupported_attribute",
  0x89: "resource_exhausted",
  0x8f: "unsupported_read",
  0x94: "timeout",
  0x9c: "busy",
  0x9d: "access_restricted",
  0xc3: "unsupported_cluster",
};

/** statusHex is a status as two hex digits, "0x86". */
export const statusHex = (s: number) => "0x" + s.toString(16).toUpperCase().padStart(2, "0");

/** imStatusName is the status's protocol name ("UNSUPPORTED_ATTRIBUTE"); undefined outside the map. */
export const imStatusName = (s: number) => IM_STATUS[s]?.toUpperCase();
