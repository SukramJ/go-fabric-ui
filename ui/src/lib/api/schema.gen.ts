// Generated from contract/openapi.yaml by `npm run gen:types`. Do not edit;
// src/lib/api/schema.test.ts fails when this file is stale.

export interface paths {
    "/node/status": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Health, counts and the commissioning window. */
        get: operations["getNodeStatus"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/fabrics": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The commissioned fabrics. */
        get: operations["listFabrics"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/fabrics/{index}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Remove one fabric. */
        delete: operations["removeFabric"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/pairing": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The pairing codes — of the window this handler opened while it is open, else the configured ones. */
        get: operations["getPairing"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/pairing/window": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Open a commissioning window. */
        post: operations["openWindow"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/pairing/window/close": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Close the open commissioning window. */
        post: operations["closeWindow"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/endpoints": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The bridged endpoints a controller sees. */
        get: operations["listEndpoints"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/surface": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * The node's surface report.
         * @description go-fabric's `surface.Report` of the node in go-fabric's own JSON
         *     encoding (`surface.Report.MarshalJSON`): the effective surface per
         *     endpoint, the decision ledger, the withheld elements and the
         *     surface changes announced in this process. Element and cluster ids
         *     are lowercase `0x` hex strings, as wide as their kind (a device
         *     type four digits, everything else eight).
         */
        get: operations["getNodeSurface"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/sessions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The open secure sessions and the session-id occupancy. */
        get: operations["listSessions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/mdns": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The live DNS-SD advertisement and its diagnosis. */
        get: operations["getMdns"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/compatibility": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The commissioned ecosystems and known incompatibilities. */
        get: operations["getCompatibility"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/events": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The pairing, session and discovery trace, newest first. */
        get: operations["listDiagnosticEvents"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/exposure/schema": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** How the selection page shows the host's items. */
        get: operations["getExposureSchema"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/exposure": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Every item the host could expose. */
        get: operations["listExposure"];
        /** Apply a batch of item changes and reassemble the endpoints. */
        put: operations["applyExposure"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/exposure/preview": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** The items as a batch of changes would leave them; nothing is written. */
        post: operations["previewExposure"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/maintenance/reassemble": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Rebuild the endpoints from the host's current selection; pairings are untouched. */
        post: operations["reassemble"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/maintenance/factory-reset": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Remove every fabric. The body must name the action. */
        post: operations["factoryReset"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/node/stream": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Server-sent events.
         * @description `event:` is one of fabric_added, fabric_removed, window_opened,
         *     window_closed, endpoints_changed, exposure_changed, surface_changed;
         *     `data:` is the JSON payload (`FabricEvent`, `WindowEvent`, `{}`,
         *     `ExposureEvent`, `SurfaceChangedEvent`).
         *     A comment line arrives every 15 s. No payload carries a passcode.
         */
        get: operations["stream"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/status": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * The controller's fabric, its node count and capabilities.
         * @description Answered only when the host serves the controller family
         *     (`Options.Controller`); without it every `/controller/*` path is
         *     404, which is how a page tells the family is absent.
         */
        get: operations["getControllerStatus"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The commissioned nodes. */
        get: operations["listControllerNodes"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        /** One node — record, properties, endpoints with clusters and device types. */
        get: operations["getControllerNode"];
        put?: never;
        post?: never;
        /** Decommission the node — remove our fabric from it and forget it. */
        delete: operations["decommissionNode"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/interview": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Read the whole node again, without data version filters. */
        post: operations["interviewNode"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/enable": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Enable a disabled node — connect and subscribe again. */
        post: operations["enableNode"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/disable": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Disable the node — no connection, no subscription, kept on the fabric. */
        post: operations["disableNode"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/auto-subscribe": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Turn the node's default subscription on or off. */
        post: operations["setNodeAutoSubscribe"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/window": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Open an enhanced commissioning window on the node for another administrator.
         * @description The answer carries the window's passcode and codes; no event and no log line does.
         */
        post: operations["openNodeWindow"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/window/close": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Revoke the node's commissioning window. */
        post: operations["closeNodeWindow"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/attributes": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        /** Write one attribute. */
        put: operations["writeNodeAttribute"];
        /**
         * Read concrete attribute paths from the node.
         * @description A targeted read: every path names its endpoint, cluster and
         *     attribute (no wildcard, at most 32 paths). The value is the JSON
         *     rendering of the attribute by the cluster definitions, a struct
         *     keyed by field name; a path the node refused carries its status
         *     instead.
         */
        post: operations["readNodeAttributes"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/commands": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Invoke one command. */
        post: operations["invokeNodeCommand"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/acl": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        /** The node's access control entries of our fabric. */
        get: operations["getNodeAcl"];
        /**
         * Replace our fabric's access control entries on the node.
         * @description The whole list is written in one write of the AccessControl `ACL`
         *     attribute. A list without an Administer entry for this controller
         *     locks it out of the node; the page warns before it sends one.
         */
        put: operations["setNodeAcl"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/bindings": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        /** The binding targets of every endpoint with a Binding cluster. */
        get: operations["getNodeBindings"];
        /** Replace our fabric's binding targets on one endpoint. */
        put: operations["setNodeBindings"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/icd": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        /** The node's intermittently-connected-device state. */
        get: operations["getNodeIcd"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/icd/register": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Register this controller as the node's check-in client. */
        post: operations["registerNodeIcd"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/icd/unregister": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Unregister this controller; with force, forget the registration without the node. */
        post: operations["unregisterNodeIcd"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/fabrics": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        /** Every fabric on the node, ours marked. */
        get: operations["listNodeFabrics"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/nodes/{id}/fabrics/{index}": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
                index: number;
            };
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /**
         * Remove another administrator's fabric from the node (RemoveFabric).
         * @description Our own fabric is removed with `DELETE /controller/nodes/{id}`; naming it here is 400.
         */
        delete: operations["removeNodeFabric"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/commission": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Commission a device from a code or QR payload, or a discovered device and its passcode.
         * @description The answer arrives when the commissioning ends; the stream reports
         *     `commissioning_progress` meanwhile. The passcode and the code are
         *     never logged and never in an event.
         */
        post: operations["commissionNode"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/discovery": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Commissionable devices on the network. */
        get: operations["discoverDevices"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/credentials": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The stored Wi-Fi and Thread credentials, by name; never a secret. */
        get: operations["listCredentials"];
        /** Store a Wi-Fi credential or a Thread dataset (write-only). */
        put: operations["setCredential"];
        post?: never;
        /** Remove a stored credential. */
        delete: operations["removeCredential"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/fabric/label": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** The controller's fabric label, pushed to every node. */
        put: operations["setFabricLabel"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/updates": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** The software update queue. */
        get: operations["listUpdates"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/updates/consent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Consent to one node's update to a version. */
        post: operations["consentUpdate"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/updates/check": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Ask every node for an available update. */
        post: operations["checkUpdates"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/controller/stream": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Server-sent events of the controller.
         * @description `event:` is one of node_added, node_updated, node_removed,
         *     connection, subscription, commissioning_progress, update; `data:`
         *     is a `ControllerEvent`. A comment line arrives every 15 s. No
         *     payload carries a passcode or a credential.
         */
        get: operations["controllerStream"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        Problem: {
            type: string;
            title: string;
            status: number;
            detail?: string;
            instance?: string;
            /** @enum {string} */
            code: "bad_request" | "forbidden" | "conflict" | "not_found" | "service_unready" | "internal" | "payload_too_large" | "node_error";
        };
        Status: {
            enabled: boolean;
            listening: boolean;
            listen_addr?: string;
            /** @description Absent when the node runs no health probe. */
            healthy?: boolean;
            health_note?: string;
            endpoint_count: number;
            fabric_count: number;
            /** @description Enabled selection items; absent without a selection. */
            enabled_count?: number;
            advertising: boolean;
            commissioning_window_open: boolean;
            commissioning_window_duration_seconds?: number;
            /**
             * @description Which window is open — the node's own on its configured credentials, a basic one on fresh credentials, an administrator's enhanced one; absent without a node.
             * @enum {string}
             */
            window_kind?: "none" | "own" | "basic" | "enhanced";
            /**
             * Format: date-time
             * @description When the open window's timer closes it
             */
            window_deadline?: string;
            /** @description Whether the host serves a selection. */
            exposure: boolean;
            /** @description The admin token of token mode; absent when the host decides admin rights. */
            csrf_token?: string;
            /** @description The host's findings about its own configuration; absent when the host reports none (no Options.Warnings), empty when it finds nothing. */
            warnings?: components["schemas"]["Warning"][];
            /** @description The policies the node's served surface is composed under (go-fabric node.Node.Surface) and the counts that tell whether to look further; absent without a node. */
            surface?: {
                /** @description The specification level the surface is composed at (go-fabric surface.Report.Level), e.g. "1.6.1". */
                level: string;
                /** @description "explicit", "profile <name>[, <name>…]" or "pin" (surface.Report.LevelOrigin). */
                level_origin: string;
                /** @description The selected controller profiles, sorted. */
                profiles: string[];
                /** @description The enabled ahead-of-matter.js register IDs, sorted. */
                extensions: string[];
                /** @description The report reads the assembled topology (after the first assembly). */
                assembled: boolean;
                /** @description Elements the level policy withheld an endpoint for. */
                withheld_count: number;
                /** @description Warn decisions: elements a policy keeps and flags (surface.Warn). */
                warning_count: number;
                /** @description Surface changes announced in this process (ConfigurationVersion raised). */
                changed_count: number;
            };
        };
        Warning: {
            /** @description The kind of finding, stable across host releases. */
            code: string;
            /** @description What is wrong, in the host's language. */
            message: string;
            /**
             * Format: date-time
             * @description When the host first saw it; absent when unknown.
             */
            since?: string;
        };
        Fabric: {
            fabric_index: number;
            /** @description Informational; use fabric_id_hex. */
            fabric_id: number;
            fabric_id_hex: string;
            /** @description Informational; use node_id_hex. */
            node_id: number;
            node_id_hex: string;
            vendor_id: number;
            vendor_name: string;
            label?: string;
            compressed_id: string;
            root_public_key: string;
        };
        FabricList: {
            fabrics: components["schemas"]["Fabric"][];
        };
        Pairing: {
            /** @description The discriminator the codes encode: of the window this handler opened on fresh credentials, else the configured one. Absent from servers before 0.1.2 for the configured source. */
            discriminator?: number;
            /** @description The passcode the codes encode: of the window this handler opened on fresh credentials, else the configured one. Absent from servers before 0.1.2 for the configured source. */
            passcode?: number;
            qr_code: string;
            manual_code: string;
            /** @enum {string} */
            source: "configured" | "window";
            window_open: boolean;
            duration_seconds?: number;
        };
        WindowRequest: {
            /** @description 180..900; 0 or absent is 900. */
            duration_seconds?: number;
        };
        Window: {
            /** @description The window's discriminator; absent from servers before 0.1.2 for a window on the configured credentials. */
            discriminator?: number;
            /** @description The window's passcode; absent from servers before 0.1.2 for a window on the configured credentials. */
            passcode?: number;
            duration_seconds: number;
            qr_code: string;
            manual_code: string;
        };
        /** @description One cluster server mounted on the endpoint, in its Descriptor's ServerList order. */
        Cluster: {
            id: number;
            /** @description The schema's cluster name; absent for a cluster the schema does not know. */
            name?: string;
            /** @description The schema's ClusterRevision; absent for a cluster the schema does not know. */
            revision?: number;
        };
        Endpoint: {
            endpoint_id: number;
            parent_endpoint_id: number;
            device_type: number;
            device_type_name: string;
            /** @description The revision the endpoint's Descriptor serves for its device type; absent for a type the schema does not know. */
            device_type_revision?: number;
            reachable: boolean;
            name: string;
            /** @description Opaque identity of the host's source; `key` is the endpoint's stable key. */
            source: {
                [key: string]: string;
            };
            clusters: components["schemas"]["Cluster"][];
        };
        EndpointList: {
            endpoints: components["schemas"]["Endpoint"][];
        };
        Session: {
            session_id: number;
            fabric_index: number;
            /** @description Informational; use peer_node_id_hex. */
            peer_node_id: number;
            peer_node_id_hex: string;
            /** @description Informational; use local_node_id_hex. */
            local_node_id: number;
            local_node_id_hex: string;
            is_pase: boolean;
            subscriptions: number;
            /** Format: date-time */
            last_activity: string;
            /** Format: date-time */
            last_peer_activity: string;
            idle_seconds: number;
            peer_idle_seconds: number;
        };
        Occupancy: {
            live: number;
            reserved: number;
            capacity: number;
            free: number;
        };
        SessionList: {
            sessions: components["schemas"]["Session"][];
            occupancy: components["schemas"]["Occupancy"];
        };
        MdnsService: {
            service_type: string;
            instance_name: string;
            host_name: string;
            port: number;
            addresses: string[];
            subtypes: string[];
            txt: {
                [key: string]: string;
            };
        };
        MdnsFinding: {
            /** @enum {string} */
            severity: "error" | "warning";
            code: string;
            message: string;
            service?: string;
        };
        Mdns: {
            advertising: boolean;
            services: components["schemas"]["MdnsService"][];
            findings: components["schemas"]["MdnsFinding"][];
        };
        Ecosystem: {
            ecosystem: string;
            vendor_id: number;
            fabric_index: number;
            label?: string;
        };
        CompatFinding: {
            ecosystem: string;
            code: string;
            message: string;
            device_type?: number;
        };
        Compatibility: {
            ecosystems: components["schemas"]["Ecosystem"][];
            endpoint_count: number;
            findings: components["schemas"]["CompatFinding"][];
        };
        DiagnosticEvent: {
            /** Format: date-time */
            at: string;
            /** @enum {string} */
            kind: "pairing" | "session" | "discovery";
            /** @enum {string} */
            severity: "info" | "warning" | "error";
            message: string;
            detail?: {
                [key: string]: string;
            };
        };
        DiagnosticEventList: {
            events: components["schemas"]["DiagnosticEvent"][];
        };
        Column: {
            key: string;
            label: string;
            /** @description Offer a value filter built from the listed values. */
            filter: boolean;
            /** @description Default false. The field is list-valued: its values are the item's lists entry for the key (or, deprecated until v0.2.0, the comma split of its fields entry when lists has no such key); shown as chips, filtered by membership, the filter's options built from the individual values. */
            multi?: boolean;
        };
        FilterOption: {
            value: string;
            label: string;
        };
        Filter: {
            field: string;
            label: string;
            options: components["schemas"]["FilterOption"][];
            /** @description Default false. Match by membership in the item's list for the field (its lists entry, or, deprecated until v0.2.0, the comma split of its fields entry, values trimmed; an empty value never matches) instead of equality. */
            multi?: boolean;
        };
        ExposureSchema: {
            columns: components["schemas"]["Column"][];
            groups: boolean;
            filters: components["schemas"]["Filter"][];
        };
        Verdict: {
            /** @enum {string} */
            state: "mappable" | "partially_mappable" | "unmappable";
            device_type: number;
            device_type_name: string;
            clusters: number[];
            reason: string;
        };
        Item: {
            key: string;
            name: string;
            group: string;
            fields: {
                [key: string]: string;
            };
            /** @description Optional. The values of list-valued (multi) fields, by field key: one entry per value, no separator semantics, so a value may contain a comma. Wins over the comma-separated fields entry of the same key. */
            lists?: {
                [key: string]: string[];
            };
            verdict: components["schemas"]["Verdict"];
            enabled: boolean;
            /** @description Where the verdict comes from */
            source: string;
        };
        ItemList: {
            items: components["schemas"]["Item"][];
        };
        Update: {
            key: string;
            enabled: boolean;
            /** @description Renames the item; absent leaves the name. */
            name?: string;
        };
        UpdateList: {
            items: components["schemas"]["Update"][];
        };
        Applied: {
            applied: number;
        };
        FactoryResetRequest: {
            /** @constant */
            confirm: "remove-all-fabrics";
        };
        FabricEvent: {
            fabric_index: number;
            vendor_id?: number;
            vendor_name?: string;
            label?: string;
        };
        WindowEvent: {
            discriminator?: number;
            duration_seconds?: number;
        };
        ExposureEvent: {
            keys: string[];
        };
        SurfaceChangedEvent: {
            /** @description The keys of the endpoints a reassembly announced a new surface change for; "" is the root endpoint. */
            endpoints: string[];
        };
        /** @description go-fabric surface.Report (surface/policy.go wireReport). */
        SurfaceReport: {
            /** @description The report reads the assembled topology. */
            assembled: boolean;
            /** @description The surface changes announced in this process, oldest first (the node keeps the last hundred). */
            changed: components["schemas"]["SurfaceChange"][];
            /** @description Every decision of the policy chain, in its order. */
            decisions: components["schemas"]["SurfaceDecision"][];
            effective: components["schemas"]["SurfaceEffective"];
            /** @description The enabled ahead-of-matter.js register IDs, sorted. */
            extensions: string[];
            /** @description The specification level the surface is composed at. */
            level: string;
            /** @description "explicit", "profile <name>[, <name>…]" or "pin". */
            level_origin: string;
            /** @description The digest of the policies, 64 hex digits. */
            policies: string;
            /** @description The selected controller profiles, sorted. */
            profiles: string[];
            /** @description The elements the level policy withheld an endpoint for. */
            withheld: components["schemas"]["SurfaceElement"][];
        };
        /** @description go-fabric surface.Surface (surface/surface.go wireSurface). */
        SurfaceEffective: {
            /** @description The endpoints other than the root, by key. */
            endpoints: components["schemas"]["SurfaceEndpoint"][];
            root: components["schemas"]["SurfaceEndpoint"];
        };
        SurfaceEndpoint: {
            /** @description The server clusters, by id. */
            clusters: components["schemas"]["SurfaceCluster"][];
            /** @description The device types in the Descriptor's order. */
            deviceTypes: components["schemas"]["SurfaceDeviceType"][];
            /** @description The endpoint's rendered stable key; "" for the root, "aggregator" for the Aggregator, "@<number>" for an endpoint without a key. */
            key: string;
            /** @description The endpoint number; absent until the endpoint is assembled, and for the root. */
            number?: number;
        };
        SurfaceDeviceType: {
            /** @description Device type id, "0x" and four hex digits. */
            id: string;
            revision: number;
        };
        SurfaceCluster: {
            /** @description AcceptedCommandList, sorted. */
            accepted: string[];
            /** @description AttributeList, sorted. */
            attributes: string[];
            /** @description The events served, sorted. */
            events: string[];
            featureMap: string;
            /** @description GeneratedCommandList, sorted. */
            generated: string[];
            id: string;
            /** @description ClusterRevision. */
            revision: number;
        };
        /** @description One candidate element of an endpoint's surface (surface.Element). */
        SurfaceElement: {
            /** @description The cluster id; "0x00000000" for a device type. */
            cluster: string;
            /** @description The endpoint's key; "" for the root. */
            endpoint: string;
            /** @description The device type id, the cluster id, the feature's one-bit mask, or the attribute, command or event id. */
            id: string;
            /** @enum {string} */
            kind: "device-type" | "cluster" | "feature" | "attribute" | "command" | "event";
            /** @description Removing the element would make the composition invalid at the pin. */
            mandatory: boolean;
            /** @enum {string} */
            origin: "cluster-mandatory" | "device-type" | "host-declared" | "kit-default";
            /** @description The ahead-of-matter.js register ID it is served against; "" for the matter.js baseline. */
            reference: string;
        };
        /** @description One policy's judgement of one element (surface.Decision). */
        SurfaceDecision: {
            cluster: string;
            endpoint: string;
            id: string;
            /** @enum {string} */
            kind: "device-type" | "cluster" | "feature" | "attribute" | "command" | "event";
            mandatory: boolean;
            /** @enum {string} */
            origin: "cluster-mandatory" | "device-type" | "host-declared" | "kit-default";
            /** @description The policy's name: "level", "reference", "profile:<name>" or a host's own. */
            policy: string;
            /** @description Empty for a silent keep. */
            reason: string;
            reference: string;
            /** @enum {string} */
            verdict: "keep" | "warn" | "remove" | "refuse";
        };
        /** @description One announced change of an endpoint's served surface (surface.SurfaceChange). */
        SurfaceChange: {
            /** @description The new surface's digest, 64 hex digits. */
            after: string;
            /**
             * Format: date-time
             * @description When the node raised the ConfigurationVersion.
             */
            at: string;
            /** @description The digest last persisted for the endpoint. */
            before: string;
            /** @description The differences; empty when only the earlier digest was available. */
            changes: components["schemas"]["SurfaceElementChange"][];
            /** @description The endpoint's key. */
            endpoint: string;
        };
        /** @description One element-level difference (surface.Change). */
        SurfaceElementChange: {
            /** @description The element's state after ("revision 6", "set", "served", "accepted", "generated"); "" when it is gone. */
            after: string;
            /** @description The state before; "" when it is new. */
            before: string;
            cluster: string;
            endpoint: string;
            id: string;
            /** @enum {string} */
            kind: "device-type" | "cluster" | "feature" | "attribute" | "command" | "event";
        };
        ControllerStatus: {
            /** @description Always true; the route is absent without a controller. */
            enabled: boolean;
            fabric_id_hex: string;
            compressed_fabric_id_hex: string;
            fabric_index: number;
            /** @description The controller's own operational node id. */
            node_id_hex: string;
            vendor_id: number;
            vendor_name: string;
            label: string;
            node_count: number;
            capabilities: components["schemas"]["ControllerCapabilities"];
            /** @description The admin token of token mode */
            csrf_token?: string;
        };
        ControllerCapabilities: {
            /** @description The controller runs an OTA provider; the update routes answer. */
            ota: boolean;
            /** @description The host stores Wi-Fi and Thread credentials; the credential routes answer. */
            credentials: boolean;
            /** @description The controller can commission over Bluetooth LE. */
            ble: boolean;
        };
        /** @enum {string} */
        ConnectionState: "connected" | "disconnected" | "reconnecting" | "waiting_for_discovery";
        ControllerNodeSummary: {
            node_id_hex: string;
            vendor_id?: number;
            vendor_name?: string;
            product_id?: number;
            product_name?: string;
            node_label?: string;
            connection: components["schemas"]["ConnectionState"];
            subscription_active: boolean;
            disabled: boolean;
            /** @description The node has ICD Management. */
            icd: boolean;
            addresses: string[];
            /** Format: date-time */
            commissioned_at?: string;
        };
        ControllerNodeList: {
            nodes: components["schemas"]["ControllerNodeSummary"][];
        };
        NodeRecord: {
            /** @description Our fabric index. */
            fabric_index: number;
            /** @description The index the node gave our fabric; absent when unknown. */
            fabric_index_on_peer?: number;
            addresses: string[];
            hostname?: string;
            /** Format: date-time */
            discovered_at?: string;
            /** Format: date-time */
            commissioned_at?: string;
            discriminator?: number;
            announced_vendor_id?: number;
            announced_product_id?: number;
            announced_device_type?: number;
            device_name?: string;
        };
        NodeProperties: {
            supports_thread: boolean;
            supports_wifi: boolean;
            supports_ethernet: boolean;
            mains_powered: boolean;
            battery_powered: boolean;
            intermittently_connected: boolean;
            long_idle_time_operating?: boolean;
            thread_sleepy_end_device?: boolean;
            idle_mode_duration_seconds?: number;
            specification_version?: number;
            serial_number?: string;
            software_version?: string;
            device_types: number[];
        };
        NodeDeviceType: {
            id: number;
            id_hex: string;
            /** @description The device type's name; empty for one the schema does not know. */
            name: string;
            revision?: number;
        };
        NodeCluster: {
            id: number;
            id_hex: string;
            /** @description The cluster's name; empty for one the schema does not know. */
            name: string;
            revision: number;
            features: string[];
        };
        NodeEndpoint: {
            endpoint: number;
            /** @description The endpoint whose parts list names this one; absent for the root and top-level endpoints. */
            parent?: number;
            name?: string;
            device_types: components["schemas"]["NodeDeviceType"][];
            clusters: components["schemas"]["NodeCluster"][];
        };
        ControllerNodeDetail: {
            node: components["schemas"]["ControllerNodeSummary"];
            record: components["schemas"]["NodeRecord"];
            properties: components["schemas"]["NodeProperties"];
            endpoints: components["schemas"]["NodeEndpoint"][];
        };
        AutoSubscribeRequest: {
            enabled: boolean;
        };
        NodeWindowRequest: {
            /** @description How long the window stays open; 0 or absent is the node's default (900 s). */
            timeout_seconds?: number;
        };
        NodeWindow: {
            manual_code: string;
            qr_code: string;
            passcode: number;
            discriminator: number;
            vendor_id: number;
            product_id: number;
            timeout_seconds: number;
        };
        AttributePath: {
            endpoint: number;
            /** @description The cluster id in hex, `0x` optional ("0x0006"). */
            cluster_hex: string;
            /** @description The attribute id in hex, `0x` optional. */
            attribute_hex: string;
        };
        ReadRequest: {
            paths: components["schemas"]["AttributePath"][];
            /** @description Read fabric-scoped lists filtered to our fabric; absent is true. */
            fabric_filtered?: boolean;
        };
        AttributeValue: {
            endpoint: number;
            cluster_hex: string;
            attribute_hex: string;
            /** @description The attribute's value as JSON; absent when the node answered a status. */
            value?: unknown;
            /** @description The node's status for the path; absent on success. */
            status?: number;
        };
        AttributeValueList: {
            values: components["schemas"]["AttributeValue"][];
        };
        WriteRequest: {
            endpoint: number;
            cluster_hex: string;
            attribute_hex: string;
            /** @description The value as JSON */
            value: unknown;
            /** @description A timed write with this timeout; absent writes untimed unless the attribute needs it. */
            timed_ms?: number;
        };
        WriteResult: {
            /** @description The node's status; 0 is success. */
            status: number;
            cluster_status?: number;
        };
        InvokeRequest: {
            endpoint: number;
            cluster_hex: string;
            command_hex: string;
            /** @description The command's fields by name; absent or null is the empty payload. */
            payload?: unknown;
            timed_ms?: number;
        };
        InvokeResult: {
            /** @description The node's status; 0 is success. */
            status: number;
            cluster_status?: number;
            /** @description The response command's id; absent without a response command. */
            response_command_hex?: string;
            /** @description The response command's fields by name; absent without one. */
            response?: unknown;
        };
        AclTarget: {
            /** @description Absent is any cluster. */
            cluster_hex?: string;
            /** @description Absent is any endpoint. */
            endpoint?: number;
            /** @description Absent is any device type. */
            device_type_hex?: string;
        };
        AclEntry: {
            /** @enum {string} */
            privilege: "view" | "proxy_view" | "operate" | "manage" | "administer";
            /** @enum {string} */
            auth_mode: "pase" | "case" | "group";
            /** @description Node ids (CASE), group ids (Group) or CATs in hex; empty is every subject of the auth mode. */
            subjects_hex: string[];
            /** @description Empty is the whole node. */
            targets: components["schemas"]["AclTarget"][];
            /** @description Set by the node; ignored in a write. */
            fabric_index?: number;
        };
        AclEntryList: {
            entries: components["schemas"]["AclEntry"][];
        };
        BindingTarget: {
            /** @description A unicast target's node; with `endpoint`. */
            node_id_hex?: string;
            /** @description A groupcast target's group id. */
            group?: number;
            endpoint?: number;
            /** @description Absent is every client cluster of the endpoint. */
            cluster_hex?: string;
            /** @description Set by the node; ignored in a write. */
            fabric_index?: number;
        };
        EndpointBindings: {
            endpoint: number;
            bindings: components["schemas"]["BindingTarget"][];
        };
        BindingList: {
            endpoints: components["schemas"]["EndpointBindings"][];
        };
        IcdState: {
            supported: boolean;
            lit_supported: boolean;
            registered: boolean;
            /** @enum {string} */
            operating_mode?: "sit" | "lit";
            awake?: boolean;
            available?: boolean;
            /** Format: date-time */
            next_check_in?: string;
            /** Format: date-time */
            last_check_in?: string;
        };
        IcdRegisterRequest: {
            /** @description Register although other administrators' vendors may not support LIT. */
            allow_multi_admin?: boolean;
        };
        IcdUnregisterRequest: {
            /** @description Forget the registration without asking the node. */
            force?: boolean;
        };
        NodeFabric: {
            fabric_index: number;
            fabric_id_hex: string;
            node_id_hex: string;
            vendor_id: number;
            vendor_name: string;
            label: string;
            /** @description This controller's fabric. */
            own: boolean;
        };
        NodeFabricList: {
            fabrics: components["schemas"]["NodeFabric"][];
        };
        DiscoveredRef: {
            instance_name: string;
            /** @description Searched for again when the controller no longer holds the discovery. */
            long_discriminator?: number;
        };
        /** @description Either `code`, or `discovered` with `passcode`. */
        CommissionRequest: {
            /** @description A QR payload ("MT:…") or a manual pairing code. */
            code?: string;
            discovered?: components["schemas"]["DiscoveredRef"];
            passcode?: number;
            /** @description The stored Wi-Fi credential to put the device on. */
            wifi_id?: string;
            /** @description The stored Thread dataset to put the device on. */
            thread_id?: string;
            /** @description Never commission over Bluetooth LE. */
            network_only?: boolean;
            /** @description The whole commissioning's bound; absent is the controller's default. */
            timeout_seconds?: number;
        };
        Discovered: {
            instance_name: string;
            hostname?: string;
            addresses: string[];
            port?: number;
            long_discriminator: number;
            vendor_id?: number;
            vendor_name?: string;
            product_id?: number;
            device_type?: number;
            device_name?: string;
            commissioning_mode: number;
            pairing_hint?: number;
            pairing_instruction?: string;
        };
        DiscoveredList: {
            devices: components["schemas"]["Discovered"][];
        };
        Credential: {
            /** @enum {string} */
            kind: "wifi" | "thread";
            id: string;
            /** @description The SSID of a Wi-Fi credential */
            name: string;
        };
        CredentialList: {
            credentials: components["schemas"]["Credential"][];
        };
        CredentialRequest: {
            /** @enum {string} */
            kind: "wifi" | "thread";
            id: string;
            /** @description Wi-Fi; required for it. */
            ssid?: string;
            /** @description Wi-Fi; write-only. */
            passphrase?: string;
            /** @description Thread; the operational dataset as hex MeshCoP TLVs */
            dataset_hex?: string;
        };
        FabricLabelRequest: {
            label: string;
        };
        PendingUpdate: {
            node_id_hex: string;
            vendor_id: number;
            product_id: number;
            target_software_version: number;
            /** @enum {string} */
            status: "queued" | "in_progress" | "stalled";
            /** @description The last status the node reported; absent before one. */
            progress_status?: string;
            /** Format: date-time */
            progress_at?: string;
            transferred?: number;
            length?: number;
        };
        UpdateQueue: {
            available: boolean;
            queue: components["schemas"]["PendingUpdate"][];
        };
        UpdateConsentRequest: {
            node_id_hex: string;
            vendor_id: number;
            product_id: number;
            target_software_version: number;
        };
        AvailableUpdate: {
            node_id_hex: string;
            vendor_id: number;
            product_id: number;
            software_version: number;
            software_version_string: string;
        };
        AvailableUpdateList: {
            updates: components["schemas"]["AvailableUpdate"][];
        };
        /** @description The payload of every controller stream event. */
        ControllerEvent: {
            /** @description The node the event is about; absent for a controller-wide one. */
            node_id_hex?: string;
            connection?: components["schemas"]["ConnectionState"];
            /** @description subscription: whether the default subscription is up. */
            active?: boolean;
            /**
             * @description commissioning_progress: the commissioning's phase.
             * @enum {string}
             */
            phase?: "started" | "succeeded" | "failed";
            /** @description commissioning_progress failed: what went wrong (no secret). */
            detail?: string;
        };
    };
    responses: {
        /** @description An RFC 9457 problem. */
        Problem: {
            headers: {
                [name: string]: unknown;
            };
            content: {
                "application/problem+json": components["schemas"]["Problem"];
            };
        };
    };
    parameters: {
        /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
        NodeID: string;
    };
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    getNodeStatus: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The status; `enabled` is false when the host serves no node. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    listFabrics: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The fabrics. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["FabricList"];
                };
            };
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    removeFabric: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                index: number;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Removed. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    getPairing: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The codes. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Pairing"];
                };
            };
            403: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    openWindow: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WindowRequest"];
            };
        };
        responses: {
            /** @description The window is open; the codes commission through it. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Window"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            409: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    closeWindow: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Closed. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    listEndpoints: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The endpoints. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["EndpointList"];
                };
            };
            503: components["responses"]["Problem"];
        };
    };
    getNodeSurface: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The report. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SurfaceReport"];
                };
            };
            503: components["responses"]["Problem"];
        };
    };
    listSessions: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The sessions. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["SessionList"];
                };
            };
            503: components["responses"]["Problem"];
        };
    };
    getMdns: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The advertisement. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Mdns"];
                };
            };
            503: components["responses"]["Problem"];
        };
    };
    getCompatibility: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The report. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Compatibility"];
                };
            };
            503: components["responses"]["Problem"];
        };
    };
    listDiagnosticEvents: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The trace. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DiagnosticEventList"];
                };
            };
            503: components["responses"]["Problem"];
        };
    };
    getExposureSchema: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The schema. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ExposureSchema"];
                };
            };
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    listExposure: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The items. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ItemList"];
                };
            };
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    applyExposure: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateList"];
            };
        };
        responses: {
            /** @description Applied. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Applied"];
                };
            };
            /** @description The batch was empty. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    previewExposure: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateList"];
            };
        };
        responses: {
            /** @description The touched items after the changes. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ItemList"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    reassemble: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Reassembled. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    factoryReset: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["FactoryResetRequest"];
            };
        };
        responses: {
            /** @description Every fabric is removed. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    stream: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The event stream. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "text/event-stream": string;
                };
            };
        };
    };
    getControllerStatus: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The status. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ControllerStatus"];
                };
            };
        };
    };
    listControllerNodes: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The nodes, by node id. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ControllerNodeList"];
                };
            };
        };
    };
    getControllerNode: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The node. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ControllerNodeDetail"];
                };
            };
            400: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
        };
    };
    decommissionNode: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The node is decommissioned. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    interviewNode: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The node was read. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    enableNode: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The node is enabled. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
        };
    };
    disableNode: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The node is disabled. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
        };
    };
    setNodeAutoSubscribe: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AutoSubscribeRequest"];
            };
        };
        responses: {
            /** @description The setting is applied. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
        };
    };
    openNodeWindow: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["NodeWindowRequest"];
            };
        };
        responses: {
            /** @description The window is open. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NodeWindow"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    closeNodeWindow: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description No window is open any more. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    writeNodeAttribute: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["WriteRequest"];
            };
        };
        responses: {
            /** @description The node's status for the write; a refusal is a status, not a problem. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WriteResult"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    readNodeAttributes: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["ReadRequest"];
            };
        };
        responses: {
            /** @description One entry per path the node answered. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AttributeValueList"];
                };
            };
            400: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    invokeNodeCommand: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["InvokeRequest"];
            };
        };
        responses: {
            /** @description The node's status and, when it answered with one, its response's fields. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["InvokeResult"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    getNodeAcl: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The entries. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AclEntryList"];
                };
            };
            400: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    setNodeAcl: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["AclEntryList"];
            };
        };
        responses: {
            /** @description The node's status for the write. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WriteResult"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    getNodeBindings: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The bindings by endpoint. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["BindingList"];
                };
            };
            400: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    setNodeBindings: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["EndpointBindings"];
            };
        };
        responses: {
            /** @description The node's status for the write. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["WriteResult"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    getNodeIcd: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The state; `supported` is false for a node without ICD Management. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["IcdState"];
                };
            };
            400: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
        };
    };
    registerNodeIcd: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["IcdRegisterRequest"];
            };
        };
        responses: {
            /** @description The state after the registration. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["IcdState"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            409: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    unregisterNodeIcd: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["IcdUnregisterRequest"];
            };
        };
        responses: {
            /** @description The state after the unregistration. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["IcdState"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    listNodeFabrics: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The fabrics. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["NodeFabricList"];
                };
            };
            400: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    removeNodeFabric: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description The node id as 16-digit hex (`node_id_hex`); case and leading zeros are free. */
                id: components["parameters"]["NodeID"];
                index: number;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The fabric is removed. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
        };
    };
    commissionNode: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CommissionRequest"];
            };
        };
        responses: {
            /** @description The commissioned node. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ControllerNodeDetail"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            502: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    discoverDevices: {
        parameters: {
            query?: {
                /** @description What `value` selects by; absent is `any`. */
                filter?: "any" | "long_discriminator" | "short_discriminator" | "vendor" | "device_type";
                /** @description The discriminator, vendor id or device type; required with every filter but `any`. */
                value?: number;
                /** @description How long to listen, in seconds; absent is 3. */
                timeout?: number;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Every device that answered, each once. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DiscoveredList"];
                };
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
        };
    };
    listCredentials: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The credentials. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CredentialList"];
                };
            };
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    setCredential: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CredentialRequest"];
            };
        };
        responses: {
            /** @description Stored. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    removeCredential: {
        parameters: {
            query: {
                kind: "wifi" | "thread";
                id: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Removed. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    setFabricLabel: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["FabricLabelRequest"];
            };
        };
        responses: {
            /** @description The label is set; the nodes get it in the background. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
        };
    };
    listUpdates: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The queue; `available` is false when the controller runs no OTA provider. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["UpdateQueue"];
                };
            };
        };
    };
    consentUpdate: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["UpdateConsentRequest"];
            };
        };
        responses: {
            /** @description The consent is recorded. */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            400: components["responses"]["Problem"];
            403: components["responses"]["Problem"];
            404: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    checkUpdates: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The updates the nodes can take. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["AvailableUpdateList"];
                };
            };
            403: components["responses"]["Problem"];
            500: components["responses"]["Problem"];
            503: components["responses"]["Problem"];
        };
    };
    controllerStream: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description The event stream. */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "text/event-stream": string;
                };
            };
        };
    };
}
