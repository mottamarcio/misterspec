// Shared sidebar navigation for every page of the misterspec
// documentation site. One source of truth for the site's own
// structure (specs/020-docs-site-readme/data-model.md) — adding a
// page means adding one entry here, never hand-editing a sidebar in
// any individual .html file.
const NAV = [
  { group: "Overview", title: "Introduction", href: "index.html", icon: "lucide:home" },
  { group: "Overview", title: "Getting Started", href: "getting-started.html", icon: "lucide:rocket" },
  { group: "Overview", title: "The Spec-Driven Workflow", href: "workflow.html", icon: "lucide:git-branch" },
  { group: "Features", title: "Core Foundation", href: "foundation.html", icon: "lucide:box" },
  { group: "Features", title: "Agent Adapters & CLI", href: "agents-cli.html", icon: "lucide:plug" },
  { group: "Features", title: "The Context Engine", href: "context-engine.html", icon: "lucide:brain" },
  { group: "Features", title: "Multi-Agent Skill Integration", href: "multi-agent-skills.html", icon: "lucide:users" },
  { group: "Features", title: "Dogfooding & Evaluation", href: "dogfooding.html", icon: "lucide:flask-conical" },
  { group: "Reference", title: "Command Reference", href: "commands.html", icon: "lucide:terminal" },
];

function renderSidebar() {
  const sidebar = document.getElementById("sidebar");
  if (!sidebar) return;

  const current = location.pathname.split("/").pop() || "index.html";
  const groups = [];
  for (const item of NAV) {
    let g = groups.find((g) => g.name === item.group);
    if (!g) {
      g = { name: item.group, items: [] };
      groups.push(g);
    }
    g.items.push(item);
  }

  let html = `
    <div class="px-4 py-5 border-b border-slate-800">
      <a href="index.html" class="flex items-center gap-2 text-white font-semibold text-lg">
        <iconify-icon icon="lucide:terminal-square" class="text-xl"></iconify-icon>
        misterspec
      </a>
      <p class="text-slate-400 text-xs mt-1">Agent-native Spec-Driven Development</p>
    </div>
    <nav class="px-2 py-4 space-y-6 overflow-y-auto">
  `;

  for (const g of groups) {
    html += `<div><p class="px-3 text-xs font-semibold uppercase tracking-wider text-slate-500 mb-1">${g.name}</p><ul class="space-y-0.5">`;
    for (const item of g.items) {
      const active = item.href === current;
      const activeClasses = active
        ? "bg-slate-800 text-white"
        : "text-slate-300 hover:bg-slate-800/60 hover:text-white";
      html += `
        <li>
          <a href="${item.href}" class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm transition-colors ${activeClasses}">
            <iconify-icon icon="${item.icon}" class="text-base shrink-0"></iconify-icon>
            <span>${item.title}</span>
          </a>
        </li>`;
    }
    html += `</ul></div>`;
  }

  html += `</nav>`;
  sidebar.innerHTML = html;
}

function setupMobileNav() {
  const toggle = document.getElementById("nav-toggle");
  const sidebar = document.getElementById("sidebar-container");
  const overlay = document.getElementById("nav-overlay");
  if (!toggle || !sidebar || !overlay) return;

  const close = () => {
    sidebar.classList.add("-translate-x-full");
    overlay.classList.add("hidden");
  };
  const open = () => {
    sidebar.classList.remove("-translate-x-full");
    overlay.classList.remove("hidden");
  };

  toggle.addEventListener("click", () => {
    sidebar.classList.contains("-translate-x-full") ? open() : close();
  });
  overlay.addEventListener("click", close);
}

document.addEventListener("DOMContentLoaded", () => {
  renderSidebar();
  setupMobileNav();
});
