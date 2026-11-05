package cmd

import (
	"encoding/json"
	"html"
)

func webConsoleHTML(lang string) string {
	t := webConsoleText(lang)
	return `<!doctype html>
<html lang="` + html.EscapeString(lang) + `"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + html.EscapeString(t.title) + `</title>
<style>` + webConsoleCSS + `</style></head><body>
<div class="wrap">
  <header class="top">
    <h1>` + html.EscapeString(t.title) + `</h1>
    <p class="hint">` + html.EscapeString(t.subtitle) + `</p>
  </header>
  <nav class="tabs">
    <button type="button" class="tab active" data-tab="profiles">` + html.EscapeString(t.tabProfiles) + `</button>
    <button type="button" class="tab" data-tab="audit">` + html.EscapeString(t.tabAudit) + `</button>
  </nav>
  <section id="panel-profiles" class="panel">
    <div class="row">
      <button type="button" class="primary" id="btn-new">` + html.EscapeString(t.newProfile) + `</button>
      <button type="button" class="ghost" id="btn-refresh-profiles">` + html.EscapeString(t.refresh) + `</button>
    </div>
    <div id="profile-list" class="list"></div>
    <div class="card form-card" id="form-card" hidden>
      <h2 id="form-title">` + html.EscapeString(t.editProfile) + `</h2>
      <div class="grid">
        <label>` + html.EscapeString(t.name) + `<input id="f-name" autocomplete="off"></label>
        <label>` + html.EscapeString(t.server) + `<input id="f-server" placeholder="http://127.0.0.1:8848"></label>
        <label>` + html.EscapeString(t.username) + `<input id="f-username"></label>
        <label>` + html.EscapeString(t.authType) + `<input id="f-authType" placeholder="nacos"></label>
        <label>` + html.EscapeString(t.namespace) + `<input id="f-namespace"></label>
        <label>` + html.EscapeString(t.contextPath) + `<input id="f-contextPath" placeholder="/nacos"></label>
        <label>` + html.EscapeString(t.format) + `<select id="f-format"><option value="json">json</option><option value="pretty">pretty</option></select></label>
        <label>` + html.EscapeString(t.password) + `<input id="f-password" type="password" autocomplete="new-password" placeholder="` + html.EscapeString(t.passwordHint) + `"></label>
        <label class="check"><input type="checkbox" id="f-clear-password"> ` + html.EscapeString(t.clearPassword) + `</label>
        <p class="hint" id="f-password-status"></p>
      </div>
      <h3>` + html.EscapeString(t.metaSection) + `</h3>
      <div class="grid">
        <label>` + html.EscapeString(t.displayName) + `<input id="f-displayName"></label>
        <label>` + html.EscapeString(t.envType) + `<select id="f-envType"><option value="">-</option><option value="lab">lab</option><option value="prod">prod</option></select></label>
        <label>` + html.EscapeString(t.color) + `<input id="f-color" placeholder="#1677ff"></label>
        <label>` + html.EscapeString(t.owner) + `<input id="f-owner"></label>
        <label class="full">` + html.EscapeString(t.notes) + `<textarea id="f-notes" rows="3"></textarea></label>
      </div>
      <div class="row">
        <button type="button" class="primary" id="btn-save">` + html.EscapeString(t.save) + `</button>
        <button type="button" class="ghost" id="btn-check-auth">` + html.EscapeString(t.checkAuth) + `</button>
        <button type="button" class="danger" id="btn-delete">` + html.EscapeString(t.delete) + `</button>
        <button type="button" class="ghost" id="btn-cancel">` + html.EscapeString(t.cancel) + `</button>
      </div>
      <p class="okmsg" id="form-ok" hidden></p>
      <p class="err" id="form-err" hidden></p>
    </div>
  </section>
  <section id="panel-audit" class="panel" hidden>
    <div class="row filters">
      <label>` + html.EscapeString(t.filterCmd) + `<input id="a-command"></label>
      <label>` + html.EscapeString(t.filterProfile) + `<input id="a-profile"></label>
      <label>` + html.EscapeString(t.filterSince) + `<input id="a-since" placeholder="RFC3339"></label>
      <button type="button" class="primary" id="btn-audit">` + html.EscapeString(t.refresh) + `</button>
    </div>
    <div class="table-wrap"><table><thead><tr>
      <th>` + html.EscapeString(t.colTime) + `</th>
      <th>` + html.EscapeString(t.colCommand) + `</th>
      <th>` + html.EscapeString(t.colRisk) + `</th>
      <th>` + html.EscapeString(t.colProfile) + `</th>
      <th>` + html.EscapeString(t.colOK) + `</th>
      <th>` + html.EscapeString(t.colError) + `</th>
    </tr></thead><tbody id="audit-body"></tbody></table></div>
  </section>
</div>
<script>
window.__NACOS_WEB_I18N__ = ` + webConsoleI18nJSON(lang) + `;
` + webConsoleJS + `
</script>
</body></html>`
}

type webText struct {
	title, subtitle, tabProfiles, tabAudit string
	newProfile, refresh, editProfile       string
	name, server, username, authType       string
	namespace, contextPath, format         string
	password, passwordHint, clearPassword  string
	passwordSet, passwordUnset             string
	metaSection, displayName, envType      string
	color, owner, notes                    string
	save, delete, cancel                   string
	checkAuth, authOK, authFail            string
	filterCmd, filterProfile, filterSince  string
	colTime, colCommand, colRisk           string
	colProfile, colOK, colError            string
	confirmDelete, saved, deleted          string
}

func webConsoleText(lang string) webText {
	if lang == "zh" {
		return webText{
			title: "nacos-cli 监管", subtitle: "仅本机 127.0.0.1 · 管理 Profile 与操作记录",
			tabProfiles: "配置文件", tabAudit: "操作记录",
			newProfile: "新建", refresh: "刷新", editProfile: "编辑 Profile",
			name: "名称", server: "Server", username: "用户名", authType: "认证类型",
			namespace: "命名空间", contextPath: "Context Path", format: "默认格式",
			password: "密码", passwordHint: "留空=沿用已保存密码；检查认证时再无则试本地 token", clearPassword: "清空已保存密码",
			passwordSet: "已设置", passwordUnset: "未设置",
			metaSection: "备注元数据", displayName: "备注名", envType: "环境类型",
			color: "颜色标签", owner: "负责人", notes: "注意事项",
			save: "保存", delete: "删除", cancel: "取消",
			checkAuth: "检查认证", authOK: "认证可用", authFail: "认证失败",
			filterCmd: "命令包含", filterProfile: "Profile", filterSince: "起始时间",
			colTime: "时间", colCommand: "命令", colRisk: "风险",
			colProfile: "Profile", colOK: "结果", colError: "错误",
			confirmDelete: "确认删除该 Profile？", saved: "已保存", deleted: "已删除",
		}
	}
	return webText{
		title: "nacos-cli Console", subtitle: "Loopback only · manage profiles and audit log",
		tabProfiles: "Profiles", tabAudit: "Audit",
		newProfile: "New", refresh: "Refresh", editProfile: "Edit profile",
		name: "Name", server: "Server", username: "Username", authType: "Auth type",
		namespace: "Namespace", contextPath: "Context path", format: "Default format",
		password: "Password", passwordHint: "Blank=keep saved password; check-auth then tries local token", clearPassword: "Clear saved password",
		passwordSet: "Set", passwordUnset: "Not set",
		metaSection: "Metadata", displayName: "Display name", envType: "Env type",
		color: "Color", owner: "Owner", notes: "Notes",
		save: "Save", delete: "Delete", cancel: "Cancel",
		checkAuth: "Check auth", authOK: "Auth OK", authFail: "Auth failed",
		filterCmd: "Command contains", filterProfile: "Profile", filterSince: "Since",
		colTime: "Time", colCommand: "Command", colRisk: "Risk",
		colProfile: "Profile", colOK: "OK", colError: "Error",
		confirmDelete: "Delete this profile?", saved: "Saved", deleted: "Deleted",
	}
}

func webConsoleI18nJSON(lang string) string {
	t := webConsoleText(lang)
	b, _ := json.Marshal(map[string]string{
		"passwordSet":   t.passwordSet,
		"passwordUnset": t.passwordUnset,
		"confirmDelete": t.confirmDelete,
		"saved":         t.saved,
		"deleted":       t.deleted,
		"editProfile":   t.editProfile,
		"newProfile":    t.newProfile,
		"checkAuth":     t.checkAuth,
		"authOK":        t.authOK,
		"authFail":      t.authFail,
	})
	return string(b)
}

const webConsoleCSS = `
*{box-sizing:border-box}
body{margin:0;background:#f5f7fa;font-family:system-ui,-apple-system,"Segoe UI",sans-serif;color:#1f1f1f}
.wrap{max-width:960px;margin:0 auto;padding:24px 16px 48px}
.top h1{margin:0 0 6px;font-size:1.45rem;font-weight:600}
.hint{color:#64748b;font-size:13px;margin:0 0 8px;line-height:1.45}
.tabs{display:flex;gap:8px;margin:16px 0}
.tab{border:1px solid #d9d9d9;background:#fff;border-radius:8px;padding:8px 14px;cursor:pointer;font-weight:500}
.tab.active{border-color:#1677ff;color:#1677ff;box-shadow:0 0 0 2px rgba(22,119,255,.12)}
.card,.form-card{background:#fff;border-radius:12px;box-shadow:0 8px 24px rgba(15,23,42,.08);padding:20px;margin-top:14px}
.list .item{display:flex;align-items:center;justify-content:space-between;gap:12px;background:#fff;border-radius:12px;
box-shadow:0 4px 14px rgba(15,23,42,.06);padding:14px 16px;margin-top:10px;cursor:pointer}
.list .item:hover{outline:2px solid rgba(22,119,255,.25)}
.badge{display:inline-block;padding:2px 8px;border-radius:999px;font-size:12px;color:#fff;margin-right:8px}
.row{display:flex;flex-wrap:wrap;gap:8px;align-items:end;margin:8px 0}
button.primary{border:0;border-radius:8px;background:#1677ff;color:#fff;font-weight:600;padding:10px 14px;cursor:pointer}
button.primary:hover{background:#4096ff}
button.ghost{border:1px solid #d9d9d9;border-radius:8px;background:#fff;padding:10px 14px;cursor:pointer}
button.danger{border:0;border-radius:8px;background:#ff4d4f;color:#fff;font-weight:600;padding:10px 14px;cursor:pointer}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.grid label{display:flex;flex-direction:column;gap:6px;font-size:13px;font-weight:500}
.grid label.full{grid-column:1/-1}
.grid label.check{flex-direction:row;align-items:center;font-weight:400}
input,select,textarea{padding:10px 12px;border:1px solid #d9d9d9;border-radius:8px;font-size:14px;font-weight:400}
input:focus,select:focus,textarea:focus{outline:none;border-color:#1677ff;box-shadow:0 0 0 2px rgba(22,119,255,.15)}
.err{color:#b00020;background:#fff2f0;border:1px solid #ffccc7;border-radius:8px;padding:10px 12px;margin:12px 0;font-size:13px}
.okmsg{color:#135200;background:#f6ffed;border:1px solid #b7eb8f;border-radius:8px;padding:10px 12px;margin:12px 0;font-size:13px}
.item .actions{display:flex;gap:6px;flex-shrink:0}
.item .actions button{padding:6px 10px;font-size:12px}
.table-wrap{overflow:auto;background:#fff;border-radius:12px;box-shadow:0 8px 24px rgba(15,23,42,.08);margin-top:12px}
table{width:100%;border-collapse:collapse;font-size:13px}
th,td{padding:10px 12px;border-bottom:1px solid #f0f0f0;text-align:left;vertical-align:top}
th{background:#fafafa;font-weight:600}
.filters label{display:flex;flex-direction:column;gap:4px;font-size:12px}
@media (max-width:720px){.grid{grid-template-columns:1fr}}
`

const webConsoleJS = `
(function(){
  const I = window.__NACOS_WEB_I18N__ || {};
  const $ = (id) => document.getElementById(id);

  document.querySelectorAll('.tab').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.tab').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      const tab = btn.getAttribute('data-tab');
      $('panel-profiles').hidden = tab !== 'profiles';
      $('panel-audit').hidden = tab !== 'audit';
      if (tab === 'audit') loadAudit();
    });
  });

  async function api(url, opts){
    const res = await fetch(url, opts);
    const data = await res.json().catch(() => ({}));
    if (!res.ok || data.ok === false) throw new Error((data && data.error) || res.statusText || 'error');
    return data;
  }

  function esc(s){
    return String(s == null ? '' : s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  }

  async function loadProfiles(){
    const data = await api('/api/profiles');
    const box = $('profile-list');
    box.innerHTML = '';
    (data.profiles || []).forEach(item => {
      const meta = item.meta || {};
      const p = item.profile || {};
      const color = meta.color || '#1677ff';
      const el = document.createElement('div');
      el.className = 'item';
      el.innerHTML = '<div><span class="badge" style="background:'+esc(color)+'">'+esc(meta.envType || '-')+
        '</span><strong>'+esc(item.name)+'</strong> '+esc(meta.displayName || '')+
        '<div class="hint">'+esc(p.server || '')+' · '+esc(p.username || '')+' · pw: '+
        (p.password_set ? esc(I.passwordSet||'set') : esc(I.passwordUnset||'unset'))+
        '</div><div class="hint" id="auth-status-'+esc(item.name)+'"></div></div>'+
        '<div class="actions"><button type="button" class="ghost btn-check" data-name="'+esc(item.name)+'">'+
        esc(I.checkAuth||'Check auth')+'</button></div>';
      el.querySelector('.btn-check').addEventListener('click', (ev) => {
        ev.stopPropagation();
        checkAuth(item.name).catch(showErr);
      });
      el.addEventListener('click', () => openEdit(item, false));
      box.appendChild(el);
    });
  }

  function openEdit(item, isNew){
    $('form-card').hidden = false;
    $('form-err').hidden = true;
    $('form-title').textContent = isNew ? (I.newProfile || 'New') : (I.editProfile || 'Edit');
    const p = (item && item.profile) || {};
    const m = (item && item.meta) || {};
    $('f-name').value = item && item.name ? item.name : '';
    $('f-name').disabled = !isNew;
    $('f-server').value = p.server || '';
    $('f-username').value = p.username || '';
    $('f-authType').value = p.authType || 'nacos';
    $('f-namespace').value = p.namespace || '';
    $('f-contextPath').value = p.contextPath || '/nacos';
    $('f-format').value = p.format || 'json';
    $('f-password').value = '';
    $('f-clear-password').checked = false;
    $('f-password-status').textContent = p.password_set ? (I.passwordSet||'set') : (I.passwordUnset||'unset');
    $('f-displayName').value = m.displayName || '';
    $('f-envType').value = m.envType || '';
    $('f-color').value = m.color || '';
    $('f-owner').value = m.owner || '';
    $('f-notes').value = m.notes || '';
    $('btn-delete').hidden = !!isNew;
  }

  $('btn-new').addEventListener('click', () => openEdit({name:'', profile:{}, meta:{}}, true));
  $('btn-cancel').addEventListener('click', () => { $('form-card').hidden = true; });
  $('btn-refresh-profiles').addEventListener('click', () => loadProfiles().catch(showErr));

  $('btn-save').addEventListener('click', async () => {
    try {
      const name = $('f-name').value.trim();
      if (!name) throw new Error('name required');
      const body = {
        server: $('f-server').value.trim(),
        username: $('f-username').value.trim(),
        authType: $('f-authType').value.trim() || 'nacos',
        namespace: $('f-namespace').value.trim(),
        contextPath: $('f-contextPath').value.trim(),
        format: $('f-format').value,
        clear_password: $('f-clear-password').checked,
        meta: {
          displayName: $('f-displayName').value.trim(),
          envType: $('f-envType').value,
          color: $('f-color').value.trim(),
          owner: $('f-owner').value.trim(),
          notes: $('f-notes').value.trim()
        }
      };
      const pw = $('f-password').value;
      if (pw) { body.password = pw; body.password_set = true; }
      await api('/api/profiles/' + encodeURIComponent(name), {
        method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify(body)
      });
      $('form-card').hidden = true;
      await loadProfiles();
    } catch (e) { showErr(e); }
  });

  $('btn-delete').addEventListener('click', async () => {
    try {
      if (!confirm(I.confirmDelete || 'Delete?')) return;
      const name = $('f-name').value.trim();
      await api('/api/profiles/' + encodeURIComponent(name) + '?confirm=1', { method: 'DELETE' });
      $('form-card').hidden = true;
      await loadProfiles();
    } catch (e) { showErr(e); }
  });

  function showErr(e){
    const ok = $('form-ok'); if (ok) ok.hidden = true;
    const el = $('form-err');
    el.hidden = false;
    el.textContent = (e && e.message) ? e.message : String(e);
  }
  function showOk(msg){
    const er = $('form-err'); if (er) er.hidden = true;
    const el = $('form-ok');
    if (!el) return;
    el.hidden = false;
    el.textContent = msg;
  }
  function formatAuth(data){
    const mode = data.auth_mode || '';
    const detail = data.detail || '';
    const err = data.error || '';
    const via = data.via || '';
    const viaBit = via ? ' · via '+via : '';
    if (data.ok) {
      return (I.authOK||'Auth OK') + (mode ? ' · '+mode : '') + viaBit + (detail ? ' · '+detail : '');
    }
    return (I.authFail||'Auth failed') + (err ? ': '+err : '') + viaBit + (mode ? ' · '+mode : '');
  }
  async function checkAuth(name){
    const res = await fetch('/api/profiles/' + encodeURIComponent(name) + '/check-auth', { method: 'POST' });
    const data = await res.json().catch(() => ({}));
    if (res.status === 404) throw new Error((data && data.error) || 'not found');
    if (!res.ok && data.ok === undefined) throw new Error((data && data.error) || res.statusText || 'error');
    const msg = formatAuth(data);
    const slot = document.getElementById('auth-status-' + name);
    if (slot) slot.textContent = msg;
    if (data.ok) showOk(msg); else showErr(new Error(msg));
    return data;
  }
  function collectAuthDraft(){
    const body = {
      server: $('f-server').value.trim(),
      username: $('f-username').value.trim(),
      authType: $('f-authType').value.trim() || 'nacos',
      namespace: $('f-namespace').value.trim(),
      contextPath: $('f-contextPath').value.trim() || '/nacos',
      format: $('f-format').value,
      clear_password: $('f-clear-password').checked
    };
    const pw = $('f-password').value;
    if (pw) { body.password = pw; body.password_set = true; }
    return body;
  }
  async function checkAuthDraft(name){
    const draft = collectAuthDraft();
    const res = await fetch('/api/profiles/' + encodeURIComponent(name) + '/check-auth', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(draft)
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok && data.ok === undefined) throw new Error((data && data.error) || res.statusText || 'error');
    const msg = formatAuth(data);
    if (data.ok) showOk(msg); else showErr(new Error(msg));
    return data;
  }
  if ($('btn-check-auth')) {
    $('btn-check-auth').addEventListener('click', async () => {
      try {
        const name = $('f-name').value.trim();
        if (!name) throw new Error('name required');
        await checkAuthDraft(name);
      } catch (e) { showErr(e); }
    });
  }

  async function loadAudit(){
    const q = new URLSearchParams();
    if ($('a-command').value.trim()) q.set('command', $('a-command').value.trim());
    if ($('a-profile').value.trim()) q.set('profile', $('a-profile').value.trim());
    if ($('a-since').value.trim()) q.set('since', $('a-since').value.trim());
    const data = await api('/api/audit?' + q.toString());
    const tb = $('audit-body');
    tb.innerHTML = '';
    (data.entries || []).forEach(e => {
      const tr = document.createElement('tr');
      tr.innerHTML = '<td>'+esc(e.time)+'</td><td>'+esc(e.command)+'</td><td>'+esc(e.risk||'')+
        '</td><td>'+esc(e.profile||'')+'</td><td>'+(e.ok?'ok':'fail')+'</td><td>'+esc(e.error||'')+'</td>';
      tb.appendChild(tr);
    });
  }
  $('btn-audit').addEventListener('click', () => loadAudit().catch(err => alert(err.message || err)));
  loadProfiles().catch(err => alert(err.message || err));
})();
`
