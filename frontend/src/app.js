// ==================================================
//  Zompy frontend ↔ Go bridge
// ==================================================

let state = {
  page: "home",
  builds: [],
  versions: [],
  servers: [],
  cfg: null,
  builder: {
    step: 0,
    name: "", desc: "", icon: "⛏️", color: "t1",
    version: "1.21.4", loader: "fabric", ram: 6,
    selectedMods: new Set(),
  },
};

// ===== INIT =====
window.addEventListener("DOMContentLoaded", async () => {
  try {
    state.cfg      = await window.go.main.App.GetConfig();
    state.builds   = await window.go.main.App.GetBuilds();
    state.servers  = await window.go.main.App.GetServers();
    state.versions = await window.go.main.App.GetVersions();
    renderHome();
    renderServersPage();
    toast("info", "Добро пожаловать в Zompy", "Ctrl+K — быстрый поиск");
  } catch (e) {
    console.error(e);
    toast("danger", "Ошибка инициализации", String(e));
  }
});

// ===== NAVIGATION =====
function goTo(page) {
  state.page = page;
  document.querySelectorAll(".page").forEach(p => p.classList.remove("active"));
  const el = document.getElementById("page-" + page);
  if (el) el.classList.add("active");
  document.querySelectorAll(".nav-item[data-page]").forEach(n => {
    n.classList.toggle("active", n.dataset.page === page);
  });
  document.querySelector("main").scrollTo({ top: 0, behavior: "smooth" });

  if (page === "builds")        renderBuilds();
  if (page === "versions")      renderVersions();
  if (page === "servers")       renderServersPage();
  if (page === "home")          renderHome();
}

// ===== HOME =====
function renderHome() {
  document.getElementById("recentGrid").innerHTML =
    state.builds.slice(0, 4).map((b, i) => cardHTML(b, i)).join("");
}

function cardHTML(b, i) {
  const loaderLabel = {forge:"Forge",fabric:"Fabric",neoforge:"NeoForge",vanilla:"Vanilla"}[b.loader] || b.loader;
  const loaderTag = {forge:"green",fabric:"rose",neoforge:"amber",vanilla:"blue"}[b.loader] || "gray";
  return `
    <div class="card" style="animation-delay:${i*45}ms" data-id="${b.id}"
         onclick="launchBuild('${b.id}')">
      <div class="card-top">
        <div class="thumb ${b.thumb || 't1'}">${b.emoji || '⛏️'}</div>
        <button class="fav-btn ${b.fav?'active':''}"
                onclick="event.stopPropagation();toggleFav('${b.id}',this)">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m12 2 3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
          </svg>
        </button>
      </div>
      <h4><span class="name">${b.name}</span></h4>
      <div class="tags">
        <span class="tag ${loaderTag}">${loaderLabel}</span>
        <span class="tag gray">${b.version}</span>
        ${b.tag ? `<span class="tag mint">${b.tag}</span>` : ""}
      </div>
      <div class="desc">${b.desc || ""}</div>
      <div class="footer">
        <span class="stat">${(b.mods||[]).length} модов</span>
        <button class="play-mini"
                onclick="event.stopPropagation();launchBuild('${b.id}')">Играть</button>
      </div>
    </div>`;
}

async function toggleFav(id, btn) {
  const b = state.builds.find(x => x.id === id);
  if (!b) return;
  b.fav = !b.fav;
  btn.classList.toggle("active", b.fav);
  await window.go.main.App.SaveBuild(b);
}

// ===== LAUNCH =====
async function launchBuild(id) {
  try {
    document.getElementById("launchOverlay").classList.add("open");
    document.getElementById("launchTitle").textContent = "Запуск...";
    await window.go.main.App.Launch(id);
    toast("success", "Игра запущена", "Minecraft стартует");
    setTimeout(() => {
      document.getElementById("launchOverlay").classList.remove("open");
    }, 1500);
  } catch (e) {
    document.getElementById("launchOverlay").classList.remove("open");
    toast("danger", "Ошибка запуска", String(e));
  }
}

// ===== SERVERS =====
function renderServersPage() {
  const grid = document.getElementById("serversGrid");
  grid.innerHTML = state.servers.map((s, i) => `
    <div class="server-card" style="animation-delay:${i*30}ms">
      <div class="server-banner" style="background:linear-gradient(135deg,hsl(${(s.id*37)%360} 40% 22%),hsl(${(s.id*37)%360} 45% 12%))">
        <div class="srv-logo">${s.emoji}</div>
      </div>
      <div class="srv-info">
        <div class="srv-name">${s.name}</div>
        <button class="srv-ip" onclick="event.stopPropagation();copyIP('${s.ip}')">
          ${s.ip}
        </button>
      </div>
      <div class="srv-tags">
        <span class="tag blue">${s.mode}</span>
        <span class="tag gray">${s.version}</span>
        <span class="tag ${s.auth==='offline'?'green':s.auth==='online'?'rose':'blue'}">${s.auth}</span>
      </div>
      <div class="srv-desc">${s.desc}</div>
      <button class="btn btn-primary btn-sm" style="width:100%;margin-top:12px"
              onclick="event.stopPropagation();pingServer('${s.ip}',${s.port||25565})">
        Проверить онлайн
      </button>
    </div>`).join("");
}

async function pingServer(host, port) {
  toast("info", "Пинг...", host);
  const r = await window.go.main.App.PingServer(host, port);
  if (r.ok) toast("success", `${host} — ${r.online}/${r.max}`, `${r.latency} мс`);
  else      toast("danger", "Не отвечает", r.error || host);
}

function copyIP(ip) {
  navigator.clipboard.writeText(ip);
  toast("success", "IP скопирован", ip);
}

// ===== VERSIONS =====
function renderVersions() {
  const grid = document.getElementById("versionGrid");
  if (!grid) return;
  grid.innerHTML = state.versions.slice(0, 40).map((v, i) => `
    <div class="card" style="animation-delay:${i*20}ms;padding:16px">
      <div style="display:flex;align-items:center;gap:12px">
        <div style="width:44px;height:44px;border-radius:11px;display:grid;place-items:center;
                    background:var(--surface-2);border:1px solid var(--line-2);color:var(--accent)">
          ${v.id.split(".").slice(0,2).join(".")}
        </div>
        <div style="flex:1">
          <div style="font-weight:600">${v.id}</div>
          <div style="font-size:11.5px;color:var(--txt-3)">${v.type} · ${v.time?.slice(0,10) || ""}</div>
        </div>
        <button class="btn ${v.installed?'btn-ghost':'btn-primary'} btn-sm"
                onclick="installVer('${v.id}')">
          ${v.installed ? "Играть" : "Установить"}
        </button>
      </div>
    </div>`).join("");
}

async function installVer(id) {
  try {
    await window.go.main.App.InstallVersion(id);
    toast("success", "Установлено", id);
    state.versions = await window.go.main.App.GetVersions();
    renderVersions();
  } catch (e) {
    toast("danger", "Ошибка", String(e));
  }
}

// ===== BUILDS =====
function renderBuilds() {
  const grid = document.getElementById("buildsGrid");
  if (!grid) return;
  grid.innerHTML = state.builds.map((b, i) => cardHTML(b, i)).join("");
}

// ===== COMMAND PALETTE =====
const cmdk = document.getElementById("cmdk");
const cmdkInput = document.getElementById("cmdkInput");
function openCmdk()  { cmdk.classList.add("open"); cmdkInput.focus(); }
function closeCmdk() { cmdk.classList.remove("open"); }
document.addEventListener("keydown", e => {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
    e.preventDefault();
    cmdk.classList.contains("open") ? closeCmdk() : openCmdk();
  }
  if (e.key === "Escape") closeCmdk();
});

// ===== TOASTS =====
const toastIcons = {
  success:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M20 6 9 17l-5-5"/></svg>',
  info:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>',
  warn:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M10.3 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.7 3.86a2 2 0 0 0-3.4 0z"/></svg>',
  danger:'<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><path d="M15 9l-6 6M9 9l6 6"/></svg>',
};
function toast(type, title, sub) {
  const wrap = document.getElementById("toasts");
  const el = document.createElement("div");
  el.className = "toast";
  el.innerHTML = `<div class="ti ${type}">${toastIcons[type]}</div>
    <div class="tc"><div class="tt">${title}</div><div class="ts">${sub||""}</div></div>`;
  wrap.appendChild(el);
  setTimeout(() => {
    el.classList.add("out");
    setTimeout(() => el.remove(), 300);
  }, 3400);
}