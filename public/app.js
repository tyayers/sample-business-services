(() => {
  const state = {
    spec: null,
    domains: [],       // auto-discovered from OpenAPI collection GET endpoints
    activeDomain: null,
    protocol: "REST",  // "REST" | "MCP"
    mcpSessionId: null,
    mcpTools: [],
    mcpResources: [],
    selectedOasOp: null,
    selectedMcpItem: null,
  };

  const $ = (sel) => document.querySelector(sel);
  const el = (tag, attrs = {}, text = "") => {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
      if (k === "class") node.className = v;
      else if (k === "dataset") Object.assign(node.dataset, v);
      else node.setAttribute(k, v);
    }
    if (text !== "") node.textContent = text;
    return node;
  };

  const camelToSnake = (str) =>
    str.replace(/[A-Z]/g, (m) => "_" + m.toLowerCase());

  const origin = () => window.location.origin;

  // --- MCP Client Helper ---
  async function mcpCall(method, params = undefined) {
    const headers = { "Content-Type": "application/json" };
    if (state.mcpSessionId) {
      headers["Mcp-Session-Id"] = state.mcpSessionId;
    }
    const payload = { jsonrpc: "2.0", id: Date.now(), method };
    if (params !== undefined) payload.params = params;

    const res = await fetch("/mcp", {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
    });
    const sid = res.headers.get("Mcp-Session-Id");
    if (sid) state.mcpSessionId = sid;
    return { status: res.status, data: await res.json(), payload };
  }

  async function initMCP() {
    await mcpCall("initialize", {
      protocolVersion: "2025-03-26",
      clientInfo: { name: "web-explorer", version: "1.0.0" },
    });
    $("#mcp-session-id").textContent = state.mcpSessionId || "stateless";

    const [toolsRes, resRes] = await Promise.all([
      mcpCall("tools/list"),
      mcpCall("resources/list"),
    ]);
    state.mcpTools = toolsRes.data?.result?.tools || [];
    state.mcpResources = resRes.data?.result?.resources || [];
    renderMCPSidebar();
  }

  // --- KPIs ---
  async function loadKPIs() {
    const res = await fetch("/api/v1/summary");
    const kpi = await res.json();
    const grid = $("#kpi-grid");
    grid.replaceChildren();

    const cards = [
      ["Customers", kpi.totalCustomers, `${kpi.activeCustomers} active`],
      ["Materials", kpi.totalMaterials, `${kpi.lowStockMaterials} low stock`],
      ["Vendors", kpi.totalVendors, "Suppliers"],
      ["Sales Orders", kpi.totalSalesOrders, `${kpi.openSalesOrders} open/confirmed`],
      ["Invoices", kpi.totalInvoices, `${kpi.openInvoices} open, ${kpi.overdueInvoices} overdue`],
      ["Outstanding AR", `$${Number(kpi.outstandingArAmount || 0).toLocaleString()}`, `Total billed $${Number(kpi.totalBilledAmount || 0).toLocaleString()}`],
      ["Purchase Orders", kpi.totalPurchaseOrders, `${kpi.openPurchaseOrders} open ($${Number(kpi.openPoCommittedSpend || 0).toLocaleString()})`],
    ];

    for (const [label, val, sub] of cards) {
      const card = el("div", { class: "kpi-card" });
      card.append(
        el("span", { class: "kpi-label" }, label),
        el("span", { class: "kpi-value" }, String(val)),
        el("span", { class: "kpi-sub" }, sub)
      );
      grid.append(card);
    }
  }

  // --- OpenAPI Discovery ---
  function discoverDomainsFromOAS(spec) {
    const domains = [];
    for (const [path, methods] of Object.entries(spec.paths || {})) {
      const getOp = methods.get;
      if (!getOp) continue;
      if (path.includes("{") || path === "/api/v1/summary") continue;

      const tag = (getOp.tags && getOp.tags[0]) || path;
      const opId = getOp.operationId || "";
      const mcpListTool = camelToSnake(opId);
      const itemPath = path + "/{id}";
      const getByIdOp = spec.paths[itemPath]?.get;
      const mcpGetTool = getByIdOp?.operationId ? camelToSnake(getByIdOp.operationId) : null;

      domains.push({
        tag,
        path,
        itemPath: getByIdOp ? itemPath : null,
        operation: getOp,
        parameters: getOp.parameters || [],
        mcpListTool,
        mcpGetTool,
      });
    }
    return domains;
  }

  // --- Mode Switching ---
  function initModeTabs() {
    document.querySelectorAll(".mode-tab").forEach((btn) => {
      btn.addEventListener("click", () => {
        const mode = btn.dataset.mode;
        document.querySelectorAll(".mode-tab").forEach((b) => {
          const active = b === btn;
          b.classList.toggle("active", active);
          b.setAttribute("aria-selected", String(active));
        });
        document.querySelectorAll(".mode-panel").forEach((p) => {
          p.classList.toggle("hidden", p.id !== `panel-${mode}`);
        });
      });
    });
  }

  // --- Mode 1: Domain Data Explorer ---
  function renderDomainTabs() {
    const container = $("#domain-tabs");
    container.replaceChildren();
    state.domains.forEach((d, idx) => {
      const btn = el(
        "button",
        {
          type: "button",
          class: "domain-tab" + (idx === 0 ? " active" : ""),
          role: "tab",
          "aria-selected": String(idx === 0),
        },
        d.tag
      );
      btn.addEventListener("click", () => selectDomain(d));
      container.append(btn);
    });
    if (state.domains.length > 0) {
      selectDomain(state.domains[0]);
    }
  }

  function selectDomain(domainObj) {
    state.activeDomain = domainObj;
    document.querySelectorAll(".domain-tab").forEach((b) => {
      const isMatch = b.textContent === domainObj.tag;
      b.classList.toggle("active", isMatch);
      b.setAttribute("aria-selected", String(isMatch));
    });
    renderDynamicFilters(domainObj.parameters);
    runExplorerQuery();
  }

  function renderDynamicFilters(params) {
    const wrap = $("#dynamic-filters");
    wrap.replaceChildren();

    if (params.length === 0) {
      wrap.append(el("span", { class: "muted" }, "No query filters defined in OpenAPI spec."));
      return;
    }

    for (const p of params) {
      const field = el("div", { class: "field" });
      const inputId = `filter-${p.name}`;
      field.append(el("label", { for: inputId }, p.name));

      const schema = p.schema || {};
      if (schema.enum) {
        const select = el("select", { id: inputId, name: p.name });
        select.append(el("option", { value: "" }, "All"));
        for (const opt of schema.enum) {
          select.append(el("option", { value: opt }, opt));
        }
        field.append(select);
      } else if (schema.type === "boolean") {
        const select = el("select", { id: inputId, name: p.name });
        select.append(
          el("option", { value: "" }, "Any"),
          el("option", { value: "true" }, "true"),
          el("option", { value: "false" }, "false")
        );
        field.append(select);
      } else {
        const input = el("input", {
          id: inputId,
          name: p.name,
          type: p.name === "q" ? "search" : "text",
          placeholder: p.description || `Filter by ${p.name}...`,
        });
        field.append(input);
      }
      wrap.append(field);
    }
  }

  function collectFilterValues() {
    const vals = {};
    const form = $("#explorer-filter-form");
    new FormData(form).forEach((v, k) => {
      const trimmed = String(v).trim();
      if (trimmed !== "") vals[k] = trimmed;
    });
    return vals;
  }

  async function runExplorerQuery() {
    const dom = state.activeDomain;
    if (!dom) return;

    const filters = collectFilterValues();
    const start = performance.now();

    let items = [];
    let rawPayload = null;
    let statusLabel = "";
    let curlCmd = "";

    if (state.protocol === "REST") {
      const qs = new URLSearchParams(filters).toString();
      const url = qs ? `${dom.path}?${qs}` : dom.path;
      curlCmd = `curl -s "${origin()}${url}"`;

      const res = await fetch(url);
      rawPayload = await res.json();
      items = rawPayload.items || [];
      statusLabel = `REST GET ${url} — HTTP ${res.status} (${Math.round(performance.now() - start)} ms)`;
    } else {
      // Convert boolean strings for MCP JSON schema
      const mcpArgs = {};
      for (const [k, v] of Object.entries(filters)) {
        const paramDef = dom.parameters.find((p) => p.name === k);
        if (paramDef?.schema?.type === "boolean") {
          mcpArgs[k] = v === "true";
        } else {
          mcpArgs[k] = v;
        }
      }

      const rpcBody = {
        jsonrpc: "2.0",
        id: 1,
        method: "tools/call",
        params: { name: dom.mcpListTool, arguments: mcpArgs },
      };
      curlCmd = `curl -s -X POST "${origin()}/mcp" \\\n  -H "Content-Type: application/json" \\\n  -d '${JSON.stringify(rpcBody)}'`;

      const { status, data } = await mcpCall("tools/call", {
        name: dom.mcpListTool,
        arguments: mcpArgs,
      });
      rawPayload = data;
      items = data?.result?.structuredContent?.items || [];
      statusLabel = `MCP tools/call (${dom.mcpListTool}) — HTTP ${status} (${Math.round(performance.now() - start)} ms)`;
    }

    $("#explorer-call-meta").textContent = statusLabel;
    $("#explorer-curl-code").textContent = curlCmd;
    $("#explorer-results-title").textContent = `${dom.tag} (${state.protocol})`;
    $("#explorer-count-badge").textContent = `${items.length} items`;
    $("#explorer-raw-json code").textContent = JSON.stringify(rawPayload, null, 2);

    renderDataTable(items, dom);
  }

  function renderDataTable(items, dom) {
    const wrap = $("#explorer-table-wrap");
    wrap.replaceChildren();

    if (!items.length) {
      wrap.append(el("div", { class: "empty-state" }, "No matching records found."));
      return;
    }

    // Pick scalar columns cleanly, keeping tables readable
    const allKeys = Object.keys(items[0]).filter((k) => {
      const sample = items[0][k];
      return typeof sample !== "object" || sample === null;
    });
    const cols = allKeys.slice(0, 8);

    const table = el("table", { class: "data-table" });
    const thead = el("thead");
    const headRow = el("tr");
    for (const c of cols) {
      headRow.append(el("th", {}, c));
    }
    thead.append(headRow);

    const tbody = el("tbody");
    for (const row of items) {
      const tr = el("tr");
      for (const c of cols) {
        const td = el("td");
        const val = row[c];
        if (c === "status" && typeof val === "string") {
          const cls =
            val === "ACTIVE" || val === "PAID" || val === "RECEIVED"
              ? "badge-ok"
              : val === "BLOCKED" || val === "OVERDUE" || val === "CANCELLED" || val === "ON_HOLD"
              ? "badge-err"
              : "badge-warn";
          td.append(el("span", { class: `badge ${cls}` }, val));
        } else {
          td.textContent = val === undefined || val === null ? "—" : String(val);
        }
        tr.append(td);
      }
      tr.addEventListener("click", () => openRecordDetail(row, dom));
      tbody.append(tr);
    }

    table.append(thead, tbody);
    wrap.append(table);
  }

  async function openRecordDetail(row, dom) {
    const dialog = $("#record-dialog");
    $("#dialog-title").textContent = `${dom.tag} — ${row.id || "Record"}`;
    let detail = row;
    let source = `Loaded via ${state.protocol}`;

    if (row.id) {
      if (state.protocol === "REST" && dom.itemPath) {
        const url = dom.itemPath.replace("{id}", encodeURIComponent(row.id));
        const res = await fetch(url);
        if (res.ok) {
          detail = await res.json();
          source = `GET ${url}`;
        }
      } else if (state.protocol === "MCP" && dom.mcpGetTool) {
        const { data } = await mcpCall("tools/call", {
          name: dom.mcpGetTool,
          arguments: { id: row.id },
        });
        if (data?.result?.structuredContent) {
          detail = data.result.structuredContent;
          source = `MCP tools/call ${dom.mcpGetTool}({ id: "${row.id}" })`;
        }
      }
    }

    $("#dialog-meta").textContent = source;
    $("#dialog-json").textContent = JSON.stringify(detail, null, 2);
    dialog.showModal();
  }

  // --- Mode 2: REST / OpenAPI Playground ---
  function renderOASSidebar() {
    const list = $("#oas-endpoint-list");
    list.replaceChildren();

    for (const [path, methods] of Object.entries(state.spec.paths || {})) {
      for (const [method, op] of Object.entries(methods)) {
        const upper = method.toUpperCase();
        const btn = el("button", { type: "button", class: "endpoint-item" });
        btn.append(
          el("span", { class: `method-badge method-${upper}` }, upper),
          el("span", {}, path)
        );
        btn.addEventListener("click", () => selectOasOperation(btn, upper, path, op));
        list.append(btn);
      }
    }
  }

  function resolveSchema(refOrSchema) {
    if (!refOrSchema) return {};
    if (refOrSchema.$ref) {
      const name = refOrSchema.$ref.split("/").pop();
      return state.spec.components?.schemas?.[name] || {};
    }
    return refOrSchema;
  }

  function buildSampleFromSchema(schema) {
    const s = resolveSchema(schema);
    if (s.example !== undefined) return s.example;
    if (s.default !== undefined) return s.default;
    if (s.type === "object" && s.properties) {
      const obj = {};
      for (const [k, prop] of Object.entries(s.properties)) {
        obj[k] = buildSampleFromSchema(prop);
      }
      return obj;
    }
    if (s.type === "array" && s.items) {
      return [buildSampleFromSchema(s.items)];
    }
    if (s.type === "integer" || s.type === "number") return 1;
    if (s.type === "boolean") return false;
    if (s.enum && s.enum.length > 0) return s.enum[0];
    return "";
  }

  function selectOasOperation(btn, method, path, op) {
    document.querySelectorAll("#oas-endpoint-list .endpoint-item").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    state.selectedOasOp = { method, path, op };

    $("#oas-empty-state").classList.add("hidden");
    $("#oas-runner").classList.remove("hidden");

    const badge = $("#oas-method-badge");
    badge.className = `method-badge method-${method}`;
    badge.textContent = method;
    $("#oas-path-display").textContent = path;
    $("#oas-summary-text").textContent = op.summary || "";

    const paramsWrap = $("#oas-params-fields");
    paramsWrap.replaceChildren();
    for (const p of op.parameters || []) {
      const field = el("div", { class: "field" });
      const id = `oas-param-${p.name}`;
      field.append(el("label", { for: id }, `${p.name} (${p.in}${p.required ? ", required" : ""})`));
      const input = el("input", {
        id,
        name: p.name,
        dataset: { in: p.in },
        value: p.example || "",
        placeholder: p.description || "",
      });
      input.addEventListener("input", updateOasCurl);
      field.append(input);
      paramsWrap.append(field);
    }

    const bodyGroup = $("#oas-body-group");
    const bodySchema = op.requestBody?.content?.["application/json"]?.schema;
    if (bodySchema) {
      bodyGroup.classList.remove("hidden");
      $("#oas-body-input").value = JSON.stringify(buildSampleFromSchema(bodySchema), null, 2);
    } else {
      bodyGroup.classList.add("hidden");
      $("#oas-body-input").value = "";
    }

    updateOasCurl();
  }

  function buildOasRequestInfo() {
    const { method, path } = state.selectedOasOp;
    let resolvedPath = path;
    const query = new URLSearchParams();

    document.querySelectorAll("#oas-params-fields input").forEach((inp) => {
      const val = inp.value.trim();
      if (inp.dataset.in === "path") {
        resolvedPath = resolvedPath.replace(`{${inp.name}}`, encodeURIComponent(val || `{${inp.name}}`));
      } else if (inp.dataset.in === "query" && val !== "") {
        query.set(inp.name, val);
      }
    });

    const qs = query.toString();
    const url = qs ? `${resolvedPath}?${qs}` : resolvedPath;
    const bodyText = $("#oas-body-input").value.trim();

    let curl = `curl -s -X ${method} "${origin()}${url}"`;
    if (bodyText) {
      curl += ` \\\n  -H "Content-Type: application/json" \\\n  -d '${bodyText.replace(/\n\s*/g, " ")}'`;
    }
    return { method, url, bodyText, curl };
  }

  function updateOasCurl() {
    if (!state.selectedOasOp) return;
    $("#oas-curl-code").textContent = buildOasRequestInfo().curl;
  }

  async function executeOasCall(e) {
    e.preventDefault();
    const { method, url, bodyText } = buildOasRequestInfo();
    const opts = { method, headers: {} };
    if (bodyText) {
      opts.headers["Content-Type"] = "application/json";
      opts.body = bodyText;
    }
    const res = await fetch(url, opts);
    const text = await res.text();
    const badge = $("#oas-status-badge");
    badge.textContent = `HTTP ${res.status}`;
    badge.className = `badge ${res.ok ? "badge-ok" : "badge-err"}`;
    try {
      $("#oas-response-code").textContent = JSON.stringify(JSON.parse(text), null, 2);
    } catch {
      $("#oas-response-code").textContent = text;
    }
    loadKPIs();
  }

  // --- Mode 3: MCP Explorer ---
  function renderMCPSidebar() {
    const toolList = $("#mcp-tool-list");
    toolList.replaceChildren();
    for (const tool of state.mcpTools) {
      const btn = el("button", { type: "button", class: "endpoint-item" });
      btn.append(el("span", { class: "method-badge method-POST" }, "TOOL"), el("span", {}, tool.name));
      btn.addEventListener("click", () => selectMcpTool(btn, tool));
      toolList.append(btn);
    }

    const resList = $("#mcp-resource-list");
    resList.replaceChildren();
    for (const r of state.mcpResources) {
      const btn = el("button", { type: "button", class: "endpoint-item" });
      btn.append(el("span", { class: "method-badge method-GET" }, "RES"), el("span", {}, r.uri));
      btn.addEventListener("click", () => selectMcpResource(btn, r));
      resList.append(btn);
    }
  }

  function selectMcpTool(btn, tool) {
    document.querySelectorAll("#panel-mcp .endpoint-item").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    state.selectedMcpItem = { kind: "tool", tool };

    $("#mcp-empty-state").classList.add("hidden");
    $("#mcp-runner").classList.remove("hidden");
    $("#mcp-tool-name").textContent = tool.name;
    $("#mcp-tool-desc").textContent = tool.description || "";
    $("#mcp-schema-details").classList.remove("hidden");
    $("#mcp-input-schema").textContent = JSON.stringify(tool.inputSchema || {}, null, 2);
    $("#mcp-output-schema").textContent = JSON.stringify(tool.outputSchema || {}, null, 2);

    const argsWrap = $("#mcp-args-fields");
    argsWrap.replaceChildren();
    const props = tool.inputSchema?.properties || {};
    const required = new Set(tool.inputSchema?.required || []);

    if (Object.keys(props).length === 0) {
      argsWrap.append(el("p", { class: "muted" }, "This tool takes no arguments."));
    } else {
      for (const [k, prop] of Object.entries(props)) {
        const field = el("div", { class: "field" });
        const id = `mcp-arg-${k}`;
        field.append(el("label", { for: id }, `${k} (${prop.type || "any"}${required.has(k) ? ", required" : ""})`));
        if (prop.enum) {
          const sel = el("select", { id, name: k, dataset: { type: prop.type || "string" } });
          sel.append(el("option", { value: "" }, "(unset)"));
          for (const opt of prop.enum) sel.append(el("option", { value: opt }, opt));
          sel.addEventListener("change", updateMcpCurl);
          field.append(sel);
        } else if (prop.type === "array" || prop.type === "object") {
          const ta = el("textarea", {
            id,
            name: k,
            rows: "3",
            dataset: { type: "json" },
            placeholder: prop.type === "array" ? '[{"materialId": "MAT-1001", "quantity": 1}]' : "{}",
          });
          ta.addEventListener("input", updateMcpCurl);
          field.append(ta);
        } else {
          const inp = el("input", {
            id,
            name: k,
            dataset: { type: prop.type || "string" },
            placeholder: prop.description || "",
          });
          inp.addEventListener("input", updateMcpCurl);
          field.append(inp);
        }
        argsWrap.append(field);
      }
    }
    updateMcpCurl();
  }

  function selectMcpResource(btn, resource) {
    document.querySelectorAll("#panel-mcp .endpoint-item").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    state.selectedMcpItem = { kind: "resource", resource };

    $("#mcp-empty-state").classList.add("hidden");
    $("#mcp-runner").classList.remove("hidden");
    $("#mcp-tool-name").textContent = resource.uri;
    $("#mcp-tool-desc").textContent = resource.description || "";
    $("#mcp-schema-details").classList.add("hidden");
    $("#mcp-args-fields").replaceChildren(el("p", { class: "muted" }, "Reads resource via resources/read."));
    updateMcpCurl();
  }

  function buildMcpArgs() {
    const args = {};
    document.querySelectorAll("#mcp-args-fields input, #mcp-args-fields select, #mcp-args-fields textarea").forEach((inp) => {
      const val = inp.value.trim();
      if (val === "") return;
      const t = inp.dataset.type;
      if (t === "integer" || t === "number") args[inp.name] = Number(val);
      else if (t === "boolean") args[inp.name] = val === "true";
      else if (t === "json") {
        try {
          args[inp.name] = JSON.parse(val);
        } catch {
          args[inp.name] = val;
        }
      } else {
        args[inp.name] = val;
      }
    });
    return args;
  }

  function buildMcpPayload() {
    if (state.selectedMcpItem.kind === "tool") {
      return {
        method: "tools/call",
        params: { name: state.selectedMcpItem.tool.name, arguments: buildMcpArgs() },
      };
    }
    return {
      method: "resources/read",
      params: { uri: state.selectedMcpItem.resource.uri },
    };
  }

  function updateMcpCurl() {
    if (!state.selectedMcpItem) return;
    const { method, params } = buildMcpPayload();
    const rpc = { jsonrpc: "2.0", id: 1, method, params };
    $("#mcp-curl-code").textContent =
      `curl -s -X POST "${origin()}/mcp" \\\n  -H "Content-Type: application/json" \\\n  -d '${JSON.stringify(rpc)}'`;
  }

  async function executeMcpCall(e) {
    e.preventDefault();
    const { method, params } = buildMcpPayload();
    const { status, data } = await mcpCall(method, params);
    const badge = $("#mcp-status-badge");
    const isErr = Boolean(data.error || data.result?.isError);
    badge.textContent = isErr ? `Error (HTTP ${status})` : `OK (HTTP ${status})`;
    badge.className = `badge ${isErr ? "badge-err" : "badge-ok"}`;
    $("#mcp-response-code").textContent = JSON.stringify(data, null, 2);
    loadKPIs();
  }

  // --- Copy Helper ---
  function bindCopy(btnId, codeId) {
    $(btnId).addEventListener("click", async () => {
      const text = $(codeId).textContent;
      await navigator.clipboard.writeText(text);
      const btn = $(btnId);
      const orig = btn.textContent;
      btn.textContent = "Copied!";
      setTimeout(() => (btn.textContent = orig), 1200);
    });
  }

  // --- Theme Toggle ---
  function initThemeToggle() {
    const modes = ["auto", "light", "dark"];
    const labels = { auto: "Theme: Auto", light: "Theme: Light", dark: "Theme: Dark" };
    const btn = $("#theme-toggle-btn");
    let current = localStorage.getItem("theme-mode") || "auto";

    const apply = (mode) => {
      document.documentElement.style.colorScheme = mode === "auto" ? "light dark" : mode;
      btn.textContent = labels[mode] || labels.auto;
    };

    apply(current);
    btn.addEventListener("click", () => {
      current = modes[(modes.indexOf(current) + 1) % modes.length];
      localStorage.setItem("theme-mode", current);
      apply(current);
    });
  }

  // --- Boot ---
  async function boot() {
    initThemeToggle();
    initModeTabs();

    $("#refresh-kpi-btn").addEventListener("click", loadKPIs);
    $("#explorer-filter-form").addEventListener("submit", (e) => {
      e.preventDefault();
      runExplorerQuery();
    });
    $("#reset-filters-btn").addEventListener("click", () => {
      $("#explorer-filter-form").reset();
      runExplorerQuery();
    });
    $("#toggle-raw-json").addEventListener("click", () => {
      $("#explorer-raw-json").classList.toggle("hidden");
    });
    $("#close-dialog-btn").addEventListener("click", () => $("#record-dialog").close());

    document.querySelectorAll(".proto-btn").forEach((btn) => {
      btn.addEventListener("click", () => {
        state.protocol = btn.dataset.protocol;
        document.querySelectorAll(".proto-btn").forEach((b) => b.classList.toggle("active", b === btn));
        runExplorerQuery();
      });
    });

    $("#oas-form").addEventListener("submit", executeOasCall);
    $("#oas-body-input").addEventListener("input", updateOasCurl);
    $("#mcp-form").addEventListener("submit", executeMcpCall);

    bindCopy("#copy-explorer-curl", "#explorer-curl-code");
    bindCopy("#copy-oas-curl", "#oas-curl-code");
    bindCopy("#copy-mcp-curl", "#mcp-curl-code");

    const oasRes = await fetch("/openapi.json");
    state.spec = await oasRes.json();
    state.domains = discoverDomainsFromOAS(state.spec);

    await Promise.all([loadKPIs(), initMCP()]);
    renderDomainTabs();
    renderOASSidebar();
  }

  boot();
})();
