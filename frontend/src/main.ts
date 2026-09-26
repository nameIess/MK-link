import "./style.css";
import { CreateLink, Validate } from "../wailsjs/go/main/App";

type LinkType = "directory-symlink" | "file-symlink" | "junction" | "hardlink";

type Description = { name: string; meta: string; body: string; target: string; link: string };

const descriptions: Record<LinkType, Description> = {
  "directory-symlink": {
    name: "Directory symbolic link", meta: "Path redirect",
    body: "A symbolic link points to another path. This directory link can cross local volumes.",
    target: "Select an existing folder.", link: "Choose the new folder link path."
  },
  "file-symlink": {
    name: "File symbolic link", meta: "Path redirect",
    body: "A symbolic link points to another path. This file link can cross local volumes.",
    target: "Select an existing file.", link: "Choose the new file link path."
  },
  junction: {
    name: "Junction", meta: "Directory reparse point",
    body: "A junction is a Windows directory reparse point. It links local directories, including across local volumes.",
    target: "Select an existing folder.", link: "Choose the new junction path."
  },
  hardlink: {
    name: "Hard link", meta: "Same file data",
    body: "A hard link gives a file another directory entry. Both paths refer to the same file data and must be on the same volume.",
    target: "Select an existing file.", link: "Choose the new file link path."
  }
};

const root = document.querySelector<HTMLDivElement>("#app")!;
root.innerHTML = `
<main class="shell">
<header class="topbar">
  <div class="brand"><img src="/icon.ico" class="logo" alt=""><div><div class="app-name">MK-Link</div><div class="tagline">Windows filesystem links</div></div></div>
  <div class="status"><span></span>Ready</div>
</header>
<section class="hero"><div><p class="eyebrow">CREATE A LINK</p><h1>Point one path at another.</h1><p class="subtitle">Choose a link type, enter the existing path and new link path, then create it directly through Windows filesystem APIs.</p></div><div class="hero-mark">↗</div></section>
<section class="card">
  <div class="section-head"><h2>Link type</h2><p id="type-description"></p></div>
  <div class="types" id="types"></div>
  <div class="paths">
    <label class="field"><span>Target</span><small id="target-help"></small><input id="target" spellcheck="false" placeholder="C:\\Projects\\Original"></label>
    <label class="field"><span>Link path</span><small id="link-help"></small><input id="link" spellcheck="false" placeholder="C:\\Projects\\Alias"></label>
  </div>
  <label class="overwrite"><input id="overwrite" type="checkbox"><span><b>Replace existing link</b><small>Only an empty directory or single existing file/link at the exact path is removed. Folder contents are never recursively deleted.</small></span></label>
  <div class="actions"><div id="validation" class="validation">Enter both paths to validate.</div><button id="create" disabled>Create link <b>↗</b></button></div>
</section>
<section class="info-grid">
<article><i>↗</i><div><h3>Symbolic link</h3><p>Path-based link for files or directories.</p></div></article>
<article><i>◇</i><div><h3>Junction</h3><p>Directory-only Windows reparse point.</p></div></article>
<article><i>=</i><div><h3>Hard link</h3><p>Second file name for the same underlying data.</p></div></article>
</section>
<footer>No shell commands are used to create links.</footer>
</main>`;

const typeEntries: Array<[LinkType, string, string]> = [
["directory-symlink","Directory symbolic link","Path redirect"],
["file-symlink","File symbolic link","Path redirect"],
["junction","Junction","Directory reparse point"],
["hardlink","Hard link","Same file data"]
];
let currentType: LinkType = "directory-symlink";
const types = document.querySelector<HTMLDivElement>("#types")!;
types.innerHTML = typeEntries.map(([value,name,meta]) => `<button class="type${value === currentType ? " active" : ""}" data-type="${value}"><b>${name}</b><small>${meta}</small></button>`).join("");

const target = document.querySelector<HTMLInputElement>("#target")!;
const link = document.querySelector<HTMLInputElement>("#link")!;
const overwrite = document.querySelector<HTMLInputElement>("#overwrite")!;
const validation = document.querySelector<HTMLDivElement>("#validation")!;
const create = document.querySelector<HTMLButtonElement>("#create")!;
const typeDescription = document.querySelector<HTMLElement>("#type-description")!;
const targetHelp = document.querySelector<HTMLElement>("#target-help")!;
const linkHelp = document.querySelector<HTMLElement>("#link-help")!;

function renderType() {
  const d = descriptions[currentType];
  typeDescription.textContent = `${d.name} — ${d.body}`;
  targetHelp.textContent = d.target;
  linkHelp.textContent = d.link;
  types.querySelectorAll<HTMLButtonElement>(".type").forEach((button) => {
    button.classList.toggle("active", button.dataset.type === currentType);
  });
}

async function refreshValidation() {
  if (!target.value.trim() || !link.value.trim()) {
    validation.textContent = "Enter both paths to validate.";
    validation.className = "validation";
    create.disabled = true;
    return;
  }
  try {
    await Validate({ type: currentType, target: target.value.trim(), link: link.value.trim(), overwrite: overwrite.checked });
    validation.textContent = "Ready to create this link.";
    validation.className = "validation ok";
    create.disabled = false;
  } catch (error) {
    validation.textContent = String(error);
    validation.className = "validation error";
    create.disabled = true;
  }
}

types.querySelectorAll<HTMLButtonElement>(".type").forEach((button) => {
  button.addEventListener("click", () => {
    currentType = button.dataset.type as LinkType;
    renderType();
    void refreshValidation();
  });
});
[target, link].forEach((field) => field.addEventListener("input", () => void refreshValidation()));
overwrite.addEventListener("change", () => void refreshValidation());
create.addEventListener("click", async () => {
  create.disabled = true;
  validation.textContent = "Creating link…";
  validation.className = "validation";
  try {
    const result = await CreateLink({ type: currentType, target: target.value.trim(), link: link.value.trim(), overwrite: overwrite.checked });
    validation.textContent = result.message;
    validation.className = "validation ok";
  } catch (error) {
    validation.textContent = String(error);
    validation.className = "validation error";
  } finally {
    await refreshValidation();
  }
});
renderType();