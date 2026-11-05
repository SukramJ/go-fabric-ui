// The node list of the node family (GET /nodes) and the node registry's
// writes (contract 0.6.0): one store per context, shared by the shell's
// selector and every node's pages (NodeScope keeps it), reloaded after each
// write and on the stream's nodes_changed.
import type { FabricClient, NodeSummary } from "./api/client";

export function createNodesStore(client: FabricClient) {
  let list = $state<NodeSummary[]>([]);
  let loaded = $state(false);
  // A reload that a newer one overtook does not overwrite it.
  let generation = 0;

  /** Reloads the list; a server before 0.5.0 (or a failed call) leaves it empty. */
  async function load() {
    const mine = ++generation;
    let next: NodeSummary[];
    try {
      next = (await client.nodes()).nodes;
    } catch {
      next = [];
    }
    if (mine !== generation) return;
    list = next;
    loaded = true;
  }

  /** Creates a node and reloads the list before it answers, so the new key is listed. */
  async function create(label: string): Promise<NodeSummary> {
    const node = await client.createNode(label);
    await load();
    return node;
  }

  async function rename(key: string, label: string) {
    await client.renameNode(key, label);
    await load();
  }

  async function remove(key: string) {
    await client.deleteNode(key);
    await load();
  }

  return {
    /** The nodes, the default first; empty until loaded. */
    get list() {
      return list;
    },
    /** Whether a load has answered. */
    get loaded() {
      return loaded;
    },
    /** The node of key, if listed. */
    find: (key: string | undefined) => (key === undefined ? undefined : list.find((n) => n.key === key)),
    load,
    create,
    rename,
    remove,
  };
}

export type NodesStore = ReturnType<typeof createNodesStore>;
