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
         *     window_closed, endpoints_changed, exposure_changed; `data:` is the
         *     JSON payload (`FabricEvent`, `WindowEvent`, `{}`, `ExposureEvent`).
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
    "/controller/commission": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Commission a device from a code or QR payload, with progress. */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Planned. */
                501: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
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
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Planned. */
                501: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
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
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Planned. */
                501: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
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
            path?: never;
            cookie?: never;
        };
        /** One node — endpoints, attributes, subscriptions, bindings, ACL, window, OTA, Thread dataset. */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path: {
                    id: string;
                };
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Planned. */
                501: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
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
        /** The controller's fabric label. */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Planned. */
                501: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content?: never;
                };
            };
        };
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
            code: "bad_request" | "forbidden" | "conflict" | "not_found" | "service_unready" | "internal" | "payload_too_large";
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
            /** @description Default false. The field is a comma-separated list: shown as chips, filtered by membership, the filter's options built from the individual values. */
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
            /** @description Default false. Match by membership in the field's comma-separated list (values trimmed; an empty value never matches) instead of equality. */
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
    parameters: never;
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
}
