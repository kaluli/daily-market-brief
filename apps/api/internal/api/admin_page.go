package api

import (
	"github.com/gofiber/fiber/v2"
)

// adminPage serves the admin panel HTML (fuentes de noticias).
// Use http://localhost:3090/admin (no depende de Next.js).
func (s *Server) adminPage(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(adminHTML)
}

const adminHTML = `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Admin — Fuentes de noticias</title>
  <style>
    * { box-sizing: border-box; }
    body { font-family: system-ui, sans-serif; background: #0f172a; color: #e2e8f0; margin: 0; padding: 24px; line-height: 1.5; }
    a { color: #60a5fa; text-decoration: none; }
    a:hover { text-decoration: underline; }
    h1 { font-size: 1.5rem; margin: 8px 0; }
    .muted { color: #94a3b8; font-size: 0.875rem; margin-top: 4px; }
    section { background: rgba(30,41,59,0.5); border: 1px solid #334155; border-radius: 8px; padding: 24px; margin-top: 24px; }
    section h2 { font-size: 1.125rem; margin: 0 0 16px 0; }
    ul { list-style: none; padding: 0; margin: 0; }
    li { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 16px; padding: 16px; margin-top: 12px; background: #1e293b; border: 1px solid #334155; border-radius: 6px; }
    li .info { flex: 1; min-width: 0; }
    li .name { font-weight: 500; }
    li .meta { font-size: 0.875rem; color: #94a3b8; margin-top: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.75rem; margin-left: 8px; }
    .badge.on { background: rgba(16,185,129,0.2); color: #6ee7b7; }
    .badge.off { background: #334155; color: #94a3b8; }
    button { padding: 8px 16px; border-radius: 6px; font-size: 0.875rem; font-weight: 500; cursor: pointer; border: none; }
    button:disabled { opacity: 0.6; cursor: not-allowed; }
    .btn-on { background: #475569; color: #fff; }
    .btn-on:hover:not(:disabled) { background: #64748b; }
    .btn-off { background: #2563eb; color: #fff; }
    .btn-off:hover:not(:disabled) { background: #3b82f6; }
    .error { color: #f87171; margin-top: 16px; }
    .loading { color: #94a3b8; }
  </style>
</head>
<body>
  <div>
    <a href="/">← API</a>
    <h1>Fuentes de noticias</h1>
    <p class="muted">Activa o desactiva fuentes. Los cambios se guardan en config/news_sources.json.</p>
  </div>
  <div id="root">
    <p class="loading">Cargando…</p>
  </div>
  <script>
    const root = document.getElementById('root');
    function err(msg) {
      root.innerHTML = '<p class="error">' + (msg || 'Error al cargar') + '</p>';
    }
    function render(data) {
      const sources = data.sources || [];
      const rss = data.rss_sources || [];
      let html = '';
      html += '<section><h2>Fuentes API (NewsAPI, Finnhub)</h2><ul>';
      if (sources.length === 0) html += '<li class="muted">No hay fuentes API.</li>';
      else sources.forEach(function(s) {
        html += '<li><div class="info"><span class="name">' + escape(s.name) + ' <span class="badge ' + (s.enabled ? 'on' : 'off') + '">' + (s.enabled ? 'Activa' : 'Inactiva') + '</span></span><div class="meta">' + escape(s.id) + ' · Peso: ' + s.weight + '</div></div><button class="' + (s.enabled ? 'btn-on' : 'btn-off') + '" data-id="' + escape(s.id) + '" data-enabled="' + s.enabled + '">' + (s.enabled ? 'Desactivar' : 'Activar') + '</button></li>';
      });
      html += '</ul></section>';
      html += '<section><h2>Fuentes RSS</h2><ul>';
      if (rss.length === 0) html += '<li class="muted">No hay fuentes RSS.</li>';
      else rss.forEach(function(s) {
        html += '<li><div class="info"><span class="name">' + escape(s.name) + ' <span class="badge ' + (s.enabled ? 'on' : 'off') + '">' + (s.enabled ? 'Activa' : 'Inactiva') + '</span></span><div class="meta" title="' + escape(s.url) + '">' + escape(s.id) + ' · ' + escape(s.url) + '</div></div><button class="' + (s.enabled ? 'btn-on' : 'btn-off') + '" data-id="' + escape(s.id) + '" data-enabled="' + s.enabled + '">' + (s.enabled ? 'Desactivar' : 'Activar') + '</button></li>';
      });
      html += '</ul></section>';
      root.innerHTML = html;
      root.querySelectorAll('button[data-id]').forEach(function(btn) {
        btn.addEventListener('click', function() {
          var id = this.getAttribute('data-id');
          var enabled = this.getAttribute('data-enabled') === 'true';
          this.disabled = true;
          fetch('/api/admin/sources/' + encodeURIComponent(id) + '/enabled', {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ enabled: !enabled })
          }).then(function(r) {
            if (r.ok) load(); else return r.json().then(function(e) { throw new Error(e.error || 'Error'); });
          }).catch(function(e) {
            err(e.message);
          });
        });
      });
    }
    function escape(s) {
      var d = document.createElement('div');
      d.textContent = s;
      return d.innerHTML;
    }
    function load() {
      root.innerHTML = '<p class="loading">Cargando…</p>';
      fetch('/api/admin/sources')
        .then(function(r) {
          if (!r.ok) return r.json().then(function(e) { throw new Error(e.error || 'Error ' + r.status); });
          return r.json();
        })
        .then(render)
        .catch(function(e) { err(e.message); });
    }
    load();
  </script>
</body>
</html>
`
